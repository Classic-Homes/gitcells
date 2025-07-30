# GitCells Documentation

GitCells includes comprehensive user documentation that can be viewed using our native documentation viewer.

## Quick Start

### View Documentation

```bash
# Install dependencies (one time only)
make docs-install

# Serve documentation locally
make docs-serve

# Or build static site
make docs-build && make docs-open
```

The documentation uses MkDocs with Material theme for:
- Beautiful, professional appearance
- Fast search and navigation
- Mobile-responsive design
- Dark/light mode toggle
- Offline support

### Alternative Methods (Optional)

If you prefer not to use the native viewer:

#### 1. Direct File Access
Browse the markdown files directly in the `docs/` directory:
- `docs/index.md` - Overview
- `docs/getting-started/` - Installation and quick start
- `docs/guides/` - User guides and tutorials
- `docs/reference/` - Command and configuration reference

#### 2. GitHub/GitLab
View rendered documentation online by browsing to the docs folder in your repository.

## Documentation Structure

```
docs/
├── index.md                    # Home page
├── getting-started/           
│   ├── installation.md         # How to install GitCells
│   ├── quickstart.md           # 5-minute quick start guide
│   └── concepts.md             # Core concepts explained
├── guides/                    
│   ├── converting.md           # Converting Excel files
│   ├── tracking.md             # Tracking changes
│   ├── collaboration.md        # Working with teams
│   ├── conflicts.md            # Resolving conflicts
│   ├── auto-sync.md           # Setting up automation
│   └── use-cases.md           # Real-world examples
└── reference/                 
    ├── commands.md             # All commands reference
    ├── configuration.md        # Configuration options
    ├── json-format.md         # JSON format specification
    └── troubleshooting.md     # Common issues and solutions
```

## For Contributors

To update documentation:
1. Edit markdown files in the `docs/` directory
2. Rebuild the viewer: `make docs-build`
3. Test locally: `make docs-run`

## How It Works

The GitCells documentation uses MkDocs - the industry standard for documentation:

1. **Markdown files** in the `docs/` directory
2. **MkDocs configuration** in `mkdocs.yml`
3. **Material theme** for beautiful appearance
4. **Static site generation** for fast loading

## Building the Documentation

```bash
# One-time setup
pip install -r requirements.txt

# Development server (auto-reload)
mkdocs serve

# Build static site
mkdocs build

# Deploy to GitHub Pages
mkdocs gh-deploy
```

## Key Features

- 🎨 **Material Design**: Professional, modern appearance
- 🔍 **Instant Search**: Fast, client-side search
- 📱 **Mobile Ready**: Responsive design for all devices
- 🌙 **Dark Mode**: Built-in dark/light theme toggle
- 🌐 **Offline Support**: Works without internet
- ⚡ **Fast**: Static site generation for speed

## Getting Help

- Check the [Troubleshooting Guide](docs/reference/troubleshooting.md)
- Browse [Common Use Cases](docs/guides/use-cases.md)
- File issues on GitHub

The documentation is designed to help both new users get started quickly and experienced users master advanced features.