package actions

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/dtylman/aitasks/tasks/ocr"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/translation"
	"github.com/ledongthuc/pdf"
	"github.com/urfave/cli/v3"
)

// PDFImportAction represents the action to import a PDF file
type PDFImportAction struct {
	project *translation.Project
	task    *ocr.Task
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
			Usage:    "Input PDF file",
			Required: true,
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
		&cli.StringFlag{
			Name:     "author",
			Aliases:  []string{"a"},
			Usage:    "Author of the document",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "title",
			Aliases:  []string{"t"},
			Usage:    "Title of the document",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "synopsis",
			Aliases:  []string{"s"},
			Usage:    "Synopsis of the document",
			Value:    "",
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
		&cli.IntFlag{
			Name:    "skip-pages",
			Aliases: []string{"p"},
			Usage:   "Number of pages to skip from the start",
			Value:   0,
		},
	}
}

func (a *PDFImportAction) Action(ctx context.Context, cmd *cli.Command) error {
	input := cmd.String("input")

	log.Printf("Importing PDF file: %s\n", input)

	// check if file exists

	_, err := os.Stat(input)
	if os.IsNotExist(err) {
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

	err = a.createProject(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	if a.project == nil {
		return fmt.Errorf("project is nil")
	}

	err = a.processPDFFile(ctx, input, cmd)
	if err != nil {
		return fmt.Errorf("failed to process PDF file: %w", err)
	}

	a.project.Normalize()

	fileName, err := a.project.Save()
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	log.Printf("Project saved to %s", fileName)

	return nil
}

func (a *PDFImportAction) processPDFFile(ctx context.Context, input string, cmd *cli.Command) error {
	log.Printf("Processing PDF file: %s\n", input)
	file, reader, err := pdf.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	skipPages := cmd.Int("skip-pages")
	pages := reader.NumPage()
	for i := skipPages; i <= pages; i++ {
		p := reader.Page(i)
		if p.V.IsNull() {
			continue
		}
		err := a.processPage(ctx, cmd, &p, i)
		if err != nil {
			return fmt.Errorf("failed to process page %d: %w", i, err)
		}
	}

	return nil
}

func (a *PDFImportAction) processPage(ctx context.Context, cmd *cli.Command, p *pdf.Page, pageNum int) error {
	log.Printf("Processing page %d", pageNum)

	if p.V.IsNull() {
		return nil
	}

	rows, err := p.GetTextByRow()
	if err != nil {
		return fmt.Errorf("failed to get text by row for page %d: %w", pageNum, err)
	}

	avgLineDetla := int64(0)
	for i := 1; i < len(rows); i++ {
		lineDelta := rows[i].Position - rows[i-1].Position
		avgLineDetla += lineDelta
	}
	if len(rows) > 1 {
		avgLineDetla /= int64(len(rows) - 1)
	}

	var pageContent strings.Builder
	for i, row := range rows {
		if i > 0 {
			lineDelta := row.Position - rows[i-1].Position
			emptyLines := int(math.Round(float64(lineDelta) / float64(avgLineDetla)))
			for j := 1; j < emptyLines; j++ {
				pageContent.WriteString("\n")
			}
		}
		for _, text := range row.Content {
			pageContent.WriteString(text.S)
		}
		pageContent.WriteString("\n")
	}

	fmt.Printf("Page %v, content: %v", pageNum, pageContent.String())

	llm, err := ai.GetLanguageModel()
	if err != nil {
		return fmt.Errorf("failed to create language model: %w", err)
	}

	ocrProjectContext := &ocr.ProjectContext{
		Title:    a.project.Title,
		Author:   a.project.Author,
		Genre:    a.project.Genre,
		Synopsis: a.project.Synopsis,
	}
	task := ocr.New(llm)
	req := &ocr.Request{
		Text:           pageContent.String(),
		ProjectContext: ocrProjectContext,
	}

	resp, err := task.Clean(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to clean page %d: %w", pageNum, err)
	}

	text := resp.Body
	if resp.Footnotes != "" {
		text += "\n\n=============\n" + resp.Footnotes
	}

	paragraph := translation.Paragraph{
		Text: text,
	}

	paragraph.ID = paragraph.CalcID()
	a.project.Source.Paragraphs = append(a.project.Source.Paragraphs, paragraph)

	log.Printf("Page %d cleaned, result: %v", pageNum, resp)

	return nil
}

func (a *PDFImportAction) createProject(ctx context.Context, cmd *cli.Command) error {
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
		return nil
	}
	log.Printf("Creating new project: %s", title)

	a.project.Title = title
	a.project.Author = cmd.String("author")
	a.project.Synopsis = cmd.String("synopsis")

	// set source and target languages
	from := cmd.String("from")
	to := cmd.String("to")

	a.project.Source.Language = from
	a.project.Target.Language = to
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

	a.project.Normalize()
	fileName, err := a.project.Save()
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}
	log.Printf("Project created and saved to %s", fileName)

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
		log.Printf("Failed to extract text from PDF: %v", err)
		return true, nil // If text extraction fails, assume OCR is needed.
	}
	if len(texts) == 0 {
		log.Println("No text extracted from PDF, OCR is likely needed")
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
		log.Println("PDF has no pages, OCR is likely needed")
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
