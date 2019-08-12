package stresscmd

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"os"
	"os/signal"

	"github.com/mitchellh/cli"
)

type Endpoint interface {
	Upload(string, io.Reader) (*http.Response, error)
	Download(string) (*http.Response, error)
}

type StressCommand struct {
	UI       cli.Ui
	Endpoint Endpoint

	DownloadWorkers int
	UploadWorkers   int
	FileSize        int
	Repo            string
}

func (c *StressCommand) Help() string {
	return `stress is a command used to test the health of an artifactory instance.
	
It collects credentials from the environment variables FOO_BAR and BAZ_BING.

It then starts a number of upload and download processes.

Upload processes upload a file full of random bytes. Download processes download a randomly
selected file that has been uploaded. The HTTP return codes and errors are logged and stored

Flags:
  --download-workers The number of processes downloading artifacts (default: 10)
  --upload-workers   The number of processes uploading artifacts (default: 2)
  --file-size        The size of file to upload in megabytes (default: 50)
  --repo             The artifactory repo to target
`
}

func (c *StressCommand) Synopsis() string {
	return "Tests the health of an artifactory instance by applying stress."
}

func (c *StressCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }

	cmdFlags.IntVar(&c.DownloadWorkers, "download-workers", 10, "The number of processes downloading artifacts desired")
	cmdFlags.IntVar(&c.UploadWorkers, "upload-workers", 2, "The number of processes uploading artifacts")
	cmdFlags.IntVar(&c.FileSize, "file-size", 50, "The size of file to upload in megabytes (default: 50)")
	cmdFlags.StringVar(&c.Repo, "repo", "", "The repo to target with files")
	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("error: error parsing flags: %s", err.Error()))
		return 1
	}

	if c.Repo == "" {
		c.UI.Error("error: required flag --repo was not set")
		return 1
	}

	dlWork := make(chan string, 5)

	// TODO: write results to a channel, output info after cancel

	for uw := 1; uw <= c.UploadWorkers; uw++ {
		go c.upload(uw, dlWork)
	}

	for dw := 1; dw <= c.DownloadWorkers; dw++ {
		go c.download(dw, dlWork)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	<-stopChan
	close(dlWork)

	return 0
}

func (c *StressCommand) upload(workerIdx int, dlWork chan<- string) {
	for {
		size := c.FileSize * 1000000
		b := make([]byte, size)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		// Create a buffer from the random file, create a second buffer for the upload writer
		buf := bytes.NewBuffer(b)
		var uploadBuf bytes.Buffer

		// Also calculate the md5 sum
		h := md5.New()

		// Write the contents of the random file to both the md5 calculator and the upload buffer
		mw := io.MultiWriter(h, &uploadBuf)
		_, err = io.Copy(mw, buf)
		if err != nil {
			panic(err)
		}

		// Upload to artifactory
		endpoint := fmt.Sprintf("%s/%d/%x", c.Repo, workerIdx, h.Sum(nil))
		res, err := c.Endpoint.Upload(endpoint, &uploadBuf)
		if err != nil {
			c.UI.Error(fmt.Sprintf(
				"5xx - upload   worker %d - %s - %s", workerIdx, endpoint, err.Error(),
			))
			continue
		}

		if res.StatusCode >= 400 {
			c.UI.Error(fmt.Sprintf(
				"%d - upload   worker %d - %s", res.StatusCode, workerIdx, endpoint,
			))
		} else {
			c.UI.Output(fmt.Sprintf(
				"%d - upload   worker %d - %s", res.StatusCode, workerIdx, endpoint,
			))
		}
		res.Body.Close()

		dlWork <- endpoint
	}
}

func (c *StressCommand) download(workerIdx int, dlWork <-chan string) {
	var assignedEndpoints []string
	for {
		// TODO: This selection algorithm doesn't balance work super efficiently, since endpoints stay
		// assigned to the same worker forever
		var selectedEndpoint string
		select {
		case endpoint := <-dlWork:
			selectedEndpoint = endpoint
			assignedEndpoints = append(assignedEndpoints, endpoint)
		default:
			if len(assignedEndpoints) == 0 {
				continue
			}
			selectedEndpoint = assignedEndpoints[mathrand.Intn(len(assignedEndpoints))]
		}

		res, err := c.Endpoint.Download(selectedEndpoint)
		if err != nil {
			c.UI.Error(fmt.Sprintf(
				"5xx - download worker %d - %s - %s", workerIdx, selectedEndpoint, err.Error(),
			))
			continue
		}

		if res.StatusCode >= 400 {
			c.UI.Error(fmt.Sprintf(
				"%d - download worker %d - %s", res.StatusCode, workerIdx, selectedEndpoint,
			))
		} else {
			c.UI.Output(fmt.Sprintf(
				"%d - download worker %d - %s", res.StatusCode, workerIdx, selectedEndpoint,
			))
		}
		res.Body.Close()
	}
}
