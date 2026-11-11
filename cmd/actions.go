package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DNSGeek/filebrowser/v3/runner"
)

func init() {
	rootCmd.AddCommand(actionsCmd)
	actionsCmd.AddCommand(actionsAddCmd, actionsRmCmd, actionsLsCmd)
}

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "Command runner actions management utility",
	Long: `Command runner actions management utility.

An action is a command users can start by name from the command runner in
the web interface, when it is enabled with --disable-exec=false. Actions
can only be managed with this command, on the machine File Browser runs
on. A user can run the actions listed in their "commands" setting. See
docs/command-execution.md.`,
	Args: cobra.NoArgs,
}

var actionsAddCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Define an action, replacing any with the same name",
	Long: `Define an action, replacing any with the same name.

The command is a program, given by its absolute path, followed by fixed
arguments. It is started directly, without a shell. One argument may be
$FILES: it is replaced by the paths the user passes, each resolved to a
file inside that user's scope. Without $FILES the user cannot pass
anything. Put "--" before $FILES when the program supports it, and quote
the command so your own shell does not expand it:

  filebrowser actions add ls '/bin/ls -la -- $FILES'`,
	Args: cobra.MinimumNArgs(2),
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		definition := strings.Join(args[1:], " ")
		if err := runner.ValidateAction(args[0], definition); err != nil {
			return err
		}

		if s.Actions == nil {
			s.Actions = map[string]string{}
		}
		s.Actions[args[0]] = definition
		if err := st.Settings.Save(s); err != nil {
			return err
		}
		printActions(s.Actions)
		return nil
	}, storeOptions{}),
}

var actionsRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove an action",
	Long:  `Remove an action.`,
	Args:  cobra.ExactArgs(1),
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		if _, ok := s.Actions[args[0]]; !ok {
			return fmt.Errorf("no action named %q", args[0])
		}
		delete(s.Actions, args[0])
		if err := st.Settings.Save(s); err != nil {
			return err
		}
		printActions(s.Actions)
		return nil
	}, storeOptions{}),
}

var actionsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all actions",
	Long:  `List all actions.`,
	Args:  cobra.NoArgs,
	RunE: withStore(func(_ *cobra.Command, _ []string, st *store) error {
		s, err := st.Settings.Get()
		if err != nil {
			return err
		}
		printActions(s.Actions)
		return nil
	}, storeOptions{}),
}

func printActions(actions map[string]string) {
	names := make([]string, 0, len(actions))
	for name := range actions {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		fmt.Printf("%s: %s\n", name, actions[name])
	}
}
