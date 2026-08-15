package generator

import (
	"sort"
	"strings"
	"unicode"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

// unwrapRef resolves a property value to a referenced schema name, handling
// both a direct $ref and the common allOf-wrapped-single-$ref pattern used
// to attach `nullable` to a reference in OpenAPI 3.0.
func unwrapRef(ref *openapi.RefOr[*openapi.Schema]) (name string, nullable, ok bool) {
	if ref.Ref != "" {
		return lastPathSegment(ref.Ref), false, true
	}

	schema := ref.Value
	if schema == nil {
		return "", false, false
	}

	if len(schema.AllOf) == 1 && schema.AllOf[0].Ref != "" {
		return lastPathSegment(schema.AllOf[0].Ref), schema.Nullable, true
	}

	return "", false, false
}

func lastPathSegment(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[i+1:]
	}

	return ref
}

func isPointerable(typ string) bool {
	return typ != goAny && !strings.HasPrefix(typ, "map[") && !strings.HasPrefix(typ, "[]")
}

// toPascalCase converts a camelCase, snake_case, or kebab-case JSON property
// name into an exported Go identifier.
func toPascalCase(s string) string {
	var b strings.Builder

	upperNext := true
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			upperNext = true
			continue
		}

		if upperNext {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// capitalizeFirst upper-cases the first rune of s, leaving the rest as-is —
// used to turn enum values like "aggressive" into valid, exported const
// name suffixes without disturbing already-uppercase values like
// "PENDING_VERIFICATION".
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

// describeSentence lowercases the first letter of a description and strips
// any trailing period, for embedding into a "X represents the <desc>."
// doc comment.
func describeSentence(desc string) string {
	d := strings.TrimSuffix(strings.TrimSpace(desc), ".")
	if d == "" {
		return ""
	}

	return strings.ToLower(d[:1]) + d[1:]
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}

	return m
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
