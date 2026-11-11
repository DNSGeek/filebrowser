package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	cmdsCmd.AddCommand(cmdsLsCmd)
	cmdsLsCmd.Flags().StringP("event", "e", "", "event name, without 'before' or 'after'")
}

var cmdsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all commands for each event",
	Long:  `List all commands for each event.`,
	Args:  cobra.NoArgs,
	RunE: withStore(func(cmd *cobra.Command, _ []string, st *store) error {
		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		evt, err := cmd.Flags().GetString("event")
		if err != nil {
			return err
		}

		if evt == "" {
			printEvents(s.Hooks)
			printLegacyCommands(s)
		} else {
			show := map[string][]string{}
			show["before_"+evt] = s.Hooks["before_"+evt]
			show["after_"+evt] = s.Hooks["after_"+evt]
			printEvents(show)
		}

		return nil
	}, storeOptions{}),
}
