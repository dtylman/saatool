# SAATool - Software Automated Translation Tool

SAATool is an Android application designed for automated translation of EPUB books using AI technology. It provides a unique reading and translation experience by translating books paragraph by paragraph as you read, with specialized features for translation workflow management.

## Features

- **Real-time Translation**: Translates EPUB books using DeepSeek AI API as you read
- **Paragraph-by-Paragraph Translation**: Smart translation that maintains context and consistency
- **Translation Management**: Fix translations, proofread automatically, and manage translation quality
- **Offline Reading**: Continue reading translated content without internet connection
- **Export Functionality**: Export translated books back to EPUB format
- **Bidirectional Text Support**: Supports both left-to-right and right-to-left languages
- **Character Context**: AI understands character names and maintains consistency throughout translation
- **Custom UI**: Specialized interface designed for translation workflows

## How It Works

### Workflow Overview

1. **Convert EPUB to SPZ**: Use `saatooltool` to convert an EPUB file into a SAATool Project (`.spz`) file
2. **Import Project**: Load the `.spz` file into the Android app
3. **Configure Translation**: Set up source and target languages, add DeepSeek API key
4. **Read & Translate**: The app automatically translates paragraphs as you read them
5. **Export**: Export the completed translation back to a `.spz` file
6. **Convert to EPUB**: Use `saatooltool` to convert the translated `.spz` back to EPUB format

### Translation Features

- **Smart Context**: AI maintains character names, terminology, and style consistency
- **Fix Translation**: Manually request re-translation of specific paragraphs
- **Auto Proofread**: Automatically proofread translations for better quality
- **Translate Ahead**: Pre-translate upcoming paragraphs for smoother reading experience
- **Progress Tracking**: Visual progress indicators and reading position management

## User Guide

### Prerequisites

- Android device (ARM64 or AMD64 architecture)
- DeepSeek AI API key ([Get one here](https://platform.deepseek.com))
- EPUB books to translate

### Installation

1. Download the latest APK from the releases page
2. Install on your Android device
3. Configure the DeepSeek API key in Settings

### Using SAATool

#### Step 1: Prepare Your Book

Use the `saatooltool` command-line utility to convert your EPUB:

```bash
./saatooltool -in "your_book.epub" -from "english" -to "hebrew" -details=true
```

This creates a `.spz` file ready for import.

#### Step 2: Import Project

1. Copy the `.spz` file to your Android device
2. Open SAATool
3. Tap "Import" and select your `.spz` file
4. The project will appear in your projects list

#### Step 3: Configure Settings

1. Go to Settings (gear icon)
2. Enter your DeepSeek API key
3. Adjust translation preferences:
   - **Translate Ahead**: Number of paragraphs to pre-translate (default: 6)
   - **Auto Proofread**: Automatically improve translations (recommended: ON)
   - **App Size Factor**: UI scaling factor
   - **Translation Doc Size**: Context size for AI (default: 3)

#### Step 4: Start Reading and Translating

1. Select your project and tap "Translate"
2. The app will display the book content
3. Tap to navigate:
   - **Left side**: Previous paragraph/word
   - **Right side**: Next paragraph/word
4. Use the language toggle to switch between source and translated text
5. Use "Fix" button to re-translate problematic paragraphs

#### Step 5: Export and Convert

1. When finished, export the project using the export button
2. Transfer the `.spz` file back to your computer
3. Use `saatooltool` to convert back to EPUB:

```bash
./saatooltool -in "translated_project.spz" -out "translated_book.epub"
```

### Screenshots

*[Screenshots will be added here]*

- Main projects view
- Translation interface
- Settings screen
- Import/Export workflow

## Development

### Prerequisites

- Go 1.24 or later
- Fyne framework
- Android SDK (for building APK)
- Git

### Required Dependencies

#### System Dependencies (Ubuntu/Debian)
```bash
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

#### Go Dependencies
```bash
go get fyne.io/fyne/v2@latest
go install fyne.io/tools/cmd/fyne@latest
```

#### Android Development
```bash
# Install Android SDK
# Set environment variables:
export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$HOME/Android/Sdk/ndk/29.0.13846066
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
```

### Building the Project

#### Desktop Application (for testing)
```bash
cd cmd/saatool
go build
./saatool
```

#### Command Line Tool
```bash
cd cmd/saatooltool
go build
./saatooltool -help
```

#### Android APK
```bash
# Using the build script
./package.sh

# Or manually with fyne
cd cmd/saatool
fyne package --target android/arm64 --app-id org.saatool.app --icon icon.png --name "SAATool"
```

#### CI/CD Build
The project includes GitHub Actions workflow that automatically builds APKs for both ARM64 and AMD64 architectures.

### Project Structure

```
saatool/
├── ai/                     # AI translation logic
│   ├── translator.go       # Main translation engine
│   ├── bookdetails.go      # Book metadata handling
│   ├── prompts.go          # AI prompts management
│   └── stats.go            # Translation statistics
├── cmd/
│   ├── saatool/            # Main Android app
│   │   ├── main.go         # App entry point
│   │   ├── icon.png        # App icon
│   │   └── FyneApp.toml    # App metadata
│   └── saatooltool/        # Command line converter
│       ├── main.go         # CLI entry point
│       ├── epubconverter.go # EPUB processing
│       └── paragraphsplitter.go # Text processing
├── config/                 # Configuration management
│   ├── options.go          # App settings
│   ├── projects.go         # Project file handling
│   └── fs.go               # File system utilities
├── translation/            # Core translation logic
│   ├── project.go          # Project data structures
│   ├── direction.go        # Text direction handling
│   └── fyne.go             # UI integration
└── ui/                     # User interface
    ├── mainwindow.go       # Main application window
    ├── translationview.go  # Translation reading interface
    ├── projectsview.go     # Project management
    ├── settingsview.go     # Settings configuration
    └── widgets/            # Custom UI components
```

### Key Technologies

- **[Fyne](https://fyne.io/)**: Cross-platform UI framework for Go
- **[DeepSeek API](https://platform.deepseek.com)**: AI translation service
- **[GoReader](https://github.com/taylorskalyo/goreader)**: EPUB file processing
- **[html2text](https://github.com/jaytaylor/html2text)**: HTML to text conversion

### Development Guidelines

#### Code Style
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Maintain thread safety for concurrent operations

#### Translation Logic
- All translation operations should be asynchronous
- Use proper error handling and user feedback
- Maintain translation state and progress tracking
- Implement proper cleanup for incomplete translations

#### UI Guidelines
- Use Fyne's responsive design principles
- Implement proper navigation and user feedback
- Support both touch and keyboard interactions
- Maintain consistent theming and icons

#### Testing
```bash
# Run tests
go test ./...

# Test specific package
go test ./translation
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

### Build Configuration

The app metadata is managed in `cmd/saatool/FyneApp.toml`:
- Version and build numbers are automatically extracted during CI builds
- App ID and name are configured for proper Android packaging
- The build process creates APKs for both ARM64 and AMD64 architectures

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and bug reports, please open an issue on the GitHub repository.

---

**Note**: This tool requires a DeepSeek API key for translation functionality. The quality of translations depends on the AI model and the complexity of the source text.

