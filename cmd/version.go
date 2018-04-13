package cmd

import (
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"
)

const versionLongMessage = `
Shows the version of Solstice.
`

func newVersionCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  versionLongMessage,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf(`go version  : %s
go compiler : %s
platform    : %s/%s
`, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}

	return cmd
}
