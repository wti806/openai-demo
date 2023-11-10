package function

import (
	"encoding/json"
	"reflect"
)

// Function is a struct that holds a function along with the reflection types of its input and output.
type Function struct {
	Func       func(in interface{}) (interface{}, error)
	InputType  reflect.Type
	OutputType reflect.Type
}

// Run executes the stored function using the provided JSON string as input.
func (f *Function) Run(inputStr string) (string, error) {
	inputPtr := reflect.New(f.InputType).Interface()

	// Convert JSON string to Input struct
	err := json.Unmarshal([]byte(inputStr), &inputPtr)
	if err != nil {
		return "", err
	}

	// Call the function with the provided input and get the output
	output, err := f.Func(reflect.ValueOf(inputPtr).Elem().Interface())
	if err != nil {
		return "", err
	}

	// Marshal the output to JSON
	outputBytes, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(outputBytes), nil
}
