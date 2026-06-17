package diag

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Report(collection *Collection) {
	for _, d := range collection.Diagnostics {
		printDiagnostic(d)
	}
}

func printDiagnostic(d Diagnostic) {
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	colorCyan := "\033[36m"
	colorBold := "\033[1m"
	colorBlue := "\033[34m"
	colorGray := "\033[90m"

	severityColor := colorRed
	if d.Severity == Warning {
		severityColor = colorYellow
	} else if d.Severity == Info {
		severityColor = colorCyan
	}

	// 1. Main Header
	fmt.Printf("%s%s%s%s: %s%s%s\n", colorBold, severityColor, d.Severity, colorReset, colorBold, d.Message, colorReset)
	fmt.Printf("  %s%s-->%s %s:%d:%d\n", colorGray, colorBold, colorReset, d.File, d.Range.Start.Line, d.Range.Start.Column)

	if d.File == "" {
		fmt.Println()
		return
	}

	// 2. Code Snippet
	file, err := os.Open(d.File)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		lineNum := 1
		found := false

		// Context lines: 1 before, 1 after (optional, but let's stick to current for now)
		for scanner.Scan() {
			if lineNum == d.Range.Start.Line {
				line := scanner.Text()
				fmt.Printf("%s%5d |%s %s\n", colorGray, lineNum, colorReset, line)

				// Caret Builder
				var paddingBuilder strings.Builder
				for i, ch := range line {
					if i >= d.Range.Start.Column-1 {
						break
					}
					if ch == '\t' {
						paddingBuilder.WriteRune('\t')
					} else {
						paddingBuilder.WriteRune(' ')
					}
				}
				padding := paddingBuilder.String()

				length := 1
				if d.Range.End.Line == d.Range.Start.Line && d.Range.End.Column > d.Range.Start.Column {
					length = d.Range.End.Column - d.Range.Start.Column
				} else if d.Range.End.Offset > d.Range.Start.Offset {
					length = d.Range.End.Offset - d.Range.Start.Offset
				}

				if length <= 0 {
					length = 1
				}
				caret := strings.Repeat("^", length)
				fmt.Printf("%s      |%s %s%s%s%s\n", colorGray, colorReset, padding, severityColor, caret, colorReset)
				found = true
				break
			}
			lineNum++
		}
		if !found {
			fmt.Printf("%s      |%s (source line %d not found)\n", colorGray, colorReset, d.Range.Start.Line)
		}
	}

	// 3. Notes
	for _, note := range d.Notes {
		fmt.Printf("  %s%s=%s %snote%s: %s\n", colorBlue, colorBold, colorReset, colorBold, colorReset, note)
	}

	// 4. Hint
	if d.Hint != "" {
		fmt.Printf("  %s%shelp%s: %s\n", colorCyan, colorBold, colorReset, d.Hint)
	}

	fmt.Println()
}
