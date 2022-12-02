package meta

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
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
	err = initSearchIndex()
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
				err = initSearchIndex()
				if err != nil {
					log.Printf("Failed to update search index: %v\n", err)
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
			_ = os.RemoveAll(path)
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

func GetTags() []Tag {
	lock.RLock()
	defer lock.RUnlock()
	return tags
}

func GetAlbumDetails(id string) (AlbumDetails, bool) {
	lock.RLock()
	defer lock.RUnlock()
	res, ok := albumIdx[AlbumIdentifier(id)]
	return res, ok
}

func GetAlbumsByTag(tag string, recursive bool) ([]AlbumDetails, bool) {
	lock.RLock()
	defer lock.RUnlock()
	if recursive {
		res, ok := tagIdx[tag]
		return res, ok
	} else {
		res, ok := tagIdxNonRecursive[tag]
		return res, ok
	}
}

func GetTagGraph() map[string][]string {
	lock.RLock()
	defer lock.RUnlock()
	return tagGraph
}

func GetAlbums() []AlbumDetails {
	lock.RLock()
	defer lock.RUnlock()
	res := make([]AlbumDetails, 0, len(albumIdx))
	for _, v := range albumIdx {
		res = append(res, v)
	}
	return res
}

func GetTrackInfo(id TrackIdentifier) TrackInfoWithAlbum {
	lock.RLock()
	defer lock.RUnlock()
	ret := TrackInfoWithAlbum{
		TrackIdentifier: id,
	}
	album, ok := albumIdx[id.AlbumID]
	if !ok {
		return ret
	}
	ret.AlbumTitle = album.Title
	if id.DiscID > uint(len(album.Discs)) {
		return ret
	}
	disc := album.Discs[id.DiscID-1]
	if id.TrackID > uint(len(disc.Tracks)) {
		return ret
	}
	track := disc.Tracks[id.TrackID-1]
	ret.TrackInfo = track.TrackInfo
	return ret
}
