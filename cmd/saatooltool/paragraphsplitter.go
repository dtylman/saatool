package main

import "strings"

type ParagraphSplitter struct {
	maxLines   int
	paragraphs []string
}

func NewParagraphSplitter(maxLines int) *ParagraphSplitter {
	return &ParagraphSplitter{
		maxLines:   maxLines,
		paragraphs: make([]string, 0),
	}
}

func (ps *ParagraphSplitter) addParagraph(text string) {
	if text == "" {
		return
	}
	ps.paragraphs = append(ps.paragraphs, text)
}

func (ps *ParagraphSplitter) Split(text string) []string {
	ps.paragraphs = make([]string, 0)
	lines := strings.Split(text, "\n")
	currentParagraph := strings.Builder{}
	count := 0
	for _, line := range lines {
		if len(line) > 1 {
			if line[0] == ' ' && line[1] != ' ' {
				ps.addParagraph(currentParagraph.String())
				currentParagraph.Reset()
			}
		}
		line = strings.TrimSpace(line)
		if line == "" {
			ps.addParagraph(currentParagraph.String())
			currentParagraph.Reset()
		} else {
			count++
			if count > ps.maxLines {
				ps.addParagraph(currentParagraph.String())
				currentParagraph.Reset()
				count = 1 // Reset count for the new paragraph
			}
			currentParagraph.WriteString(line + "\n")
		}
	}
	ps.addParagraph(currentParagraph.String())
	return ps.paragraphs
}
