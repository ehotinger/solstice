package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ehotinger/solstice/client"
	"github.com/spf13/cobra"
)

type listCmd struct {
	subscriptionID    string
	resourceGroupName string
	registryName      string
	out               io.Writer
}

func newListCmd(out io.Writer) *cobra.Command {
	listCmd := &listCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List builds",
		Long:  "List builds",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
			defer cancel()

			fmt.Println("Getting client...")
			c, err := client.GetBuildsClient(listCmd.subscriptionID)
			if err != nil {
				return fmt.Errorf("Errored while creating client. Err: %v", err)
			}

			filter := ""
			var top int32
			top = 100
			skipToken := ""

			fmt.Println("Listing builds...")
			page, err := c.List(ctx, listCmd.resourceGroupName, listCmd.registryName, filter, &top, skipToken)
			if err != nil {
				return fmt.Errorf("Errored while listing builds. Err: %v", err)
			}

			vals := page.Values()

			fmt.Printf("Values: %v", vals)
			return err
		},
	}

	f := cmd.Flags()
	f.StringVar(&listCmd.subscriptionID, "s", "", "The subscription ID to use for auth")
	f.StringVar(&listCmd.resourceGroupName, "rg", "", "The resource group to use for auth")
	f.StringVar(&listCmd.registryName, "n", "", "The name of the registry")

	return cmd
}
