package meta

import (
	"errors"
	"github.com/pelletier/go-toml/v2"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
)

var albumIdx map[AlbumIdentifier]AlbumDetails
var tagIdx map[string][]AlbumDetails
var tagIdxNonRecursive map[string][]AlbumDetails
var tags []Tag
var tagGraph map[string][]string
var dbAvailable = false

func DBAvailable() bool {
	return dbAvailable
}

func Read(p string) error {
	albums, err := readAlbums(path.Join(p, "album"))
	if err != nil {
		return err
	}
	V, E, err := readTags(path.Join(p, "tag"))
	if err != nil {
		return err
	}

	err = checkTags(V, E)
	if err != nil {
		return err
	}

	albumIdx = make(map[AlbumIdentifier]AlbumDetails)
	tagIdx = make(map[string][]AlbumDetails)
	tags = V
	tagGraph = E

	for _, v := range albums {
		albumIdx[v.AlbumID] = v
	}

	tmp := make(map[string]map[AlbumIdentifier]bool)

	tagIdxNonRecursive = make(map[string][]AlbumDetails)

	for _, v := range tags {
		tmp[v.Name] = make(map[AlbumIdentifier]bool)
		tagIdxNonRecursive[v.Name] = make([]AlbumDetails, 0)
	}

	for _, album := range albums {
		for _, tag := range album.Tags {
			if tmp[tag] == nil {
				return errors.New("unknown tag: " + tag)
			}
			tmp[tag][album.AlbumID] = true
			tagIdxNonRecursive[tag] = append(tagIdxNonRecursive[tag], album)
		}
	}

	vis := make(map[string]bool)
	var dfs func(x string)
	dfs = func(x string) {
		vis[x] = true
		for _, nx := range E[x] {
			if !vis[nx] {
				dfs(nx)
			}
			for album := range tmp[nx] {
				if tmp[x] == nil {
					tmp[x] = make(map[AlbumIdentifier]bool)
				}
				tmp[x][album] = true
			}
		}
	}
	for _, x := range V {
		if !vis[x.Name] {
			dfs(x.Name)
		}
	}

	for k, v := range tmp {
		tagIdx[k] = make([]AlbumDetails, 0)
		for albumId := range v {
			tagIdx[k] = append(tagIdx[k], albumIdx[albumId])
		}
	}

	err = generateAnniDb()
	if err != nil {
		log.Printf("Failed to generate anni db: %v\n", err)
		dbAvailable = false
	} else {
		dbAvailable = true
	}

	return nil
}

func checkTags(V []Tag, E map[string][]string) error {
	VSet := make(map[string]int)
	for _, v := range V {
		VSet[v.Name]++
	}
	// Node constraints check
	if len(VSet) != len(V) {
		msg := "duplicated tags detected: "
		first := true
		for k, v := range VSet {
			if v != 1 {
				if !first {
					msg += ","
				} else {
					first = false
				}
				msg += k
			}
		}
		return errors.New(msg)
	}
	for k, v := range E {
		_, e := VSet[k]
		if !e {
			return errors.New("unknown tag: " + k)
		}
		for _, t := range v {
			_, e := VSet[t]
			if !e {
				return errors.New("unknown tag: " + t)
			}
		}
	}
	acc := make(map[string]bool)
	vis := make(map[string]bool)
	// Loop detect
	var dfs func(string) error
	dfs = func(x string) error {
		acc[x] = true
		vis[x] = true
		for _, nx := range E[x] {
			if acc[nx] {
				return errors.New("loop detected")
			}
			if !vis[nx] {
				err := dfs(nx)
				if err != nil {
					return err
				}
			}
		}
		acc[x] = false
		return nil
	}

	for _, v := range V {
		if !vis[v.Name] {
			err := dfs(v.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func readTags(p string) ([]Tag, map[string][]string, error) {
	E := make(map[string][]string)
	V := make([]Tag, 0)

	f, err := os.ReadDir(p)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range f {
		f, err := os.Open(path.Join(p, v.Name()))
		if err != nil {
			return nil, nil, errors.New(v.Name() + ": " + err.Error())
		}
		tags := make(map[string][]tagDef, 0)
		err = toml.NewDecoder(f).Decode(&tags)
		if err != nil {
			return nil, nil, errors.New(v.Name() + ": " + err.Error())
		}
		f.Close()
		for _, tag := range tags["tag"] {
			if !validateTagType(tag.Type) {
				return nil, nil, errors.New(v.Name() + ": " + "invalid tag type: " + tag.Type)
			}
			V = append(V, Tag{
				Name:  tag.Name,
				Type:  tag.Type,
				Alias: tag.Alias,
			})
			for typ, child := range tag.Includes {
				if !validateTagType(typ) {
					return nil, nil, errors.New(v.Name() + ": " + "invalid tag type: " + typ)
				}
				for _, v := range child {
					V = append(V, Tag{
						Name: v,
						Type: typ,
					})
					E[tag.Name] = append(E[tag.Name], v)
				}
			}
			for _, parent := range tag.IncludedBy {
				E[parent] = append(E[parent], tag.Name)
			}
		}
	}
	return V, E, nil
}

func readAlbums(p string) ([]AlbumDetails, error) {
	ret := make([]AlbumDetails, 0)
	f, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	for _, v := range f {
		record := record{}
		date := ""

		f, err := os.Open(path.Join(p, v.Name()))
		if err != nil {
			return nil, errors.New(v.Name() + ":" + err.Error())
		}
		err = toml.NewDecoder(f).Decode(&record)
		if err != nil {
			return nil, errors.New(v.Name() + ":" + err.Error())
		}
		f.Close()

		localDate, ok := record.Album.Date.(toml.LocalDate)
		if ok {
			date = localDate.String()
		} else if str, ok := record.Album.Date.(string); ok {
			date = str
		} else {
			return nil, errors.New(v.Name() + ": invalid date")
		}

		album := AlbumDetails{
			AlbumInfo: AlbumInfo{
				AlbumID: record.Album.AlbumID,
				Title:   record.Album.Title,
				Edition: record.Album.Edition,
				Catalog: record.Album.Catalog,
				Artist:  record.Album.Artist,
				Date:    date,
				Type:    record.Album.Type,
			},
			Artists: record.Album.Artists,
			Discs:   record.Discs,
		}

		albumTags := map[string]bool{}
		for _, v := range record.Album.Tags {
			albumTags[v] = true
		}
		for _, disc := range album.Discs {
			if disc.Type == nil {
				disc.Type = &album.Type
			}
			if disc.Artist == nil {
				disc.Artist = &album.Artist
			}
			if disc.Artists == nil {
				disc.Artists = album.Artists
			}
			discTags := map[string]bool{}
			for _, v := range disc.Tags {
				discTags[v] = true
			}
			for _, track := range disc.Tracks {
				if track.Artist == nil {
					track.Artist = disc.Artist
				}
				if track.Type == nil {
					track.Type = disc.Type
				}
				if track.Artists == nil {
					track.Artists = disc.Artists
				}
				for _, tag := range track.Tags {
					discTags[tag] = true
					albumTags[tag] = true
				}
			}
			disc.Tags = toArray(discTags)
		}
		album.Tags = toArray(albumTags)

		ret = append(ret, album)
	}
	return ret, nil
}

var validTypes = []string{"artist", "group", "animation", "series", "project", "game", "organization", "default", "category"}

func validateTagType(p string) bool {
	for _, v := range validTypes {
		if p == v {
			return true
		}
	}
	return false
}

func toArray(s map[string]bool) []string {
	ret := make([]string, 0, len(s))
	for v := range s {
		ret = append(ret, v)
	}
	return ret
}

func generateAnniDb() error {
	_, err := exec.LookPath("anni")
	if err != nil {
		return err
	}
	_ = os.Mkdir("./tmp/prebuilt", fs.ModePerm)
	cmd := exec.Command("anni", "repo", "--root", "./tmp/meta", "db", "./tmp/prebuilt")
	return cmd.Run()
}
