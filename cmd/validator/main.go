package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	openapi3 "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

// ParseOpenAPI reads a local JSON file and parses it into the OpenAPI struct.
func ParseOpenAPI(filePath string) (*openapi3.OpenAPI, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var doc openapi3.OpenAPI
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("JSON syntax/schema mismatch: %w", err)
	}

	return &doc, nil
}

// ValidateOASObject runs standard checks against the parsed OpenAPI object.
// It reports structural violations according to the OpenAPI 3.0 specification.
func ValidateOASObject(doc *openapi3.OpenAPI) []string {
	var errors []string

	// 1. Validate OpenAPI Version
	if doc.OpenAPI == "" {
		errors = append(errors, "REQUIRED root field 'openapi' is missing or empty")
	} else if !strings.HasPrefix(doc.OpenAPI, "3.0.") {
		errors = append(errors, fmt.Sprintf("Unsupported OpenAPI version %q. This validator supports version 3.0.x", doc.OpenAPI))
	}

	// 2. Validate Info Object
	if doc.Info.Title == "" {
		errors = append(errors, "REQUIRED field 'info.title' is missing or empty")
	}
	if doc.Info.Version == "" {
		errors = append(errors, "REQUIRED field 'info.version' is missing or empty")
	}

	// 3. Validate Paths Object
	if doc.Paths == nil {
		errors = append(errors, "REQUIRED root field 'paths' is missing")
	} else {
		for path := range doc.Paths {
			if !strings.HasPrefix(path, "/") {
				errors = append(errors, fmt.Sprintf("Paths map key %q is invalid: must begin with a forward slash '/'", path))
			}
		}
	}

	return errors
}

func main() {
	// Setup command-line flags
	fileFlag := flag.String("file", "", "Path to the OpenAPI JSON file to validate")
	verboseFlag := flag.Bool("verbose", false, "Print verbose details about paths, operations, and schemas")
	flag.Parse()

	// If no file flag is provided, fall back to the first positional argument
	filePath := *fileFlag
	if filePath == "" && len(flag.Args()) > 0 {
		filePath = flag.Arg(0)
	}

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: missing input file path.")
		fmt.Println("Usage:")
		fmt.Println("  go run openapi_cli.go -file <path-to-openapi.json>")
		fmt.Println("  go run openapi_cli.go <path-to-openapi.json>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Analyzing OpenAPI Document: %s...\n", filePath)

	doc, err := ParseOpenAPI(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[FATAL ERROR] Parsing failed:\n  %v\n", err)
		os.Exit(1)
	}

	// Run structural checks
	validationErrors := ValidateOASObject(doc)

	fmt.Println("\n==========================================")
	fmt.Println("         OpenAPI Validation Report        ")
	fmt.Println("==========================================")

	if len(validationErrors) > 0 {
		fmt.Printf("❌ Validation Failed with %d structural error(s):\n", len(validationErrors))
		for i, vErr := range validationErrors {
			fmt.Printf("  %d. %s\n", i+1, vErr)
		}
		fmt.Println("\nResult: INVALID document according to OAS 3.0 rules.")
		os.Exit(1)
	}

	fmt.Println("✅ Validation Succeeded! The document meets basic OAS 3.0 structural rules.")

	// Output API Summary Statistics
	fmt.Println("\nAPI Metadata Summary:")
	fmt.Printf("  - Title:       %s\n", doc.Info.Title)
	fmt.Printf("  - Version:     %s\n", doc.Info.Version)
	fmt.Printf("  - OAS Version: %s\n", doc.OpenAPI)
	if doc.Info.Description != "" {
		desc := doc.Info.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		fmt.Printf("  - Description: %s\n", desc)
	}

	// Count paths and operations
	pathCount := len(doc.Paths)
	opCount := 0
	for _, pathItem := range doc.Paths {
		if pathItem == nil {
			continue
		}
		if pathItem.Get != nil {
			opCount++
		}
		if pathItem.Put != nil {
			opCount++
		}
		if pathItem.Post != nil {
			opCount++
		}
		if pathItem.Delete != nil {
			opCount++
		}
		if pathItem.Options != nil {
			opCount++
		}
		if pathItem.Head != nil {
			opCount++
		}
		if pathItem.Patch != nil {
			opCount++
		}
		if pathItem.Trace != nil {
			opCount++
		}
	}

	fmt.Println("\nAPI Endpoint Metrics:")
	fmt.Printf("  - Total Endpoints:  %d\n", pathCount)
	fmt.Printf("  - Total Operations: %d\n", opCount)

	// Count components
	if doc.Components != nil {
		fmt.Println("\nReusable Components:")
		fmt.Printf("  - Schemas:          %d\n", len(doc.Components.Schemas))
		fmt.Printf("  - Responses:        %d\n", len(doc.Components.Responses))
		fmt.Printf("  - Parameters:       %d\n", len(doc.Components.Parameters))
	}

	// Print detailed structural walkthrough if requested
	if *verboseFlag {
		fmt.Println("\n==========================================")
		fmt.Println("       Detailed Structural Walkthrough    ")
		fmt.Println("==========================================")
		for path, pathItem := range doc.Paths {
			fmt.Printf("\nPath: %s\n", path)
			if pathItem == nil {
				continue
			}
			if pathItem.Summary != "" {
				fmt.Printf("  Summary: %s\n", pathItem.Summary)
			}
			printOps(pathItem)
		}
	}
}

func printOps(pi *openapi3.PathItem) {
	if pi.Get != nil {
		printOpDetails("GET", pi.Get)
	}
	if pi.Post != nil {
		printOpDetails("POST", pi.Post)
	}
	if pi.Put != nil {
		printOpDetails("PUT", pi.Put)
	}
	if pi.Delete != nil {
		printOpDetails("DELETE", pi.Delete)
	}
	if pi.Patch != nil {
		printOpDetails("PATCH", pi.Patch)
	}
}

func printOpDetails(method string, op *openapi3.Operation) {
	fmt.Printf("  - %s: %s (ID: %s)\n", method, op.Summary, op.OperationID)
	if len(op.Tags) > 0 {
		fmt.Printf("    Tags: [%s]\n", strings.Join(op.Tags, ", "))
	}
}
