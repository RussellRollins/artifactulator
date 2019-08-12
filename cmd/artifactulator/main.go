package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"

	"github.com/hashicorp/artifactulator/pkg/artifactory"
	"github.com/hashicorp/artifactulator/pkg/stresscmd"
)

func main() {
	if err := inner(); err != nil {
		log.Printf("citool error: %s\n", err)
	}
}

func inner() error {
	ui := &cli.ConcurrentUi{
		Ui: &cli.ColoredUi{
			Ui: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			OutputColor: cli.UiColorGreen,
			ErrorColor:  cli.UiColorRed,
		},
	}

	afy, err := artifactory.NewArtifactory(
		os.Getenv("ARTIFACTORY_HOST"),
		os.Getenv("ARTIFACTORY_USER"),
		os.Getenv("ARTIFACTORY_TOKEN"),
	)
	if err != nil {
		return err
	}

	c := cli.NewCLI("artifactulator", "0.1")

	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"stress": func() (cli.Command, error) {
			return &stresscmd.StressCommand{UI: ui, Endpoint: afy}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		return err
	}

	os.Exit(exitStatus)
	return nil
}
