package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

const (
	blue  = "\033[34m"
	reset = "\033[0m"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // Fallback to 80 columns
	}
	return width
}

func separator() string {
	width := getTerminalWidth()
	return strings.Repeat("─", width)
}

func header(title string) string {
	width := getTerminalWidth()
	prefix := "──["
	suffix := "]"
	headerText := prefix + title + suffix
	headerLen := len(headerText)

	if headerLen >= width {
		return headerText
	}

	remaining := width - headerLen
	return headerText + strings.Repeat("─", remaining)
}

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func truncateLine(line string, maxWidth int) string {
	visibleLen := len(stripAnsi(line))
	if visibleLen <= maxWidth {
		return line
	}

	// Truncate at maxWidth-1 to leave room for '>'
	targetLen := maxWidth - 1
	result := ""
	visibleCount := 0
	inAnsi := false
	hasAnsi := ansiRegex.MatchString(line)

	for i := 0; i < len(line); i++ {
		if line[i] == '\x1b' {
			inAnsi = true
		}

		if inAnsi {
			result += string(line[i])
			if line[i] == 'm' {
				inAnsi = false
			}
		} else {
			if visibleCount >= targetLen {
				break
			}
			result += string(line[i])
			visibleCount++
		}
	}

	// If line had color codes, reset before adding '>'
	if hasAnsi {
		return result + reset + ">"
	}
	return result + ">"
}

func alignDescriptions(text string) string {
	lines := strings.Split(text, "\n")

	// Find the position where descriptions start (after blue color code)
	// Format: "  command    [34mdescription[0m"
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.WriteString(line + "\n")
			continue
		}

		// Find the blue color code which marks the start of description
		blueIdx := strings.Index(line, blue)
		if blueIdx == -1 {
			result.WriteString(line + "\n")
			continue
		}

		// Extract command part (before blue) and description part (from blue onwards)
		commandPart := line[:blueIdx]
		descriptionPart := line[blueIdx:]

		// Pad command part to 24 characters (aligns all descriptions)
		paddedCommand := commandPart
		visibleLen := len(stripAnsi(commandPart))
		if visibleLen < 24 {
			paddedCommand += strings.Repeat(" ", 24-visibleLen)
		}

		result.WriteString(paddedCommand + descriptionPart + "\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

func printTruncated(text string) {
	width := getTerminalWidth()
	aligned := alignDescriptions(text)
	lines := strings.Split(aligned, "\n")
	for _, line := range lines {
		fmt.Println(truncateLine(line, width))
	}
}

func printParagraph(text string) {
	// Paragraphs wrap naturally, no truncation
	fmt.Println(text)
}

func printInfo(main, description string) {
	width := getTerminalWidth()
	// Add 2 spaces indent and pad to 24 characters for alignment
	mainPart := "  " + main
	visibleLen := len(mainPart)
	if visibleLen < 24 {
		mainPart += strings.Repeat(" ", 24-visibleLen)
	}
	line := mainPart + blue + description + reset
	fmt.Println(truncateLine(line, width))
}

func Help(args []string) {
	// Handle subcommands
	if len(args) > 0 {
		switch args[0] {
		case "config":
			showConfigHelp()
			return
		}
	}

	// Main help
	fmt.Println("\n" + header("Usage"))
	fmt.Println("\n  yap <command> [options]\n")

	printInfo("Location:", "~/.config/yappers-of-linux/config.toml")
	printInfo("Edit:", "yap config (opens in $EDITOR)")
	fmt.Println()

	fmt.Println("\n" + header("Commands"))
	commands := `
  start [options]     ` + blue + `Start voice typing (default: --model tiny)` + reset + `
  toggle [options]    ` + blue + `Smart pause/resume/start` + reset + `
  pause               ` + blue + `Pause listening` + reset + `
  resume              ` + blue + `Resume listening` + reset + `
  stop (kill)         ` + blue + `Stop voice typing` + reset + `
  models              ` + blue + `Show installed models` + reset + `
  config              ` + blue + `Open config file in $EDITOR` + reset
	printTruncated(commands)

	fmt.Println("\n" + header("Options"))
	options := `
  --model X           ` + blue + `Model size: tiny, base, small, medium, large` + reset + `
  --cpu / --gpu       ` + blue + `Device selection` + reset + `
  --language X        ` + blue + `Language code (default: en)` + reset + `
  --lang X            ` + blue + `Short alias for --language` + reset + `
  --tcp [PORT]        ` + blue + `Enable TCP server (default port: 12322)` + reset + `
  --fast              ` + blue + `Use fast mode (int8, less accurate but faster)` + reset + `
  --no-typing         ` + blue + `Disable keyboard typing (only print to terminal)` + reset
	printTruncated(options)

	fmt.Println("\n" + header("Modes"))
	modes := `
  default             ` + blue + `Accurate mode (float32, better quality)` + reset + `
  --fast              ` + blue + `Fast mode (int8, lower quality but faster)` + reset
	printTruncated(modes)

	fmt.Println()
	printParagraph("Models automatically download on first use")
	fmt.Println("\n" + header("For more help"))
	helpLine := `
  yap help config       ` + blue + `Configuration file syntax` + reset
	printTruncated(helpLine)
}

func showConfigHelp() {
	fmt.Println("\nConfiguration File")
	printInfo("Location:", "~/.config/yappers-of-linux/config.toml")
	printInfo("Edit:", "yap config (opens in $EDITOR)")

	fmt.Println("\n" + header("Notifications"))
	fmt.Println()
	printParagraph("Control desktop notifications with comma-separated events and optional 'urgent' modifier. Events: start (ready to listen), pause, stop. Urgent makes notifications persistent and ignores Do Not Disturb mode.")

	fmt.Println("\n" + header("Examples"))
	examples := `
  "start,pause,stop"        ` + blue + `All events, normal priority` + reset + `
  "start,pause,stop,urgent" ` + blue + `All events, urgent (persistent, ignores DND)` + reset + `
  "start"                   ` + blue + `Only when ready to listen` + reset + `
  "urgent"                  ` + blue + `Shorthand for "start,urgent"` + reset + `
  "false" / ""              ` + blue + `Disabled (false or empty string)` + reset
	printTruncated(examples)
}
