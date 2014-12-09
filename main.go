package main

import (
	"flag"
	"fmt"

	"github.com/maxmind/geoip2-csv-converter/convert"
)

func main() {
	required := "REQUIRED"
	input := flag.String("block-file", required, "The path to the block CSV file to use as input")
	output := flag.String("output-file", required, "The path to the output CSV")
	ipRange := flag.Bool("include-range", false, "Include the IP range of the network in string format")
	intRange := flag.Bool("include-integer-range", false, "Include the IP range of the network in integer format")
	cidr := flag.Bool("include-cidr", false, "Include the network in CIDR format")

	flag.Parse()

	if *input == required || *output == required {
		printHelp()
		return
	}

	if !*ipRange && !*intRange && !*cidr {
		printHelp()
		return
	}

	err := convert.ConvertFile(*input, *output, *cidr, *ipRange, *intRange)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func printHelp() {
	flag.PrintDefaults()
	fmt.Println("\nAt least one of -include-* param is required")

}
