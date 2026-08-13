package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
	"github.com/google/go-cmp/cmp"
)

func TestGenerate(t *testing.T) {
	update := os.Getenv("UPDATE_GOLDEN") != ""

	tests := []struct {
		name      string
		inputFile string
		refFile   string
	}{
		{
			name:      "CustomerDetailCompanyResponseDto",
			inputFile: "testdata/CustomerDetailCompanyResponseDto/components.schemas.CustomerDetailCompanyResponseDto.json",
			refFile:   "testdata/CustomerDetailCompanyResponseDto/generated.ref.go",
		},
		{
			name:      "Car",
			inputFile: "testdata/Car/Car.json",
			refFile:   "testdata/Car/generated.ref.go",
		},
		{
			name:      "CompanyDetailResponseDto",
			inputFile: "testdata/CompanyDetailResponseDto/components.schemas.CompanyDetailResponseDto.json",
			refFile:   "testdata/CompanyDetailResponseDto/generated.ref.go",
		},
		{
			name:      "Bicycle",
			inputFile: "testdata/Bicycle/Bicycle.json",
			refFile:   "testdata/Bicycle/generated.ref.go",
		},
		{
			name:      "BaseModel",
			inputFile: "testdata/BaseModel/BaseModel.json",
			refFile:   "testdata/BaseModel/generated.ref.go",
		},
		{
			name:      "ArraysAndCollections",
			inputFile: "testdata/ArraysAndCollections/ArraysAndCollections.json",
			refFile:   "testdata/ArraysAndCollections/generated.ref.go",
		},
		{
			name:      "CompositionAllOf",
			inputFile: "testdata/CompositionAllOf/CompositionAllOf.json",
			refFile:   "testdata/CompositionAllOf/generated.ref.go",
		},
		{
			name:      "OneOfPrimitives",
			inputFile: "testdata/OneOfPrimitives/OneOfPrimitives.json",
			refFile:   "testdata/OneOfPrimitives/generated.ref.go",
		},
		{
			name:      "OneOfObjectsWithoutDiscriminator",
			inputFile: "testdata/OneOfObjectsWithoutDiscriminator/OneOfObjectsWithoutDiscriminator.json",
			refFile:   "testdata/OneOfObjectsWithoutDiscriminator/generated.ref.go",
		},
		{
			name:      "PolymorphicCat",
			inputFile: "testdata/PolymorphicCat/PolymorphicCat.json",
			refFile:   "testdata/PolymorphicCat/generated.ref.go",
		},
		{
			name:      "PolymorphicPet",
			inputFile: "testdata/PolymorphicPet/PolymorphicPet.json",
			refFile:   "testdata/PolymorphicPet/generated.ref.go",
		},
		{
			name:      "NullableFields",
			inputFile: "testdata/NullableFields/NullableFields.json",
			refFile:   "testdata/NullableFields/generated.ref.go",
		},
		{
			name:      "MapsAndDictionaries",
			inputFile: "testdata/MapsAndDictionaries/MapsAndDictionaries.json",
			refFile:   "testdata/MapsAndDictionaries/generated.ref.go",
		},
		{
			name:      "PrimitivesAndFormats",
			inputFile: "testdata/PrimitivesAndFormats/PrimitivesAndFormats.json",
			refFile:   "testdata/PrimitivesAndFormats/generated.ref.go",
		},
		{
			name:      "Formats",
			inputFile: "testdata/Formats/Formats.json",
			refFile:   "testdata/Formats/generated.ref.go",
		},
		{
			name:      "Vehicle",
			inputFile: "testdata/Vehicle/Vehicle.json",
			refFile:   "testdata/Vehicle/generated.ref.go",
		},
		{
			name:      "VehicleUnion",
			inputFile: "testdata/VehicleUnion/VehicleUnion.json",
			refFile:   "testdata/VehicleUnion/generated.ref.go",
		},
		{
			name:      "IndividualResponseDto",
			inputFile: "testdata/IndividualResponseDto/IndividualResponseDto.json",
			refFile:   "testdata/IndividualResponseDto/generated.ref.go",
		},
		{
			name:      "BasePath",
			inputFile: "testdata/BasePath/BasePath.json",
			refFile:   "testdata/BasePath/generated.ref.go",
		},
		{
			name:      "PathParameters",
			inputFile: "testdata/PathParameters/PathParameters.json",
			refFile:   "testdata/PathParameters/generated.ref.go",
		},
		{
			name:      "Customer",
			inputFile: "testdata/Customer/Customer.json",
			refFile:   "testdata/Customer/generated.ref.go",
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

			got, supportFiles, err := Generate(&spec, "generated")
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}

			dir := filepath.Dir(tt.refFile)

			if update {
				if err := os.WriteFile(tt.refFile, []byte(got), 0o644); err != nil {
					t.Fatalf("updating golden file %s: %v", tt.refFile, err)
				}

				for name, content := range supportFiles {
					if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
						t.Fatalf("updating golden support file %s: %v", name, err)
					}
				}

				return
			}

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
		})
	}
}
