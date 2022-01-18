package main

import (
	"fmt"
	"github.com/ProjectAnni/anniv-go/meta"
	"os"
)

func main() {
	p := os.Args[1]
	err := meta.Read(p)
	checkErr(err)
	fmt.Println("Read complete, no error detected.")
	fmt.Printf("%d albums, %d tags in total.\n", len(meta.GetAlbums()), len(meta.GetTags()))
}

func checkErr(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
