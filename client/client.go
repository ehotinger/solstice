package client

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2018-02-01-preview/containerregistry"
	"github.com/ehotinger/solstice/iam"
)

// GetRegistriesClient returns a client to interact with registry resources.
func GetRegistriesClient(subID string) (c containerregistry.RegistriesClient, err error) {
	registriesClient := containerregistry.NewRegistriesClient(subID)
	auth, err := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	if err != nil {
		return c, fmt.Errorf("Failed to get client. Err: %v", err)
	}
	registriesClient.Authorizer = auth
	registriesClient.AddToUserAgent(containerregistry.UserAgent())
	return registriesClient, nil
}

// GetBuildsClient returns a client to interact with builds.
func GetBuildsClient(subID string) (c containerregistry.BuildsClient, err error) {
	buildsClient := containerregistry.NewBuildsClient(subID)
	auth, err := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	if err != nil {
		return c, fmt.Errorf("Failed to get client. Err: %v", err)
	}
	buildsClient.Authorizer = auth
	buildsClient.AddToUserAgent(containerregistry.UserAgent())
	return buildsClient, nil
}
