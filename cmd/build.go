package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2018-02-01-preview/containerregistry"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/ehotinger/solstice/client"
	"github.com/spf13/cobra"
)

type buildCmd struct {
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

			subscription, err := getSubscriptionFromProfile()
			if err != nil {
				return fmt.Errorf("There was an error while grabbing the subscription: %v", err)
			}

			c, err := client.GetRegistriesClient(subscription.ID)
			if err != nil {
				return fmt.Errorf("could not get registry client: %v", err)
			}

			// Get the authorizer for auth access
			tokenPath, err := cli.AccessTokensPath()
			if err != nil {
				return fmt.Errorf("There was an error while grabbing the access token path: %v", err)
			}
			var authorizer autorest.Authorizer
			tokens, err := cli.LoadTokens(tokenPath)
			if err != nil {
				return fmt.Errorf("There was an error loading the tokens from %s: %v", tokenPath, err)
			}
			for _, token := range tokens {
				adalToken, err := token.ToADALToken()
				if err != nil {
					continue
				}
				if adalToken.IsExpired() {
					continue
				}
				authorizer = autorest.NewBearerAuthorizer(&adalToken)
				break
			}
			if authorizer == nil {
				return fmt.Errorf("run `az login` to get started")
			}
			c.Authorizer = authorizer

			// TODO: make all this configurable...

			req := containerregistry.QuickBuildRequest{
				ImageName:      to.StringPtr("acr-builder"),
				SourceLocation: to.StringPtr("https://github.com/deis/example-dockerfile-http/archive/master.tar.gz"),
				BuildArguments: nil,
				IsPushEnabled:  to.BoolPtr(true),
				Timeout:        to.Int32Ptr(600),
				Platform: &containerregistry.PlatformProperties{
					OsType: containerregistry.Linux,
					// NB: CPU isn't required right now, possibly want to make this configurable
					// It'll actually default to 2 from the server
					// CPU: to.Int32Ptr(1),
				},
				DockerFilePath: to.StringPtr("Dockerfile"),
				Type:           containerregistry.TypeQuickBuild,
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
	f.StringVar(&buildCmd.resourceGroupName, "rg", "", "The resource group to use for auth")
	f.StringVar(&buildCmd.registryName, "n", "", "The name of the registry")

	return cmd
}

func getSubscriptionFromProfile() (*cli.Subscription, error) {
	profilePath, err := cli.ProfilePath()
	if err != nil {
		return nil, err
	}
	profile, err := cli.LoadProfile(profilePath)
	if err != nil {
		return nil, err
	}
	var subscription *cli.Subscription
	for _, sub := range profile.Subscriptions {
		if sub.IsDefault {
			subscription = &sub
		}
	}
	if subscription == nil {
		return nil, fmt.Errorf("could not find a default subscription ID from %s", profilePath)
	}
	return subscription, nil
}
