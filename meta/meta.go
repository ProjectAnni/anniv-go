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

	tagNameIdx = make(map[string]Tag)
	for _, v := range tags {
		tagNameIdx[v.Name] = v
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
				trackTags := map[string]bool{}
				for _, v := range track.Tags {
					trackTags[v] = true
				}
				for _, v := range disc.Tags {
					trackTags[v] = true
				}
				if track.Artist == nil {
					track.Artist = disc.Artist
				}
				if track.Type == nil {
					track.Type = disc.Type
				}
				if track.Artists == nil {
					track.Artists = disc.Artists
				}
				track.Tags = toArray(trackTags)
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
