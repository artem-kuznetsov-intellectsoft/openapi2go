package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/generator"
	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

const usage = "usage: openapi2go generate <openapi-spec-path> [-o output.go] [-pkg name]\n       openapi2go version"

const generatedFileMode = 0o644

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

	code, supportFiles, clientCode, err := generator.Generate(&spec, *pkg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to generate Go code:", err)
		os.Exit(1)
	}

	if *output == "" {
		fmt.Print(code)

		if len(supportFiles) > 0 {
			fmt.Fprintln(os.Stderr, "note: generated code needs the support types below; pass -o to also write them as files:")
			for name := range supportFiles {
				fmt.Fprintln(os.Stderr, " -", name)
			}
		}

		if clientCode != "" {
			fmt.Fprintln(os.Stderr, "note: a client.go can also be generated; pass -o to write it as a file")
		}

		return
	}

	if err := os.WriteFile(*output, []byte(code), generatedFileMode); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write output file:", err)
		os.Exit(1)
	}

	dir := filepath.Dir(*output)
	for name, content := range supportFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), generatedFileMode); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write support file:", err)
			os.Exit(1)
		}
	}

	if clientCode != "" {
		if err := os.WriteFile(filepath.Join(dir, "client.go"), []byte(clientCode), generatedFileMode); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write client file:", err)
			os.Exit(1)
		}
	}
}
