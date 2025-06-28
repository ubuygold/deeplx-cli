# DeepLX CLI

> This project is a CLI wrapper for the [OwO-Network/DeepLX](https://github.com/OwO-Network/DeepLX) project, providing command-line access to the DeepLX translation API.

This is a command-line tool for the DeepLX translation API.

## ✨ New Features

### Auto-Copy to Clipboard
After successful translation, the result is automatically copied to your system clipboard. 
This allows for instant pasting into any application.

To use:
```bash
deeplx-cli -text "Hello world" -s EN -t ZH
# Translation is now in your clipboard - just paste!
```

### Comprehensive Testing
We've added full test coverage including:
- Configuration loading tests
- Translation logic tests
- Clipboard integration tests
- End-to-end workflow tests

Run tests with:
```bash
make test
# or
go test -v ./...
```

## 📥 Installation

### Using Go Install
```bash
go install github.com/ubuygold/deeplx-cli@latest
```

### From Source
```bash
git clone https://github.com/ubuygold/deeplx-cli.git
cd deeplx-cli
make build
sudo make install  # Installs to /usr/local/bin
```

## 🛠️ Usage

### Translate via Command-Line Arguments
```bash
deeplx-cli -source_lang en -target_lang zh Hello,world!
# Or using shorthand
deeplx-cli -s en -t zh Hello,world!
```

### Translate via Standard Input
```bash
echo "Hello, world!" | deeplx-cli -source_lang en -target_lang zh
# Or using shorthand
echo "Hello, world!" | deeplx-cli -s en -t zh
```

### Clipboard Integration
The translation result is automatically copied to your clipboard. 
You can immediately paste it anywhere after the command completes.

## 📋 Configuration File
DeepLX CLI attempts to read configuration from `~/.deeplx-cli.yml`. 
If the file does not exist, a default configuration will be automatically generated.
 
> **Note**: The underlying API is provided by the [OwO-Network/DeepLX](https://github.com/OwO-Network/DeepLX) project.
 
**Configuration Example (`~/.deeplx-cli.yml`):**
```yaml
# DeepLX CLI Configuration
# DeepLX API Address
deeplx_api: "https://deeplx.vercel.app/translate"
# Default Source Language
source_lang: "auto"
# Default Target Language
target_lang: "EN"
```

## 🧪 Development

### Building from Source
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests with coverage report
make test

# Show version
./deeplx-cli -version
```

### Testing
We provide comprehensive test coverage:
```bash
# Run all tests
make test

# Run specific test
go test -v -run TestEndToEnd
```

Test coverage includes:
- Configuration loading
- Translation logic
- Clipboard operations
- End-to-end workflow
- Helper functions
