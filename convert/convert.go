package convert

import (
	"encoding/csv"
	"io"
	"math/big"
	"net"
	"os"

	"github.com/mikioh/ipaddr"
)

type headerFunc func([]string) []string
type lineFunc func(ipaddr.Prefix, []string) []string

// ConvertFile converts the MaxMind GeoIP2 or GeoLite2 CSV file `inputFile` to
// `outputFile` file using a different representation of the network. The
// representation can be specified by setting one or more of `cidr`,
// `ipRange`, or `intRange` to true. If none of these are set to true, it will
// strip off the network information.
func ConvertFile(
	inputFile string,
	outputFile string,
	cidr bool,
	ipRange bool,
	intRange bool,
) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	return Convert(inFile, outFile, cidr, ipRange, intRange)
}

// Convert writes the MaxMind GeoIP2 or GeoLite2 CSV in the `input` io.Reader
// to the Writer `output` using the network represenation specified by setting
// `cidr`, ipRange`, or `intRange` to true. If none of these are set to true,
// it will strip off the network information.
func Convert(
	input io.Reader,
	output io.Writer,
	cidr bool,
	ipRange bool,
	intRange bool,
) error {

	makeHeader := func(orig []string) []string { return orig }
	makeLine := func(_ ipaddr.Prefix, orig []string) []string { return orig }

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

func addHeaderFunc(first headerFunc, second headerFunc) headerFunc {
	return func(header []string) []string {
		return second(first(header))
	}
}

func addLineFunc(first lineFunc, second lineFunc) lineFunc {
	return func(network ipaddr.Prefix, line []string) []string {
		return second(network, first(network, line))
	}
}

func cidrHeader(orig []string) []string {
	return append([]string{"network"}, orig...)
}

func cidrLine(network ipaddr.Prefix, orig []string) []string {
	return append([]string{network.String()}, orig...)
}

func rangeHeader(orig []string) []string {
	return append([]string{"start_ip", "end_ip"}, orig...)
}

func rangeLine(network ipaddr.Prefix, orig []string) []string {
	return append(
		[]string{network.Addr().String(), network.LastAddr().String()},
		orig...,
	)
}

func intRangeHeader(orig []string) []string {
	return append([]string{"start_integer", "end_integer"}, orig...)
}

func intRangeLine(network ipaddr.Prefix, orig []string) []string {
	startInt := new(big.Int)
	startInt.SetBytes(network.Addr())

	endInt := new(big.Int)
	endInt.SetBytes(network.LastAddr())

	return append(
		[]string{startInt.String(), endInt.String()},
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
		return err
	}

	newHeader := makeHeader(header[1:])
	err = writer.Write(newHeader)
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		p, err := makePrefix(record[0])
		if err != nil {
			return err
		}
		writer.Write(makeLine(p, record[1:]))
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}
	return nil
}

func makePrefix(network string) (ipaddr.Prefix, error) {
	_, ipn, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}
	nbits, _ := ipn.Mask.Size()
	return ipaddr.NewPrefix(ipn.IP, nbits)
}
