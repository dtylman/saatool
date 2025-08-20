package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/translation"
)

type ToolTool struct {
	ec          *EPubConverter
	inFile      string
	outFile     string
	maxLines    int
	fromLang    string
	toLang      string
	deepSeekKey string
	getDetails  bool
}

func (tt *ToolTool) Run(ctx context.Context) error {
	log.Printf("converting epub: '%s'", tt.inFile)
	tt.ec = NewEPubConverter()
	err := tt.ec.ConvertEPub(tt.inFile)
	if err != nil {
		return err
	}

	if tt.getDetails {
		err = tt.getBookDetails(ctx, tt.ec.Project)
		if err != nil {
			return err
		}
	}

	return tt.saveProject()

}

func (tt *ToolTool) saveProject() error {
	log.Printf("saving project to: %s", tt.outFile)
	data, err := json.MarshalIndent(tt.ec.Project, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(tt.outFile, data, 0644)
	if err != nil {
		return err
	}
	log.Printf("Project saved to %s", tt.outFile)
	return nil
}

func (tt *ToolTool) getBookDetails(ctx context.Context, project *translation.Project) error {
	log.Printf("Getting book details for project: %s", project.Title)
	translator := ai.NewTranslator(tt.deepSeekKey)

	bookDetails, err := translator.GetBookDetails(ctx, ai.NewBookDetails(tt.ec.Project))
	if err != nil {
		return err
	}

	project.Author = bookDetails.Author
	project.Genre = bookDetails.Genre
	project.Synopsis = bookDetails.Synopsis
	project.Characters = bookDetails.MainCharacters

	log.Printf("Book details retrieved: %s by %s", bookDetails.Title, bookDetails.Author)

	return nil
}

func main() {
	tt := &ToolTool{}
	flag.StringVar(&tt.inFile, "in", "", "Input EPUB file")
	flag.StringVar(&tt.outFile, "out", "project.json", "Output project file")
	flag.IntVar(&tt.maxLines, "maxlines", 5, "Maximum lines per paragraph")
	flag.StringVar(&tt.fromLang, "from", "english", "Source language")
	flag.StringVar(&tt.toLang, "to", "hebrew", "Target language")
	flag.StringVar(&tt.deepSeekKey, "key", "", "DeepSeek API key")
	flag.BoolVar(&tt.getDetails, "details", false, "Get book details")
	flag.Parse()

	if tt.inFile == "" {
		log.Fatal("Input file is required")
	}

	err := tt.Run(context.Background())
	if err != nil {
		log.Fatalf("Error running ToolTool: %v", err)
	}
	log.Println("ToolTool completed successfully")
}
