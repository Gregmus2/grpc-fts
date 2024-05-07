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
TestCaseLoop:
	for _, testCase := range r.testCases {
		if failed, dependency := failedTestCases.HasDependencyFailed(testCase.DependsOn); failed {
			r.logger.Infof("test case %s skipped due to failed dependency %s", testCase.Name, dependency)
			failedTestCases.Add(testCase.Name)

			continue
		}

		for i, step := range testCase.Steps {
			// todo add timeout option to request and apply it for stream and unary
			md, request, err := r.prepareRequest(step.Metadata, step.Service.Metadata, step.Request)
			if err != nil {
				return errors.Wrapf(err, "for step %d of test case %s", i, testCase.Name)
			}

			fullName := protoreflect.FullName(fmt.Sprintf("%s.%s", step.Service.Service, step.Method))
			client := r.clients.GetClient(step.ServiceName)
			response, err := client.Invoke(fullName, request, metadata.New(md))
			if err != nil {
				return errors.Wrapf(err, "error on calling service %s", step.ServiceName)
			}

			expectedResponse, err := r.prepareResponse(step.Response)
			if err != nil {
				return errors.Wrapf(err, "error on preparing expected response for step %d of test case %s", i, testCase.Name)
			}

			fails, err := r.check(step.Status, expectedResponse, response)
			if errors.Is(err, ErrValidationFailed) {
				failedTestCases.Add(testCase.Name)
				r.failed(fails, testCase.Name, i)

				break TestCaseLoop
			}
			if err != nil {
				return errors.Wrapf(err, "response validation error")
			}
		}

		r.logger.Infof("test case %s was finished successfully", testCase.Name)
	}

	return nil
}

func (r *runner) check(expectedStatus *config.Status, expectedResponse map[string]any, response *proto.GRPCResponse) ([]models.ValidationFail, error) {
	statusFails, err := r.checker.CheckStatus(response.Status, expectedStatus)
	if err != nil {
		return statusFails, err
	}

	if !response.IsStream {
		return r.checker.CheckResponse(response.Response, expectedResponse)
	}

	for i := 0; ; i++ {
		err := response.StreamReceive()
		if err != nil {
			return nil, errors.Wrap(err, "error on stream receiving")
		}

		statusFails, err := r.checker.CheckStatus(response.Status, expectedStatus)
		if err != nil {
			return statusFails, err
		}

		fails, err := r.checker.CheckResponse(response.Response, expectedResponse)
		if err != nil && !errors.Is(err, ErrValidationFailed) {
			return nil, errors.Wrapf(err, "error checking stream message #%d", i)
		}

		// successful exit
		if len(fails) == 0 {
			return nil, nil
		}
	}
}

func (r *runner) failed(fails []models.ValidationFail, testCase string, step int) {
	entry := r.logger
	for _, fail := range fails {
		entry = entry.WithFields(logrus.Fields{
			"field":    fail.Field,
			"function": fail.Function,
			"expected": fail.Expectation,
			"actual":   fail.ActualValue,
		})
	}
	entry.Warnf("test case %s, step %d finished with some fails", testCase, step)
}

func (r *runner) prepareRequest(stepMD, serviceMD config.Metadata, request json.RawMessage) (map[string]string, json.RawMessage, error) {
	err := r.variables.ReplaceMap(stepMD)
	if err != nil {
		return nil, nil, errors.Wrap(err, "metadata build error")
	}
	md := serviceMD.MergeWith(stepMD)

	req, err := r.variables.ReplaceInJson(request)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error on replacing variables in request")
	}

	return md, req, nil
}

func (r *runner) prepareResponse(response json.RawMessage) (map[string]any, error) {
	if len(response) == 0 {
		return nil, nil
	}

	rawResponse, err := r.variables.ReplaceInJson(response)
	if err != nil {
		return nil, errors.Wrap(err, "error on replacing variables in response")
	}

	var responseMap map[string]any
	err = json.Unmarshal(rawResponse, &responseMap)
	if err != nil {
		return nil, errors.Wrap(err, "error on unmarshalling response")
	}

	return responseMap, nil
}
