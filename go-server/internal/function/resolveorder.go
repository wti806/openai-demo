package function

import "fmt"

type Order struct {
	Name string `json:"name,omitempty"`
}

type ResolveResp struct {
	Status string `json:"status,omitempty"`
}

func ResolveOrder(in interface{}) (interface{}, error) {
	input, ok := in.(Order)
	if !ok {
		return nil, fmt.Errorf("input is not of expected type")
	}
	fmt.Printf(input.Name)
	return ResolveResp{Status: "good"}, nil
}
