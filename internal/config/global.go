package config

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"os"
)

type Global struct {
	ProtoRoot    string   `json:"proto_root"`
	ProtoImports []string `json:"proto_imports"`
	ProtoSources []string `json:"proto_sources"`
	Format       string   `json:"format"`
	Timestamp    bool     `json:"timestamp"`
}

func NewGlobal(ctx ContextWrapper) (*Global, error) {
	file, err := os.ReadFile(ctx.ConfigFlag() + "/global.yaml")
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
