package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

type goldenTestCase struct {
	name          string
	inputFile     string
	refFile       string
	clientRefFile string // "" skips the client.go comparison — see Customer's entry
}

func TestGenerate(t *testing.T) {
	update := os.Getenv("UPDATE_GOLDEN") != ""

	tests := []goldenTestCase{
		{
			name:      "CustomerDetailCompanyResponseDto",
			inputFile: "fixtures/CustomerDetailCompanyResponseDto/components.schemas.CustomerDetailCompanyResponseDto.json",
			refFile:   "fixtures/CustomerDetailCompanyResponseDto/types.gen.go",
		},
		{
			name:      "Car",
			inputFile: "fixtures/Car/Car.json",
			refFile:   "fixtures/Car/types.gen.go",
		},
		{
			name:      "CompanyDetailResponseDto",
			inputFile: "fixtures/CompanyDetailResponseDto/components.schemas.CompanyDetailResponseDto.json",
			refFile:   "fixtures/CompanyDetailResponseDto/types.gen.go",
		},
		{
			name:      "Bicycle",
			inputFile: "fixtures/Bicycle/Bicycle.json",
			refFile:   "fixtures/Bicycle/types.gen.go",
		},
		{
			name:      "BaseModel",
			inputFile: "fixtures/BaseModel/BaseModel.json",
			refFile:   "fixtures/BaseModel/types.gen.go",
		},
		{
			name:      "ArraysAndCollections",
			inputFile: "fixtures/ArraysAndCollections/ArraysAndCollections.json",
			refFile:   "fixtures/ArraysAndCollections/types.gen.go",
		},
		{
			name:      "CompositionAllOf",
			inputFile: "fixtures/CompositionAllOf/CompositionAllOf.json",
			refFile:   "fixtures/CompositionAllOf/types.gen.go",
		},
		{
			name:      "OneOfPrimitives",
			inputFile: "fixtures/OneOfPrimitives/OneOfPrimitives.json",
			refFile:   "fixtures/OneOfPrimitives/types.gen.go",
		},
		{
			name:      "OneOfObjectsWithoutDiscriminator",
			inputFile: "fixtures/OneOfObjectsWithoutDiscriminator/OneOfObjectsWithoutDiscriminator.json",
			refFile:   "fixtures/OneOfObjectsWithoutDiscriminator/types.gen.go",
		},
		{
			name:      "PolymorphicCat",
			inputFile: "fixtures/PolymorphicCat/PolymorphicCat.json",
			refFile:   "fixtures/PolymorphicCat/types.gen.go",
		},
		{
			name:      "PolymorphicPet",
			inputFile: "fixtures/PolymorphicPet/PolymorphicPet.json",
			refFile:   "fixtures/PolymorphicPet/types.gen.go",
		},
		{
			name:      "NullableFields",
			inputFile: "fixtures/NullableFields/NullableFields.json",
			refFile:   "fixtures/NullableFields/types.gen.go",
		},
		{
			name:      "MapsAndDictionaries",
			inputFile: "fixtures/MapsAndDictionaries/MapsAndDictionaries.json",
			refFile:   "fixtures/MapsAndDictionaries/types.gen.go",
		},
		{
			name:      "PrimitivesAndFormats",
			inputFile: "fixtures/PrimitivesAndFormats/PrimitivesAndFormats.json",
			refFile:   "fixtures/PrimitivesAndFormats/types.gen.go",
		},
		{
			name:      "Formats",
			inputFile: "fixtures/Formats/Formats.json",
			refFile:   "fixtures/Formats/types.gen.go",
		},
		{
			name:      "Vehicle",
			inputFile: "fixtures/Vehicle/Vehicle.json",
			refFile:   "fixtures/Vehicle/types.gen.go",
		},
		{
			name:      "VehicleUnion",
			inputFile: "fixtures/VehicleUnion/VehicleUnion.json",
			refFile:   "fixtures/VehicleUnion/types.gen.go",
		},
		{
			name:      "IndividualResponseDto",
			inputFile: "fixtures/IndividualResponseDto/IndividualResponseDto.json",
			refFile:   "fixtures/IndividualResponseDto/types.gen.go",
		},
		{
			name:      "OmitZero",
			inputFile: "fixtures/OmitZero/OmitZero.json",
			refFile:   "fixtures/OmitZero/types.gen.go",
		},
		{
			name:          "BasePath",
			inputFile:     "fixtures/BasePath/BasePath.json",
			refFile:       "fixtures/BasePath/types.gen.go",
			clientRefFile: "fixtures/BasePath/client.gen.go",
		},
		{
			name:          "PathParameters",
			inputFile:     "fixtures/PathParameters/PathParameters.json",
			refFile:       "fixtures/PathParameters/types.gen.go",
			clientRefFile: "fixtures/PathParameters/client.gen.go",
		},
		{
			name:          "Customer",
			inputFile:     "fixtures/Customer/Customer.json",
			refFile:       "fixtures/Customer/types.gen.go",
			clientRefFile: "fixtures/Customer/client.gen.go",
		},
		{
			name:          "CustomerPost",
			inputFile:     "fixtures/CustomerPost/CustomerPost.json",
			refFile:       "fixtures/CustomerPost/types.gen.go",
			clientRefFile: "fixtures/CustomerPost/client.gen.go",
		},
		{
			name:          "InlinedSchemas",
			inputFile:     "fixtures/InlinedSchemas/InlinedSchemas.json",
			refFile:       "fixtures/InlinedSchemas/types.gen.go",
			clientRefFile: "fixtures/InlinedSchemas/client.gen.go",
		},
		{
			name:          "EnumWithSpecialSymbols",
			inputFile:     "fixtures/EnumWithSpecialSymbols/EnumWithSpecialSymbols.json",
			refFile:       "fixtures/EnumWithSpecialSymbols/types.gen.go",
			clientRefFile: "fixtures/EnumWithSpecialSymbols/client.gen.go",
		},
		{
			name:          "OperationEdgeCases",
			inputFile:     "fixtures/OperationEdgeCases/OperationEdgeCases.json",
			refFile:       "fixtures/OperationEdgeCases/types.gen.go",
			clientRefFile: "fixtures/OperationEdgeCases/client.gen.go",
		},
		{
			name:      "SchemaReuseAndEdgeCases",
			inputFile: "fixtures/SchemaReuseAndEdgeCases/SchemaReuseAndEdgeCases.json",
			refFile:   "fixtures/SchemaReuseAndEdgeCases/types.gen.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.inputFile)
			if err != nil {
				t.Fatalf("reading input fixture: %v", err)
			}

			var spec openapi.OpenAPI
			if err := json.Unmarshal(data, &spec); err != nil {
				t.Fatalf("unmarshaling input fixture: %v", err)
			}

			got, supportFiles, clientCode, err := Generate(&spec, "generated")
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}

			if update {
				writeGoldenFiles(t, tt, got, supportFiles, clientCode)
				return
			}

			compareGoldenFiles(t, tt, got, supportFiles, clientCode)
		})
	}
}

func writeGoldenFiles(t *testing.T, tt goldenTestCase, got string, supportFiles map[string]string, clientCode string) {
	t.Helper()

	dir := filepath.Dir(tt.refFile)

	if err := os.WriteFile(tt.refFile, []byte(got), 0o644); err != nil {
		t.Fatalf("updating golden file %s: %v", tt.refFile, err)
	}

	for name, content := range supportFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("updating golden support file %s: %v", name, err)
		}
	}

	if tt.clientRefFile == "" {
		return
	}

	if clientCode != "" {
		if err := os.WriteFile(tt.clientRefFile, []byte(clientCode), 0o644); err != nil {
			t.Fatalf("updating golden client file %s: %v", tt.clientRefFile, err)
		}
		return
	}

	// No physical golden file for "no client.go generated" — an empty file
	// isn't valid Go and trips gofmt/lint on the fixture tree. Absence of
	// the file is itself the signal.
	if err := os.Remove(tt.clientRefFile); err != nil && !os.IsNotExist(err) {
		t.Fatalf("removing stale golden client file %s: %v", tt.clientRefFile, err)
	}
}

func compareGoldenFiles(t *testing.T, tt goldenTestCase, got string, supportFiles map[string]string, clientCode string) {
	t.Helper()

	dir := filepath.Dir(tt.refFile)

	want, err := os.ReadFile(tt.refFile)
	if err != nil {
		t.Fatalf("reading expected fixture: %v", err)
	}

	if diff := cmp.Diff(string(want), got); diff != "" {
		t.Errorf("generated output does not match %s (-want +got):\n%s", tt.refFile, diff)
	}

	for name, content := range supportFiles {
		wantSupport, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("reading expected support file %s: %v", name, err)
		}

		if diff := cmp.Diff(string(wantSupport), content); diff != "" {
			t.Errorf("generated support file %s does not match (-want +got):\n%s", name, diff)
		}
	}

	if tt.clientRefFile == "" {
		return
	}

	wantClient, err := os.ReadFile(tt.clientRefFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("reading expected client fixture: %v", err)
	}
	// A missing file means the fixture expects no client.go at all (e.g. no
	// operationId) — wantClient stays "" in that case.

	if diff := cmp.Diff(string(wantClient), clientCode); diff != "" {
		t.Errorf("generated client output does not match %s (-want +got):\n%s", tt.clientRefFile, diff)
	}
}
