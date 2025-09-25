package actions

import (
	"context"
	"fmt"

	"github.com/ledongthuc/pdf"
	"github.com/urfave/cli/v3"
)

// PDFImportAction represents the action to import a PDF file
type PDFImportAction struct {
}

// Name returns the name of the action
func (a *PDFImportAction) Name() string {
	return "pdf"
}

// Usage returns the usage string for the action
func (a *PDFImportAction) Usage() string {
	return "Import a PDF file"
}

// Flags returns the flags for the action
func (a *PDFImportAction) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Aliases:  []string{"i"},
			Usage:    "Input EPUB file",
			Required: true,
		},
		&cli.IntFlag{
			Name:    "max-words",
			Aliases: []string{"m"},
			Usage:   "Maximum words per paragraph before considering a split",
			Value:   200,
		},
		&cli.IntFlag{
			Name:    "max-words-tolerance",
			Aliases: []string{"t"},
			Usage:   "Maximum words per paragraph before forcing a split",
			Value:   300,
		},
		&cli.BoolFlag{
			Name:    "strip-to-ascii",
			Aliases: []string{"s"},
			Usage:   "Strip non-ASCII characters from the text",
			Value:   false,
		},
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "Source language (e.g. english, french, german)",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "to",
			Aliases:  []string{"o"},
			Usage:    "Target language code (e.g. polish, spanish, italian)",
			Value:    "",
			Required: true,
		},
	}
}

func (a *PDFImportAction) Action(ctx context.Context, cmd *cli.Command) error {
	input := cmd.String("input")

	fmt.Printf("Importing PDF file: %s\n", input)

	file, reader, err := pdf.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	totalPage := reader.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := reader.Page(pageIndex)
		if p.V.IsNull() || p.V.Key("Contents").Kind() == pdf.Null {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			// Each row may contain multiple words
			for _, word := range row.Content {
				fmt.Println(word.S)
			}
		}
	}
	return nil

	// sentences, err := reader.GetStyledTexts()
	// if err != nil {
	// 	return fmt.Errorf("failed to extract text from PDF: %w", err)
	// }

	// for _, s := range sentences {

	// }

	return nil
}
