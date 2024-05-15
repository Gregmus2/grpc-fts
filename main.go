package main

import (
	"github.com/res-am/grpc-fts/internal"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	cApp := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run all test cases",
				Flags: []cli.Flag{
					config.ConfigsFlagSetup,
					config.VarFlagSetup,
					config.TargetFlagSetup,
					config.VerboseFlagSetup,
				},
				Action: func(ctx *cli.Context) error {
					return internal.NewContainer(ctx).RunTestCase()
				},
			},
			{
				Name:  "validate",
				Usage: "validate configuration",
				Flags: []cli.Flag{
					config.ConfigsFlagSetup,
					config.VerboseFlagSetup,
				},
				Action: func(ctx *cli.Context) error {
					return internal.NewContainer(ctx).Validate()
				},
			},
		},
	}

	err := cApp.Run(os.Args)
	if err != nil {
		println(err.Error()) //nolint:forbidigo

		os.Exit(1)
	}
}
