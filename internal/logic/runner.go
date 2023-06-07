package logic

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/models"
	"github.com/res-am/grpc-fts/internal/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type runner struct {
	testCases config.TestCases
	clients   proto.ClientsManager
	logger    *logrus.Entry
	checker   ResponseChecker
	variables Variables
}

func NewRunner(testCases config.TestCases, clients proto.ClientsManager, logger *logrus.Entry, validator ResponseChecker, variables Variables) Runner {
	return &runner{testCases: testCases, clients: clients, logger: logger, checker: validator, variables: variables}
}

func (r *runner) RunTestCases() (err error) {
	failedTestCases := make(failedDependencies)
	for _, testCase := range r.testCases {
		if failed, dependency := failedTestCases.HasDependencyFailed(testCase.DependsOn); failed {
			r.logger.Infof("test case %s skipped due to failed dependency %s", testCase.Name, dependency)

			continue
		}

		for i, step := range testCase.Steps {
			md, request, err := r.prepareRequest(step.Metadata, step.Service.Metadata, step.Request)
			if err != nil {
				return errors.Wrapf(err, "for step %d of test case %s", i, testCase.Name)
			}

			fullName := protoreflect.FullName(fmt.Sprintf("%s.%s", step.Service.Service, step.Method))
			client := r.clients.GetClient(step.ServiceName)
			response, stat, err := client.Invoke(fullName, request, metadata.New(md))
			if err != nil {
				return errors.Wrapf(err, "error on calling service %s", step.ServiceName)
			}

			fails, err := r.check(step, stat, response)
			if errors.Is(err, ErrValidationFailed) {
				failedTestCases.Add(testCase.Name)
				r.failed(fails, testCase.Name, i)

				break
			}
			if err != nil {
				return errors.Wrapf(err, "response validation error")
			}
		}

		r.logger.Infof("test case %s was finished successfully", testCase.Name)
	}

	return nil
}

func (r *runner) check(step config.Step, status *status.Status, response map[string]interface{}) ([]models.ValidationFail, error) {
	statusFails, err := r.checker.CheckStatus(status, step.Status)
	if err != nil {
		return statusFails, err
	}

	return r.checker.CheckResponse(response, step.Response)
}

func (r *runner) failed(fails []models.ValidationFail, testCase string, step int) {
	r.logger.Warnf("test case %s, step %d finished with some fails:", testCase, step)
	for _, fail := range fails {
		r.logger.Warnf("Field: %s, Function: %s", fail.Field, fail.Function)
		r.logger.Warnf("Expected: %v, Actual: %s", fail.Expectation, fail.ActualValue)
	}
}

func (r *runner) prepareRequest(stepMD, serviceMD config.Metadata, request json.RawMessage) (map[string]string, json.RawMessage, error) {
	err := r.variables.ReplaceMap(stepMD)
	if err != nil {
		return nil, nil, errors.Wrap(err, "metadata build error")
	}
	md := serviceMD.MergeWith(stepMD)

	req, err := r.variables.ReplaceRequest(request)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error on replacing variables in request")
	}

	return md, req, nil
}