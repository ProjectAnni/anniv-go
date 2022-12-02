package meta

import (
	"encoding/json"
	"github.com/blevesearch/bleve/v2"
	"log"
	"sort"
	"time"
)

var tracksIdx bleve.Index
var albumsIdx bleve.Index

func initSearchIndex() error {
	log.Println("Building search index...")
	start := time.Now()
	var err error
	lock.RLock()
	defer lock.RUnlock()
	mapping := bleve.NewIndexMapping()
	tracksIdx, err = bleve.New("", mapping)
	if err != nil {
		return err
	}
	tracksBatch := tracksIdx.NewBatch()
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
				if err := tracksBatch.Index(string(key), t); err != nil {
					return err
				}
				trackId++
			}
			discId++
		}
	}

	err = tracksIdx.Batch(tracksBatch)
	if err != nil {
		return err
	}

	albumsIdx, err = bleve.New("", mapping)
	albumsBatch := albumsIdx.NewBatch()
	for _, v := range albumIdx {
		key, _ := json.Marshal(v)
		if err := albumsBatch.Index(string(key), v); err != nil {
			return err
		}
	}

	err = albumsIdx.Batch(albumsBatch)
	if err != nil {
		return err
	}

	log.Printf("Done, took %d ms.\n", time.Now().Sub(start).Milliseconds())

	return nil
}

func SearchAlbums(keyword string) []AlbumDetails {
	query := bleve.NewMatchQuery(keyword)
	search := bleve.NewSearchRequest(query)
	searchResults, err := albumsIdx.Search(search)
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
	searchResults, err := tracksIdx.Search(search)
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
