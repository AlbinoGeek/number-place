# Number Place

[![MIT License](https://img.shields.io/github/license/AlbinoGeek/number-place.svg)](https://github.com/AlbinoGeek/number-place/blob/master/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AlbinoGeek/number-place.svg)](https://github.com/AlbinoGeek/number-place)
[![Maintainability](https://api.codeclimate.com/v1/badges/be0523753694eee85927/maintainability)](https://codeclimate.com/github/AlbinoGeek/number-place/maintainability)  
[![CI](https://github.com/AlbinoGeek/number-place/workflows/CI/badge.svg?branch=main)](#)
[![GoReportCard](https://goreportcard.com/badge/github.com/AlbinoGeek/number-place)](https://goreportcard.com/report/github.com/AlbinoGeek/number-place)

## Features

### Puzzles Supported

- [ ] Sodoku
  - [ ] Standard Grid (9x9, 3x3 subgrids)
    - [ ] Classic
  - [ ] Mini Grid (6x6, 3x2 subgrids)
  - [ ] Giant Grid (16x16, 4x4 subgrids)
- [ ] Constrainted-Based Gridded Sodoku:
  - [ ] Standard Grid (9x9, 3x3 subgrids)
    - [ ] Greater Than
    - [ ] XV

If you know of a puzzle type not listed above, please [request it!](https://github.com/AlbinoGeek/number-place/issues/new?assignees=AlbinoGeek&labels=enhancement&template=feature-request.md&title=%5BFEATURE+REQUEST%5D)

## Usage

[Download a Release](https://github.com/AlbinoGeek/number-place/releases) or [Get the Sources](#building-from-source)

## Building From Source

1. Install Golang
1. Clone the repository
1. Run `make`

```bash
# Install Golang [apt / yum / dnf install golang]
$ git clone https://github.com/AlbinoGeek/number-place
$ cd number-place
$ make all
```

The binary will be built into the `_dist` directory.

## License

This project is licensed under the terms of the [MIT License](/LICENSE).
