# GitCells Documentation

This directory contains the GitCells documentation built with MkDocs and Material theme.

## Viewing Documentation

### Online
The documentation is automatically deployed to GitHub Pages:
https://classic-homes.github.io/gitcells/

### Local Development

#### Using Docker (Recommended)
```bash
# Start the documentation server
docker compose up docs

# Visit http://localhost:8000
```

#### Using Python
```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
pip install mkdocs mkdocs-material

# Start development server
mkdocs serve

# Visit http://localhost:8000
```

## Documentation Structure

```
docs/
├── index.md                    # Home page
├── getting-started/           # Installation and quick start guides
├── user-guide/               # Detailed user documentation
├── reference/                # API and command references
└── development/              # Developer documentation
```

## Making Changes

1. Edit markdown files in the appropriate directory
2. Changes are automatically reloaded if using `mkdocs serve`
3. Test the build: `mkdocs build --strict`
4. Submit a pull request

## Writing Guidelines

- Use clear, simple language
- Include code examples
- Add screenshots for UI features
- Test all commands and examples
- Keep non-technical users in mind

## Building for Production

```bash
# Build static site
mkdocs build

# Output will be in site/ directory
```

## CI/CD

- Documentation is automatically built and deployed on push to main branch
- Pull requests trigger documentation validation
- See `.github/workflows/docs.yml` for the deployment pipeline