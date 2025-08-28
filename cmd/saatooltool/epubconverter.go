package main

import (
	"fmt"
	"io"
	"log"
	"path"
	"strings"

	"github.com/dtylman/saatool/translation"
	"github.com/taylorskalyo/goreader/epub"
	"jaytaylor.com/html2text"
)

// EPubConverter handles converting EPUB files to text and preparing them for translation
type EPubConverter struct {
	rc      *epub.ReadCloser
	Project *translation.Project
}

// NewEPubConverter creates a new EPubConverter instance
func NewEPubConverter() *EPubConverter {
	return &EPubConverter{
		rc:      nil,
		Project: nil,
	}

}

// ConvertEPub converts the given EPUB file to text and prepares it for translation
func (ec *EPubConverter) ConvertEPub(fileName string) error {
	log.Printf("Converting EPUB file: %v", fileName)

	ec.Project = translation.NewProject(fileName)

	var err error
	ec.rc, err = epub.OpenReader(fileName)
	if err != nil {
		return fmt.Errorf("failed to open EPUB file %s: %w", fileName, err)
	}
	defer ec.rc.Close()

	if len(ec.rc.Rootfiles) == 0 {
		return fmt.Errorf("no root files found in EPUB")
	}

	_, name := path.Split(fileName)
	ec.Project = translation.NewProject(name)
	book := ec.rc.Rootfiles[0]

	ec.Project.Title = book.Title
	ec.Project.Source.Language = book.Language
	ec.Project.Source.Paragraphs = make([]translation.Paragraph, 0)
	ec.Project.Target.Language = ""
	ec.Project.Target.Paragraphs = make([]translation.Paragraph, 0)

	for _, item := range book.Spine.Itemrefs {
		err = ec.processItem(item)
		if err != nil {
			return fmt.Errorf("error processing item %s: %w", item.ID, err)
		}
	}

	ec.Project.Normalize()

	return nil
}

func (ec *EPubConverter) processItem(item epub.Itemref) error {
	reader, err := item.Open()
	if err != nil {
		return fmt.Errorf("failed to open item %s: %w", item.ID, err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content of item %s: %w", item.ID, err)
	}

	text, err := html2text.FromString(string(content), html2text.Options{PrettyTables: true, OmitLinks: true, TextOnly: false})
	if err != nil {
		return fmt.Errorf("failed to convert HTML to text for item %s: %w", item.ID, err)
	}

	if strings.TrimSpace(text) == "" {
		return nil
	}

	text = removeEmptyLines(text)

	splitter := NewParagraphSplitter()

	paragraphs := splitter.Split(text)
	for _, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph) == "" {
			continue
		}
		p := translation.Paragraph{
			Text: paragraph,
		}
		ec.Project.Source.Paragraphs = append(ec.Project.Source.Paragraphs, p)
	}
	return nil
}

func removeEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	nonEmptyLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	return strings.Join(nonEmptyLines, "\n")
}
