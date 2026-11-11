package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/DNSGeek/filebrowser/v3/runner"
)

func init() {
	cmdsCmd.AddCommand(cmdsAddCmd)
}

var cmdsAddCmd = &cobra.Command{
	Use:   "add <event> <command>",
	Short: "Add a command to run on a specific event",
	Long: `Add a command to run on a specific event, such as before_upload or
after_delete.

The command is started directly, without a shell: there are no pipes,
redirections or command substitutions. An argument that is exactly $FILE,
$SCOPE, $TRIGGER, $USERNAME or $DESTINATION is replaced by its value, and
the same variables are set in the command's environment. Quote the command
so your own shell does not expand them first:

  filebrowser cmds add after_upload '/usr/local/bin/scan $FILE'

End the command with & to run it without waiting for it to finish.`,
	Args: cobra.MinimumNArgs(2),
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		s, err := st.Settings.Get()
		if err != nil {
			return err
		}
		if err := checkHookEvent(args[0]); err != nil {
			return err
		}
		command := strings.Join(args[1:], " ")
		if _, _, err := runner.SplitCommandAndArgs(strings.TrimSuffix(command, "&")); err != nil {
			return err
		}
		s.Hooks[args[0]] = append(s.Hooks[args[0]], command)
		err = st.Settings.Save(s)
		if err != nil {
			return err
		}
		printEvents(s.Hooks)
		return nil
	}, storeOptions{}),
}
