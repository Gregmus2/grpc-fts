package config

import "github.com/urfave/cli/v2"

const (
	ConfigsFlag = "configs"
	VarFlag     = "var"
	TargetFlag  = "target"
	VerboseFlag = "verbose"
)

var ConfigsFlagSetup = &cli.StringFlag{
	Name:  "configs",
	Value: ".",
	Usage: "path to configs directory",
}
var VarFlagSetup = &cli.StringSliceFlag{
	Name:  "var",
	Value: cli.NewStringSlice(),
	Usage: "user defined variables, format: key=value",
}
var TargetFlagSetup = &cli.StringFlag{
	Name:  "target",
	Usage: "to run only specific test case, format: test_case_name",
}
var VerboseFlagSetup = &cli.BoolFlag{
	Name:  "verbose",
	Usage: "verbose output",
}

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
