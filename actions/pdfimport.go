package actions

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/urfave/cli/v3"
)

type TextBlock struct {
	Text          string
	totalFontSize float64
	items         int
}

func (t *TextBlock) deltaY(t pdf.Text) float64 {

}

// PDFImportAction represents the action to import a PDF file
type PDFImportAction struct {
	textBlocks []TextBlock
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
		&cli.BoolFlag{
			Name:    "ocr",
			Aliases: []string{"c"},
			Usage:   "Force OCR even if text is detected",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "ocr-langs",
			Aliases: []string{"l"},
			Usage:   "Languages for OCR, comma separated (e.g. eng,pol)",
			Value:   "eng",
		},
	}
}

func (a *PDFImportAction) Action(ctx context.Context, cmd *cli.Command) error {
	input := cmd.String("input")

	log.Printf("Importing PDF file: %s\n", input)

	var err error

	// check if file exists
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", input)
	}

	needOCR := cmd.Bool("ocr")
	if !needOCR {

		needOCR, err = a.needsOCR(input)
		if err != nil {
			return fmt.Errorf("failed to determine if OCR is needed: %w", err)
		}
	}
	if needOCR {
		input, err = a.runOCR(input, ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to run OCR: %w", err)
		}
		fmt.Printf("OCR completed, new file: %s\n", input)
	}

	return a.createProject(ctx, input, cmd)

	// totalPage := reader.NumPage()

	// texts, err := reader.GetStyledTexts()
	// if err != nil {
	// 	return fmt.Errorf("failed to extract text from PDF: %w", err)
	// }

	// markdown := ""

	// for _, text := range texts {

	// 	if text.FontSize > 20 {
	// 		markdown += "# "
	// 	} else if text.FontSize > 15 {
	// 		markdown += "## "
	// 	} else if text.FontSize > 12 {
	// 		markdown += "### "
	// 	}
	// 	markdown += text.S
	// }

	// fmt.Printf("Total pages: %d\n", totalPage)
	// fmt.Printf("Extracted text length: %d\n", len(markdown))

	// fmt.Println(markdown)
	// sentences, err := reader.GetStyledTexts()
	// if err != nil {
	// 	return fmt.Errorf("failed to extract text from PDF: %w", err)
	// }

	// for _, s := range sentences {

	// }

	return nil
}

func (a *PDFImportAction) getLastBlock(t pdf.Text) *TextBlock {
	if len(a.textBlocks) == 0 {
		a.textBlocks = append(a.textBlocks, TextBlock{})
	}
	return &a.textBlocks[len(a.textBlocks)-1]
}

func (a *PDFImportAction) createProject(ctx context.Context, input string, cmd *cli.Command) error {
	file, reader, err := pdf.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	// pages := reader.NumPage()
	p := reader.Page(2)

	texts := p.Content().Text

	text := ""
	totalFontSize := 0.0
	numFonts := 0
	lastY := -1.0
	for _, t := range texts {

		deltaY := math.Abs(t.Y - lastY)
		if deltaY > 2.0 {
			log.Printf("New line detected at Y=%.2f (last Y=%.2f), inserting line break", t.Y, lastY)
			text += "\n"
			lastY = t.Y
		}
		numFonts++
		totalFontSize += math.Abs(t.FontSize)
		averageFontSize := float64(totalFontSize) / float64(numFonts)
		fontDelta := math.Abs(t.FontSize) - averageFontSize
		if fontDelta > 2.0 {
			log.Printf("Font size change detected: %.2f -> %.2f (delta %.2f), inserting line break", averageFontSize, t.FontSize, fontDelta)
			text += "\n"
			numFonts = 0
			totalFontSize = 0.0
		}
		text += t.S
	}

	fmt.Println(text)
	return nil
}

// needsOCR checks if the PDF file likely needs OCR by analyzing its content.
func (a *PDFImportAction) needsOCR(fileName string) (bool, error) {
	log.Printf("Checking if PDF %s needs OCR", fileName)
	file, reader, err := pdf.Open(fileName)
	if err != nil {
		return false, fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	// heuristic: if the PDF has very little extractable text compared to the number of pages, OCR is likely needed.
	texts, err := reader.GetStyledTexts()
	if err != nil {
		return true, nil // If text extraction fails, assume OCR is needed.
	}
	if len(texts) == 0 {
		return true, nil
	}
	// Count total words extracted
	wordCount := 0
	for _, t := range texts {
		wordCount += len(strings.Fields(t.S))
	}
	// If average words per page is very low, assume mostly images
	numPages := reader.NumPage()
	if numPages == 0 {
		return true, nil
	}
	avgWordsPerPage := float64(wordCount) / float64(numPages)
	log.Printf("PDF %s has %d pages and %d words (avg %.2f words/page)", fileName, numPages, wordCount, avgWordsPerPage)
	return avgWordsPerPage < 10, nil // threshold: less than 10 words per page
}

// runOCR runs an external OCR tool on the PDF file and returns the path to the new OCRed PDF.
func (a *PDFImportAction) runOCR(fileName string, ctx context.Context, cmd *cli.Command) (string, error) {
	log.Printf("Running OCR on PDF %s", fileName)

	// check if ocrmypdf is installed
	_, err := exec.LookPath("ocrmypdf")
	if err != nil {
		return "", fmt.Errorf("ocrmypdf not found in PATH, please install it to use OCR functionality")
	}

	langs := cmd.String("ocr-langs")
	if langs == "" {
		langs = "eng"
	}

	outputFile := strings.TrimSuffix(fileName, ".pdf") + "_ocr.pdf"
	args := []string{"-l", langs, "--rotate-pages", "--deskew", fileName, outputFile}

	if cmd.Bool("strip-to-ascii") {
		args = append(args, "--remove-background")
	}

	if cmd.Bool("ocr") {
		args = append(args, "--force-ocr")
	}

	log.Printf("Executing command: ocrmypdf %s", strings.Join(args, " "))
	ocrCmd := exec.CommandContext(ctx, "ocrmypdf", args...)
	ocrCmd.Stdout = os.Stdout
	ocrCmd.Stderr = os.Stderr
	ocrCmd.Stdin = os.Stdin
	err = ocrCmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run ocrmypdf: %w", err)
	}
	return outputFile, nil

}
