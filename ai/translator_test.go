package translator

import (
	"fmt"
	"testing"

	"github.com/dtylman/saatool/translation"
)

func TestBookDetails(t *testing.T) {
	project := &translation.Project{Title: "The Mote in God's Eye"}
	resp, err := GetBookDetails(t.Context(), project)
	if err != nil {
		t.Errorf("BookDetails failed: %v", err)
	} else {
		t.Log("BookDetails executed successfully")
	}
	fmt.Printf("Book Details: %+v\n", resp)
}

func TestTranslate2(t *testing.T) {
	project := &translation.Project{
		Title:    "The Mote in God's Eye",
		Author:   "Larry Niven",
		Synopsis: "A science fiction novel about first contact with an alien species.",
		Genre:    "Science Fiction",
		Characters: []translation.Character{
			{Name: "Bob", Gender: "Male"},
			{Name: "Alice", Gender: "Female"},
		},
		Source: translation.Unit{
			Language: "English",
			Paragraphs: []translation.Paragraph{
				translation.Paragraph{
					Text: "The Mote in God's Eye is a classic science fiction novel. It explores the complexities of first contact with an alien species. The story is set in a future where humanity has expanded into the galaxy, and it raises questions about communication, trust, and the unknown. The characters navigate through these challenges, making difficult decisions that impact their understanding of the universe and themselves.",
					ID:   "1",
				},
			},
		},
		Target: translation.Unit{
			Language:   "Hebrew",
			Paragraphs: []translation.Paragraph{},
		},
	}

	text, err := Translate2(t.Context(), project, 0)
	if err != nil {
		t.Errorf("Translate2 failed: %v", err)
	} else {
		t.Log("Translate2 executed successfully")
	}
	fmt.Printf("Translated Text: %s\n", text)
}
