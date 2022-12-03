package meta

import (
	"encoding/json"
	"github.com/blevesearch/bleve/v2"
	"log"
	"sort"
	"time"
)

type trackDetails struct {
	TrackInfoWithAlbum
	AlbumTags []Tag
	//Tags  []Tag        `json:"tags"`
}

type albumDetails struct {
	AlbumDetails
	Tags []Tag
}

var tracksSearchIdx bleve.Index
var albumsSearchIdx bleve.Index

func initSearchIndex() error {
	log.Println("Building search index...")
	start := time.Now()
	var err error
	lock.RLock()
	defer lock.RUnlock()
	mapping := bleve.NewIndexMapping()
	tracksSearchIdx, err = bleve.New("", mapping)
	if err != nil {
		return err
	}
	tracksBatch := tracksSearchIdx.NewBatch()
	for _, album := range albumIdx {
		discId := uint(1)
		for _, disc := range album.Discs {
			trackId := uint(1)
			for _, track := range disc.Tracks {
				t := TrackInfoWithAlbum{
					TrackIdentifier: TrackIdentifier{
						DiscIdentifier: DiscIdentifier{
							AlbumID: album.AlbumID,
							DiscID:  discId,
						},
						TrackID: trackId,
					},
					TrackInfo:  track.TrackInfo,
					AlbumTitle: album.Title,
				}
				key, _ := json.Marshal(t)
				val := trackDetails{
					TrackInfoWithAlbum: t,
				}
				for _, tag := range album.Tags {
					val.AlbumTags = append(val.AlbumTags, tagNameIdx[tag])
				}
				//for _, tagName := range albumIdx[t.AlbumID].Discs[t.DiscID-1].Tracks[t.TrackID-1].Tags {
				//	val.Tags = append(val.Tags, tagNameIdx[tagName])
				//}
				if err := tracksBatch.Index(string(key), val); err != nil {
					return err
				}
				trackId++
			}
			discId++
		}
	}

	err = tracksSearchIdx.Batch(tracksBatch)
	if err != nil {
		return err
	}

	albumsSearchIdx, err = bleve.New("", mapping)
	albumsBatch := albumsSearchIdx.NewBatch()
	for _, v := range albumIdx {
		key, _ := json.Marshal(v)
		val := albumDetails{AlbumDetails: v}
		for _, tagName := range v.Tags {
			val.Tags = append(val.Tags, tagNameIdx[tagName])
		}
		if err := albumsBatch.Index(string(key), val); err != nil {
			return err
		}
	}

	err = albumsSearchIdx.Batch(albumsBatch)
	if err != nil {
		return err
	}

	log.Printf("Done, took %d ms.\n", time.Now().Sub(start).Milliseconds())

	return nil
}

func SearchAlbums(keyword string) []AlbumDetails {
	query := bleve.NewMatchQuery(keyword)
	search := bleve.NewSearchRequest(query)
	searchResults, err := albumsSearchIdx.Search(search)
	if err != nil {
		return nil
	}

	var res []AlbumDetails
	sort.Sort(searchResults.Hits)

	for _, v := range searchResults.Hits {
		entry := AlbumDetails{}
		if err := json.Unmarshal([]byte(v.ID), &entry); err != nil {
			panic(err)
		}
		res = append(res, entry)
	}

	return res
}

func SearchTracks(keyword string) []TrackInfoWithAlbum {
	query := bleve.NewMatchQuery(keyword)
	search := bleve.NewSearchRequest(query)
	searchResults, err := tracksSearchIdx.Search(search)
	if err != nil {
		return nil
	}

	var res []TrackInfoWithAlbum
	sort.Sort(searchResults.Hits)

	for _, v := range searchResults.Hits {
		entry := TrackInfoWithAlbum{}
		if err := json.Unmarshal([]byte(v.ID), &entry); err != nil {
			panic(err)
		}
		res = append(res, entry)
	}

	return res
}
