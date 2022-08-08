// Package convert transforms a GeoIP2/GeoLite2 CSV to various formats.
package convert

import (
	"encoding/csv"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"

	"github.com/mikioh/ipaddr"
	"github.com/pkg/errors"
)

type (
	headerFunc func([]string) []string
	lineFunc   func(*ipaddr.Prefix, []string) []string
)

// ConvertFile converts the MaxMind GeoIP2 or GeoLite2 CSV file `inputFile` to
// `outputFile` file using a different representation of the network. The
// representation can be specified by setting one or more of `cidr`,
// `ipRange`, `intRange` or `hexRange` to true. If none of these are set to true, it will
// strip off the network information.
func ConvertFile( //nolint: revive // stutters, should fix
	inputFile string,
	outputFile string,
	cidr bool,
	ipRange bool,
	intRange bool,
	hexRange bool,
) error {
	outFile, err := os.Create(filepath.Clean(outputFile))
	if err != nil {
		return errors.Wrapf(err, "error creating output file (%s)", outputFile)
	}
	defer outFile.Close() //nolint: gosec

	inFile, err := os.Open(inputFile) //nolint: gosec
	if err != nil {
		return errors.Wrapf(err, "error opening input file (%s)", inputFile)
	}
	defer inFile.Close() //nolint: gosec

	err = Convert(inFile, outFile, cidr, ipRange, intRange, hexRange)
	if err != nil {
		return err
	}
	err = outFile.Sync()
	if err != nil {
		return errors.Wrapf(err, "error syncing file (%s)", outputFile)
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
	makeLine := func(_ *ipaddr.Prefix, orig []string) []string { return orig }

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
	return func(network *ipaddr.Prefix, line []string) []string {
		return second(network, first(network, line))
	}
}

func cidrHeader(orig []string) []string {
	return append([]string{"network"}, orig...)
}

func cidrLine(network *ipaddr.Prefix, orig []string) []string {
	return append([]string{network.String()}, orig...)
}

func rangeHeader(orig []string) []string {
	return append([]string{"network_start_ip", "network_last_ip"}, orig...)
}

func rangeLine(network *ipaddr.Prefix, orig []string) []string {
	return append(
		[]string{network.IP.String(), network.Last().String()},
		orig...,
	)
}

func intRangeHeader(orig []string) []string {
	return append([]string{"network_start_integer", "network_last_integer"}, orig...)
}

func intRangeLine(network *ipaddr.Prefix, orig []string) []string {
	startInt := new(big.Int)

	startInt.SetBytes(canonicalizeIP(network.IP))

	endInt := new(big.Int)
	endInt.SetBytes(canonicalizeIP(network.Last()))

	return append(
		[]string{startInt.String(), endInt.String()},
		orig...,
	)
}

func hexRangeHeader(orig []string) []string {
	return append([]string{"network_start_hex", "network_last_hex"}, orig...)
}

func hexRangeLine(network *ipaddr.Prefix, orig []string) []string {
	startInt := new(big.Int)

	startInt.SetBytes(canonicalizeIP(network.IP))

	endInt := new(big.Int)
	endInt.SetBytes(canonicalizeIP(network.Last()))

	return append(
		[]string{startInt.Text(16), endInt.Text(16)},
		orig...,
	)
}

func canonicalizeIP(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
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
		return errors.Wrap(err, "error reading CSV header")
	}

	newHeader := makeHeader(header[1:])
	err = writer.Write(newHeader)
	if err != nil {
		return errors.Wrap(err, "error writing CSV header")
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "error reading CSV")
		}

		p, err := makePrefix(record[0])
		if err != nil {
			return err
		}
		err = writer.Write(makeLine(p, record[1:]))
		if err != nil {
			return errors.Wrap(err, "error writing CSV")
		}
	}

	writer.Flush()
	return errors.Wrap(writer.Error(), "error writing CSV")
}

func makePrefix(network string) (*ipaddr.Prefix, error) {
	_, ipn, err := net.ParseCIDR(network)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing network (%s)", network)
	}
	return ipaddr.NewPrefix(ipn), nil
}
