package ui

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/translation"
)

// ProjectSaver saves the project periodically
type ProjectSaver struct {
	translator  *ai.Translator
	project     *translation.Project
	dirty       bool
	cancel      context.CancelFunc
	progressBar *widget.ProgressBar
}

// NewProjectSaver creates a new ProjectSaver for the given project.
func NewProjectSaver(translator *ai.Translator, project *translation.Project) *ProjectSaver {
	ps := &ProjectSaver{
		translator: translator,
		project:    project,
		dirty:      false,
	}
	return ps
}

// SetDirty marks the project as dirty or clean
func (ps *ProjectSaver) SetDirty(dirty bool) {
	log.Printf("Project dirty state changed: %v", dirty)
	ps.dirty = dirty
}

// Start begins the periodic saving of the project
func (ps *ProjectSaver) Start() {
	log.Println("Starting project saver")
	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	go func() {
		lastSave := time.Time{}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(4 * time.Second):
				ps.onStatInterval()
				if time.Since(lastSave) > 30*time.Second {
					ps.onSaveInterval()
					lastSave = time.Now()
				}
			}
		}
	}()
}

// onSaveInterval saves the project if there are unsaved changes
func (ps *ProjectSaver) onSaveInterval() {
	log.Printf("Auto-save interval triggered, dirty=%v", ps.dirty)
	if ps.dirty {
		log.Println("Auto-saving project...")
		_, err := ps.project.Save()
		if err != nil {
			log.Printf("Failed to save project: %v", err)
		} else {
			ps.dirty = false
			log.Println("Project auto-saved.")
		}
	}
}

// SetProgressBar sets the progress bar to update with ETA.
func (ps *ProjectSaver) SetProgressBar(bar *widget.ProgressBar) {
	ps.progressBar = bar
}

// onStatInterval updates the status in the header
func (ps *ProjectSaver) onStatInterval() {
	eta, total := ps.translator.Stats()
	if total == 0 {
		Main.SetStatus("")
		if ps.progressBar != nil {
			ps.progressBar.SetValue(0)
		}
	} else {
		text := fmt.Sprintf("%d translating (%s)", total, eta.Round(time.Second).String())
		Main.SetStatus(text)
		if ps.progressBar != nil {
			const maxETA = 45.0
			secs := eta.Seconds()
			if secs > maxETA {
				secs = maxETA
			}
			ps.progressBar.SetValue(secs / maxETA)
		}
	}
}

func (ps *ProjectSaver) Stop() {
	log.Println("Stopping project saver")
	ps.cancel()
	ps.onSaveInterval()
	Main.SetStatus("")
}
