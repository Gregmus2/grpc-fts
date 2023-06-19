package config

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/models"
	"github.com/urfave/cli/v2"
	"os"
)

type Services map[string]Service

type Service struct {
	Address  string
	Service  string
	TLS      *models.TLS
	Metadata Metadata
}

func NewServices(ctx *cli.Context) (Services, error) {
	file, err := os.ReadFile(ctx.String("configs") + "/services.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "error reading service config")
	}

	var config Services
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing service config")
	}

	return config, nil
}
