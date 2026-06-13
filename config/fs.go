package config

import (
	"log"
	"os"
	"path"

	"fyne.io/fyne/v2"
)

// AppDir returns the application's storage directory.
func AppDir() string {
	if fyne.CurrentApp() != nil {
		dir := fyne.CurrentApp().Storage().RootURI().Path()
		if dir != "" {
			log.Printf("using fyne app storage dir: %v", dir)
			return dir
		}
	}
	dir := os.Getenv("FILESDIR")
	var err error
	if dir != "" {
		log.Printf("using FILESDIR environment variable for app dir: %v", dir)
	}
	if dir == "" {
		dir, err = os.UserConfigDir()
		if err == nil && dir != "" {
			log.Printf("using user config dir for app dir: %v", dir)
		}
	}
	if dir == "" {
		dir, err = os.UserHomeDir()
		if err == nil && dir != "" {
			log.Printf("using user home dir for app dir: %v", dir)
		}
	}

	if dir != "" {
		appDir := path.Join(dir, "saatool")
		err = os.MkdirAll(appDir, 0755)
		if err != nil {
			log.Fatalf("failed to create app dir %s: %v", appDir, err)
		}
		return appDir
	}

	log.Fatal("failed to determine app directory, set FILESDIR environment variable to specify the directory")
	return ""
}

// ConfigDir returns the configuration directory path, creating it if necessary.
func ConfigDir() string {
	appDir := AppDir()
	configDir := path.Join(appDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		log.Fatalf("failed to create config dir %s: %v", configDir, err)
	}
	return configDir
}

// ProjectsDir returns the projects directory path, creating it if necessary.
func ProjectsDir() string {
	appDir := AppDir()
	projectsDir := path.Join(appDir, "projects")
	err := os.MkdirAll(projectsDir, 0755)
	if err != nil {
		log.Fatalf("failed to create projects dir %s: %v", projectsDir, err)
	}
	return projectsDir
}
