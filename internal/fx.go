package internal

import (
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/logic"
	"github.com/res-am/grpc-fts/internal/proto"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

func buildDIContainer() []fx.Option {
	return []fx.Option{
		fx.Provide(
			config.NewServices,
			config.NewTestCases,
			config.NewGlobal,
			config.NewLogrusEntry,
			proto.NewDescriptorsManager,
			proto.NewClientsManager,
			logic.NewVariables,
			logic.NewResponseChecker,
			logic.NewRunner,
		),
		fx.Invoke(func(variables logic.Variables, services config.Services) error {
			return variables.ReplaceServicesMetadata(services)
		}),
	}
}

func injectContext(ctx *cli.Context) fx.Option {
	return fx.Provide(func() *cli.Context { return ctx })
}

func RunTestCase(ctx *cli.Context) error {
	options := buildDIContainer()
	options = append(
		options,
		injectContext(ctx),
		fx.Invoke(func(runner logic.Runner) error {
			return runner.RunTestCases()
		}),
	)

	return fx.New(options...).Start(ctx.Context)
}
