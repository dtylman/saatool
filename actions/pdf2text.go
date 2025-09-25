package actions

import (
	"context"
	"fmt"

	"github.com/ledongthuc/pdf"
	"github.com/urfave/cli/v3"
)

// PDFToText represents the action to add a user
type PDFToText struct {
}

// Name returns the name of the action
func (a *PDFToText) Name() string {
	return "to-text"
}

// Usage returns the usage string for the action
func (a *PDFToText) Usage() string {
	return "Convert a PDF file to text"
}

// Flags returns the flags for the action
func (a *PDFToText) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Aliases:  []string{"i"},
			Usage:    "Input PDF file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "output",
			Aliases:  []string{"o"},
			Usage:    "Output text file",
			Required: true,
		},
	}
}

func (a *PDFToText) Action(ctx context.Context, cmd *cli.Command) error {
	input := cmd.String("input")
	output := cmd.String("output")

	fmt.Printf("Converting PDF file: %s to text file: %s\n", input, output)

	file, reader, err := pdf.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	sentences, err := reader.GetStyledTexts()
	if err != nil {
		return fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	for _, s := range sentences {
		fmt.Println(s.S)
	}

	return nil
}
