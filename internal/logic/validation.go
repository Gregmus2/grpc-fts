package logic

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type validator struct {
	clientsManager proto.ClientsManager
	manager        proto.DescriptorsManager
	checker        ResponseChecker
}

func NewValidator(clientsManager proto.ClientsManager, manager proto.DescriptorsManager, checker ResponseChecker) Validator {
	return &validator{
		clientsManager: clientsManager,
		manager:        manager,
		checker:        checker,
	}
}

func (v validator) Validate(testCases config.TestCases) error {
	for _, testCase := range testCases {
		if err := v.validateTestCase(testCase); err != nil {
			return errors.Wrapf(err, "test case %s", testCase.Name)
		}
	}

	return nil
}

func (v validator) validateTestCase(testCase config.TestCase) error {
	for i, step := range testCase.Steps {
		if err := v.validateStep(step); err != nil {
			return errors.Wrapf(err, "step %d", i+1)
		}
	}

	return nil
}

func (v validator) validateStep(step config.Step) error {
	fullName := step.BuildProtoFullName()
	descriptor := v.manager.GetDescriptor(fullName)

	if err := v.validateRequest(step.ServiceName, descriptor.Input(), step.Request); err != nil {
		return errors.Wrap(err, "request")
	}

	if err := v.validateResponse(descriptor.Output().Fields(), step.Response); err != nil {
		return errors.Wrap(err, "response")
	}

	return nil
}

func (v validator) validateRequest(service string, input protoreflect.MessageDescriptor, request json.RawMessage) error {
	_, err := v.clientsManager.GetClient(service).BuildRequest(input, request)
	if err != nil {
		return err
	}

	return nil
}

func (v validator) validateResponse(fields protoreflect.FieldDescriptors, response json.RawMessage) error {
	if len(response) == 0 {
		return nil
	}

	var responseMap map[string]any
	err := json.Unmarshal(response, &responseMap)
	if err != nil {
		return errors.Wrap(err, "error on unmarshalling response")
	}

	return newResponseValidator(v.checker.FunctionExists).validate(fields, responseMap)
}

type responseValidator struct {
	check func(function string) bool
}

func newResponseValidator(check func(function string) bool) responseValidator {
	return responseValidator{check: check}
}

func (v responseValidator) validate(fields protoreflect.FieldDescriptors, response map[string]any) error {
	for key, value := range response {
		// todo allow to specify snake case
		field := fields.ByJSONName(key)
		if field == nil && !v.check(key) {
			return fmt.Errorf("unexpected key %s", key)
		}
		if field != nil && field.Kind() == protoreflect.MessageKind {
			fields = field.Message().Fields()
		}

		return v.validateValue(fields, value)
	}

	return nil
}

func (v responseValidator) validateValue(fields protoreflect.FieldDescriptors, value any) error {
	switch t := value.(type) {
	case map[string]any:
		return v.validate(fields, t)
	case []any:
		for _, item := range t {
			err := v.validateValue(fields, item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
