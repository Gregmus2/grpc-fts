package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/models"
	"os"
)

type Services map[string]Service

type Service struct {
	Address  string
	Service  string
	TLS      *models.TLS
	Metadata Metadata
}

func NewServices(ctx ContextWrapper) (Services, error) {
	file, err := os.ReadFile(ctx.ConfigFlag() + "/services.yaml")
	if errors.Is(err, os.ErrNotExist) {
		return nil, models.NewErr("services.yaml file not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "error reading service config")
	}

	var config Services
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, models.NewErr(fmt.Sprintf("error parsing service config: %s", err.Error()))
	}

	return config, nil
}
