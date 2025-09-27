package translation

import (
	"regexp"
	"strings"
	"unicode"
)

// ParagraphSplitter splits text into paragraphs based on specified rules.
type ParagraphSplitter struct {
	// Accumulated paragraphs
	paragraphs []string
	// Maximum words per paragraph before considering a split
	MaxWords int
	// Maximum words per paragraph before forcing a split
	MaxWordsTolerance int
	// Flag to strip non-ASCII characters
	StripToAscii bool
}

// NewParagraphSplitter creates a new ParagraphSplitter with the specified maximum words per paragraph.
func NewParagraphSplitter(maxWords, maxWordsTolerance int, stripToAscii bool) *ParagraphSplitter {
	return &ParagraphSplitter{
		paragraphs:        make([]string, 0),
		MaxWords:          maxWords,
		MaxWordsTolerance: maxWordsTolerance,
		StripToAscii:      stripToAscii,
	}
}

func (ps *ParagraphSplitter) addParagraph(text string) {
	if text == "" {
		return
	}
	ps.paragraphs = append(ps.paragraphs, text)
}

func isSentenceEnd(word string) bool {
	if len(word) == 0 {
		return false
	}
	last := word[len(word)-1]
	return last == '.' || last == '!' || last == '?'
}

/*
Split the given text into paragraphs based on the following rules:

Pre process: Trim all non-ascii text and whitespace except new lines, can use a regular expression. I want a clean text with letters, numbers, punctuation and new lines only.

Split the text into words, an empty line should be represented as an empty string in the words array.

Run on the words array and split them into paragraphs, use the 'addParagraph' method to add a new paragraph.

New paragraphs should be created when:

1. An empty line is encountered.
2. A line that starts with a space followed by a non-space character is encountered (indicating an indented paragraph).
3. The number of words in the current paragraph exceeds 'maxWords' and the current word ends with a sentence-ending punctuation mark (., !, ?).
4. The number of words in the current paragraph exceeds maxWordsTolerance, in this case, split at the last space before maxWords if possible, otherwise split at maxWords.
*/
func (ps *ParagraphSplitter) Split(text string) []string {
	var clean = text
	if ps.StripToAscii {
		// Preprocess: keep only ASCII letters, numbers, punctuation, and newlines
		re := regexp.MustCompile(`[^\x20-\x7E\n]`)
		clean = re.ReplaceAllString(text, "")
	}
	// Normalize whitespace except newlines
	clean = regexp.MustCompile(`[ \t\r\f\v]+`).ReplaceAllStringFunc(clean, func(s string) string {
		if strings.Contains(s, "\n") {
			return "\n"
		}
		return " "
	})

	lines := strings.Split(clean, "\n")
	words := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimRightFunc(line, unicode.IsSpace)
		if trimmed == "" {
			words = append(words, "")
			continue
		}
		// preserve leading spaces for indented lines
		if len(line) > 1 && line[0] == ' ' && line[1] != ' ' {
			words = append(words, line)
			continue
		}

		for _, w := range strings.Fields(line) {
			words = append(words, w)
		}
		words[len(words)-1] += "\n" // mark end of line
	}

	ps.paragraphs = make([]string, 0)
	var (
		currentParagraph []string
		wordCount        int
	)

	for i := 0; i < len(words); i++ {
		word := words[i]

		// 1. Empty line triggers paragraph split
		if word == "" {
			ps.addParagraph(strings.Join(currentParagraph, " "))
			currentParagraph = nil
			wordCount = 0
			continue
		}

		// 2. Indented line triggers paragraph split
		if len(word) > 1 && word[0] == ' ' && word[1] != ' ' {
			ps.addParagraph(strings.Join(currentParagraph, " "))
			currentParagraph = []string{strings.TrimLeft(word, " ")}
			wordCount = len(strings.Fields(word))
			continue
		}

		currentParagraph = append(currentParagraph, word)
		wordCount++

		// 3. Split if wordCount > maxWords and ends with sentence-ending punctuation
		if wordCount > ps.MaxWords && isSentenceEnd(word) {
			ps.addParagraph(strings.Join(currentParagraph, " "))
			currentParagraph = nil
			wordCount = 0
			continue
		}

		// 4. Split if wordCount > maxWordsTolerance
		if wordCount > ps.MaxWordsTolerance {
			// Try to split at last space before maxWords
			splitIdx := -1
			count := 0
			for j := range currentParagraph {
				count++
				if count == ps.MaxWords {
					splitIdx = j
					break
				}
			}
			if splitIdx == -1 {
				splitIdx = ps.MaxWords
			}
			ps.addParagraph(strings.Join(currentParagraph[:splitIdx], " "))
			currentParagraph = currentParagraph[splitIdx:]
			wordCount = len(currentParagraph)
		}
	}
	ps.addParagraph(strings.Join(currentParagraph, " "))
	return ps.paragraphs
}
