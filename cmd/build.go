package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2018-02-01-preview/containerregistry"
	"github.com/ehotinger/solstice/client"
	"github.com/spf13/cobra"
)

type buildCmd struct {
	subscriptionID    string
	resourceGroupName string
	registryName      string
	out               io.Writer
}

func newBuildCmd(out io.Writer) *cobra.Command {

	buildCmd := &buildCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Queue a build",
		Long:  "Queue a build",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
			defer cancel()

			fmt.Println("Getting client...")
			c, err := client.GetRegistriesClient(buildCmd.subscriptionID)
			if err != nil {
				return fmt.Errorf("Errored while creating client. Err: %v", err)
			}

			// TODO: make all this configurable...

			imageName := "acr-builder"
			sourceLocation := "https://github.com/Azure/acr-builder"
			platform := containerregistry.PlatformProperties{OsType: "Linux"}
			push := true
			var timeout int32
			timeout = 600
			dockerFilePath := "."
			t := containerregistry.TypeQuickBuild

			req := containerregistry.QuickBuildRequest{
				ImageName:      &imageName,
				SourceLocation: &sourceLocation,
				BuildArguments: nil,
				IsPushEnabled:  &push,
				Timeout:        &timeout,
				Platform:       &platform,
				DockerFilePath: &dockerFilePath,
				Type:           t,
			}
			fmt.Println("Creating quick build request...")
			bas, ok := req.AsBasicQueueBuildRequest()
			if !ok {
				return errors.New("Failed to create quick build request")
			}

			fmt.Println("Queuing build...")
			future, err := c.QueueBuild(ctx, buildCmd.resourceGroupName, buildCmd.registryName, bas)
			if err != nil {
				return fmt.Errorf("Errored while queuing build. Err: %v", err)
			}

			fmt.Println("Waiting for completion...")
			err = future.WaitForCompletion(ctx, c.Client)
			if err != nil {
				return fmt.Errorf("Errored while waiting for completion")
			}

			fmt.Println()
			fin, err := future.Result(c)

			fmt.Printf("Build ID: %s\n", *fin.BuildID)

			return err
		},
	}

	f := cmd.Flags()
	f.StringVar(&buildCmd.subscriptionID, "s", "", "The subscription ID to use for auth")
	f.StringVar(&buildCmd.resourceGroupName, "rg", "", "The resource group to use for auth")
	f.StringVar(&buildCmd.registryName, "n", "", "The name of the registry")

	return cmd
}
