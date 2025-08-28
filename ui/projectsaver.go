package ui

import (
	"context"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectSaver saves the project periodically
type ProjectSaver struct {
	project *translation.Project
	dirty   bool
	label   *widget.Label
	View    fyne.CanvasObject
	cancel  context.CancelFunc
}

// NewProjectSaver creates a new ProjectSaver for the given project.
func NewProjectSaver(project *translation.Project) *ProjectSaver {
	ps := &ProjectSaver{
		project: project,
		dirty:   false,
		label:   widget.NewLabel("Test"),
	}
	panel := widgets.NewPanel(ps.label, fyne.NewSize(150, 20))
	panel.Border = 3
	ps.View = panel

	return ps
}

func (ps *ProjectSaver) SetDirty(dirty bool) {
	ps.dirty = dirty
	if dirty {
		ps.label.SetText("Unsaved changes")
	} else {
		ps.label.SetText("All changes saved")
	}
}

func (ps *ProjectSaver) Start() {
	log.Println("Starting project saver")
	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				log.Println("Project saver tick")
				timeStr := time.Now().Format("15:04:05")
				fyne.Do(func() {
					ps.label.SetText(timeStr)
				})

			}
		}
	}()
}

func (ps *ProjectSaver) Stop() {
	log.Println("Stopping project saver")
	ps.cancel()
}
