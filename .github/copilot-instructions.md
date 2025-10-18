# SAATool AI Coding Instructions

SAATool is a Go-based Android translation app built with Fyne that converts EPUB books into translation projects and provides AI-powered paragraph-by-paragraph translation using DeepSeek API.

## Architecture Overview

### Dual Application Structure
- **`cmd/saatool/`**: Main Android app (Fyne-based GUI) for reading and translating
- **`cmd/saatooltool/`**: CLI tool for EPUB ↔ SPZ conversion using action pattern

### Core Data Flow
1. CLI converts EPUB → SPZ (compressed JSON project file)  
2. Android app imports SPZ → translates paragraphs → exports SPZ
3. CLI converts translated SPZ → EPUB

### Key Components
- **`translation/project.go`**: Core data structures (Project, Unit, Paragraph, Character)
- **`ai/translator.go`**: DeepSeek API integration with async translation queue
- **`config/`**: Global options with JSON persistence (`Options` struct)
- **`ui/`**: Fyne widgets with custom theme and navigation patterns

## Critical Patterns

### Project File Format (.spz)
```go
// Projects are gzipped JSON with source/target language units
type Project struct {
    Source Unit `json:"source"`  // Original paragraphs
    Target Unit `json:"target"`  // Translated paragraphs  
    Characters []Character       // For translation context
    // ... metadata fields
}
```

### Translation State Management
- Use `translator.inTranslation` map to prevent duplicate API calls
- All translation operations are async with callbacks
- Paragraphs identified by MD5 hash of source text
- Thread-safe with mutex protection on Project operations

### UI Navigation Pattern
```go
// All main views implement WindowContent interface
type WindowContent interface {
    View() fyne.CanvasObject
    Close()
    Load()
}
// Switch views via Main.SetContent()
```

### Action Command Pattern
```go
// CLI commands implement Action interface
type Action interface {
    Name() string
    Usage() string  
    Flags() []cli.Flag
    Action(context.Context, *cli.Command) error
}
// Register with actions.AddAction(cmd, "import", &EPubImportAction{})
```

## Development Workflows

### Building
```bash
# Desktop testing
cd cmd/saatool && go build && ./saatool

# Android APKs (requires Android SDK)
./package.sh  # Creates both ARM64 and AMD64 APKs

# CLI tool
cd cmd/saatooltool && go build && ./saatooltool -help
```

### Testing Translation Workflow
```bash
# Convert EPUB to project
./saatooltool -in book.epub -from english -to hebrew -details=true

# Creates book.epub.spz ready for Android import
```

### CI/CD Integration
- GitHub Actions auto-builds APKs from `FyneApp.toml` metadata
- Build numbers auto-increment in CI
- Artifacts uploaded for ARM64/AMD64 architectures

## Project Conventions

### Configuration Management
- Global config in `config.Options` struct with JSON persistence
- DeepSeek API key, translation settings, UI scaling factors
- Thread-safe access patterns throughout

### Custom Fyne Patterns
- Embedded fonts (`SimpleCLM-Medium.ttf`) for consistent rendering
- Custom theme with configurable base sizing
- Icon resources embedded as SVG → Go resources
- Bidirectional text support for RTL languages

### Error Handling
- Extensive logging throughout (`log.Printf` patterns)
- UI error dialogs for user-facing errors
- Context cancellation for AI operations
- Graceful degradation when API unavailable

### File Extensions
- `.spz`: SAATool project files (gzipped JSON)
- `.epub`: Standard EPUB format
- UI file dialogs filtered by these extensions

## AI Integration Specifics

### DeepSeek API Usage
- JSON mode for structured responses (book details, character info)
- System prompts with template substitution (`ai/prompts.go`)
- Context management with configurable paragraph window
- Rate limiting and concurrent request management

### Translation Context
- Character gender/details maintained for consistent pronouns
- Book synopsis and genre inform translation style  
- Configurable context window (`TranslationDocSize`)
- Auto-proofreading with separate API calls

### Paragraph Splitting Logic
The `ParagraphSplitter` breaks text into translation units using sophisticated rules:
- **ASCII normalization**: Optional stripping of non-ASCII characters for clean processing
- **Smart boundary detection**: Splits on empty lines, indented paragraphs, sentence endings
- **Word count limits**: `MaxWords` (soft limit at sentence boundaries) + `MaxWordsTolerance` (hard limit)
- **Preserves formatting**: Maintains line breaks and indentation for context

Split triggers (in order of precedence):
1. Empty lines (natural paragraph breaks)
2. Indented lines (dialogue/quoted text)  
3. Sentence endings when over `MaxWords` limit
4. Hard split at `MaxWordsTolerance` with backtrack to `MaxWords` if possible

## Key Files for Understanding
- `translation/project.go`: Core data model
- `ai/translator.go`: Translation engine and API integration  
- `ui/mainwindow.go`: Main application orchestration
- `actions/action.go`: CLI command pattern
- `cmd/saatool/main.go`: Android app entry point
- `config/options.go`: Global configuration management

When working on this codebase, always consider the dual Android/CLI nature, maintain thread safety for concurrent operations, and follow the established Fyne UI patterns for consistent user experience.