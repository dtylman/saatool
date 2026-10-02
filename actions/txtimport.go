package actions

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

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
		// avoid duplicating paragraphs on re-import
		a.project.Source.Paragraphs = nil
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

// readTextParagraphs reads a text file and splits it into paragraphs. Paragraphs are
// separated by one or more blank lines; lines within a paragraph are joined with a
// single space and runs of whitespace are collapsed.
func readTextParagraphs(fileName string, dehyphenate bool) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var paragraphs []string
	var current strings.Builder

	flush := func() {
		text := strings.TrimSpace(current.String())
		if text != "" {
			paragraphs = append(paragraphs, text)
		}
		current.Reset()
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.Join(strings.Fields(strings.ReplaceAll(scanner.Text(), "\f", " ")), " ")
		if line == "" {
			flush()
			continue
		}
		if current.Len() == 0 {
			current.WriteString(line)
			continue
		}
		prev := current.String()
		if dehyphenate && joinsHyphenated(prev, line) {
			current.Reset()
			current.WriteString(strings.TrimSuffix(prev, "-"))
			current.WriteString(line)
			continue
		}
		current.WriteString(" ")
		current.WriteString(line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	flush()
	return paragraphs, nil
}

// joinsHyphenated returns true when prev ends with a letter followed by a hyphen
// and next starts with a lowercase letter (a word broken across lines).
func joinsHyphenated(prev, next string) bool {
	if !strings.HasSuffix(prev, "-") || len(prev) < 2 {
		return false
	}
	runs := []rune(prev)
	if !unicode.IsLetter(runs[len(runs)-2]) {
		return false
	}
	return unicode.IsLower([]rune(next)[0])
}
