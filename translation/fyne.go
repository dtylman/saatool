package translation

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2"
)

func ImportProject(reader fyne.URIReadCloser) (string, error) {
	log.Printf("importing project from %s", reader.URI().String())
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read project file: %v", err)
	}

	project := NewProject(reader.URI().Name())
	err = json.Unmarshal(data, &project)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal project file: %v", err)
	}

	project.Name = reader.URI().Name()
	project.Normalize()

	return project.Save()
}
