package internal

import (
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/logic"
	"github.com/res-am/grpc-fts/internal/models"
	"github.com/res-am/grpc-fts/internal/proto"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

type Container struct {
	ctx *cli.Context
}

func NewContainer(ctx *cli.Context) Container {
	return Container{ctx: ctx}
}

func (c Container) RunTestCase() error {
	return c.runApp(
		fx.Invoke(
			func(runner logic.Runner) error {
				return runner.RunTestCases()
			},
			func(variables logic.Variables, services config.Services) error {
				return variables.ReplaceServicesMetadata(services)
			},
		),
	)
}

func (c Container) Validate() error {
	return c.runApp(
		fx.Invoke(func(validator *logic.Validator, testCases config.TestCases) error {
			return validator.Validate(testCases)
		}),
	)
}

func (c Container) buildDIContainer() fx.Option {
	return fx.Provide(
		config.NewServices,
		config.NewTestCases,
		config.NewGlobal,
		config.NewLogrusEntry,
		proto.NewDescriptorsManager,
		proto.NewClientsManager,
		logic.NewVariables,
		logic.NewResponseChecker,
		logic.NewRunner,
		logic.NewValidator,
		c.contextWrapper(c.ctx),
	)
}

func (c Container) contextWrapper(ctx *cli.Context) func() config.ContextWrapper {
	return func() config.ContextWrapper { return config.NewContextWrapper(ctx) }
}

func (c Container) runApp(invokes fx.Option) error {
	providers := c.buildDIContainer()
	err := fx.New(providers, invokes, fx.NopLogger).Start(c.ctx.Context)
	var userErr models.UserErr
	if !c.ctx.Bool("verbose") && errors.As(err, &userErr) {
		return userErr
	}

	return err
}
