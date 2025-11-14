package actions

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtylman/saatool/translation"
	"github.com/urfave/cli/v3"
)

// RTFExportAction handles exporting translation projects to RTF format (compatible with LibreOffice/OpenOffice)
type RTFExportAction struct {
}

func (re *RTFExportAction) Name() string {
	return "rtf"
}

func (re *RTFExportAction) Usage() string {
	return "Export a translation project to RTF format (compatible with LibreOffice/OpenOffice)"
}

func (re *RTFExportAction) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Aliases:  []string{"i"},
			Usage:    "Input project file (.spz)",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Output RTF file path (optional, defaults to input filename with .rtf extension)",
		},
	}
}

func (re *RTFExportAction) Action(ctx context.Context, cmd *cli.Command) error {
	inputPath := cmd.String("input")
	outputPath := cmd.String("output")

	// Load the translation project
	project, err := translation.LoadProject(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine output path if not provided
	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		outputPath = strings.TrimSuffix(inputPath, ext) + ".rtf"
	}

	log.Printf("Exporting project '%s' to RTF format: %s", project.Name, outputPath)

	// Create the RTF document
	err = re.createRTFDocument(project, outputPath)
	if err != nil {
		return fmt.Errorf("failed to create RTF document: %w", err)
	}

	log.Printf("Successfully exported to %s", outputPath)
	return nil
}

func (re *RTFExportAction) createRTFDocument(project *translation.Project, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Determine text direction based on target language
	targetLang := project.GetTargetLanguage()
	isRTL := translation.GetTextDirection(targetLang) == translation.RightToLeft

	// Start RTF document
	var rtfContent strings.Builder

	// RTF header with font table and document info
	rtfContent.WriteString("{\\rtf1\\ansi\\deff0")

	// Font table - using Unicode-compatible fonts
	rtfContent.WriteString("{\\fonttbl")
	rtfContent.WriteString("{\\f0\\froman Times New Roman;}")
	rtfContent.WriteString("{\\f1\\fswiss Arial;}")
	if isRTL {
		rtfContent.WriteString("{\\f2\\froman David;}")  // Hebrew font
		rtfContent.WriteString("{\\f3\\froman Miriam;}") // Alternative Hebrew font
	}
	rtfContent.WriteString("}")

	// Document info
	rtfContent.WriteString("{\\info")
	rtfContent.WriteString(fmt.Sprintf("{\\title %s}", re.escapeRTF(project.Title)))
	if project.Author != "" {
		rtfContent.WriteString(fmt.Sprintf("{\\author %s}", re.escapeRTF(project.Author)))
	}
	rtfContent.WriteString("{\\creatim\\yr2024\\mo11\\dy14}")
	rtfContent.WriteString("{\\subject Translation created with SAATool}")
	rtfContent.WriteString("}")

	// Set default paragraph direction if RTL
	if isRTL {
		rtfContent.WriteString("\\rtldoc")
	}

	// Title page
	re.addTitlePageRTF(&rtfContent, project, isRTL)

	// Page break
	rtfContent.WriteString("\\page")

	// Book information page
	re.addBookInfoPageRTF(&rtfContent, project, isRTL)

	// Page break
	rtfContent.WriteString("\\page")

	// Main content
	re.addTranslatedContentRTF(&rtfContent, project, isRTL)

	// Close RTF document
	rtfContent.WriteString("}")

	// Write to file
	_, err = file.WriteString(rtfContent.String())
	return err
}

func (re *RTFExportAction) addTitlePageRTF(rtf *strings.Builder, project *translation.Project, isRTL bool) {
	// Title
	rtf.WriteString("\\pard\\qc") // Center alignment
	if isRTL {
		rtf.WriteString("\\rtlpar")
	}
	rtf.WriteString("\\fs48\\b ") // 24pt bold
	rtf.WriteString(re.escapeRTF(project.Title))
	rtf.WriteString("\\b0\\fs24\\par\\par") // Reset formatting and add spacing

	// Author
	if project.Author != "" {
		rtf.WriteString("\\fs32\\b ") // 16pt bold
		rtf.WriteString("By: ")
		rtf.WriteString(re.escapeRTF(project.Author))
		rtf.WriteString("\\b0\\fs24\\par\\par")
	}

	// Translation info
	rtf.WriteString("\\i ") // Italic
	rtf.WriteString("Translated to ")
	rtf.WriteString(re.escapeRTF(strings.Title(project.GetTargetLanguage())))
	rtf.WriteString("\\i0\\par\\par")

	// SAATool credit
	rtf.WriteString("\\fs20\\i ") // 10pt italic
	rtf.WriteString("Created with SAATool")
	rtf.WriteString("\\i0\\fs24\\par")
}

func (re *RTFExportAction) addBookInfoPageRTF(rtf *strings.Builder, project *translation.Project, isRTL bool) {
	// Book Information header
	rtf.WriteString("\\pard")
	if isRTL {
		rtf.WriteString("\\qr\\rtlpar") // Right align for RTL
	} else {
		rtf.WriteString("\\ql") // Left align for LTR
	}
	rtf.WriteString("\\fs36\\b Book Information\\b0\\fs24\\par\\par")

	// Synopsis
	if project.Synopsis != "" {
		rtf.WriteString("\\b Synopsis:\\b0\\par")
		rtf.WriteString(re.escapeRTF(project.Synopsis))
		rtf.WriteString("\\par\\par")
	}

	// Genre
	if project.Genre != "" {
		rtf.WriteString("\\b Genre: \\b0")
		rtf.WriteString(re.escapeRTF(project.Genre))
		rtf.WriteString("\\par\\par")
	}

	// Characters
	if len(project.Characters) > 0 {
		rtf.WriteString("\\b Characters:\\b0\\par")
		for _, char := range project.Characters {
			rtf.WriteString("\\bullet ")
			rtf.WriteString(re.escapeRTF(char.Name))
			if char.Gender != "" {
				rtf.WriteString(" (")
				rtf.WriteString(re.escapeRTF(char.Gender))
				rtf.WriteString(")")
			}
			if char.Role != "" {
				rtf.WriteString(" - ")
				rtf.WriteString(re.escapeRTF(char.Role))
			}
			if char.Description != "" {
				rtf.WriteString(": ")
				rtf.WriteString(re.escapeRTF(char.Description))
			}
			rtf.WriteString("\\par")
		}
	}
}

func (re *RTFExportAction) addTranslatedContentRTF(rtf *strings.Builder, project *translation.Project, isRTL bool) {
	// Content header
	rtf.WriteString("\\pard")
	if isRTL {
		rtf.WriteString("\\qr\\rtlpar")
	} else {
		rtf.WriteString("\\ql")
	}
	rtf.WriteString("\\fs36\\b Translated Text\\b0\\fs24\\par\\par")

	// Add each translated paragraph
	for i, targetPara := range project.Target.Paragraphs {
		// Skip empty paragraphs
		if strings.TrimSpace(targetPara.Text) == "" {
			continue
		}

		// Set paragraph properties
		rtf.WriteString("\\pard")
		if isRTL {
			rtf.WriteString("\\qr\\rtlpar")
		} else {
			rtf.WriteString("\\ql")
		}
		rtf.WriteString("\\fs24 ") // 12pt font

		// Add paragraph text
		rtf.WriteString(re.escapeRTF(targetPara.Text))

		// Add spacing between paragraphs
		if i < len(project.Target.Paragraphs)-1 {
			rtf.WriteString("\\par\\par")
		} else {
			rtf.WriteString("\\par")
		}
	}

	// Add footer information
	rtf.WriteString("\\par\\par")
	rtf.WriteString("\\pard\\qc\\fs20\\i ") // Center, 10pt, italic
	rtf.WriteString("Translated from ")
	rtf.WriteString(re.escapeRTF(strings.Title(project.GetSourceLanguage())))
	rtf.WriteString(" to ")
	rtf.WriteString(re.escapeRTF(strings.Title(project.GetTargetLanguage())))
	rtf.WriteString(" using SAATool")
	rtf.WriteString("\\i0\\fs24")
}

// escapeRTF escapes special RTF characters
func (re *RTFExportAction) escapeRTF(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	text = strings.ReplaceAll(text, "\n", "\\par ")
	text = strings.ReplaceAll(text, "\r", "")

	// Handle Unicode characters by converting to RTF Unicode escapes
	var result strings.Builder
	for _, r := range text {
		if r > 127 {
			result.WriteString(fmt.Sprintf("\\u%d?", r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
