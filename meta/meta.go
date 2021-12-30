package meta

import (
	"errors"
	"github.com/pelletier/go-toml/v2"
	"os"
	"path"
)

var albumIdx map[string]AlbumInfo
var tagIdx map[string][]AlbumInfo
var tagIdxNonRecursive map[string][]AlbumInfo
var tags []string
var tagGraph map[string][]string

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

	albumIdx = make(map[string]AlbumInfo)
	tagIdx = make(map[string][]AlbumInfo)
	tags = V
	tagGraph = E

	for _, v := range albums {
		albumIdx[v.AlbumID] = v
	}

	tmp := make(map[string]map[string]bool)

	for _, v := range tags {
		tmp[v] = make(map[string]bool)
	}

	tagIdxNonRecursive = make(map[string][]AlbumInfo)
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
					tmp[x] = make(map[string]bool)
				}
				tmp[x][album] = true
			}
		}
	}
	for _, x := range V {
		if !vis[x] {
			dfs(x)
		}
	}

	for k, v := range tmp {
		tagIdx[k] = nil
		for albumId := range v {
			tagIdx[k] = append(tagIdx[k], albumIdx[albumId])
		}
	}

	return nil
}

func checkTags(V []string, E map[string][]string) error {
	VSet := make(map[string]interface{})
	for _, v := range V {
		VSet[v] = nil
	}
	// Node constraints check
	if len(VSet) != len(V) {
		return errors.New("invalid includes detected")
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
			err := dfs(nx)
			if err != nil {
				return err
			}
		}
		acc[x] = false
		return nil
	}

	for _, v := range V {
		if !vis[v] {
			err := dfs(v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func readTags(p string) ([]string, map[string][]string, error) {
	E := make(map[string][]string)
	V := make([]string, 0)

	f, err := os.ReadDir(p)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range f {
		f, err := os.Open(path.Join(p, v.Name()))
		if err != nil {
			return nil, nil, err
		}
		tags := make(map[string][]Tag, 0)
		err = toml.NewDecoder(f).Decode(&tags)
		if err != nil {
			return nil, nil, err
		}
		f.Close()
		for _, tag := range tags["tag"] {
			V = append(V, tag.Name)
			for _, child := range tag.Includes {
				V = append(V, child)
				E[tag.Name] = append(E[tag.Name], child)
			}
			for _, parent := range tag.IncludedBy {
				E[parent] = append(E[parent], tag.Name)
			}
		}
	}
	return V, E, nil
}

func readAlbums(p string) ([]AlbumInfo, error) {
	ret := make([]AlbumInfo, 0)
	f, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	for _, v := range f {
		record := Record{}

		f, err := os.Open(path.Join(p, v.Name()))
		if err != nil {
			return nil, err
		}
		err = toml.NewDecoder(f).Decode(&record)
		if err != nil {
			return nil, err
		}
		f.Close()
		album := AlbumInfo{
			AlbumID: record.Album.AlbumID,
			Title:   record.Album.Title,
			Edition: record.Album.Edition,
			Catalog: record.Album.Catalog,
			Artist:  record.Album.Artist,
			Date:    record.Album.Date,
			Tags:    record.Album.Tags,
			Type:    record.Album.Type,
			Discs:   record.Discs,
		}

		for _, disc := range album.Discs {
			if disc.Type == "" {
				disc.Type = album.Type
			}
			if disc.Artist == "" {
				disc.Artist = album.Artist
			}
			if disc.Title == "" {
				disc.Title = album.Title
			}
			if disc.Tags == nil {
				disc.Tags = []string{}
			}
			for _, track := range disc.Tracks {
				if track.Title == "" {
					track.Type = disc.Type
				}
				if track.Artist == "" {
					track.Artist = disc.Artist
				}
				if track.Type == "" {
					track.Type = disc.Type
				}
				if track.Tags == nil {
					track.Tags = []string{}
				}
			}
		}
		ret = append(ret, album)
	}
	return ret, nil
}
