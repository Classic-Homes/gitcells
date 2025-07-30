# GitCells Documentation

This directory contains comprehensive user documentation for GitCells.

## Viewing Documentation

### Primary Method: MkDocs (Recommended)

GitCells uses MkDocs with Material theme for professional documentation:

```bash
# From project root - install dependencies (one time)
pip install -r requirements.txt

# Serve locally with live reload
mkdocs serve

# Or build static site
mkdocs build
```

**Professional and powerful!** MkDocs provides:
- Beautiful Material Design interface
- Instant search functionality
- Mobile-responsive design
- Dark/light mode toggle
- Live reload during development

### Alternative Viewing Methods

#### GitHub/GitLab
If your repository is on GitHub or GitLab, the markdown files will be automatically rendered when browsing the repository online.

#### Markdown Viewer
Use any markdown viewer or editor:
- VS Code (with Markdown Preview)
- MacDown (macOS)
- Typora
- Obsidian

#### Command Line
For quick reference in the terminal:

```bash
# Using mdless (install with: gem install mdless)
mdless docs/getting-started/quickstart.md

# Using glow (install with: brew install glow)
glow docs/getting-started/quickstart.md

# Using bat (install with: brew install bat)
bat docs/getting-started/quickstart.md
```

## Documentation Structure

```
docs/
├── index.md                    # Home page
├── README.md                   # This file
├── serve.py                    # Documentation server
│
├── getting-started/           # New user guides
│   ├── installation.md        # Installation instructions
│   ├── quickstart.md          # 5-minute quick start
│   └── concepts.md            # Core concepts
│
├── guides/                    # How-to guides
│   ├── converting.md          # Converting files
│   ├── tracking.md            # Tracking changes
│   ├── collaboration.md       # Team workflows
│   ├── conflicts.md           # Conflict resolution
│   ├── auto-sync.md          # Automation setup
│   └── use-cases.md          # Real-world examples
│
└── reference/                 # Reference material
    ├── commands.md            # Command reference
    ├── configuration.md       # Config options
    ├── json-format.md        # JSON specification
    └── troubleshooting.md    # Problem solving
```

## For Documentation Contributors

### Style Guide

- Use clear, concise language
- Include practical examples
- Start with the most common use case
- Use proper markdown formatting
- Test all code examples

### Adding New Pages

1. Create a new `.md` file in the appropriate directory
2. Add navigation entry in `mkdocs.yml` under the `nav:` section
3. Update the index.md with a link if needed
4. Test with live reload: `mkdocs serve`

### Building for Distribution

Create a static site for deployment:

```bash
# Build static site
mkdocs build

# Files are generated in site/ directory
# Deploy site/ to any web server or GitHub Pages
```

## Development Workflow

For contributors working on documentation:

```bash
# Install dependencies (one time)
pip install -r requirements.txt

# Start development server
mkdocs serve

# Documentation is available at http://localhost:8000
# Changes to .md files reload automatically
```

## Questions?

- Check the [Troubleshooting Guide](reference/troubleshooting.md)
- File an issue on GitHub
- Visit gitcells.com