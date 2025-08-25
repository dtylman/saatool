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

// ToolToolOptions holds the command line options for the ToolTool
var ToolToolOptions struct {
	//InFile is the input EPUB file
	InFile string
	//OutFile is the output file
	OutFile string
	//MaxWords is the maximum words per paragraph for translation
	MaxWords int
	//MaxWordsTolerance is the maximum words tolerance per paragraph for translation (used to split long paragraphs)
	MaxWordsTolerance int
	//FromLang is the source language
	FromLang string
	//ToLang is the target language
	ToLang string
	//GetDetails indicates whether to get book details
	GetDetails bool
	//StripToAscii indicates whether to strip text to ASCII only
	StripToAscii bool
}

// ToolTool is a command line tool for converting EPUB files and getting book details.
type ToolTool struct {
	ec *EPubConverter
}

func (tt *ToolTool) Run(ctx context.Context) error {
	log.Printf("converting epub: '%s'", ToolToolOptions.InFile)
	tt.ec = NewEPubConverter()
	err := tt.ec.ConvertEPub(ToolToolOptions.InFile)
	if err != nil {
		return err
	}

	if ToolToolOptions.GetDetails {
		err = tt.getBookDetails(ctx, tt.ec.Project)
		if err != nil {
			return err
		}
	}

	tt.ec.Project.Source.Language = ToolToolOptions.FromLang
	tt.ec.Project.Target.Language = ToolToolOptions.ToLang

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
	flag.StringVar(&ToolToolOptions.InFile, "in", "", "Input EPUB file")
	flag.IntVar(&ToolToolOptions.MaxWords, "maxwords", 200, "Maximum words per paragraph")
	flag.IntVar(&ToolToolOptions.MaxWordsTolerance, "maxtolerance", 300, "Maximum words tolerance per paragraph")
	flag.BoolVar(&ToolToolOptions.StripToAscii, "ascii", false, "Strip text to ASCII only")
	flag.StringVar(&ToolToolOptions.FromLang, "from", "english", "Source language")
	flag.StringVar(&ToolToolOptions.ToLang, "to", "", "Target language")
	flag.StringVar(&config.Options.DeepSeekAPIKey, "key", "", "DeepSeek API key")
	flag.BoolVar(&ToolToolOptions.GetDetails, "details", false, "Get book details")
	flag.Parse()

	os.Setenv("FILESDIR", ".")

	if ToolToolOptions.InFile == "" {
		log.Fatal("Input file is required")
	}

	if ToolToolOptions.ToLang == "" {
		log.Fatal("Target language is required")
	}

	err := tt.Run(context.Background())
	if err != nil {
		log.Fatalf("Error running ToolTool: %v", err)
	}
	log.Println("ToolTool completed successfully")
}
