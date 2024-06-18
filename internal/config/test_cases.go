package config

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/models"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"
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

type Step struct {
	ServiceName string `json:"service"`
	Method      string
	Request     json.RawMessage
	Response    json.RawMessage
	Status      *Status
	Metadata    Metadata
	Store       map[string]interface{}
	Stream      bool
	Service     Service `json:"-"`
}

func (s Step) BuildProtoFullName() protoreflect.FullName {
	return protoreflect.FullName(fmt.Sprintf("%s.%s", s.Service.Service, s.Method))
}

type Status struct {
	Code    *string
	Message *string
}

func NewTestCases(ctx ContextWrapper, logger *logrus.Entry, services Services) (TestCases, error) {
	files, err := os.ReadDir(ctx.ConfigFlag() + "/test-cases")
	if err != nil {
		logger.Errorf("error on reading test-cases dir: %s", err)

		return nil, errors.Wrap(err, "error on reading test-cases dir")
	}

	testCases, err := collectTestCases(ctx.ConfigFlag(), files, services)
	if err != nil {
		return nil, err
	}

	if ctx.TargetFlag() != "" {
		testCases, err = testCases.Filter(ctx.TargetFlag())
		if err != nil {
			return nil, errors.Wrapf(err, "error filtering test cases")
		}
	}

	testCases, err = Sort(testCases)
	if err != nil {
		return nil, errors.Wrapf(err, "test case dependency error")
	}

	return testCases, nil
}

func collectTestCases(configDir string, files []os.DirEntry, services Services) (TestCases, error) {
	testCases := make(TestCases, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := configDir + "/test-cases/" + file.Name()
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
			service, exists := services[testCase.Steps[i].ServiceName]
			if !exists {
				return nil, models.NewErr("service '" + testCase.Steps[i].ServiceName + "' not found")
			}

			testCase.Steps[i].Service = service
		}

		if testCase.Name == "" {
			fileName := file.Name()
			testCase.Name = fileName[:len(fileName)-len(filepath.Ext(fileName))]
		}
		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

func (t TestCases) Filter(target string) (TestCases, error) {
	var targetTestCase TestCase
	for _, testCase := range t {
		if testCase.Name == target {
			targetTestCase = testCase

			break
		}
	}

	if targetTestCase.Name == "" {
		return nil, errors.New("target test case not found")
	}

	return append(t.CollectDependencies(targetTestCase), targetTestCase), nil
}

func (t TestCases) CollectDependencies(target TestCase) TestCases {
	result := make(TestCases, 0)
	for _, testCase := range t {
		for _, dependsOn := range target.DependsOn {
			if dependsOn == testCase.Name {
				result = append(result, testCase)
				result = append(result, t.CollectDependencies(testCase)...)

				break
			}
		}
	}

	return result
}
