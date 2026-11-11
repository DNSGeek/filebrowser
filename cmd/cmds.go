package cmd

import (
	"fmt"
	"strings"

	"github.com/DNSGeek/filebrowser/v3/settings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdsCmd)
}

var cmdsCmd = &cobra.Command{
	Use:   "cmds",
	Short: "Event hooks management utility",
	Long: `Event hooks management utility.

Hooks are commands that run before or after a file event (save, copy,
rename, upload, delete). They can only be managed with this command, on
the machine File Browser runs on, and are started directly, without a
shell. See docs/command-execution.md.`,
	Args: cobra.NoArgs,
}

// checkHookEvent verifies that evt is "before_" or "after_" followed by one of
// the known events.
func checkHookEvent(evt string) error {
	for _, e := range settings.HookEvents {
		if evt == "before_"+e || evt == "after_"+e {
			return nil
		}
	}
	return fmt.Errorf("unknown event %q: use before_ or after_ followed by one of %s",
		evt, strings.Join(settings.HookEvents, ", "))
}

// printLegacyCommands lists the hooks left over from when they were editable
// in the web interface. They are not run.
func printLegacyCommands(s *settings.Settings) {
	if !s.HasLegacyCommands() {
		return
	}
	fmt.Println("\nNot run: hooks defined before they became command-line only.")
	fmt.Println("Review them and re-add the ones you want with 'filebrowser cmds add'.")
	printEvents(s.Commands)
}

func printEvents(m map[string][]string) {
	for evt, cmds := range m {
		for i, cmd := range cmds {
			fmt.Printf("%s(%d): %s\n", evt, i, cmd)
		}
	}
}
