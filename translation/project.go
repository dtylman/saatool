package translation

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2/storage"
)

// Character represents a character in a translation project.
type Character struct {
	// Name is the name of the character.
	Name string `json:"name"`
	// Gender is the gender of the character
	Gender string `json:"gender"`
	// Age is the age of the character.
	Age int `json:"age"`
	// Role is the role of the character in the story.
	Role string `json:"role"`
	// Description is a brief description of the character.
	Description string `json:"description"`
}

// Paragraph represents a single paragraph in a translation project.
type Paragraph struct {
	// ID is a unique identifier for the paragraph.
	ID string `json:"id"`
	// Text is the content of the paragraph.
	Text string `json:"text"`
}

// CalcID calculates a unique ID for the paragraph based on its text.
func (p *Paragraph) CalcID() string {
	sum := md5.Sum([]byte(p.Text))
	return hex.EncodeToString(sum[:])
}

// Unit represents a collection of paragraphs in a specific language.
type Unit struct {
	// Language is the language of the paragraphs.
	Language string `json:"language"`
	// Paragraphs is a list of paragraphs in this unit.
	Paragraphs []Paragraph `json:"paragraphs"`
}

// Project represents a translation project with source and target languages and paragraphs.
type Project struct {
	// Name is the name of the project.
	Name string `json:"name"`
	// Title is the title of the text being translated.
	Title string `json:"title"`
	// Author is the author of the text being translated.
	Author string `json:"author"`
	// Synopsis is a brief summary of the text.
	Synopsis string `json:"synopsis"`
	// Genre is the genre of the text.
	Genre string `json:"genre"`
	// Characters is a list of characters involved in the text.
	Characters []Character `json:"characters"`
	// Source is the source language unit containing paragraphs to be translated.
	Source Unit `json:"source"`
	// Target is the target language unit where the translated paragraphs will be stored.
	Target Unit `json:"target"`
	// Prompt is the translation prompt or instructions for the translator.
	Prompt string `json:"prompt"`
}

// NewProject creates a new translation project with empty source and target units.
func NewProject() *Project {
	return &Project{
		Characters: make([]Character, 0),
	}
}

func (p *Project) Save(activeProject string) error {
	writer, err := storage.Writer(storage.NewFileURI(activeProject))
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}
	defer writer.Close()
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal project data: %w", err)
	}
	//todo: make sure we write all data
	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write project data: %w", err)
	}
	log.Printf("project saved to %s", activeProject)
	return nil
}

func LoadProjectFromReader(reader io.ReadCloser) (*Project, error) {
	log.Printf("loading project from reader")
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading project data: %w", err)
	}

	var project Project
	err = json.Unmarshal(data, &project)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling project data: %w", err)
	}

	project.Normalize()

	return &project, nil
}

// Normalize ensures that the project has valid data.
func (p *Project) Normalize() {
	if len(p.Source.Paragraphs) > len(p.Target.Paragraphs) {
		diff := len(p.Source.Paragraphs) - len(p.Target.Paragraphs)
		log.Printf("normalizing project: adding %d paragraphs to target", diff)
		for i := 0; i < diff; i++ {
			p.Target.Paragraphs = append(p.Target.Paragraphs, Paragraph{})
		}
	}

	for i := range p.Source.Paragraphs {
		if p.Source.Paragraphs[i].ID == "" {
			p.Source.Paragraphs[i].ID = p.Source.Paragraphs[i].CalcID()
			p.Target.Paragraphs[i].ID = p.Source.Paragraphs[i].ID
		}
	}
}
