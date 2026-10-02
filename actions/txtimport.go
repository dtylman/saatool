package actions

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/translation"
	"github.com/urfave/cli/v3"
)

// TextImportAction represents the action to import a plain text file (e.g. created by pdf2txt)
type TextImportAction struct {
	project *translation.Project
}

// Name returns the name of the action
func (a *TextImportAction) Name() string {
	return "txt"
}

// Usage returns the usage string for the action
func (a *TextImportAction) Usage() string {
	return "Import a plain text file (paragraphs are separated by blank lines)"
}

// Flags returns the flags for the action
func (a *TextImportAction) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Aliases:  []string{"i"},
			Usage:    "Input text file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "Source language (e.g. english, french, german)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "to",
			Aliases:  []string{"o"},
			Usage:    "Target language (e.g. polish, spanish, italian)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "author",
			Aliases:  []string{"a"},
			Usage:    "Author of the document",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "title",
			Aliases:  []string{"t"},
			Usage:    "Title of the document",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "synopsis",
			Aliases:  []string{"s"},
			Usage:    "Synopsis of the document",
			Required: false,
		},
		&cli.BoolFlag{
			Name:    "details",
			Aliases: []string{"d"},
			Usage:   "Get book details from DeepSeek",
			Value:   true,
		},
		&cli.StringFlag{
			Name:  "style",
			Usage: "Translation prompt style (strict, academic, literary, archaic, rap)",
			Value: "strict",
		},
		&cli.BoolFlag{
			Name:  "dehyphenate",
			Usage: "Join words hyphenated at the end of a line (e.g. func-\\ntioned -> functioned)",
			Value: true,
		},
		&cli.IntFlag{
			Name:  "skip-paragraphs",
			Usage: "Number of paragraphs to skip from the start",
			Value: 0,
		},
	}
}

// Action executes the import
func (a *TextImportAction) Action(ctx context.Context, cmd *cli.Command) error {
	input := cmd.String("input")
	log.Printf("Importing text file: %s", input)

	if _, err := os.Stat(input); err != nil {
		return fmt.Errorf("input file error: %w", err)
	}

	if err := a.createProject(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	paragraphs, err := readTextParagraphs(input, cmd.Bool("dehyphenate"))
	if err != nil {
		return fmt.Errorf("failed to read text file: %w", err)
	}

	skip := cmd.Int("skip-paragraphs")
	if skip < 0 || skip > len(paragraphs) {
		return fmt.Errorf("invalid skip-paragraphs %d (file has %d paragraphs)", skip, len(paragraphs))
	}
	paragraphs = paragraphs[skip:]

	for _, text := range paragraphs {
		paragraph := translation.Paragraph{Text: text}
		paragraph.ID = paragraph.CalcID()
		a.project.Source.Paragraphs = append(a.project.Source.Paragraphs, paragraph)
	}
	log.Printf("Imported %d paragraphs", len(paragraphs))

	a.project.Normalize()
	for i := range a.project.Source.Paragraphs {
		a.project.Target.Paragraphs[i].ID = a.project.Source.Paragraphs[i].ID
	}

	fileName, err := a.project.Save()
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}
	log.Printf("Project saved to %s", fileName)
	return nil
}

func (a *TextImportAction) createProject(ctx context.Context, cmd *cli.Command) error {
	title := cmd.String("title")
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	a.project = translation.NewProject(title)

	projectFileName := a.project.ProjectFileName()
	existingProject, err := translation.LoadProject(projectFileName)
	if err == nil && existingProject != nil {
		log.Printf("Project file %s already exists, loading existing project", projectFileName)
		a.project = existingProject
		// start from scratch on re-import to avoid duplicates and stale translations
		a.project.Source.Paragraphs = nil
		a.project.Target.Paragraphs = nil
		return nil
	}
	log.Printf("Creating new project: %s", title)

	a.project.Title = title
	a.project.Author = cmd.String("author")
	a.project.Synopsis = cmd.String("synopsis")
	a.project.Source.Language = cmd.String("from")
	a.project.Target.Language = cmd.String("to")
	a.project.Style = cmd.String("style")

	if !cmd.Bool("details") {
		return nil
	}

	translator, err := ai.NewTranslator(a.project)
	if err != nil {
		return fmt.Errorf("failed to create translator: %w", err)
	}
	bookDetails, err := translator.PopulateBookDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to get book details: %w", err)
	}
	a.project.Title = bookDetails.Title
	a.project.Author = bookDetails.Author
	a.project.Synopsis = bookDetails.Synopsis
	a.project.Genre = bookDetails.Genre
	return nil
}

var (
	pageNumberRegex  = regexp.MustCompile(`^(\d{1,4}|[ivxlcIVXLC]{1,7})$`)
	sentenceEndRegex = regexp.MustCompile(`[.!?:;\x22\x27)\]\x{201D}\x{2019}]\d*$`)
)

// readTextParagraphs reads a text file and splits it into paragraphs.
//
// Text produced by pdf2txt is hard-wrapped and contains spurious blank lines, page
// numbers and running headers, so the splitting works as follows:
//   - page numbers (a line with only a number) and running headers (short lines that
//     repeat many times) are dropped
//   - blank lines separate blocks; lines inside a block are joined with a single space
//   - a block that starts with a lowercase letter, following a block that does not end
//     with sentence punctuation, continues that block (page/line breaks mid-sentence)
//   - optionally, words hyphenated at the end of a line are joined
func readTextParagraphs(fileName string, dehyphenate bool) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	counts := make(map[string]int)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.Join(strings.Fields(strings.ReplaceAll(scanner.Text(), "\f", " ")), " ")
		lines = append(lines, line)
		counts[line]++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// build blocks separated by blank lines
	var paragraphs []string
	var current string
	var block string
	flushBlock := func() {
		if block == "" {
			return
		}
		if current != "" && continuesParagraph(current, block) {
			current = joinLines(current, block, dehyphenate)
		} else {
			if current != "" {
				paragraphs = append(paragraphs, current)
			}
			current = block
		}
		block = ""
	}
	for _, line := range lines {
		if line == "" {
			flushBlock()
			continue
		}
		if pageNumberRegex.MatchString(line) || (len(line) < 80 && counts[line] >= 5) {
			continue // page number or running header
		}
		if block == "" {
			block = line
		} else {
			block = joinLines(block, line, dehyphenate)
		}
	}
	flushBlock()
	if current != "" {
		paragraphs = append(paragraphs, current)
	}
	return paragraphs, nil
}

// continuesParagraph returns true when next looks like the continuation of prev.
func continuesParagraph(prev, next string) bool {
	first, _ := utf8.DecodeRuneInString(next)
	if !unicode.IsLower(first) {
		return false
	}
	return !sentenceEndRegex.MatchString(prev)
}

// joinLines joins two pieces of text, merging a hyphenated word break if requested.
func joinLines(prev, next string, dehyphenate bool) string {
	if dehyphenate && strings.HasSuffix(prev, "-") && len(prev) >= 2 {
		runs := []rune(prev)
		first, _ := utf8.DecodeRuneInString(next)
		if unicode.IsLetter(runs[len(runs)-2]) && unicode.IsLower(first) {
			return strings.TrimSuffix(prev, "-") + next
		}
	}
	return prev + " " + next
}
