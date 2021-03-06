# keyring

[![Build Status](https://github.com/martinohmann/keyring/workflows/build/badge.svg)](https://github.com/martinohmann/keyring/actions?query=workflow%3Abuild)
[![codecov](https://codecov.io/gh/martinohmann/keyring/branch/master/graph/badge.svg)](https://codecov.io/gh/martinohmann/keyring)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinohmann/keyring)](https://goreportcard.com/report/github.com/martinohmann/keyring)
[![GoDoc](https://godoc.org/github.com/martinohmann/keyring?status.svg)](https://godoc.org/github.com/martinohmann/keyring)
![GitHub](https://img.shields.io/github/license/martinohmann/keyring?color=orange)

Simple commandline wrapper around [zalando/go-keyring](https://github.com/zalando/go-keyring).

## Installation

Download the [latest binary release](https://github.com/martinohmann/keyring/releases)
for your platform or install via `go get`:

```bash
$ go install github.com/martinohmann/keyring/cmd/keyring@latest
```

## Usage

Write secret from keyring to stdout:

```bash
$ keyring get myservice myuser
```

Pipe secret from keyring into another command:

```bash
$ keyring get myservice myuser | cat
```

Store secret in keyring, read secret from stdin:

```bash
$ echo -n "supersecret" | keyring set myservice myuser
```

Store secret in keyring, interactive secret prompt:

```bash
$ keyring set myservice myuser

enter secret:
```

Delete secret from keyring:

```bash
$ keyring delete myservice myuser
```

## License

The source code of keyring is released under the MIT License. See the bundled
LICENSE file for details.
