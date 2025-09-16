package translation

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2"
)

// ExportProject exports the given project to a .saatool file using the provided writer.
func ExportProject(project *Project, writer fyne.URIWriteCloser) error {
	log.Printf("exporting project to %s", writer.URI().String())
	n, err := project.SaveToWriter(writer)
	if err != nil {
		return fmt.Errorf("failed to save project: %v", err)
	}
	log.Printf("wrote %d bytes to %s", n, writer.URI().String())
	return nil
}

// ImportProject imports a project from a .saatool file and returns the project path
func ImportProject(reader fyne.URIReadCloser) (string, error) {
	log.Printf("importing project from %s", reader.URI().String())

	zipReader, err := gzip.NewReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer zipReader.Close()

	data, err := io.ReadAll(zipReader)
	if err != nil {
		return "", fmt.Errorf("failed to read project file: %v", err)
	}

	project := NewProject(reader.URI().Name())
	err = json.Unmarshal(data, &project)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal project file: %v", err)
	}

	project.SetName(reader.URI().Name())
	project.Normalize()

	return project.Save()
}
