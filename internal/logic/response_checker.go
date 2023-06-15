package logic

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"strings"
)

var ErrValidationFailed = errors.New("validation failed")
var statusOk = codes.OK.String()

type function struct {
	action         func(expectation any, val reflect.Value) (bool, error)
	supportedTypes []reflect.Kind
}

type responseChecker struct {
	functions map[string]function
	variables Variables
}

func NewResponseChecker(variables Variables) ResponseChecker {
	validator := &responseChecker{variables: variables}
	numericTypes := []reflect.Kind{
		reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8,
		reflect.Int16, reflect.Int32, reflect.Int64,
	}
	scalarTypes := append(numericTypes, reflect.String) //nolint:gocritic

	validator.functions = map[string]function{
		"len": {
			action:         validator.lenCheck,
			supportedTypes: []reflect.Kind{reflect.Slice},
		},
		"gt": {
			action:         validator.gtCheck,
			supportedTypes: numericTypes,
		},
		"gte": {
			action:         validator.gteCheck,
			supportedTypes: numericTypes,
		},
		"lt": {
			action:         validator.ltCheck,
			supportedTypes: numericTypes,
		},
		"lte": {
			action:         validator.lteCheck,
			supportedTypes: numericTypes,
		},
		"one_of": {
			action:         validator.oneOfCheck,
			supportedTypes: append(scalarTypes, reflect.Map),
		},
		"any": {
			action:         validator.anyCheck,
			supportedTypes: []reflect.Kind{reflect.Slice},
		},
		"first": {
			action:         validator.firstCheck,
			supportedTypes: []reflect.Kind{reflect.Slice},
		},
		"all": {
			action:         validator.allCheck,
			supportedTypes: []reflect.Kind{reflect.Slice},
		},
		"store": {
			action:         validator.store,
			supportedTypes: scalarTypes,
		},
	}

	return validator
}

func (c *responseChecker) executeFunction(function string, expectation any, val reflect.Value) (bool, error) {
	model, ok := c.functions[function]
	if !ok {
		return false, fmt.Errorf("function %s is not exist", function)
	}

	for _, kind := range model.supportedTypes {
		if kind == val.Kind() {
			return model.action(expectation, val)
		}
	}

	return false, fmt.Errorf("unsupported type %s for function %s", val.Kind(), function)
}

func (c *responseChecker) CheckResponse(response map[string]interface{}, expectations map[string]any) ([]models.ValidationFail, error) {
	return c.checkObject("", expectations, reflect.ValueOf(response))
}

func (c *responseChecker) CheckStatus(actual *status.Status, expectation *config.Status) ([]models.ValidationFail, error) {
	if actual == nil && expectation == nil {
		return nil, nil
	}
	if expectation == nil {
		expectation = &config.Status{Code: &statusOk, Message: new(string)}
	}

	fails := make([]models.ValidationFail, 0)
	if expectation.Code != nil && !strings.EqualFold(*expectation.Code, actual.Code().String()) {
		fails = append(fails, models.Fail("response.status.code", "", *expectation.Code, actual.Code().String()))
	}
	if expectation.Message != nil && *expectation.Message != actual.Message() {
		fails = append(fails, models.Fail("response.status.message", "", *expectation.Message, actual.Message()))
	}
	if len(fails) > 0 {
		return fails, ErrValidationFailed
	}

	return fails, nil
}

func (c *responseChecker) checkObject(path string, expectations map[string]any, object reflect.Value) ([]models.ValidationFail, error) {
	if object.Kind() == reflect.Interface {
		object = object.Elem()
	}
	if !object.IsValid() {
		return []models.ValidationFail{
			models.Fail(path, "", expectations, "nil"),
		}, ErrValidationFailed
	}

	fails := make([]models.ValidationFail, 0)
	for field, expectation := range expectations {
		path = path + "." + field
		if _, ok := c.functions[field]; ok {
			fail, err := c.checkFunction(path, field, expectation, object)
			if errors.Is(err, ErrValidationFailed) {
				fails = append(fails, fail)

				continue
			}
			if err != nil {
				return nil, errors.Wrapf(err, "error validating %s", path)
			}

			continue
		}

		val := ExtractValueByField(object, field)
		if !val.IsValid() {
			return nil, fmt.Errorf("field %s is not function, neither field", field)
		}
		val = val.Elem()
		embeddedFails, err := c.checkValue(path, expectation, val)
		if err != nil {
			if errors.Is(err, ErrValidationFailed) {
				fails = append(fails, embeddedFails...)

				continue
			}

			return nil, errors.Wrapf(err, "error validating %s", path)
		}
	}

	if len(fails) > 0 {
		return fails, ErrValidationFailed
	}

	return nil, nil
}

func ExtractValueByField(object reflect.Value, key string) reflect.Value {
	switch object.Kind() { //nolint:exhaustive
	case reflect.Map:
		return object.MapIndex(reflect.ValueOf(key))
	default:
		return reflect.Value{}
	}
}

//nolint:godox
func (c *responseChecker) checkValue(path string, expectation any, val reflect.Value) ([]models.ValidationFail, error) {
	// todo cover proto enums
	condition, isEmbeddedCondition := expectation.(map[string]any)
	switch {
	case isSlice(val) && !isEmbeddedCondition:
		return c.checkSlice(path, expectation, val)
	case isEmbeddedCondition:
		return c.checkObject(path, condition, val)
	case !isEmbeddedCondition:
		return c.checkScalar(path, expectation, val)
	default:
		return nil, fmt.Errorf("unexpected case. expectation: %+v, value: %+v", expectation, val)
	}
}

func (c *responseChecker) checkScalar(path string, expectation any, val reflect.Value) ([]models.ValidationFail, error) {
	isValid, err := c.equalCheck(expectation, val)
	if err != nil {
		return nil, errors.Wrapf(err, "exact check validation error for field %s", path)
	}

	if !isValid {
		return []models.ValidationFail{
			models.Fail(path, "equal", expectation, fmt.Sprintf("%v", val.Interface())),
		}, ErrValidationFailed
	}

	return nil, nil
}

func (c *responseChecker) checkFunction(path, function string, expectation any, val reflect.Value) (models.ValidationFail, error) {
	isValid, err := c.executeFunction(function, expectation, val)
	if err != nil {
		return models.ValidationFail{}, errors.Wrapf(err, "error on validate function %s for field %s", function, path)
	}

	if !isValid {
		return models.Fail(
			path, function, expectation, fmt.Sprintf("%v", val),
		), ErrValidationFailed
	}

	return models.ValidationFail{}, nil
}

// checkSlice checks that all elements of the slice satisfy expectations, makes all to all comparison to
// provide an order independent check
func (c *responseChecker) checkSlice(path string, expectation any, val reflect.Value) ([]models.ValidationFail, error) {
	expectedSlice, ok := expectation.([]any)
	if !ok {
		return nil, fmt.Errorf("expected value is not a slice")
	}

	if len(expectedSlice) != val.Len() {
		return nil, fmt.Errorf("expected slice length %d is not equal to actual slice length %d", len(expectedSlice), val.Len())
	}

	usedIndexes := make(map[int]struct{})
	for i := 0; i < len(expectedSlice); i++ {
		for j := 0; j < len(expectedSlice); j++ {
			if _, ok := usedIndexes[j]; ok {
				continue
			}

			embeddedFails, err := c.checkValue(path, expectedSlice[i], val.Index(j))
			if err != nil && !errors.Is(err, ErrValidationFailed) {
				return nil, errors.Wrapf(err, "error validating %s", path)
			}

			if len(embeddedFails) == 0 {
				usedIndexes[j] = struct{}{}

				break
			}
		}
	}

	if len(usedIndexes) != len(expectedSlice) {
		return []models.ValidationFail{
			models.Fail(path, "slice", expectedSlice, stringifyReflectedSlice(val)),
		}, ErrValidationFailed
	}

	return nil, nil
}

func (c *responseChecker) lenCheck(expectation any, val reflect.Value) (bool, error) {
	switch t := expectation.(type) {
	case float64:
		return val.Len() == int(t), nil
	case map[string]any: // in case of embedded function (len: gte: 10)
		if len(t) == 0 {
			return false, errors.New("no expectations")
		}

		for function, expectation := range t {
			isValid, err := c.executeFunction(function, expectation, reflect.ValueOf(float64(val.Len())))
			if err != nil {
				return false, err
			}
			if !isValid {
				return false, nil
			}
		}

		return true, nil
	default:
		return false, fmt.Errorf("unsupported type %s with value %c", t, expectation)
	}
}

func (c *responseChecker) float64Check(check func(float64) bool, val reflect.Value) (bool, error) {
	var f float64
	switch val.Kind() { //nolint:exhaustive
	case reflect.Float32, reflect.Float64:
		f = val.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f = float64(val.Int())
	default:
		return false, fmt.Errorf("numeric type was expected, got %s", val.Kind())
	}

	return check(f), nil
}

func (c *responseChecker) gtCheck(expectation any, val reflect.Value) (bool, error) {
	return c.float64Check(func(f float64) bool {
		return f > expectation.(float64)
	}, val)
}

func (c *responseChecker) gteCheck(expectation any, val reflect.Value) (bool, error) {
	return c.float64Check(func(f float64) bool {
		return f >= expectation.(float64)
	}, val)
}

func (c *responseChecker) ltCheck(expectation any, val reflect.Value) (bool, error) {
	return c.float64Check(func(f float64) bool {
		return f < expectation.(float64)
	}, val)
}

func (c *responseChecker) lteCheck(expectation any, val reflect.Value) (bool, error) {
	return c.float64Check(func(f float64) bool {
		return f <= expectation.(float64)
	}, val)
}

func (c *responseChecker) oneOfCheck(expectation any, val reflect.Value) (bool, error) {
	values, ok := expectation.([]any)
	if !ok {
		return false, errors.New("array was expected")
	}

	for _, value := range values {
		if fails, err := c.checkValue("", value, val); err == nil && len(fails) == 0 {
			return true, nil
		}
	}

	return false, nil
}

func (c *responseChecker) anyCheck(expectation any, val reflect.Value) (bool, error) {
	for i := 0; i < val.Len(); i++ {
		fails, err := c.checkValue("", expectation, val.Index(i).Elem())
		if err != nil && !errors.Is(err, ErrValidationFailed) {
			return false, errors.Wrapf(err, "error checking `any`, index %d", i)
		}

		if len(fails) == 0 {
			return true, nil
		}
	}

	return false, nil
}

func (c *responseChecker) firstCheck(expectation any, val reflect.Value) (bool, error) {
	if val.Len() == 0 {
		return false, errors.New("array is empty")
	}

	fails, err := c.checkValue("", expectation, val.Index(0).Elem())
	if err != nil && !errors.Is(err, ErrValidationFailed) {
		return false, errors.Wrap(err, "error checking `first`")
	}

	if len(fails) == 0 {
		return true, nil
	}

	return false, nil
}

func (c *responseChecker) allCheck(expectation any, val reflect.Value) (bool, error) {
	for i := 0; i < val.Len(); i++ {
		fails, err := c.checkValue("", expectation, val.Index(i).Elem())
		if err != nil && !errors.Is(err, ErrValidationFailed) {
			return false, errors.Wrapf(err, "error checking `any`, index %d", i)
		}

		if len(fails) > 0 {
			return false, nil
		}
	}

	return true, nil
}

func (c *responseChecker) store(variable any, val reflect.Value) (bool, error) {
	variableName, ok := variable.(string)
	if !ok {
		return false, errors.New("variable name was expected")
	}

	c.variables[variableName] = fmt.Sprintf("%v", val.Interface())

	return true, nil
}

func (c *responseChecker) equalCheck(expectation any, val reflect.Value) (bool, error) {
	switch val.Kind() { //nolint:exhaustive
	case reflect.Float32, reflect.Float64:
		return expectation.(float64) == val.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return expectation.(float64) == float64(val.Int()), nil
	case reflect.String:
		return expectation.(string) == val.String(), nil
	case reflect.Bool:
		return expectation.(bool) == val.Bool(), nil
	default:
		return false, fmt.Errorf("unsopported type %s", val.Kind())
	}
}

func isSlice(val reflect.Value) bool {
	return val.Kind() == reflect.Slice
}

func stringifyReflectedSlice(slice reflect.Value) string {
	fragments := make([]string, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		fragments[i] = fmt.Sprintf("%v", slice.Index(i).Interface())
	}

	return strings.Join(fragments, ", ")
}
