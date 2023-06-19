package internal

import (
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/logic"
	"github.com/res-am/grpc-fts/internal/proto"
	"github.com/urfave/cli/v2"
	"go.uber.org/dig"
)

type App struct {
	container *dig.Container
}

func NewApp(ctx *cli.Context) (*App, error) {
	a := &App{}
	c, err := a.buildContainer()
	if err != nil {
		return nil, err
	}

	if err := c.Provide(func() *cli.Context { return ctx }); err != nil {
		return nil, err
	}

	err = a.init(c)
	if err != nil {
		return nil, err
	}

	a.container = c

	return a, nil
}

func (a *App) RunTestCases() error {
	return a.container.Invoke(func(runner logic.Runner) error {
		return runner.RunTestCases()
	})
}

func (a *App) buildContainer() (*dig.Container, error) {
	c := dig.New()

	if err := c.Provide(logic.NewVariables); err != nil {
		return nil, err
	}
	if err := c.Provide(config.NewServices); err != nil {
		return nil, err
	}
	if err := c.Provide(config.NewTestCases); err != nil {
		return nil, err
	}
	if err := c.Provide(config.NewGlobal); err != nil {
		return nil, err
	}
	if err := c.Provide(config.NewLogrusEntry); err != nil {
		return nil, err
	}
	if err := c.Provide(proto.NewDescriptorsManager); err != nil {
		return nil, err
	}
	if err := c.Provide(proto.NewClientsManager); err != nil {
		return nil, err
	}
	if err := c.Provide(logic.NewResponseChecker); err != nil {
		return nil, err
	}
	if err := c.Provide(logic.NewRunner); err != nil {
		return nil, err
	}

	return c, nil
}

func (a *App) init(c *dig.Container) error {
	return c.Invoke(func(variables logic.Variables, services config.Services) error {
		return variables.ReplaceServicesMetadata(services)
	})
}
