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

// PDFImportAction represents the action to import a PDF file
type PDFImportAction struct {
	textBlocks []TextBlock
}

// TextBlock represents a block of text in the PDF
type TextBlock struct {
	Text          string
	count         int
	totalFontSize float64

	x, y          float64
	width, height float64
}

func newTextBlock(t pdf.Text) *TextBlock {
	return &TextBlock{
		Text:          t.S,
		count:         1,
		totalFontSize: math.Abs(t.FontSize),
		x:             t.X,
		y:             t.Y,
		width:         t.W,
		height:        math.Abs(t.FontSize),
	}
}

// addText adds a text element to the block and updates its bounding box
func (tb *TextBlock) addText(t pdf.Text) {
	tb.Text += t.S
	tb.count++
	tb.totalFontSize += math.Abs(t.FontSize)
}

func (tb *TextBlock) avgFontSize() float64 {
	if tb.count == 0 {
		return 0
	}
	return tb.totalFontSize / float64(tb.count)
}

// contains checks if the point (x, y) is within the block's bounding box with a threshold
func (tb *TextBlock) contains(x, y float64) bool {
	xThreshold := tb.avgFontSize() * 2
	yThreshold := tb.avgFontSize() * 2
	minX := tb.x - xThreshold
	maxX := tb.x + tb.width + xThreshold
	minY := tb.y - yThreshold
	maxY := tb.y + tb.height + yThreshold
	return x >= minX && x <= maxX && y >= minY && y <= maxY
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

func (a *PDFImportAction) createProject(ctx context.Context, input string, cmd *cli.Command) error {
	file, reader, err := pdf.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	// pages := reader.NumPage()
	p := reader.Page(2)

	texts := p.Content().Text

	maxHeight := 0.0
	maxWidth := 0.0
	for i, t := range texts {
		h := math.Abs(t.FontSize) + t.Y
		if h > maxHeight {
			maxHeight = h
		}
		if t.W == 0.0 {
			t.W = math.Abs(t.FontSize) * 0.6 * float64(len(t.S))
			texts[i] = t
		}
		if t.X+t.W > maxWidth {
			maxWidth = t.X + t.W
		}
	}

	for i, t := range texts {
		// invert the Y coordinate
		t.Y = maxHeight - t.Y
		texts[i] = t
		a.addTextBlock(t)
	}

	for i, tb := range a.textBlocks {
		fmt.Printf("Block %d: (x=%.2f,y=%.2f,w=%.2f,h=%.2f,fs=%.2f) %s\n", i, tb.x, tb.y, tb.width, tb.height, tb.avgFontSize(), tb.Text)
	}
	return nil
}

func (a *PDFImportAction) findBlock(t pdf.Text) *TextBlock {
	for i, tb := range a.textBlocks {
		if tb.contains(t.X, t.Y) {
			return &a.textBlocks[i] // return pointer to the block in the slice
		}
	}
	return nil
}

func (a *PDFImportAction) addTextBlock(t pdf.Text) {
	fmt.Println(t.X, t.Y, t.X+t.W, t.Y+math.Abs(t.FontSize), t.S)
	block := a.findBlock(t)
	if block != nil {
		block.addText(t)
		return
	}
	block = newTextBlock(t)
	a.textBlocks = append(a.textBlocks, *block)
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
