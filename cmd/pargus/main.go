package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dspasibenko/pargus/pkg/generator"
	"github.com/dspasibenko/pargus/pkg/parser"
)

func main() {
	var (
		output    = flag.String("o", "", "Output file (default: input.h for C++, input.go for Go)")
		namespace = flag.String("n", "", "C++ namespace name (required for C++)")
		pkg       = flag.String("p", "", "Go package name (required for Go)")
		genType   = flag.String("t", "cpp", "Generator type: cpp or go")
		help      = flag.Bool("help", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.pa\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate C++ code:\n")
		fmt.Fprintf(os.Stderr, "  %s -t cpp -n MyNamespace -o output.h input.pa\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate Go code:\n")
		fmt.Fprintf(os.Stderr, "  %s -t go -p mypackage -o output.go input.pa\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate generator type
	if *genType != "cpp" && *genType != "go" {
		fmt.Fprintf(os.Stderr, "Error: generator type must be 'cpp' or 'go'\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check for required parameters based on generator type
	if *genType == "cpp" && *namespace == "" {
		fmt.Fprintf(os.Stderr, "Error: -n (namespace) parameter is required for C++ generator\n")
		flag.Usage()
		os.Exit(1)
	}

	if *genType == "go" && *pkg == "" {
		fmt.Fprintf(os.Stderr, "Error: -p (package) parameter is required for Go generator\n")
		flag.Usage()
		os.Exit(1)
	}

	// Get input file from command line arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Error: only one input file is allowed\n")
		flag.Usage()
		os.Exit(1)
	}

	inputFile := args[0]

	// Set default output file if not specified
	if *output == "" {
		ext := filepath.Ext(inputFile)
		base := filepath.Base(inputFile)
		if ext != "" {
			base = base[:len(base)-len(ext)]
		}
		if *genType == "cpp" {
			*output = base
		} else {
			*output = base + ".go"
		}
	}

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	// Parse the input
	device, err := parser.Parse(string(inputData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input: %v\n", err)
		os.Exit(1)
	}

	// Generate code
	if *genType == "cpp" {
		// Remove extension from output if it was specified
		outputBase := *output
		if ext := filepath.Ext(outputBase); ext == ".h" || ext == ".hpp" || ext == ".cpp" {
			outputBase = outputBase[:len(outputBase)-len(ext)]
		}

		hppFileName := outputBase + ".h"
		cppFileName := outputBase + ".cpp"

		// Use only the base filename (without directory path) for includes and guards
		baseHppFileName := filepath.Base(hppFileName)
		hpp, cpp, err := generator.GenerateHppCpp(device, *namespace, strings.ReplaceAll(baseHppFileName, ".", "_"), baseHppFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
			os.Exit(1)
		}
		// Write output file
		err = os.WriteFile(hppFileName, []byte(hpp), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", hppFileName, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully generated %s\n", hppFileName)
		err = os.WriteFile(cppFileName, []byte(cpp), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", cppFileName, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully generated %s\n", cppFileName)
		return
	}
	code, err := generator.GenerateGo(device, *pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	err = os.WriteFile(*output, []byte(code), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", *output, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s\n", *output)
}
