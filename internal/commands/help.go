package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/gohelp"
)

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
	gohelp.PrintHeader("Usage")
	fmt.Println("  yap <command> [options]")
	fmt.Println()

	gohelp.Item("Location:", "~/.config/yappers-of-linux/config.toml")
	gohelp.Item("Edit:", "yap config (opens in $EDITOR)")

	gohelp.PrintHeader("Commands")
	gohelp.Item("start [options]", "Start voice typing (default: --model tiny)")
	gohelp.Item("toggle [options]", "Smart pause/resume/start")
	gohelp.Item("pause", "Pause listening")
	gohelp.Item("resume", "Resume listening")
	gohelp.Item("stop (kill)", "Stop voice typing")
	gohelp.Item("models", "Show installed models")
	gohelp.Item("config", "Open config file in $EDITOR")
	gohelp.Item("version", "Show version and check for updates")
	gohelp.Item("update [--force]", "Update to latest version")

	gohelp.PrintHeader("Options")
	gohelp.Item("--model X", "Model size: tiny, base, small, medium, large")
	gohelp.Item("--cpu / --gpu", "Device selection")
	gohelp.Item("--language X", "Language code (default: en)")
	gohelp.Item("--lang X", "Short alias for --language")
	gohelp.Item("--tcp [PORT]", "Enable TCP server (default port: 12322)")
	gohelp.Item("--fast", "Use fast mode (int8, less accurate but faster)")
	gohelp.Item("--no-typing", "Disable keyboard typing (only print to terminal)")

	gohelp.PrintHeader("Modes")
	gohelp.Item("default", "Accurate mode (float32, better quality)")
	gohelp.Item("--fast", "Fast mode (int8, lower quality but faster)")

	gohelp.Paragraph("Models automatically download on first use")

	gohelp.PrintHeader("For more help")
	gohelp.Item("yap help config", "Configuration file syntax")
}

func showConfigHelp() {
	gohelp.PrintHeader("Configuration File")
	gohelp.Item("Location:", "~/.config/yappers-of-linux/config.toml")
	gohelp.Item("Edit:", "yap config (opens in $EDITOR)")

	gohelp.PrintHeader("Notifications")
	gohelp.Paragraph("Control desktop notifications with comma-separated events and optional 'urgent' modifier. Events: start (yapping started), pause, stop. Urgent makes notifications persistent and ignores Do Not Disturb mode.")

	gohelp.PrintHeader("Examples")
	gohelp.Item(`"start,pause,stop"`, "All events, normal priority")
	gohelp.Item(`"start,pause,stop,urgent"`, "All events, urgent (persistent, ignores DND)")
	gohelp.Item(`"start"`, "Only when yapping starts")
	gohelp.Item(`"urgent"`, `Shorthand for "start,urgent"`)
	gohelp.Item(`"false" / ""`, "Disabled (false or empty string)")

	gohelp.PrintHeader("Output File")
	gohelp.Paragraph("Write transcriptions to output.txt for piping to other scripts or automation. File is ephemeral - deleted on each start for fresh sessions. Location: ~/.config/yappers-of-linux/output.txt")
	gohelp.Item("output_file = true", "Enable file output")
	gohelp.Item("output_file = false", "Disable (default)")
}
