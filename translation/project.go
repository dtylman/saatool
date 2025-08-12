package translation

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
