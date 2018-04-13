package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const globalUsageMessage = `Solstice.

To start working with solstice, run 'solstice help'.
`

// Execute executes the root command.
func Execute() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "solstice",
		Short:        "A CLI for ACR Build.",
		Long:         globalUsageMessage,
		SilenceUsage: true,
	}

	flags := cmd.PersistentFlags()

	out := cmd.OutOrStdout()

	cmd.AddCommand(
		newVersionCmd(out),
	)

	flags.Parse(args)

	return cmd
}
