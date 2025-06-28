# DeepLX CLI

This is a command-line tool for the DeepLX translation API.

## Features

*   Supports reading configuration from `~/.deeplx-cli.yml`.
*   Command-line arguments (including `-s` for source language and `-t` for target language) override configuration settings.
*   Supports getting translation text from standard input or direct command-line arguments.
*   Automatically generates a default configuration file if it does not exist.
*   Mandatory specification of source and target languages.

## Installation

To build and install DeepLX CLI, you have two options:

### Using `go install` (Recommended)

Ensure you have Go installed on your system. Then, you can install the DeepLX CLI directly using `go install`:

```bash
go install github.com/ubuygold/deeplx-cli@latest
```

This command will download the source code, compile it, and place the executable in your `$GOPATH/bin` or `$GOBIN` directory. Make sure this directory is included in your system's PATH environment variable to run `deeplx-cli` from any location.

### Using `go build`

Alternatively, you can build the executable manually:

```bash
go build -o deeplx-cli
```

This will generate an executable named `deeplx-cli` in your current directory. You can then move it to a directory included in your PATH to run it from any location.

## Usage

### Translate via Command-Line Arguments

You can use either the full parameter names or their shorthand versions:

```bash
./deeplx-cli -source_lang en -target_lang zh Hello,world!
# Or using shorthand
./deeplx-cli -s en -t zh Hello,world!
```

### Translate via Standard Input

```bash
echo "Hello, world!" | ./deeplx-cli -source_lang en -target_lang zh
# Or using shorthand
echo "Hello, world!" | ./deeplx-cli -s en -t zh
```

### Configuration File

DeepLX CLI attempts to read configuration from `~/.deeplx-cli.yml`. If the file does not exist, a default configuration will be automatically generated.

**Configuration Example (`~/.deeplx-cli.yml`):**

```yaml
# DeepLX CLI Configuration Example
# DeepLX API Address
deeplx_api: "https://deeplx.vercel.app/translate"
# Default Source Language
source_lang: "auto"
# Default Target Language
target_lang: "EN"
```

### Mandatory Language Parameters

The `source_lang` (`-s`) and `target_lang` (`-t`) parameters are mandatory. You must specify them.

### Text Input Considerations

When providing text directly as command-line arguments, quotes are generally not required for single words or text without spaces. For example:

```bash
./deeplx-cli -s en -t zh Hello
```

However, if the text contains spaces or special characters, it should be enclosed in quotes to be treated as a single argument:

```bash
./deeplx-cli -s auto -t EN "Hello World"
```

## Development

### Building from Source

#### Prerequisites
- Go 1.22 or later

#### Local Build
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Show version
./deeplx-cli -version
```

#### Using Makefile
This project includes a Makefile with various useful targets:

```bash
make help          # Show all available targets
make build         # Build binary for current platform
make build-all     # Cross-compile for all platforms
make test          # Run tests
make test-race     # Run tests with race detector
make vet           # Run go vet
make fmt           # Format code
make deps          # Download and verify dependencies
make checksums     # Generate checksums for all binaries
make install       # Install binary to /usr/local/bin
make clean         # Remove build artifacts
```

### CI/CD

This project uses GitHub Actions for automated building and releasing:

#### Workflows

1. **Release Workflow** (`.github/workflows/go.yml`)
   - Triggers on tag pushes (e.g., `v1.0.0`)
   - Cross-compiles for multiple platforms:
     - Linux (amd64, arm64)
     - Windows (amd64, arm64)
     - macOS (amd64, arm64)
     - FreeBSD (amd64)
   - Generates SHA256 checksums for all binaries
   - Creates GitHub releases with all artifacts

2. **Test Workflow** (`.github/workflows/test.yml`)
   - Triggers on pushes to main branches and pull requests
   - Runs tests and code quality checks
   - Tests cross-compilation for all target platforms

#### Creating a Release

To create a new release:

1. Create and push a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The GitHub Actions workflow will automatically:
   - Build binaries for all platforms
   - Generate checksums
   - Create a GitHub release
   - Upload all artifacts

#### Supported Platforms

The CI automatically builds for these platforms:
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`
- `windows/arm64`
- `darwin/amd64` (macOS Intel)
- `darwin/arm64` (macOS Apple Silicon)
- `freebsd/amd64`

All binaries include version information and are optimized for size.