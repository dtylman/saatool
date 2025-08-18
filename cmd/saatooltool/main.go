package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/taylorskalyo/goreader/epub"
)

func printItem(item epub.Itemref, bp *bluemonday.Policy) {
	fmt.Printf("============[ %v %v %v] =========:\n", item.ID, item.HREF, item.MediaType)
	reader, err := item.Open()
	if err != nil {
		log.Printf("Error opening item %s: %v", item.ID, err)
		return
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Error reading item %s: %v", item.ID, err)
		return
	}

	text := string(content)
	text = bp.Sanitize(text)
	text = html.UnescapeString(text)
	text = strings.TrimSpace(text)
	fmt.Println(text)
}

func work() error {
	file := "luciper.epub"

	rc, err := epub.OpenReader(file)
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	// The rootfile (content.opf) lists all of the contents of an epub file.
	// There may be multiple rootfiles, although typically there is only one.
	book := rc.Rootfiles[0]

	// Print book title.
	fmt.Println("Title: ", book.Title)
	bp := bluemonday.StrictPolicy()
	// List the IDs of files in the book's spine.
	for _, item := range book.Spine.Itemrefs {
		printItem(item, bp)
		fmt.Scanln()
	}
	return nil
}

func main() {
	if err := work(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
