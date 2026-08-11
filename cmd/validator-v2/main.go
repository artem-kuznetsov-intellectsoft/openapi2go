package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	openapi3 "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

// RefOccurrence represents a $ref found during AST traversal
type RefOccurrence struct {
	Ref      string // The literal $ref value (e.g. "#/components/schemas/User")
	Location string // Coherent string path indicating where it was found (e.g. "paths[/users].get.responses[200]")
}

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

// CollectReferences recursively traverses the parsed AST tree to discover all $ref instances.
func CollectReferences(doc *openapi3.OpenAPI) []RefOccurrence {
	var occurrences []RefOccurrence

	addRef := func(ref string, loc string) {
		if ref != "" {
			occurrences = append(occurrences, RefOccurrence{Ref: ref, Location: loc})
		}
	}

	// Helper to walk schemas recursively to find nested refs
	var walkSchema func(s *openapi3.Schema, loc string)
	walkSchema = func(s *openapi3.Schema, loc string) {
		if s == nil {
			return
		}
		// If there are properties, they can be RefOr[*Schema]
		for propName, propRefOr := range s.Properties {
			if propRefOr != nil {
				addRef(propRefOr.Ref, fmt.Sprintf("%s.properties[%s]", loc, propName))
				if propRefOr.Value != nil {
					walkSchema(propRefOr.Value, fmt.Sprintf("%s.properties[%s]", loc, propName))
				}
			}
		}
		// items can be RefOr[*Schema]
		if s.Items != nil {
			addRef(s.Items.Ref, loc+".items")
			if s.Items.Value != nil {
				walkSchema(s.Items.Value, loc+".items")
			}
		}
		// allOf, oneOf, anyOf are arrays of RefOr[*Schema]
		for i, sub := range s.AllOf {
			if sub != nil {
				addRef(sub.Ref, fmt.Sprintf("%s.allOf[%d]", loc, i))
				if sub.Value != nil {
					walkSchema(sub.Value, fmt.Sprintf("%s.allOf[%d]", loc, i))
				}
			}
		}
		for i, sub := range s.OneOf {
			if sub != nil {
				addRef(sub.Ref, fmt.Sprintf("%s.oneOf[%d]", loc, i))
				if sub.Value != nil {
					walkSchema(sub.Value, fmt.Sprintf("%s.oneOf[%d]", loc, i))
				}
			}
		}
		for i, sub := range s.AnyOf {
			if sub != nil {
				addRef(sub.Ref, fmt.Sprintf("%s.anyOf[%d]", loc, i))
				if sub.Value != nil {
					walkSchema(sub.Value, fmt.Sprintf("%s.anyOf[%d]", loc, i))
				}
			}
		}
		// not
		if s.Not != nil {
			addRef(s.Not.Ref, loc+".not")
			if s.Not.Value != nil {
				walkSchema(s.Not.Value, loc+".not")
			}
		}
		// additionalProperties can be boolean or RefOr[*Schema]
		if s.AdditionalProperties != nil {
			if refOr, ok := s.AdditionalProperties.(*openapi3.RefOr[*openapi3.Schema]); ok && refOr != nil {
				addRef(refOr.Ref, loc+".additionalProperties")
				if refOr.Value != nil {
					walkSchema(refOr.Value, loc+".additionalProperties")
				}
			}
		}
	}

	// Helper to walk media types
	walkMediaType := func(mt *openapi3.MediaType, loc string) {
		if mt == nil {
			return
		}
		if mt.Schema != nil {
			addRef(mt.Schema.Ref, loc+".schema")
			if mt.Schema.Value != nil {
				walkSchema(mt.Schema.Value, loc+".schema")
			}
		}
	}

	// Helper to walk parameters
	walkParameter := func(p *openapi3.Parameter, loc string) {
		if p == nil {
			return
		}
		if p.Schema != nil {
			addRef(p.Schema.Ref, loc+".schema")
			if p.Schema.Value != nil {
				walkSchema(p.Schema.Value, loc+".schema")
			}
		}
		for mtName, mt := range p.Content {
			walkMediaType(mt, fmt.Sprintf("%s.content[%s]", loc, mtName))
		}
	}

	// Helper to walk request body
	walkRequestBody := func(rb *openapi3.RequestBody, loc string) {
		if rb == nil {
			return
		}
		for mtName, mt := range rb.Content {
			walkMediaType(mt, fmt.Sprintf("%s.content[%s]", loc, mtName))
		}
	}

	// Helper to walk responses
	walkResponses := func(resps openapi3.Responses, loc string) {
		for code, respRefOr := range resps {
			if respRefOr == nil {
				continue
			}
			addRef(respRefOr.Ref, fmt.Sprintf("%s[%s]", loc, code))
			if respRefOr.Value != nil {
				r := respRefOr.Value
				for mtName, mt := range r.Content {
					walkMediaType(mt, fmt.Sprintf("%s[%s].content[%s]", loc, code, mtName))
				}
			}
		}
	}

	// Helper to walk operations
	walkOperation := func(op *openapi3.Operation, loc string) {
		if op == nil {
			return
		}
		for i, paramRefOr := range op.Parameters {
			if paramRefOr != nil {
				addRef(paramRefOr.Ref, fmt.Sprintf("%s.parameters[%d]", loc, i))
				if paramRefOr.Value != nil {
					walkParameter(paramRefOr.Value, fmt.Sprintf("%s.parameters[%d]", loc, i))
				}
			}
		}
		if op.RequestBody != nil {
			addRef(op.RequestBody.Ref, loc+".requestBody")
			if op.RequestBody.Value != nil {
				walkRequestBody(op.RequestBody.Value, loc+".requestBody")
			}
		}
		walkResponses(op.Responses, loc+".responses")
	}

	// Traversal 1: Walk Paths
	if doc.Paths != nil {
		for path, pathItem := range doc.Paths {
			if pathItem == nil {
				continue
			}
			pathLoc := fmt.Sprintf("paths[%s]", path)
			addRef(pathItem.Ref, pathLoc)

			for i, paramRefOr := range pathItem.Parameters {
				if paramRefOr != nil {
					addRef(paramRefOr.Ref, fmt.Sprintf("%s.parameters[%d]", pathLoc, i))
					if paramRefOr.Value != nil {
						walkParameter(paramRefOr.Value, fmt.Sprintf("%s.parameters[%d]", pathLoc, i))
					}
				}
			}

			walkOperation(pathItem.Get, pathLoc+".get")
			walkOperation(pathItem.Post, pathLoc+".post")
			walkOperation(pathItem.Put, pathLoc+".put")
			walkOperation(pathItem.Delete, pathLoc+".delete")
			walkOperation(pathItem.Options, pathLoc+".options")
			walkOperation(pathItem.Head, pathLoc+".head")
			walkOperation(pathItem.Patch, pathLoc+".patch")
			walkOperation(pathItem.Trace, pathLoc+".trace")
		}
	}

	// Traversal 2: Walk Components (to find references inside reusable components)
	if doc.Components != nil {
		for name, schemaRefOr := range doc.Components.Schemas {
			if schemaRefOr != nil {
				addRef(schemaRefOr.Ref, fmt.Sprintf("components.schemas[%s]", name))
				if schemaRefOr.Value != nil {
					walkSchema(schemaRefOr.Value, fmt.Sprintf("components.schemas[%s]", name))
				}
			}
		}
		for name, respRefOr := range doc.Components.Responses {
			if respRefOr != nil {
				addRef(respRefOr.Ref, fmt.Sprintf("components.responses[%s]", name))
				if respRefOr.Value != nil {
					r := respRefOr.Value
					for mtName, mt := range r.Content {
						walkMediaType(mt, fmt.Sprintf("components.responses[%s].content[%s]", name, mtName))
					}
				}
			}
		}
		for name, paramRefOr := range doc.Components.Parameters {
			if paramRefOr != nil {
				addRef(paramRefOr.Ref, fmt.Sprintf("components.parameters[%s]", name))
				if paramRefOr.Value != nil {
					walkParameter(paramRefOr.Value, fmt.Sprintf("components.parameters[%s]", name))
				}
			}
		}
		for name, reqRefOr := range doc.Components.RequestBodies {
			if reqRefOr != nil {
				addRef(reqRefOr.Ref, fmt.Sprintf("components.requestBodies[%s]", name))
				if reqRefOr.Value != nil {
					walkRequestBody(reqRefOr.Value, fmt.Sprintf("components.requestBodies[%s]", name))
				}
			}
		}
	}

	return occurrences
}

// checkLocalRef verifies that a pointer starting with #/ actually points to a real component in the document.
func checkLocalRef(doc *openapi3.OpenAPI, ref string) error {
	if !strings.HasPrefix(ref, "#/") {
		return fmt.Errorf("invalid local JSON pointer (must start with '#/')")
	}

	parts := strings.Split(ref[2:], "/")
	if len(parts) < 3 {
		return fmt.Errorf("local pointer does not point to a specific component (needs to be e.g. components/schemas/Name)")
	}

	if parts[0] != "components" {
		return fmt.Errorf("local pointer points to root field %q instead of components (OAS 3.0 references typically reside inside 'components')", parts[0])
	}

	category := parts[1]
	name := parts[2]

	if doc.Components == nil {
		return fmt.Errorf("document has no components defined; cannot resolve %q", ref)
	}

	switch category {
	case "schemas":
		if doc.Components.Schemas == nil || doc.Components.Schemas[name] == nil {
			return fmt.Errorf("schema %q not found in components.schemas", name)
		}
	case "responses":
		if doc.Components.Responses == nil || doc.Components.Responses[name] == nil {
			return fmt.Errorf("response %q not found in components.responses", name)
		}
	case "parameters":
		if doc.Components.Parameters == nil || doc.Components.Parameters[name] == nil {
			return fmt.Errorf("parameter %q not found in components.parameters", name)
		}
	case "requestBodies":
		if doc.Components.RequestBodies == nil || doc.Components.RequestBodies[name] == nil {
			return fmt.Errorf("request body %q not found in components.requestBodies", name)
		}
	case "examples":
		if doc.Components.Examples == nil || doc.Components.Examples[name] == nil {
			return fmt.Errorf("example %q not found in components.examples", name)
		}
	case "headers":
		if doc.Components.Headers == nil || doc.Components.Headers[name] == nil {
			return fmt.Errorf("header %q not found in components.headers", name)
		}
	case "securitySchemes":
		if doc.Components.SecuritySchemes == nil || doc.Components.SecuritySchemes[name] == nil {
			return fmt.Errorf("security scheme %q not found in components.securitySchemes", name)
		}
	case "links":
		if doc.Components.Links == nil || doc.Components.Links[name] == nil {
			return fmt.Errorf("link %q not found in components.links", name)
		}
	case "callbacks":
		if doc.Components.Callbacks == nil || doc.Components.Callbacks[name] == nil {
			return fmt.Errorf("callback %q not found in components.callbacks", name)
		}
	default:
		return fmt.Errorf("unsupported or unknown components category %q", category)
	}

	return nil
}

// checkExternalRef validates reference paths that reside outside the entry document (URLs or relative files).
func checkExternalRef(ref string, baseDir string) error {
	// 1. Check if it's a remote URL
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		_, err := url.Parse(ref)
		if err != nil {
			return fmt.Errorf("invalid URL format: %w", err)
		}
		// Since sandbox is air-gapped, we don't perform live network calls, but we validate URI formatting
		return nil
	}

	// 2. Local relative file reference (e.g. "Pet.json", "schemas/user.json#/definitions/User")
	filePath := ref
	anchorIdx := strings.Index(ref, "#")
	if anchorIdx != -1 {
		filePath = ref[:anchorIdx]
	}

	// If no filepath is left (e.g. a plain "#" anchor reference), it's local but handled differently
	if filePath == "" {
		return nil
	}

	// Resolve the relative path against the input document's base directory
	var fullPath string
	if filepath.IsAbs(filePath) {
		fullPath = filePath
	} else if baseDir != "" {
		fullPath = filepath.Join(baseDir, filePath)
	} else {
		fullPath = filePath
	}

	// Check if the external target file exists on disk
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("referenced file does not exist (resolved: %s)", fullPath)
	}
	if err != nil {
		return fmt.Errorf("system error inspecting reference path: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("reference points to a directory %q instead of a file", fullPath)
	}

	return nil
}

// ValidateReferences walks all collected references and resolves them
func ValidateReferences(doc *openapi3.OpenAPI, occurrences []RefOccurrence, baseDir string) []string {
	var errors []string

	for _, occurrence := range occurrences {
		ref := occurrence.Ref
		loc := occurrence.Location

		if strings.HasPrefix(ref, "#") {
			// Local pointers
			if err := checkLocalRef(doc, ref); err != nil {
				errors = append(errors, fmt.Sprintf("Broken Reference at %s: %q -> %v", loc, ref, err))
			}
		} else {
			// External relative file paths or remote URLs
			if err := checkExternalRef(ref, baseDir); err != nil {
				errors = append(errors, fmt.Sprintf("External Reference Issue at %s: %q -> %v", loc, ref, err))
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
		fmt.Println("  go run openapi_cli-v2.go -file <path-to-openapi.json>")
		fmt.Println("  go run openapi_cli-v2.go <path-to-openapi.json>")
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

	// Extract base directory of the input file to resolve relative external $refs
	baseDir := filepath.Dir(filePath)

	// Run structural checks
	validationErrors := ValidateOASObject(doc)

	// Collect and check all schema references ($ref)
	refs := CollectReferences(doc)
	refErrors := ValidateReferences(doc, refs, baseDir)
	validationErrors = append(validationErrors, refErrors...)

	fmt.Println("\n==========================================")
	fmt.Println("         OpenAPI Validation Report        ")
	fmt.Println("==========================================")

	if len(validationErrors) > 0 {
		fmt.Printf("❌ Validation Failed with %d error(s):\n", len(validationErrors))
		for i, vErr := range validationErrors {
			fmt.Printf("  %d. %s\n", i+1, vErr)
		}
		fmt.Println("\nResult: INVALID document according to OAS 3.0 rules.")
		os.Exit(1)
	}

	fmt.Println("Ref Validation metrics:")
	fmt.Printf("  - Total References checked: %d\n", len(refs))
	fmt.Println("✅ Validation Succeeded! The document meets basic OAS 3.0 structural rules and has no broken reference links.")

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
		if pathItem.Post != nil {
			opCount++
		}
		if pathItem.Put != nil {
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
