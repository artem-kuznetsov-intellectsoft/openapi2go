package generator

import (
	"fmt"
	"strings"
)

func (g *generator) render(pkgName string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	for _, e := range g.enums {
		renderEnum(&b, e)
	}

	for _, s := range g.structs {
		renderStruct(&b, s)
	}

	return b.String()
}

func renderEnum(b *strings.Builder, e *enumDef) {
	if sentence := describeSentence(e.description); sentence != "" {
		fmt.Fprintf(b, "// %s represents the %s.\n", e.name, sentence)
	}

	fmt.Fprintf(b, "type %s string\n\n", e.name)

	b.WriteString("const (\n")
	for _, v := range e.values {
		fmt.Fprintf(b, "%s%s %s = %q\n", e.name, capitalizeFirst(v), e.name, v)
	}
	b.WriteString(")\n\n")
}

func renderStruct(b *strings.Builder, s *structDef) {
	if s.comment != "" {
		fmt.Fprintf(b, "// %s\n", s.comment)
	}

	switch {
	case s.alias != "":
		fmt.Fprintf(b, "type %s = %s\n\n", s.name, s.alias)
	case len(s.fields) == 0:
		fmt.Fprintf(b, "type %s struct{}\n\n", s.name)
	default:
		fmt.Fprintf(b, "type %s struct {\n", s.name)
		for _, f := range s.fields {
			renderField(b, f)
		}
		b.WriteString("}\n\n")
	}

	if s.errorBody != "" {
		fmt.Fprintf(b, "func (r %s) Error() string {\n%s\n}\n\n", s.name, s.errorBody)
	}
}

func renderField(b *strings.Builder, f fieldDef) {
	switch {
	case f.embedded:
		fmt.Fprintf(b, "%s\n", f.typ)
	case f.jsonTag == "":
		fmt.Fprintf(b, "%s %s\n", f.name, f.typ)
	default:
		fmt.Fprintf(b, "%s %s `json:%q`\n", f.name, f.typ, f.jsonTag)
	}
}
