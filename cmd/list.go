package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/ehotinger/solstice/client"
	"github.com/spf13/cobra"
)

type listCmd struct {
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
			return listCmd.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&listCmd.resourceGroupName, "rg", "", "The resource group to use for auth")
	f.StringVar(&listCmd.registryName, "n", "", "The name of the registry")

	return cmd
}

func (c *listCmd) run() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
	defer cancel()

	subscription, err := getSubscriptionFromProfile()
	if err != nil {
		return fmt.Errorf("There was an error while grabbing the subscription: %v", err)
	}

	fmt.Println("Getting client...")
	client, err := client.GetBuildsClient(subscription.ID)
	if err != nil {
		return fmt.Errorf("Errored while creating client. Err: %v", err)
	}

	fmt.Println("Listing builds...")
	page, err := client.List(ctx, c.resourceGroupName, c.registryName, "", nil, "")
	if err != nil {
		return fmt.Errorf("Errored while listing builds. Err: %v", err)
	}

	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "Build ID\tCreate Time\tStart Time\tFinish Time")

	builds := page.Values()
	for _, b := range builds {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", *b.BuildID, *b.CreateTime, *b.StartTime, *b.FinishTime)
	}

	w.Flush()
	return err
}
