package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
	"github.com/ehotinger/solstice/client"
	"github.com/ehotinger/solstice/pkg/blob"
	"github.com/spf13/cobra"
)

type logsCmd struct {
	resourceGroupName string
	registryName      string
	buildID           string
	out               io.Writer
}

func newLogsCmd(out io.Writer) *cobra.Command {
	logsCmd := &logsCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show logs",
		Long:  "Show logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return logsCmd.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&logsCmd.resourceGroupName, "rg", "", "The resource group to use for auth")
	f.StringVar(&logsCmd.registryName, "n", "", "The name of the registry")
	f.StringVar(&logsCmd.buildID, "b", "", "The build id to look for logs")

	return cmd
}

func (cmd *logsCmd) run() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
	defer cancel()

	subscription, err := getSubscriptionFromProfile()
	if err != nil {
		return fmt.Errorf("There was an error while grabbing the subscription: %v", err)
	}

	fmt.Println("Getting client...")
	c, err := client.GetBuildsClient(subscription.ID)
	if err != nil {
		return fmt.Errorf("Errored while creating client. Err: %v", err)
	}

	logResult, err := c.GetLogLink(ctx, cmd.resourceGroupName, cmd.registryName, cmd.buildID)
	if err != nil {
		return fmt.Errorf("Errored while getting log link. Err: %v", err)
	}

	logSAS := *logResult.LogLink

	if logSAS == "" {
		return errors.New("Unable to create a link to the logs")
	}

	fmt.Printf("Log URL: %s\n", logSAS)
	blobURL := blob.GetAppendBlobURL(logSAS)

	contentLength := int64(0) // Used for progress reporting to report the total number of bytes being downloaded.

	// NewGetRetryStream creates an intelligent retryable stream around a blob; it returns an io.ReadCloser.
	rs := azblob.NewDownloadStream(context.Background(),
		// We pass more tha "blobUrl.GetBlob" here so we can capture the blob's full
		// content length on the very first internal call to Read.
		func(ctx context.Context, blobRange azblob.BlobRange, ac azblob.BlobAccessConditions, rangeGetContentMD5 bool) (*azblob.GetResponse, error) {
			get, err := blobURL.GetBlob(ctx, blobRange, ac, rangeGetContentMD5)
			if err == nil && contentLength == 0 {
				// If 1st successful Get, record blob's full size for progress reporting
				contentLength = get.ContentLength()
			}
			return get, err
		},
		azblob.DownloadStreamOptions{})

	// NewResponseBodyStream wraps the GetRetryStream with progress reporting; it returns an io.ReadCloser.
	stream := pipeline.NewResponseBodyProgress(rs,
		func(bytesTransferred int64) {
			fmt.Printf("Downloaded %d of %d bytes.\n", bytesTransferred, contentLength)
		})
	defer stream.Close() // The client must close the response body when finished with it

	fmt.Println("\nLogs:")
	_, err = io.Copy(os.Stdout, stream)
	if err != nil {
		log.Fatal(err)
	}

	return err
}
