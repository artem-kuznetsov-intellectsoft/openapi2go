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

// minArgs is the fewest os.Args entries needed to have a subcommand: the
// binary name itself plus the subcommand (e.g. "generate" or "version").
const minArgs = 2

func main() {
	if len(os.Args) < minArgs {
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
// shortRevisionLen is the number of leading characters of a full VCS
// revision hash to print, matching git's default abbreviated-hash length.
const shortRevisionLen = 7

func runVersion() {
	revision := "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				revision = s.Value
				if len(revision) > shortRevisionLen {
					revision = revision[:shortRevisionLen]
				}

				break
			}
		}
	}

	// A failed write to stdout isn't actionable here, same as the
	// fmt.Fprintln(os.Stderr, ...) calls elsewhere in this file.
	_, _ = fmt.Fprintln(os.Stdout, revision)
}

// reorderFlagsFirst moves recognized valued flags (and their values) to the
// front of args, since flag.FlagSet.Parse stops at the first non-flag token
// and would otherwise ignore flags placed after the positional spec path.
func reorderFlagsFirst(args []string, valuedFlags ...string) []string {
	// Each valued flag is indexed under both "name" and "-name", hence *2.
	const aliasesPerFlag = 2

	isValuedFlag := make(map[string]bool, len(valuedFlags)*aliasesPerFlag)
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
	output, pkg, specPath := parseGenerateFlags(args)
	spec := loadSpec(specPath)

	code, supportFiles, clientCode, err := generator.Generate(&spec, pkg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to generate Go code:", err)
		os.Exit(1)
	}

	if output == "" {
		writeGeneratedStdout(code, supportFiles, clientCode)
		return
	}

	writeGeneratedFiles(output, code, supportFiles, clientCode)
}

// parseGenerateFlags parses the "generate" subcommand's flags and positional
// spec-path argument, exiting the process on a missing spec path (fs itself
// uses flag.ExitOnError, so Parse only returns after a successful parse).
func parseGenerateFlags(args []string) (output, pkg, specPath string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	o := fs.String("o", "", "output file path for generated Go code (default: stdout)")
	p := fs.String("pkg", "generated", "package name for generated Go code")
	_ = fs.Parse(reorderFlagsFirst(args, "-o", "-pkg"))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	return *o, *p, fs.Arg(0)
}

// loadSpec reads and unmarshals the OpenAPI spec at specPath, exiting the
// process on either failure.
func loadSpec(specPath string) openapi.OpenAPI {
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

	return spec
}

// writeGeneratedStdout prints the generated code to stdout and, since no
// output path was given to also write the support/client files to, notes
// their availability on stderr instead.
func writeGeneratedStdout(code string, supportFiles map[string]string, clientCode string) {
	_, _ = fmt.Fprint(os.Stdout, code)

	if len(supportFiles) > 0 {
		fmt.Fprintln(os.Stderr, "note: generated code needs the support types below; pass -o to also write them as files:")

		for name := range supportFiles {
			fmt.Fprintln(os.Stderr, " -", name)
		}
	}

	if clientCode != "" {
		fmt.Fprintln(os.Stderr, "note: a client.go can also be generated; pass -o to write it as a file")
	}
}

// writeGeneratedFiles writes the generated code, support files, and (if any)
// client.go alongside output, exiting the process on any write failure.
func writeGeneratedFiles(output, code string, supportFiles map[string]string, clientCode string) {
	if err := os.WriteFile(output, []byte(code), generatedFileMode); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write output file:", err)
		os.Exit(1)
	}

	dir := filepath.Dir(output)
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
