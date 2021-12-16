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
	log.Println("Meta initialization complete.")
	go func() {
		ticker := time.NewTicker(time.Hour)
		for {
			<-ticker.C
			log.Println("Syncing meta repo...")
			err := updateRepo(path, url)
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("Failed to update repo: %v\n", err)
			}
			if err == git.NoErrAlreadyUpToDate {
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
