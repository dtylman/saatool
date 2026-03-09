package ai

import "github.com/dtylman/saatool/translation"

// BookDetails represents the details of a book.
type BookDetails struct {
	Title          string                  `json:"title"`
	Author         string                  `json:"author"`
	Synopsis       string                  `json:"synopsis"`
	Genre          string                  `json:"genre"`
	// WritingStyle describes the author's narrative voice, tone, and pacing.
	// E.g. "third-person omniscient, dark and introspective, lyrical prose".
	WritingStyle   string                  `json:"writing_style"`
	MainCharacters []translation.Character `json:"main_characters"`
}

// NewBookDetails creates a new BookDetails instance from a translation.Project.
func NewBookDetails(project *translation.Project) *BookDetails {
	project.Lock()
	defer project.Unlock()

	return &BookDetails{
		Title:          project.Title,
		Author:         project.Author,
		Synopsis:       project.Synopsis,
		Genre:          project.Genre,
		WritingStyle:   project.WritingStyle,
		MainCharacters: project.Characters,
	}
}
