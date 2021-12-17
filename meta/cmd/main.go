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
	graph := meta.GetTagGraph()
	fmt.Println("digraph {")
	for u, t := range graph {
		for _, v := range t {
			fmt.Printf("    \"%s\" -> \"%s\"\n", u, v)
		}
	}
	fmt.Println("}")
}

func checkErr(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
