package logic

import (
	"bytes"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/urfave/cli/v2"
	"os"
	"regexp"
	"strings"
)

var ErrVariableNotFound = errors.New("variable not found")
var replacerRegExp = regexp.MustCompile(`\$\w+`)

type Variables map[string]string

func NewVariables(ctx *cli.Context) (Variables, error) {
	file, err := os.ReadFile(ctx.String("configs") + "/variables.yaml")
	if errors.Is(err, os.ErrNotExist) {
		return make(Variables), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "error reading service config")
	}

	variables := make(Variables)
	err = yaml.Unmarshal(file, &variables)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing service config")
	}

	for _, variable := range ctx.StringSlice("var") {
		kv := strings.SplitN(variable, "=", 2)
		variables[kv[0]] = kv[1]
	}

	return variables, nil
}

func (v Variables) ReplaceServicesMetadata(services config.Services) error {
	for key := range services {
		err := v.ReplaceMap(services[key].Metadata)
		if err != nil {
			return errors.Wrap(err, "error replacing services metadata")
		}
	}

	return nil
}

func (v Variables) Find(source string) (string, error) {
	placeholder, found := strings.CutPrefix(source, "$")
	if !found {
		return source, nil
	}

	variable, found := v[placeholder]
	if found {
		return variable, nil
	}

	return "", errors.Wrap(ErrVariableNotFound, variable)
}

func (v Variables) ReplaceRequest(source []byte) ([]byte, error) {
	matches := replacerRegExp.FindAll(source, -1)
	for _, match := range matches {
		variable, err := v.Find(string(match))
		if err != nil {
			return nil, err
		}

		source = bytes.ReplaceAll(source, match, []byte(variable))
	}

	return source, nil
}

func (v Variables) ReplaceMap(md map[string]string) error {
	for key, value := range md {
		replaced, err := v.Find(value)
		if err != nil {
			return err
		}

		md[key] = replaced
	}

	return nil
}
