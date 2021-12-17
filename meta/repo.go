package meta

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var lock = &sync.RWMutex{}

func Init(path, url string) error {
	log.Println("Initializing meta index...")
	err := updateRepo(path, url)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	err = updateIndex(path)
	if err != nil {
		return err
	}
	log.Println("Meta initialization complete.")
	go func() {
		ticker := time.NewTicker(time.Hour)
		for {
			<-ticker.C
			log.Println("Syncing meta repo...")
			err := updateRepo(path, url)
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("Failed to update repo: %v\n", err)
			} else if err == git.NoErrAlreadyUpToDate {
				log.Println("Already up to date.")
			} else {
				err = updateIndex(path)
				if err != nil {
					log.Printf("Failed to index repo: %v\n", err)
				}
			}
		}
	}()
	return nil
}

func updateIndex(path string) error {
	lock.Lock()
	defer lock.Unlock()
	log.Println("Indexing repo...")
	start := time.Now()
	err := Read(path)
	if err == nil {
		took := time.Now().Sub(start)
		log.Printf("Index done, took %d ms.\n", took.Milliseconds())
	}
	return err
}

func updateRepo(path, url string) error {
	s, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = initRepo(path, url)
		if err != nil {
			return err
		}
	} else {
		if !s.IsDir() {
			return errors.New("must be a dir")
		}
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	return err
}

func initRepo(path, url string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})
	return err
}

func GetTags() []string {
	lock.RLock()
	defer lock.RUnlock()
	return tags
}

func GetAlbumInfo(id string) (AlbumInfo, bool) {
	lock.RLock()
	defer lock.RUnlock()
	res, ok := albumIdx[id]
	return res, ok
}

func GetAlbumsByTag(tag string) ([]AlbumInfo, bool) {
	lock.RLock()
	defer lock.RUnlock()
	res, ok := tagIdx[tag]
	return res, ok
}

func SearchAlbums(keyword string) (ret []AlbumInfo) {
	lock.RLock()
	defer lock.RUnlock()
	ret = make([]AlbumInfo, 0)
	if keyword == "" {
		return
	}
	for _, v := range albumIdx {
		if strings.Contains(strings.ToUpper(v.Title), strings.ToUpper(keyword)) {
			ret = append(ret, v)
		}
	}
	return
}

func SearchTracks(keyword string) (ret []*TrackInfoWithAlbum) {
	lock.RLock()
	defer lock.RUnlock()
	ret = make([]*TrackInfoWithAlbum, 0)
	if keyword == "" {
		return
	}
	for _, album := range albumIdx {
		for discId, disc := range album.Discs {
			for trackId, track := range disc.Tracks {
				if strings.Contains(strings.ToUpper(track.Title), strings.ToUpper(keyword)) {
					ret = append(ret, &TrackInfoWithAlbum{
						Title:   track.Title,
						Artist:  track.Artist,
						Type:    track.Type,
						Tags:    track.Tags,
						TrackID: trackId + 1,
						DiscID:  discId + 1,
						AlbumID: album.AlbumID,
					})
				}
			}
		}
	}
	return
}

func GetTagGraph() map[string][]string {
	return tagGraph
}
