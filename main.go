package main

import (
	"errors"
	"github.com/res-am/grpc-fts/internal"
	"github.com/res-am/grpc-fts/internal/models"
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
			&cli.StringFlag{
				Name:  "target",
				Usage: "to run only specific test case, format: test_case_name",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "verbose output",
			},
		},
		Action: func(ctx *cli.Context) error {
			err := internal.RunTestCase(ctx)
			if err != nil {
				return err
			}
			var userErr models.UserErr
			if !ctx.Bool("verbose") && errors.As(err, &userErr) {
				return userErr
			}

			return err
		},
	}

	err := cApp.Run(os.Args)
	if err != nil {
		println(err.Error()) //nolint:forbidigo

		os.Exit(1)
	}
}
