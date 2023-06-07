package config

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
)

type Global struct {
	ProtoRoot    string   `json:"proto_root"`
	ProtoImports []string `json:"proto_imports"`
	ProtoSources []string `json:"proto_sources"`
}

func NewGlobal(ctx *cli.Context) (*Global, error) {
	file, err := os.ReadFile(ctx.String("configs") + "/global.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "error reading service config")
	}

	var config Global
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing service config")
	}

	return &config, nil
}
