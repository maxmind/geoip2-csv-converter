// Package convert transforms a GeoIP2/GeoLite2 CSV to various formats.
package convert

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"os"

	"go4.org/netipx"
)

type (
	headerFunc func([]string) []string
	lineFunc   func(netip.Prefix, []string) []string
)

// ConvertFile converts the MaxMind GeoIP2 or GeoLite2 CSV file `inputFile` to
// `outputFile` file using a different representation of the network. The
// representation can be specified by setting one or more of `cidr`,
// `ipRange`, `intRange` or `hexRange` to true. If none of these are set to true, it will
// strip off the network information.
func ConvertFile( // nolint: golint
	inputFile string,
	outputFile string,
	cidr bool,
	ipRange bool,
	intRange bool,
	hexRange bool,
) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output file (%s): %w", outputFile, err)
	}
	defer outFile.Close() // nolint: gosec

	inFile, err := os.Open(inputFile) // nolint: gosec
	if err != nil {
		return fmt.Errorf("opening input file (%s): %w", inputFile, err)
	}
	defer inFile.Close() // nolint: gosec

	err = Convert(inFile, outFile, cidr, ipRange, intRange, hexRange)
	if err != nil {
		return err
	}
	err = outFile.Sync()
	if err != nil {
		return fmt.Errorf("syncing file (%s): %w", outputFile, err)
	}
	return nil
}

// Convert writes the MaxMind GeoIP2 or GeoLite2 CSV in the `input` io.Reader
// to the Writer `output` using the network representation specified by setting
// `cidr`, ipRange`, or `intRange` to true. If none of these are set to true,
// it will strip off the network information.
func Convert(
	input io.Reader,
	output io.Writer,
	cidr bool,
	ipRange bool,
	intRange bool,
	hexRange bool,
) error {
	makeHeader := func(orig []string) []string { return orig }
	makeLine := func(_ netip.Prefix, orig []string) []string { return orig }

	if hexRange {
		makeHeader = addHeaderFunc(makeHeader, hexRangeHeader)
		makeLine = addLineFunc(makeLine, hexRangeLine)
	}

	if intRange {
		makeHeader = addHeaderFunc(makeHeader, intRangeHeader)
		makeLine = addLineFunc(makeLine, intRangeLine)
	}

	if ipRange {
		makeHeader = addHeaderFunc(makeHeader, rangeHeader)
		makeLine = addLineFunc(makeLine, rangeLine)
	}

	if cidr {
		makeHeader = addHeaderFunc(makeHeader, cidrHeader)
		makeLine = addLineFunc(makeLine, cidrLine)
	}

	return convert(input, output, makeHeader, makeLine)
}

func addHeaderFunc(first, second headerFunc) headerFunc {
	return func(header []string) []string {
		return second(first(header))
	}
}

func addLineFunc(first, second lineFunc) lineFunc {
	return func(network netip.Prefix, line []string) []string {
		return second(network, first(network, line))
	}
}

func cidrHeader(orig []string) []string {
	return append([]string{"network"}, orig...)
}

func cidrLine(network netip.Prefix, orig []string) []string {
	return append([]string{network.String()}, orig...)
}

func rangeHeader(orig []string) []string {
	return append([]string{"network_start_ip", "network_last_ip"}, orig...)
}

func rangeLine(network netip.Prefix, orig []string) []string {
	return append(
		[]string{network.Addr().String(), netipx.PrefixLastIP(network).String()},
		orig...,
	)
}

func intRangeHeader(orig []string) []string {
	return append([]string{"network_start_integer", "network_last_integer"}, orig...)
}

func intRangeLine(network netip.Prefix, orig []string) []string {
	startInt := new(big.Int)

	startInt.SetBytes(network.Addr().AsSlice())

	endInt := new(big.Int)
	endInt.SetBytes(netipx.PrefixLastIP(network).AsSlice())

	return append(
		[]string{startInt.String(), endInt.String()},
		orig...,
	)
}

func hexRangeHeader(orig []string) []string {
	return append([]string{"network_start_hex", "network_last_hex"}, orig...)
}

func hexRangeLine(network netip.Prefix, orig []string) []string {
	startInt := new(big.Int)

	startInt.SetBytes(network.Addr().AsSlice())

	endInt := new(big.Int)
	endInt.SetBytes(netipx.PrefixLastIP(network).AsSlice())

	return append(
		[]string{startInt.Text(16), endInt.Text(16)},
		orig...,
	)
}

func convert(
	input io.Reader,
	output io.Writer,
	makeHeader headerFunc,
	makeLine lineFunc,
) error {
	reader := csv.NewReader(input)
	writer := csv.NewWriter(output)

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("reading CSV header: %w", err)
	}

	newHeader := makeHeader(header[1:])
	err = writer.Write(newHeader)
	if err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("reading CSV: %w", err)
		}

		p, err := makePrefix(record[0])
		if err != nil {
			return err
		}
		err = writer.Write(makeLine(p, record[1:]))
		if err != nil {
			return fmt.Errorf("writing CSV: %w", err)
		}
	}

	writer.Flush()
	return nil
}

func makePrefix(network string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(network)
	if err != nil {
		return prefix, fmt.Errorf("parsing network (%s): %w", network, err)
	}
	return prefix, nil
}
