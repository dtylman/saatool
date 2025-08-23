package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
)

type ToolTool struct {
	ec         *EPubConverter
	inFile     string
	outFile    string
	maxWords   int
	fromLang   string
	toLang     string
	getDetails bool
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

	tt.ec.Project.Source.Language = tt.fromLang
	tt.ec.Project.Target.Language = tt.toLang

	fileName, err := tt.ec.Project.Save()
	if err != nil {
		return err
	}
	log.Printf("project saved to: '%s'", fileName)
	return nil

}

func (tt *ToolTool) getBookDetails(ctx context.Context, project *translation.Project) error {
	log.Printf("Getting book details for project: %s", project.Title)
	translator, err := ai.NewTranslator(project)
	if err != nil {
		return err
	}

	bookDetails, err := translator.GetBookDetails(ctx)
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
	flag.IntVar(&tt.maxWords, "maxwords", 200, "Maximum words per paragraph")
	flag.StringVar(&tt.fromLang, "from", "english", "Source language")
	flag.StringVar(&tt.toLang, "to", "hebrew", "Target language")
	flag.StringVar(&config.Options.DeepSeekAPIKey, "key", "", "DeepSeek API key")
	flag.BoolVar(&tt.getDetails, "details", false, "Get book details")
	flag.Parse()

	os.Setenv("FILESDIR", ".")

	if tt.inFile == "" {
		log.Fatal("Input file is required")
	}

	if tt.toLang == "" {
		log.Fatal("Target language is required")
	}

	err := tt.Run(context.Background())
	if err != nil {
		log.Fatalf("Error running ToolTool: %v", err)
	}
	log.Println("ToolTool completed successfully")
}
