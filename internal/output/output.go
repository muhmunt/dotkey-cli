package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var Quiet bool

var (
	successMark = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("✓")
	errorMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✗")
	infoMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("→")
	warnMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("!")

	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle   = lipgloss.NewStyle().Bold(true)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	addedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	changedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	removedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	sameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func Success(format string, args ...any) {
	if Quiet {
		return
	}
	fmt.Println(successMark + " " + fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	fmt.Fprintln(os.Stderr, errorMark+" "+fmt.Sprintf(format, args...))
}

func Fatal(format string, args ...any) {
	Error(format, args...)
	os.Exit(1)
}

func Info(format string, args ...any) {
	if Quiet {
		return
	}
	fmt.Println(infoMark + " " + fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	if Quiet {
		return
	}
	fmt.Println(warnMark + " " + fmt.Sprintf(format, args...))
}

func Println(msg string) {
	if Quiet {
		return
	}
	fmt.Println(msg)
}

func Dim(msg string) string    { return dimStyle.Render(msg) }
func Bold(msg string) string   { return boldStyle.Render(msg) }

// Table prints a clean headerless-border table to stdout.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		Println(Dim("  (none)"))
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := range widths {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	// header row
	var hdr strings.Builder
	hdr.WriteString("  ")
	for i, h := range headers {
		col := headerStyle.Render(padRight(strings.ToUpper(h), widths[i]))
		hdr.WriteString(col)
		if i < len(headers)-1 {
			hdr.WriteString("   ")
		}
	}
	fmt.Println(hdr.String())
	fmt.Println()

	for _, row := range rows {
		var line strings.Builder
		line.WriteString("  ")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if i < len(headers)-1 {
				line.WriteString(padRight(cell, widths[i]))
				line.WriteString("   ")
			} else {
				line.WriteString(cell)
			}
		}
		fmt.Println(line.String())
	}
}

// DiffTable prints the diff with colored status symbols.
func DiffTable(headers []string, rows [][]string, statusCol int) {
	if len(rows) == 0 {
		Println(Dim("  no differences"))
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := range widths {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var hdr strings.Builder
	hdr.WriteString("  ")
	for i, h := range headers {
		hdr.WriteString(headerStyle.Render(padRight(strings.ToUpper(h), widths[i])))
		if i < len(headers)-1 {
			hdr.WriteString("   ")
		}
	}
	fmt.Println(hdr.String())
	fmt.Println()

	for _, row := range rows {
		var line strings.Builder
		line.WriteString("  ")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}

			var rendered string
			if i == statusCol {
				rendered = colorStatus(cell)
			} else if i < len(headers)-1 {
				rendered = padRight(cell, widths[i])
			} else {
				rendered = cell
			}

			line.WriteString(rendered)
			if i < len(headers)-1 {
				line.WriteString("   ")
			}
		}
		fmt.Println(line.String())
	}
}

func colorStatus(status string) string {
	switch status {
	case "changed":
		return changedStyle.Render("~ changed     ")
	case "missing_in_b":
		return removedStyle.Render("+ missing in B")
	case "missing_in_a":
		return addedStyle.Render("+ missing in A")
	case "same":
		return sameStyle.Render("✓ same        ")
	default:
		return status
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Confirm asks the user a yes/no question and returns true for "y" or "Y".
func Confirm(question string) bool {
	fmt.Printf("%s (y/N): ", question)
	var answer string
	fmt.Scanln(&answer)
	return strings.ToLower(strings.TrimSpace(answer)) == "y"
}

// Prompt asks the user for input with a label. Returns the typed string.
func Prompt(label string) string {
	fmt.Printf("%s: ", label)
	var answer string
	fmt.Scanln(&answer)
	return strings.TrimSpace(answer)
}
