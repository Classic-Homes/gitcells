# GitCells Documentation Viewer

A native Go application that serves GitCells documentation locally with a beautiful web interface.

## Features

- 🚀 **Fast**: Built with Go for instant startup
- 📦 **Self-contained**: Documentation embedded in the binary
- 🎨 **Beautiful UI**: Clean, modern interface with navigation
- 🔍 **Searchable**: Built-in search functionality
- 💻 **Cross-platform**: Works on macOS, Windows, and Linux
- 🌐 **No internet required**: Everything runs locally

## Quick Start

### Using Make (Recommended)

```bash
# Build and run documentation
make docs

# Or separately:
make docs-build  # Build the viewer
make docs-run    # Run the viewer
```

### Manual Build

```bash
# Download dependencies
go mod download

# Build the documentation viewer
go build -o gitcells-docs cmd/gitcells-docs-simple/main.go

# Run it
./gitcells-docs
```

The documentation will open in your default browser at http://localhost:8080

## Installation

### System-wide Installation

```bash
# Using make (requires sudo)
sudo make docs-install

# Then run from anywhere
gitcells-docs
```

### Manual Installation

```bash
# Build
go build -o gitcells-docs cmd/gitcells-docs-simple/main.go

# Install
sudo cp gitcells-docs /usr/local/bin/
sudo chmod 755 /usr/local/bin/gitcells-docs
```

## Usage

### Command Line Options

```bash
# Default port (8080)
gitcells-docs

# Custom port
gitcells-docs 3000
```

### Navigation

- Use the sidebar to browse documentation sections
- Click any link to navigate
- Use the search box (press Enter to search)
- The current page is highlighted in the sidebar

## Architecture

The documentation viewer:
1. Embeds all markdown files at compile time
2. Starts a local web server
3. Converts markdown to HTML on request
4. Opens your default browser automatically

No external files or internet connection needed!

## Building for Distribution

### All Platforms

```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -o gitcells-docs-mac-intel cmd/gitcells-docs-simple/main.go
GOOS=darwin GOARCH=arm64 go build -o gitcells-docs-mac-arm cmd/gitcells-docs-simple/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o gitcells-docs.exe cmd/gitcells-docs-simple/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o gitcells-docs-linux cmd/gitcells-docs-simple/main.go
```

## Development

### Updating Documentation

1. Edit markdown files in `/docs`
2. Rebuild the viewer: `make docs-build`
3. Documentation is automatically embedded

### Adding New Pages

1. Create new `.md` file in appropriate directory
2. Update navigation in `cmd/gitcells-docs-simple/main.go`
3. Rebuild

## Troubleshooting

### Port Already in Use

If port 8080 is busy:
```bash
gitcells-docs 8090  # Use different port
```

### Browser Doesn't Open

The viewer will print the URL - open it manually:
```
http://localhost:8080
```

### Can't Find Documentation

Documentation is embedded at build time. If you see 404 errors, rebuild:
```bash
make clean
make docs-build
```

## Technical Details

- Built with Go's `embed` package
- Uses `goldmark` for markdown rendering
- Simple HTTP server with template rendering
- No JavaScript framework - just vanilla JS
- Responsive design with CSS

## License

Same as GitCells - MIT License