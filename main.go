package main

import (
	"github.com/res-am/grpc-fts/internal"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	cApp := &cli.App{
		Name:  "run",
		Usage: "run all test cases",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "configs",
				Value: ".",
				Usage: "path to configs directory",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Value: cli.NewStringSlice(),
				Usage: "user defined variables, format: key=value",
			},
		},
		Action: func(ctx *cli.Context) error {
			app, err := internal.NewApp(ctx)
			if err != nil {
				return err
			}

			return app.RunTestCases()
		},
	}

	err := cApp.Run(os.Args)
	if err != nil {
		println(err.Error()) //nolint:forbidigo

		os.Exit(1)
	}
}
