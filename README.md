GeoIP2 CSV Format Converter
---------------------------

This is a simple utility for converting the MaxMind GeoIP2 and GeoLite2 CSVs
to different formats for representing IP addresses such as IP ranges or
integer ranges.

Compiled binaries for Linux/x86_64, Windows, and Mac OS X can be downloaded
from the GitHub releases page.

Usage
=====


Required:

* -block-file=[FILENAME] - The name of the block CSV file to use as input.
* -output-file=[FILENAME] - The file name to the output CSV

In addition, at least one of these is required:

* -include-cidr - Include the network in CIDR format
* -include-range - Include the IP range of the network in string format
* -include-integer-range - Include the IP range of the network in integer format
* -include-hex-range - Include the IP range of the network in hexadecimal format

Output
======

### CIDR (-include-cidr)

This will include the network in CIDR notation in the `network` column as it
is in the original CSV.

### Range (-include-range)

This adds `network_start_ip` and `network_last_ip` columns. These
are string representations of the first and last IP address in the network.

### Integer Range (-include-integer-range)

This adds `network_start_integer` and `network_last_integer` columns. These
are integer representations of the first and last IP address in the network.

### Integer Range (-include-hex-range)

This adds `network_start_hex` and `network_last_hex` columns. These
are hexadecimal representations of the first and last IP address in the network.

Copyright and License
=====================

This software is Copyright (c) 2014 by MaxMind, Inc.

This is free software, licensed under the Apache License, Version 2.0.
