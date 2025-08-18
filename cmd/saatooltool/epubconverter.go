package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"strings"

	"github.com/dtylman/saatool/translation"
	"github.com/microcosm-cc/bluemonday"
	"github.com/taylorskalyo/goreader/epub"
)

type EPubConverter struct {
	rc       *epub.ReadCloser
	Project  *translation.Project
	bp       *bluemonday.Policy
	maxLines int
}

func NewEPubConverter() *EPubConverter {
	return &EPubConverter{
		rc:       nil,
		Project:  translation.NewProject(),
		bp:       bluemonday.StrictPolicy(),
		maxLines: 5,
	}
}

func (ec *EPubConverter) ConvertEPub(fileName string) error {
	log.Printf("Converting EPUB file: %v", fileName)

	var err error
	ec.rc, err = epub.OpenReader(fileName)
	if err != nil {
		return fmt.Errorf("failed to open EPUB file %s: %w", fileName, err)
	}
	defer ec.rc.Close()

	if len(ec.rc.Rootfiles) == 0 {
		return fmt.Errorf("no root files found in EPUB")
	}

	ec.Project = translation.NewProject()
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

	text := string(content)
	text = ec.bp.Sanitize(text)
	text = html.UnescapeString(text)
	text = strings.TrimSpace(text)

	splitter := NewParagraphSplitter(ec.maxLines)
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
