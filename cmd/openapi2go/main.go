package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/internal/generator"
	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

const usage = "usage: openapi2go generate <openapi-spec-path> [-o output.go] [-pkg name]\n       openapi2go version"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		runGenerate(os.Args[2:])
	case "version":
		runVersion()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n%s\n", os.Args[1], usage)
		os.Exit(1)
	}
}

// runVersion prints the short commit hash Go's toolchain stamps into the
// binary from VCS info at build time (go install/build from a git checkout),
// available via debug.ReadBuildInfo without any custom ldflags.
func runVersion() {
	revision := "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				revision = s.Value
				if len(revision) > 7 {
					revision = revision[:7]
				}

				break
			}
		}
	}

	fmt.Println(revision)
}

// reorderFlagsFirst moves recognized valued flags (and their values) to the
// front of args, since flag.FlagSet.Parse stops at the first non-flag token
// and would otherwise ignore flags placed after the positional spec path.
func reorderFlagsFirst(args []string, valuedFlags ...string) []string {
	isValuedFlag := make(map[string]bool, len(valuedFlags)*2)
	for _, f := range valuedFlags {
		isValuedFlag[f] = true
		isValuedFlag["-"+f] = true
	}

	var flags, rest []string
	for i := 0; i < len(args); i++ {
		if isValuedFlag[args[i]] {
			flags = append(flags, args[i])
			if i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}

			continue
		}

		rest = append(rest, args[i])
	}

	return append(flags, rest...)
}

func runGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	output := fs.String("o", "", "output file path for generated Go code (default: stdout)")
	pkg := fs.String("pkg", "generated", "package name for generated Go code")
	fs.Parse(reorderFlagsFirst(args, "-o", "-pkg"))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	specPath := fs.Arg(0)

	data, err := os.ReadFile(specPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read openapi spec:", err)
		os.Exit(1)
	}

	var spec openapi.OpenAPI
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintln(os.Stderr, "failed to unmarshal openapi spec:", err)
		os.Exit(1)
	}

	code, err := generator.Generate(&spec, *pkg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to generate Go code:", err)
		os.Exit(1)
	}

	if *output == "" {
		fmt.Print(code)
		return
	}

	if err := os.WriteFile(*output, []byte(code), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write output file:", err)
		os.Exit(1)
	}
}
