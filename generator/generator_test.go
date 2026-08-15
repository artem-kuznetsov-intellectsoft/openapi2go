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
			refFile:   "fixtures/CustomerDetailCompanyResponseDto/generated.ref.go",
		},
		{
			name:      "Car",
			inputFile: "fixtures/Car/Car.json",
			refFile:   "fixtures/Car/generated.ref.go",
		},
		{
			name:      "CompanyDetailResponseDto",
			inputFile: "fixtures/CompanyDetailResponseDto/components.schemas.CompanyDetailResponseDto.json",
			refFile:   "fixtures/CompanyDetailResponseDto/generated.ref.go",
		},
		{
			name:      "Bicycle",
			inputFile: "fixtures/Bicycle/Bicycle.json",
			refFile:   "fixtures/Bicycle/generated.ref.go",
		},
		{
			name:      "BaseModel",
			inputFile: "fixtures/BaseModel/BaseModel.json",
			refFile:   "fixtures/BaseModel/generated.ref.go",
		},
		{
			name:      "ArraysAndCollections",
			inputFile: "fixtures/ArraysAndCollections/ArraysAndCollections.json",
			refFile:   "fixtures/ArraysAndCollections/generated.ref.go",
		},
		{
			name:      "CompositionAllOf",
			inputFile: "fixtures/CompositionAllOf/CompositionAllOf.json",
			refFile:   "fixtures/CompositionAllOf/generated.ref.go",
		},
		{
			name:      "OneOfPrimitives",
			inputFile: "fixtures/OneOfPrimitives/OneOfPrimitives.json",
			refFile:   "fixtures/OneOfPrimitives/generated.ref.go",
		},
		{
			name:      "OneOfObjectsWithoutDiscriminator",
			inputFile: "fixtures/OneOfObjectsWithoutDiscriminator/OneOfObjectsWithoutDiscriminator.json",
			refFile:   "fixtures/OneOfObjectsWithoutDiscriminator/generated.ref.go",
		},
		{
			name:      "PolymorphicCat",
			inputFile: "fixtures/PolymorphicCat/PolymorphicCat.json",
			refFile:   "fixtures/PolymorphicCat/generated.ref.go",
		},
		{
			name:      "PolymorphicPet",
			inputFile: "fixtures/PolymorphicPet/PolymorphicPet.json",
			refFile:   "fixtures/PolymorphicPet/generated.ref.go",
		},
		{
			name:      "NullableFields",
			inputFile: "fixtures/NullableFields/NullableFields.json",
			refFile:   "fixtures/NullableFields/generated.ref.go",
		},
		{
			name:      "MapsAndDictionaries",
			inputFile: "fixtures/MapsAndDictionaries/MapsAndDictionaries.json",
			refFile:   "fixtures/MapsAndDictionaries/generated.ref.go",
		},
		{
			name:      "PrimitivesAndFormats",
			inputFile: "fixtures/PrimitivesAndFormats/PrimitivesAndFormats.json",
			refFile:   "fixtures/PrimitivesAndFormats/generated.ref.go",
		},
		{
			name:      "Formats",
			inputFile: "fixtures/Formats/Formats.json",
			refFile:   "fixtures/Formats/generated.ref.go",
		},
		{
			name:      "Vehicle",
			inputFile: "fixtures/Vehicle/Vehicle.json",
			refFile:   "fixtures/Vehicle/generated.ref.go",
		},
		{
			name:      "VehicleUnion",
			inputFile: "fixtures/VehicleUnion/VehicleUnion.json",
			refFile:   "fixtures/VehicleUnion/generated.ref.go",
		},
		{
			name:      "IndividualResponseDto",
			inputFile: "fixtures/IndividualResponseDto/IndividualResponseDto.json",
			refFile:   "fixtures/IndividualResponseDto/generated.ref.go",
		},
		{
			name:      "OmitZero",
			inputFile: "fixtures/OmitZero/OmitZero.json",
			refFile:   "fixtures/OmitZero/generated.ref.go",
		},
		{
			name:          "BasePath",
			inputFile:     "fixtures/BasePath/BasePath.json",
			refFile:       "fixtures/BasePath/generated.ref.go",
			clientRefFile: "fixtures/BasePath/client.ref.go",
		},
		{
			name:          "PathParameters",
			inputFile:     "fixtures/PathParameters/PathParameters.json",
			refFile:       "fixtures/PathParameters/generated.ref.go",
			clientRefFile: "fixtures/PathParameters/client.ref.go",
		},
		{
			name:          "Customer",
			inputFile:     "fixtures/Customer/Customer.json",
			refFile:       "fixtures/Customer/generated.ref.go",
			clientRefFile: "fixtures/Customer/client.ref.go",
		},
		{
			name:          "CustomerPost",
			inputFile:     "fixtures/CustomerPost/CustomerPost.json",
			refFile:       "fixtures/CustomerPost/generated.ref.go",
			clientRefFile: "fixtures/CustomerPost/client.ref.go",
		},
		{
			name:          "InlinedSchemas",
			inputFile:     "fixtures/InlinedSchemas/InlinedSchemas.json",
			refFile:       "fixtures/InlinedSchemas/generated.ref.go",
			clientRefFile: "fixtures/InlinedSchemas/client.ref.go",
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
