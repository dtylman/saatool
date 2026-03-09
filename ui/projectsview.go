package ui

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/actions"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectsView lists the translation projects.
type ProjectsView struct {
	selectedProject *translation.Project
	lstProjects     *widget.List
	wrapper         *fyne.Container // stable outer container — tab content pointer never changes
	projects        []config.ProjectFile
}

// NewProjectsView creates a new ProjectsView.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{
		wrapper: container.NewStack(),
	}
	pv.lstProjects = widget.NewList(pv.lstProjectsLen, pv.lstProjectsCreateItem, pv.lstProjectsUpdateItem)
	pv.lstProjects.OnSelected = pv.onProjectSelected
	return pv
}

// View returns the stable wrapper used as the Library tab content.
func (pl *ProjectsView) View() fyne.CanvasObject {
	return pl.wrapper
}

// Load is called when the Library tab becomes active.
// It sets up the action toolbar and refreshes the project list.
func (pl *ProjectsView) Load() {
	Main.ClearActions()
	Main.AddAction("Translate", widgets.IconTranslate, pl.onTranslateTapped)
	Main.AddAction("Import", widgets.IconOpen, pl.onImportTapped)
	Main.AddAction("Export", widgets.IconSave, pl.onExportTapped)
	Main.AddAction("Delete", widgets.IconDelete, pl.onDeleteTapped)
	pl.listProjects()
}

func (pl *ProjectsView) Close() {}

// buildHeader returns the "My Library" heading with book count.
func (pl *ProjectsView) buildHeader() fyne.CanvasObject {
	title := canvas.NewText("My Library", theme.Color(theme.ColorNameForeground))
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = theme.Size(theme.SizeNameText) + 8

	count := fmt.Sprintf("%d book", len(pl.projects))
	if len(pl.projects) != 1 {
		count += "s"
	}
	subtitle := canvas.NewText(count, theme.Color(theme.ColorNamePlaceHolder))
	subtitle.TextSize = theme.Size(theme.SizeNameText)

	sep := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	sep.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(
		container.NewPadded(container.NewVBox(title, subtitle)),
		sep,
	)
}

// buildEmptyState returns a centered call-to-action when there are no projects.
func (pl *ProjectsView) buildEmptyState() fyne.CanvasObject {
	icon := widget.NewIcon(widgets.IconOpen)

	msg := canvas.NewText("No books yet", theme.Color(theme.ColorNameForeground))
	msg.TextStyle = fyne.TextStyle{Bold: true}
	msg.TextSize = theme.Size(theme.SizeNameText) + 4
	msg.Alignment = fyne.TextAlignCenter

	hint := canvas.NewText("Tap Import above to add your first EPUB", theme.Color(theme.ColorNamePlaceHolder))
	hint.TextSize = theme.Size(theme.SizeNameText)
	hint.Alignment = fyne.TextAlignCenter

	return container.NewCenter(
		container.NewVBox(
			container.NewCenter(icon),
			container.NewCenter(msg),
			container.NewCenter(hint),
		),
	)
}

// setView rebuilds the inner content of the stable wrapper.
func (pl *ProjectsView) setView() {
	header := pl.buildHeader()
	var body fyne.CanvasObject
	if pl.lstProjects.Length() == 0 {
		body = pl.buildEmptyState()
	} else {
		body = pl.lstProjects
	}
	pl.wrapper.RemoveAll()
	pl.wrapper.Add(container.NewBorder(header, nil, nil, nil, body))
	pl.wrapper.Refresh()
}

func (pl *ProjectsView) listProjects() {
	var err error
	pl.projects, err = config.ListProjects()
	if err != nil {
		log.Printf("failed to list projects: %v", err)
		return
	}
	pl.lstProjects.Refresh()
	pl.setView()
}

func (pl *ProjectsView) onProjectSelected(id widget.ListItemID) {
	if id < 0 || id >= len(pl.projects) {
		return
	}
	project, err := translation.LoadProject(pl.projects[id].Path)
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to load project file '%s': %v", pl.projects[id].Path, err))
		return
	}
	pl.setProject(project)
}

func (pl *ProjectsView) lstProjectsLen() int {
	return len(pl.projects)
}

func (pl *ProjectsView) lstProjectsCreateItem() fyne.CanvasObject {
	return widgets.NewListItem(widget.NewIcon(widgets.IconProject), "Project", "", nil)
}

func (pl *ProjectsView) lstProjectsUpdateItem(id widget.ListItemID, obj fyne.CanvasObject) {
	item := obj.(*widgets.ListItem)
	if id < 0 || id >= len(pl.projects) {
		log.Printf("invalid project id: %d", id)
		return
	}

	pf := pl.projects[id]
	displayName := strings.TrimSuffix(pf.Name, config.ProjectFileExt)
	subtitle := ""

	isSelected := pl.selectedProject != nil && pl.selectedProject.Name == pf.Name
	if isSelected {
		p := pl.selectedProject
		if p.Title != "" {
			displayName = p.Title
		}
		subtitle = fmt.Sprintf("%s → %s", titleCase(p.Source.Language), titleCase(p.Target.Language))
		item.SetSelected(true)
	} else {
		item.SetSelected(false)
	}

	item.SetTitle(displayName)
	item.SetSubtitle(subtitle)
}

// titleCase capitalises the first letter of a word.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// onExportTapped handles the Export action for the project.
func (pl *ProjectsView) onExportTapped() {
	if pl.selectedProject == nil {
		Main.ShowError("No project selected to export.")
		return
	}
	Main.OpenProjectSaveDialog(pl.onProjectFileExported, pl.selectedProject)
}

func (pl *ProjectsView) onProjectFileExported(writer fyne.URIWriteCloser, err error) {
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to export project file: %v", err))
		return
	}
	if writer == nil {
		Main.ShowError("No file selected to export project.")
		return
	}
	err1 := translation.ExportProject(pl.selectedProject, writer)
	if err1 != nil {
		Main.ShowError(fmt.Sprintf("Failed to export project file: %v", err1))
		return
	}
	writer.Close()
	Main.ShowMessage(fmt.Sprintf("Project exported to %s", writer.URI().Name()))
}

// onImportTapped handles the Import action.
func (pl *ProjectsView) onImportTapped() {
	Main.OpenProjectLoadDialog(pl.onProjectFileOpened)
}

func (pl *ProjectsView) onProjectFileOpened(reader fyne.URIReadCloser, err error) {
	fyne.Do(func() {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		if strings.HasSuffix(strings.ToLower(reader.URI().Name()), ".epub") {
			pl.onEPUBFileOpened(reader)
			return
		}

		projectPath, err := translation.ImportProject(reader)
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to import project file: %v", err))
			return
		}
		log.Printf("imported project from %s", projectPath)
		pl.listProjects()
	})
}

func (pl *ProjectsView) onEPUBFileOpened(reader fyne.URIReadCloser) {
	epubPath := reader.URI().Path()
	// On Windows, Fyne URI paths start with a leading slash before the drive letter (e.g. /C:/...)
	if runtime.GOOS == "windows" && len(epubPath) > 2 && epubPath[0] == '/' {
		epubPath = epubPath[1:]
	}

	fromEntry := widget.NewEntry()
	fromEntry.SetText(config.Options.SourceLanguage)
	fromEntry.SetPlaceHolder("e.g. english")

	toEntry := widget.NewEntry()
	toEntry.SetText(config.Options.TargetLanguage)
	toEntry.SetPlaceHolder("e.g. hebrew")

	items := []*widget.FormItem{
		widget.NewFormItem("Source Language", fromEntry),
		widget.NewFormItem("Target Language", toEntry),
	}

	finalPath := epubPath
	dlg := dialog.NewForm("Import EPUB", "Import", "Cancel", items, func(confirmed bool) {
		if !confirmed {
			return
		}
		from := strings.TrimSpace(fromEntry.Text)
		to := strings.TrimSpace(toEntry.Text)
		if from == "" || to == "" {
			Main.ShowError("Source and target languages are required.")
			return
		}

		config.Options.SourceLanguage = from
		config.Options.TargetLanguage = to

		project, err := actions.ImportEPUBFile(finalPath, from, to)
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to import EPUB: %v", err))
			return
		}
		projectPath, err := project.Save()
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to save project: %v", err))
			return
		}
		log.Printf("imported EPUB from %s, saved to %s", finalPath, projectPath)
		tv, err := NewTranslationView(project)
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to open translation view: %v", err))
			return
		}
		Main.SetContent(tv)
	}, Main.window)
	dlg.Show()
}

// setProject stores the selected project and refreshes the list.
func (ed *ProjectsView) setProject(project *translation.Project) {
	ed.selectedProject = project
	ed.lstProjects.Refresh()
}

func (ed *ProjectsView) onTranslateTapped() {
	if ed.selectedProject == nil {
		Main.ShowError("No project selected to translate.")
		return
	}
	tv, err := NewTranslationView(ed.selectedProject)
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to create translation view: %v", err))
		return
	}
	Main.SetContent(tv)
}

func (ed *ProjectsView) onDeleteTapped() {
	if ed.selectedProject == nil {
		Main.ShowError("No project selected to delete.")
		return
	}

	msg := fmt.Sprintf("Delete \"%s\"?", ed.selectedProject.Title)
	confirm := dialog.NewConfirm("Delete Book", msg, func(confirmed bool) {
		if !confirmed {
			return
		}
		err := translation.DeleteProject(ed.selectedProject)
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to delete project: %v", err))
			return
		}
		ed.selectedProject = nil
		ed.listProjects()
	}, Main.window)
	confirm.Show()
}
