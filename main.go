// geoip2-csv-converter is a utility for converting the MaxMind GeoIP2 and
// GeoLite2 CSVs to different formats for representing IP addresses such as IP
// ranges or integer ranges.
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
	hexRange := flag.Bool("include-hex-range", false, "Include the IP range of the network in hexadecimal format")
	cidr := flag.Bool("include-cidr", false, "Include the network in CIDR format")

	flag.Parse()

	var errors []string

	if *input == "" {
		errors = append(errors, "-block-file is required")
	}

	if *output == "" {
		errors = append(errors, "-output-file is required")
	}

	if *input != "" && *output != "" && *output == *input {
		errors = append(errors, "Your output file must be different than your block file(input file).")
	}

	if !*ipRange && !*intRange && !*cidr && !*hexRange {
		errors = append(errors, "-include-cidr, -include-range, -include-integer-range,"+
			" or -include-hex-range is required")
	}

	args := flag.Args()
	if len(args) > 0 {
		errors = append(errors, "unknown argument(s): "+strings.Join(args, ", "))
	}

	if len(errors) != 0 {
		printHelp(errors)
		os.Exit(1)
	}

	err := convert.ConvertFile(*input, *output, *cidr, *ipRange, *intRange, *hexRange)
	if err != nil {
		//nolint:errcheck // We are exiting and there isn't much we can do.
		fmt.Fprintf(flag.CommandLine.Output(), "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp(errors []string) {
	var passedFlags []string
	flag.Visit(func(f *flag.Flag) {
		passedFlags = append(passedFlags, "-"+f.Name)
	})

	if len(passedFlags) > 0 {
		errors = append(errors, "flags passed: "+strings.Join(passedFlags, ", "))
	}

	for _, message := range errors {
		//nolint:errcheck // There isn't much to do if we can't print to the output.
		fmt.Fprintln(flag.CommandLine.Output(), message)
	}

	flag.Usage()
}
