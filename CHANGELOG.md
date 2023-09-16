# CHANGELOG

## 1.4.0

* Use goreleaser to release.
* Add arm64 architecture builds.

## 1.3.0 (2021-01-15)

* Added `-include-hex-range` flag. If set, this will include the IP range
  in hexadecimal format. Pull request by Alexander Sinitsyn. GitHub #33.

## 1.2.0 (2020-12-03)

* The output file is now synced before it is closed and the program exits.
  Requested by orang3-juic3. GitHub #30.
* Dependencies have been updated.

## 1.1.0 (2018-12-06)

* The help output is now improved on errors.

## 1.0.0 (2016-11-04)

* Compiled with Go 1.7.3. This fixes issues on macOS Sierra. Closes #6.
* Updated to new version of github.com/mikioh/ipaddr

## 0.0.1 (2014-12-09)

* Initial release.
