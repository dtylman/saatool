package config

import (
	"fmt"
	"log"
	"os"
	"path"
)

// ProjectFileExt is the file extension for project files
const ProjectFileExt = ".json"

// ProjectFile represents a translation project file
type ProjectFile struct {
	//Path is the file path of the project
	Path string
	//Name is the name of the project
	Name string
}

// ListProjects lists all project files in the projects directory
func ListProjects() ([]ProjectFile, error) {
	log.Println("listing projects")
	var projects []ProjectFile

	projectsDir := ProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects dir %s: %v", projectsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if path.Ext(entry.Name()) != ProjectFileExt {
			continue
		}
		project := ProjectFile{
			Path: path.Join(projectsDir, entry.Name()),
			Name: entry.Name(),
		}
		projects = append(projects, project)
	}
	return projects, nil
}
