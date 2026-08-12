package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

func init() {
	os.Setenv("OPENAPI_FILEPATH", "/Users/artemkuznetsov/Projects/openapi2go/tests/test1/components.schemas.CustomerDetailCompanyResponseDto.json")
}

func main() {
	data, err := os.ReadFile(os.Getenv("OPENAPI_FILEPATH"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read openapi.json:", err)
		os.Exit(1)
	}

	var spec openapi.OpenAPI
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintln(os.Stderr, "failed to unmarshal openapi.json:", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", string(spec.JSON()))
}
