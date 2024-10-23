# Admina SysUtils

Admina SysUtils is a command-line tool for automating management tasks.

## Installation

```
go install github.com/moneyforward-i/admina-sysutils/cmd/admina-sysutils@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Usage

```
admina-sysutils --help
```

## Development

### Requirements

- Go 1.20 or higher

### Running during development

To run the program during development, use the following command:

```
go run ./cmd/admina-sysutils/main.go
```

### Building

To build for your local environment:

```
make build
```

To build for all platforms (Windows, Mac, Linux):

```
make build-all
```

### Build output location

- Local build: The `admina-sysutils` binary (or `admina-sysutils.exe` for Windows) will be generated in the `bin` directory of the project.
- Cross-compilation: The following files will be generated in the `bin` directory of the project:
  - Linux: `admina-sysutils-linux-amd64`
  - Mac: `admina-sysutils-darwin-amd64`
  - Windows: `admina-sysutils-windows-amd64.exe`

### Running the built file

To run the built file, use the following command in the terminal:

# For Linux and Mac

```
./bin/admina-sysutils
```

# For Windows

```
.\bin\admina-sysutils.exe
```

For cross-compiled files, replace the file name accordingly.

### Testing

```
make test
```

## License

This project is released under the MIT License. See the [LICENSE](LICENSE) file for details.

go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
