package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/maxmind/geoip2-csv-converter/convert"
)

func main() {
	input := flag.String("block-file", "", "The path to the block CSV file to use as input (REQUIRED)")
	output := flag.String("output-file", "", "The path to the output CSV (REQUIRED)")
	ipRange := flag.Bool("include-range", false, "Include the IP range of the network in string format")
	intRange := flag.Bool("include-integer-range", false, "Include the IP range of the network in integer format")
	cidr := flag.Bool("include-cidr", false, "Include the network in CIDR format")

	flag.Parse()

	var errors []string

	if *input == "" {
		errors = append(errors, "-block-file is required")
	}

	if *output == "" {
		errors = append(errors, "-output-file is required")
	}

	if !*ipRange && !*intRange && !*cidr {
		errors = append(errors, "-include-cidr, -include-range, or -include-integer-range is required")
	}

	args := flag.Args()
	if len(args) > 0 {
		errors = append(errors, "unknown argument(s): "+strings.Join(args, ", "))
	}

	if len(errors) != 0 {
		printHelp(errors)
		os.Exit(1)
	}

	err := convert.ConvertFile(*input, *output, *cidr, *ipRange, *intRange)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: %v\n", err)
	}
}

func printHelp(errors []string) {
	for _, message := range errors {
		fmt.Fprintln(flag.CommandLine.Output(), message)
	}
	flag.PrintDefaults()
}
