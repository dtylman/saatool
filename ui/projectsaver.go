package ui

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectSaver saves the project periodically
type ProjectSaver struct {
	translator *ai.Translator
	project    *translation.Project
	dirty      bool
	label      *widget.Label
	View       fyne.CanvasObject
	cancel     context.CancelFunc
}

// NewProjectSaver creates a new ProjectSaver for the given project.
func NewProjectSaver(translator *ai.Translator, project *translation.Project) *ProjectSaver {
	ps := &ProjectSaver{
		translator: translator,
		project:    project,
		dirty:      false,
		label:      widget.NewLabel("ETA"),
	}
	panel := widgets.NewPanel(ps.label, fyne.NewSize(150, 20))
	panel.Border = 3
	ps.View = panel

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

// onStatInterval updates the statistics label
func (ps *ProjectSaver) onStatInterval() {
	eta, total := ps.translator.Stats()
	text := fmt.Sprintf("%d (ETA: %s)", total, eta.Truncate(time.Second))
	fyne.Do(func() {
		ps.label.SetText(text)
	})
}

func (ps *ProjectSaver) Stop() {
	log.Println("Stopping project saver")
	ps.cancel()
	ps.onSaveInterval()
}
