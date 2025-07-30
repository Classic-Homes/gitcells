#!/bin/bash
# Build a standalone HTML documentation file

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DOCS_DIR="$SCRIPT_DIR/../docs"
OUTPUT_FILE="$SCRIPT_DIR/../dist/gitcells-docs.html"

# Create dist directory if it doesn't exist
mkdir -p "$SCRIPT_DIR/../dist"

# Start building the HTML file
cat > "$OUTPUT_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GitCells Documentation</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
        }
        .container {
            display: flex;
            max-width: 1400px;
            margin: 0 auto;
            min-height: 100vh;
            background: white;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .sidebar {
            width: 300px;
            background: #2c3e50;
            color: white;
            padding: 20px;
            overflow-y: auto;
            position: sticky;
            top: 0;
            height: 100vh;
        }
        .sidebar h1 {
            font-size: 24px;
            margin-bottom: 20px;
            color: #3498db;
        }
        .sidebar h2 {
            font-size: 16px;
            margin-top: 20px;
            margin-bottom: 10px;
            color: #ecf0f1;
        }
        .sidebar ul {
            list-style: none;
        }
        .sidebar li {
            margin-bottom: 5px;
        }
        .sidebar a {
            color: #bdc3c7;
            text-decoration: none;
            display: block;
            padding: 5px 10px;
            border-radius: 3px;
            cursor: pointer;
        }
        .sidebar a:hover {
            background: #34495e;
            color: white;
        }
        .sidebar a.active {
            background: #3498db;
            color: white;
        }
        .content {
            flex: 1;
            padding: 40px;
            max-width: 900px;
        }
        .doc-section {
            display: none;
        }
        .doc-section.active {
            display: block;
        }
        h1 { font-size: 2.5em; color: #2c3e50; margin-bottom: 20px; }
        h2 { font-size: 2em; color: #34495e; margin: 30px 0 15px; }
        h3 { font-size: 1.5em; color: #34495e; margin: 25px 0 10px; }
        p { margin-bottom: 16px; }
        pre {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 16px;
            overflow-x: auto;
            margin-bottom: 16px;
        }
        code {
            background: #f8f9fa;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Consolas', 'Monaco', monospace;
            font-size: 0.9em;
        }
        pre code {
            background: none;
            padding: 0;
        }
        blockquote {
            border-left: 4px solid #3498db;
            padding-left: 16px;
            margin: 16px 0;
            color: #666;
        }
        ul, ol {
            margin-bottom: 16px;
            padding-left: 24px;
        }
        li { margin-bottom: 4px; }
        .content a { color: #3498db; }
        .search-box {
            margin-bottom: 20px;
        }
        .search-box input {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid #34495e;
            border-radius: 4px;
            background: #34495e;
            color: white;
            font-size: 14px;
        }
        .search-box input::placeholder {
            color: #95a5a6;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 16px;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background: #f8f9fa;
            font-weight: 600;
        }
        @media print {
            .sidebar { display: none; }
            .content { margin: 0; padding: 20px; }
            .doc-section { display: block !important; page-break-after: always; }
        }
    </style>
</head>
<body>
    <div class="container">
        <nav class="sidebar">
            <h1>GitCells Docs</h1>
            <div class="search-box">
                <input type="text" placeholder="Search documentation..." onkeyup="searchDocs(event)">
            </div>
            
            <h2>Getting Started</h2>
            <ul>
                <li><a onclick="showDoc('index')">Overview</a></li>
                <li><a onclick="showDoc('installation')">Installation</a></li>
                <li><a onclick="showDoc('quickstart')">Quick Start</a></li>
                <li><a onclick="showDoc('concepts')">Basic Concepts</a></li>
            </ul>
            
            <h2>User Guides</h2>
            <ul>
                <li><a onclick="showDoc('converting')">Converting Files</a></li>
                <li><a onclick="showDoc('tracking')">Tracking Changes</a></li>
                <li><a onclick="showDoc('collaboration')">Team Collaboration</a></li>
                <li><a onclick="showDoc('conflicts')">Resolving Conflicts</a></li>
                <li><a onclick="showDoc('auto-sync')">Auto-sync Setup</a></li>
                <li><a onclick="showDoc('use-cases')">Common Use Cases</a></li>
            </ul>
            
            <h2>Reference</h2>
            <ul>
                <li><a onclick="showDoc('commands')">Commands</a></li>
                <li><a onclick="showDoc('configuration')">Configuration</a></li>
                <li><a onclick="showDoc('json-format')">JSON Format</a></li>
                <li><a onclick="showDoc('troubleshooting')">Troubleshooting</a></li>
            </ul>
        </nav>
        
        <main class="content">
EOF

echo "Building documentation HTML..."

# Function to convert markdown to HTML (basic conversion)
convert_md_to_html() {
    local content="$1"
    # Escape HTML
    content=$(echo "$content" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')
    
    # Convert headers
    content=$(echo "$content" | sed -E 's/^### (.*)/\<h3\>\1\<\/h3\>/g')
    content=$(echo "$content" | sed -E 's/^## (.*)/\<h2\>\1\<\/h2\>/g')
    content=$(echo "$content" | sed -E 's/^# (.*)/\<h1\>\1\<\/h1\>/g')
    
    # Convert code blocks
    content=$(echo "$content" | sed -E 's/```([a-z]*)/\<pre\>\<code\>/g' | sed 's/```/\<\/code\>\<\/pre\>/g')
    
    # Convert inline code
    content=$(echo "$content" | sed -E 's/`([^`]+)`/\<code\>\1\<\/code\>/g')
    
    # Convert bold
    content=$(echo "$content" | sed -E 's/\*\*([^*]+)\*\*/\<strong\>\1\<\/strong\>/g')
    
    # Convert links
    content=$(echo "$content" | sed -E 's/\[([^\]]+)\]\(([^)]+)\)/\<a href="\2"\>\1\<\/a\>/g')
    
    # Convert line breaks to paragraphs
    content=$(echo "$content" | awk 'BEGIN{p=0} /^$/{if(p){print "</p>"}p=0;next} {if(!p){print "<p>";p=1}print}END{if(p)print "</p>"}')
    
    echo "$content"
}

# Add index/overview
echo '<div id="index" class="doc-section active">' >> "$OUTPUT_FILE"
if [ -f "$DOCS_DIR/index.md" ]; then
    convert_md_to_html "$(cat "$DOCS_DIR/index.md")" >> "$OUTPUT_FILE"
fi
echo '</div>' >> "$OUTPUT_FILE"

# Add Getting Started docs
for file in installation quickstart concepts; do
    echo "<div id=\"$file\" class=\"doc-section\">" >> "$OUTPUT_FILE"
    if [ -f "$DOCS_DIR/getting-started/$file.md" ]; then
        convert_md_to_html "$(cat "$DOCS_DIR/getting-started/$file.md")" >> "$OUTPUT_FILE"
    fi
    echo '</div>' >> "$OUTPUT_FILE"
done

# Add Guide docs
for file in converting tracking collaboration conflicts auto-sync use-cases; do
    echo "<div id=\"$file\" class=\"doc-section\">" >> "$OUTPUT_FILE"
    if [ -f "$DOCS_DIR/guides/$file.md" ]; then
        convert_md_to_html "$(cat "$DOCS_DIR/guides/$file.md")" >> "$OUTPUT_FILE"
    fi
    echo '</div>' >> "$OUTPUT_FILE"
done

# Add Reference docs
for file in commands configuration json-format troubleshooting; do
    echo "<div id=\"$file\" class=\"doc-section\">" >> "$OUTPUT_FILE"
    if [ -f "$DOCS_DIR/reference/$file.md" ]; then
        convert_md_to_html "$(cat "$DOCS_DIR/reference/$file.md")" >> "$OUTPUT_FILE"
    fi
    echo '</div>' >> "$OUTPUT_FILE"
done

# Add JavaScript
cat >> "$OUTPUT_FILE" << 'EOF'
        </main>
    </div>
    
    <script>
        // Show specific documentation section
        function showDoc(docId) {
            // Hide all sections
            document.querySelectorAll('.doc-section').forEach(section => {
                section.classList.remove('active');
            });
            
            // Show selected section
            const section = document.getElementById(docId);
            if (section) {
                section.classList.add('active');
            }
            
            // Update active link
            document.querySelectorAll('.sidebar a').forEach(link => {
                link.classList.remove('active');
            });
            event.target.classList.add('active');
            
            // Scroll to top
            window.scrollTo(0, 0);
            
            // Update URL hash
            window.location.hash = docId;
        }
        
        // Search functionality
        function searchDocs(event) {
            const searchTerm = event.target.value.toLowerCase();
            
            if (searchTerm.length < 2) {
                // Reset highlighting
                document.querySelectorAll('.doc-section').forEach(section => {
                    section.innerHTML = section.innerHTML.replace(/<mark>/g, '').replace(/<\/mark>/g, '');
                });
                return;
            }
            
            // Search in all sections
            let found = false;
            document.querySelectorAll('.doc-section').forEach(section => {
                const content = section.textContent.toLowerCase();
                if (content.includes(searchTerm)) {
                    if (!found) {
                        // Show first matching section
                        showDoc(section.id);
                        found = true;
                    }
                    
                    // Highlight search term
                    const regex = new RegExp('(' + searchTerm + ')', 'gi');
                    section.innerHTML = section.innerHTML.replace(/<mark>/g, '').replace(/<\/mark>/g, '');
                    section.innerHTML = section.innerHTML.replace(regex, '<mark>$1</mark>');
                }
            });
        }
        
        // Handle hash navigation
        window.addEventListener('load', function() {
            const hash = window.location.hash.substring(1);
            if (hash && document.getElementById(hash)) {
                showDoc(hash);
            }
        });
        
        // Set first link as active
        document.querySelector('.sidebar a').classList.add('active');
    </script>
</body>
</html>
EOF

echo "✅ Documentation built: $OUTPUT_FILE"
echo "📄 Open in your browser: file://$(realpath "$OUTPUT_FILE")"