package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/generator"
	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

const usage = "usage: openapi2go generate <openapi-spec-path> [-o output.go] [-pkg name]\n       openapi2go version"

const generatedFileMode = 0o644

var (
	errUsage          = errors.New(usage)
	errUnknownCommand = errors.New("unknown command")
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the single point that converts a subcommand's error into a printed
// message and an exit code, so every function below it can be exercised with
// bytes.Buffer stand-ins for stdout/stderr instead of calling os.Exit itself.
func run(args []string, stdout, stderr io.Writer) int {
	if err := dispatch(args, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func dispatch(args []string, stdout, stderr io.Writer) error {
	if len(args) < 1 {
		return errUsage
	}

	switch args[0] {
	case "generate":
		return runGenerate(args[1:], stdout, stderr)
	case "version":
		return runVersion(stdout)
	default:
		return fmt.Errorf("%w %q\n%s", errUnknownCommand, args[0], usage)
	}
}

// runVersion prints the short commit hash Go's toolchain stamps into the
// binary from VCS info at build time (go install/build from a git checkout),
// available via debug.ReadBuildInfo without any custom ldflags.
// shortRevisionLen is the number of leading characters of a full VCS
// revision hash to print, matching git's default abbreviated-hash length.
const shortRevisionLen = 7

func runVersion(stdout io.Writer) error {
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

	_, _ = fmt.Fprintln(stdout, revision)

	return nil
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

func runGenerate(args []string, stdout, stderr io.Writer) error {
	output, pkg, specPath, err := parseGenerateFlags(args)
	if err != nil {
		return err
	}

	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}

	code, supportFiles, clientCode, err := generator.Generate(&spec, pkg)
	if err != nil {
		return fmt.Errorf("failed to generate Go code: %w", err)
	}

	if output == "" {
		writeGeneratedStdout(stdout, stderr, code, supportFiles, clientCode)
		return nil
	}

	return writeGeneratedFiles(output, code, supportFiles, clientCode)
}

// parseGenerateFlags parses the "generate" subcommand's flags and positional
// spec-path argument. It never prints or exits itself: flag.ContinueOnError
// with output discarded keeps callers in control of the single error-message
// print site in run.
func parseGenerateFlags(args []string) (output, pkg, specPath string, err error) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	o := fs.String("o", "", "output file path for generated Go code (default: stdout)")
	p := fs.String("pkg", "generated", "package name for generated Go code")

	if err := fs.Parse(reorderFlagsFirst(args, "-o", "-pkg")); err != nil {
		return "", "", "", err
	}

	if fs.NArg() < 1 {
		return "", "", "", errUsage
	}

	return *o, *p, fs.Arg(0), nil
}

// loadSpec reads and unmarshals the OpenAPI spec at specPath.
func loadSpec(specPath string) (openapi.OpenAPI, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return openapi.OpenAPI{}, fmt.Errorf("failed to read openapi spec: %w", err)
	}

	var spec openapi.OpenAPI
	if err := json.Unmarshal(data, &spec); err != nil {
		return openapi.OpenAPI{}, fmt.Errorf("failed to unmarshal openapi spec: %w", err)
	}

	return spec, nil
}

// writeGeneratedStdout prints the generated code to stdout and, since no
// output path was given to also write the support/client files to, notes
// their availability on stderr instead.
func writeGeneratedStdout(stdout, stderr io.Writer, code string, supportFiles map[string]string, clientCode string) {
	_, _ = fmt.Fprint(stdout, code)

	if len(supportFiles) > 0 {
		fmt.Fprintln(stderr, "note: generated code needs the support types below; pass -o to also write them as files:")

		for name := range supportFiles {
			fmt.Fprintln(stderr, " -", name)
		}
	}

	if clientCode != "" {
		fmt.Fprintln(stderr, "note: a client.go can also be generated; pass -o to write it as a file")
	}
}

// writeGeneratedFiles writes the generated code, support files, and (if any)
// client.go alongside output.
func writeGeneratedFiles(output, code string, supportFiles map[string]string, clientCode string) error {
	if err := os.WriteFile(output, []byte(code), generatedFileMode); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	dir := filepath.Dir(output)
	for name, content := range supportFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), generatedFileMode); err != nil {
			return fmt.Errorf("failed to write support file: %w", err)
		}
	}

	if clientCode != "" {
		if err := os.WriteFile(filepath.Join(dir, "client.go"), []byte(clientCode), generatedFileMode); err != nil {
			return fmt.Errorf("failed to write client file: %w", err)
		}
	}

	return nil
}
