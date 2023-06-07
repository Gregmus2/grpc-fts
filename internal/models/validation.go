package models

type ValidationFail struct {
	Field       string
	Function    string
	Expectation interface{}
	ActualValue string
}

func Fail(field, function string, expectation interface{}, actualValue string) ValidationFail {
	return ValidationFail{
		Field:       field,
		Function:    function,
		Expectation: expectation,
		ActualValue: actualValue,
	}
}
