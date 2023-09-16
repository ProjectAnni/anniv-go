package meta

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/pelletier/go-toml/v2"
)

var albumIdx map[AlbumIdentifier]*AlbumDetails
var tagSet *TagSet
var dbAvailable = false

func DBAvailable() bool {
	return dbAvailable
}

func Read(p string) error {
	var err error
	// read all albums
	albums, err := readAlbums(path.Join(p, "album"))
	if err != nil {
		return err
	}
	// read all tags
	tagSet, err = readTags(path.Join(p, "tag"))
	if err != nil {
		return err
	}

	// build album id -> details index
	albumIdx = make(map[AlbumIdentifier]*AlbumDetails)
	for idx, v := range albums {
		albumIdx[v.AlbumID] = &albums[idx]
	}

	// add album tag relations
	for idx, album := range albums {
		for _, tag := range album.Tags {
			tagRef, err := tagSet.FindTag(tag)
			if err != nil {
				return err
			}
			tagRef.AddAlbum(&albums[idx])
		}
	}

	// expand tags
	for idx := range albums {
		err := tagSet.ExpandTagsDef(&albums[idx])
		if err != nil {
			return err
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

func readTags(p string) (*TagSet, error) {
	var tagArray []Tag

	f, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	for _, v := range f {
		f, err := os.Open(path.Join(p, v.Name()))
		if err != nil {
			return nil, errors.New(v.Name() + ": " + err.Error())
		}
		tags := make(map[string][]tagDef, 0)
		err = toml.NewDecoder(f).Decode(&tags)
		if err != nil {
			return nil, errors.New(v.Name() + ": " + err.Error())
		}
		f.Close()

		// First, read and all tags
		for _, tag := range tags["tag"] {
			tagArray = append(tagArray, Tag{
				Name:  tag.Name,
				Type:  tag.Type,
				Names: tag.Names,
			})
			for _, child := range tag.Includes {
				childTag, err := TagFromStr(child)
				if err != nil {
					return nil, err
				}
				childTag.parentTags = append(childTag.parentTags, tag.Str())
				tagArray = append(tagArray, *childTag)
			}
			for _, parent := range tag.IncludedBy {
				tag.parentTags = append(tag.parentTags, parent)
			}
		}
	}

	return NewTagSet(tagArray)
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
	output, err := exec.Command("anni", "repo", "--root", "./tmp/meta", "db", "./tmp/prebuilt").CombinedOutput()
	if err != nil {
		log.Println(string(output))
	}
	return err
}
