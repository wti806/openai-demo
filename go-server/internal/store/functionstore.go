package store

import (
	"fmt"
	"reflect"
	"sync"

	fun "github.com/wti806/openai-demo/go-server/internal/function"
)

// FunctionStore holds a map of functions that can be executed by name.
type FunctionStore struct {
	funcMap sync.Map
}

// NewFunctionStore creates a new FunctionStore.
func NewFunctionStore() *FunctionStore {
	return &FunctionStore{
		funcMap: sync.Map{},
	}
}

// RegisterFunc registers a new function with the specified name.
func (s *FunctionStore) RegisterFunc(funcName string, function func(in interface{}) (interface{}, error), inputType, outputType reflect.Type) {
	s.funcMap.Store(funcName, fun.Function{
		Func:       function,
		InputType:  inputType,
		OutputType: outputType,
	})
}

// ExecFunc looks up a function by name and executes it with the provided JSON string as input.
func (s *FunctionStore) ExecFunc(funcName string, jsonStr string) (string, error) {
	value, ok := s.funcMap.Load(funcName)
	if !ok {
		return "", fmt.Errorf("function %q not found", funcName)
	}

	function, ok := value.(fun.Function)
	if !ok {
		return "", fmt.Errorf("stored value is not a Function")
	}

	return function.Run(jsonStr)
}
