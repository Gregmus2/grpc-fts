package config

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

type TestCases []TestCase

type TestCase struct {
	Steps     []Step
	DependsOn []string `json:"depends_on"`
	Name      string
}

type Function string

// todo configuration validation, especially service reference and method existence
type Step struct {
	ServiceName string `json:"service"`
	Method      string
	Request     json.RawMessage
	Response    map[string]interface{}
	Status      *Status
	Metadata    Metadata
	Store       map[string]interface{}
	Stream      bool
	Service     Service `json:"-"`
}

type Status struct {
	Code    *string
	Message *string
}

func NewTestCases(ctx *cli.Context, logger *logrus.Entry, services Services) (TestCases, error) {
	files, err := os.ReadDir(ctx.String("configs") + "/test-cases")
	if err != nil {
		logger.Errorf("error on reading test-cases dir: %s", err)

		return nil, errors.Wrap(err, "error on reading test-cases dir")
	}

	testCases := make(TestCases, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := ctx.String("configs") + "/test-cases/" + file.Name()
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading %s", filePath)
		}

		var testCase TestCase
		err = yaml.Unmarshal(content, &testCase)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing %s", filePath)
		}

		for i := range testCase.Steps {
			testCase.Steps[i].Service = services[testCase.Steps[i].ServiceName]
		}

		if testCase.Name == "" {
			fileName := file.Name()
			testCase.Name = fileName[:len(fileName)-len(filepath.Ext(fileName))]
		}
		testCases = append(testCases, testCase)
	}

	testCases, err = Sort(testCases)
	if err != nil {
		return nil, errors.Wrapf(err, "test case dependency error")
	}

	return testCases, nil
}
