// Package convert transforms a GeoIP2/GeoLite2 CSV to various formats.
package convert

import (
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

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
func ConvertFile( //nolint: revive // too late to change name
	inputFile string,
	outputFile string,
	cidr bool,
	ipRange bool,
	intRange bool,
	hexRange bool,
) error {
	outFile, err := os.Create(filepath.Clean(outputFile))
	if err != nil {
		return fmt.Errorf("creating output file (%s): %w", outputFile, err)
	}

	inFile, err := os.Open(filepath.Clean(inputFile))
	if err != nil {
		outFile.Close()
		return fmt.Errorf("opening input file (%s): %w", inputFile, err)
	}

	err = Convert(inFile, outFile, cidr, ipRange, intRange, hexRange)
	if err != nil {
		inFile.Close()
		outFile.Close()
		return err
	}
	err = outFile.Sync()
	if err != nil {
		inFile.Close()
		outFile.Close()
		return fmt.Errorf("syncing file (%s): %w", outputFile, err)
	}
	if err := inFile.Close(); err != nil {
		return fmt.Errorf("closing file (%s): %w", inputFile, err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("closing file (%s): %w", outputFile, err)
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
	return append(
		[]string{
			toHex(network.Addr()),
			toHex(netipx.PrefixLastIP(network)),
		},
		orig...,
	)
}

func toHex(ip netip.Addr) string {
	return strings.TrimPrefix(hex.EncodeToString(ip.AsSlice()), "0")
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
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("reading CSV: %w", err)
		}

		prefix, err := netip.ParsePrefix(record[0])
		if err != nil {
			return fmt.Errorf("parsing network (%s): %w", record[0], err)
		}

		err = writer.Write(makeLine(prefix, record[1:]))
		if err != nil {
			return fmt.Errorf("writing CSV: %w", err)
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("flushing CSV: %w", err)
	}

	return nil
}
