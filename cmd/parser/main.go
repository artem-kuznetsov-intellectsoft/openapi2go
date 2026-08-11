package main

import (
	"encoding/json"
	"fmt"
	"os"

	openapi3 "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

// ParseOpenAPIFile reads a local JSON or YAML file (converted to JSON)
// and unmarshals it into our strongly-typed OpenAPI root struct.
func ParseOpenAPIFile(filePath string) (*openapi3.OpenAPI, error) {
	// Read the file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON bytes directly into the OpenAPI structure
	var doc openapi3.OpenAPI
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &doc, nil
}

func main() {
	filePath := "boilerplate.json"
	fmt.Printf("Loading OpenAPI Document from: %s...\n", filePath)

	// Parse the file
	doc, err := ParseOpenAPIFile(filePath)
	if err != nil {
		fmt.Printf("Error during parsing: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSuccessfully parsed OpenAPI Document!")
	fmt.Printf("OAS Version: %s\n", doc.OpenAPI)
	fmt.Printf("API Title:   %s\n", doc.Info.Title)
	fmt.Printf("API Version: %s\n", doc.Info.Version)
	fmt.Printf("Description: %s\n", doc.Info.Description)

	// Display configured API target environments (servers)
	fmt.Println("\n--- Servers ---")
	for i, server := range doc.Servers {
		fmt.Printf("[%d] URL: %s (%s)\n", i+1, server.URL, server.Description)
	}

	// Iterate over available paths and their documented operations
	fmt.Println("\n--- Paths/Endpoints ---")
	for path, item := range doc.Paths {
		fmt.Printf("Endpoint: %s\n", path)

		// GET Operation details
		if item.Get != nil {
			fmt.Printf("  - GET: %s (OperationID: %s)\n", item.Get.Summary, item.Get.OperationID)
			for code, respRef := range item.Get.Responses {
				if respRef.Ref != "" {
					fmt.Printf("    * Response [%s]: Referenced schema: %s\n", code, respRef.Ref)
				} else if respRef.Value != nil {
					fmt.Printf("    * Response [%s]: %s\n", code, respRef.Value.Description)
				}
			}
		}

		// POST Operation details
		if item.Post != nil {
			fmt.Printf("  - POST: %s (OperationID: %s)\n", item.Post.Summary, item.Post.OperationID)
			if item.Post.RequestBody != nil {
				if item.Post.RequestBody.Ref != "" {
					fmt.Printf("    * Request Body: Referenced schema: %s\n", item.Post.RequestBody.Ref)
				} else if item.Post.RequestBody.Value != nil {
					fmt.Printf("    * Request Body Required: %t\n", item.Post.RequestBody.Value.Required)
				}
			}
		}
	}

	// Print reusable schemas defined in components
	fmt.Println("\n--- Components (Reusable Schemas) ---")
	if doc.Components != nil && doc.Components.Schemas != nil {
		for name, schemaRef := range doc.Components.Schemas {
			if schemaRef.Ref != "" {
				fmt.Printf("  - Schema [%s]: Reference to %s\n", name, schemaRef.Ref)
			} else if schemaRef.Value != nil {
				fmt.Printf("  - Schema [%s]: type=%s, required properties=%v\n", name, schemaRef.Value.Type, schemaRef.Value.Required)
			}
		}
	}
}
