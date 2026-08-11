package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

func main() {
	data, err := os.ReadFile("api/customer.post.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read openapi.json:", err)
		os.Exit(1)
	}

	var spec openapi.OpenAPI
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintln(os.Stderr, "failed to unmarshal openapi.json:", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", spec.Paths["/customer"].Post)
}
