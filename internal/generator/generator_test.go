package generator

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
	"github.com/google/go-cmp/cmp"
)

func TestGenerate_CustomerDetailCompanyResponseDto(t *testing.T) {
	data, err := os.ReadFile("testdata/CustomerDetailCompanyResponseDto/components.schemas.CustomerDetailCompanyResponseDto.json")
	if err != nil {
		t.Fatalf("reading input fixture: %v", err)
	}

	var spec openapi.OpenAPI
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("unmarshaling input fixture: %v", err)
	}

	got, err := Generate(&spec, "generated")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want, err := os.ReadFile("testdata/CustomerDetailCompanyResponseDto/output.go")
	if err != nil {
		t.Fatalf("reading expected fixture: %v", err)
	}

	if diff := cmp.Diff(string(want), got); diff != "" {
		t.Errorf("generated output does not match testdata/types_example.go (-want +got):\n%s", diff)
	}
}
