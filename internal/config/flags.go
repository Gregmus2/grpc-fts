package config

import "github.com/urfave/cli/v2"

const (
	ConfigsFlag   = "configs"
	VarFlag       = "var"
	TargetFlag    = "target"
	VerboseFlag   = "verbose"
	DirectoryFlag = "directory"
)

var (
	ConfigsFlagSetup = &cli.StringFlag{
		Name:  "configs",
		Value: ".",
		Usage: "path to configs directory",
	}
	VarFlagSetup = &cli.StringSliceFlag{
		Name:  "var",
		Value: cli.NewStringSlice(),
		Usage: "user defined variables, format: key=value",
	}
	TargetFlagSetup = &cli.StringFlag{
		Name:  "target",
		Usage: "to run only specific test case, format: test_case_name",
	}
	VerboseFlagSetup = &cli.BoolFlag{
		Name:  "verbose",
		Usage: "verbose output",
	}
	DirectoryFlagSetup = &cli.StringFlag{
		Name:     "directory",
		Usage:    "directory to setup the project",
		Value:    ".",
		Required: true,
	}
)

type ContextWrapper struct {
	*cli.Context
}

func NewContextWrapper(ctx *cli.Context) ContextWrapper {
	return ContextWrapper{ctx}
}

func (ctx ContextWrapper) ConfigFlag() string {
	return ctx.String(ConfigsFlag)
}

func (ctx ContextWrapper) VarFlag() []string {
	return ctx.StringSlice(VarFlag)
}

func (ctx ContextWrapper) TargetFlag() string {
	return ctx.String(TargetFlag)
}

func (ctx ContextWrapper) VerboseFlag() bool {
	return ctx.Bool(VerboseFlag)
}

func (ctx ContextWrapper) DirectoryFlag() string {
	return ctx.String(DirectoryFlag)
}
