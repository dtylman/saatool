package config

import (
	"fmt"
	"log"
	"os"
	"path"

	"fyne.io/fyne/v2"
)

func ReadWriteCreate() error {
	appDir := fyne.CurrentApp().Storage().RootURI().Path()
	projectsDir := path.Join(appDir, "projects")
	someText := "This is a test file."
	testFile := path.Join(projectsDir, "testfile.txt")

	err := os.MkdirAll(path.Dir(testFile), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", testFile, err)
	}

	err = os.WriteFile(testFile, []byte(someText), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", testFile, err)
	}
	log.Printf("Successfully wrote to file %s", testFile)

	data, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read from file %s: %w", testFile, err)
	}

	log.Printf("Successfully read from file %s: %s", testFile, string(data))

	err = os.Remove(testFile)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", testFile, err)
	}
	log.Printf("Successfully deleted file %s", testFile)
	return nil

}
