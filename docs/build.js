#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Simple markdown to HTML converter
function convertMarkdownToHtml(markdown) {
    let html = markdown;
    
    // Escape HTML
    html = html.replace(/&/g, '&amp;')
               .replace(/</g, '&lt;')
               .replace(/>/g, '&gt;');
    
    // Headers
    html = html.replace(/^#### (.*?)$/gm, '<h4>$1</h4>');
    html = html.replace(/^### (.*?)$/gm, '<h3>$1</h3>');
    html = html.replace(/^## (.*?)$/gm, '<h2>$1</h2>');
    html = html.replace(/^# (.*?)$/gm, '<h1>$1</h1>');
    
    // Bold and italic
    html = html.replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>');
    html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
    html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');
    
    // Code blocks
    html = html.replace(/```(\w*)\n([\s\S]*?)```/g, (match, lang, code) => {
        return `<pre><code class="${lang}">${code}</code></pre>`;
    });
    
    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
    
    // Links - convert internal .md links to JavaScript navigation
    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (match, text, url) => {
        // Handle internal documentation links
        if (url.endsWith('.md')) {
            // Convert path to doc ID
            let docId = url.replace(/\.md$/, '').replace(/.*\//, '');
            return `<a href="#" onclick="showDoc('${docId}'); return false;">${text}</a>`;
        } else if (url.startsWith('getting-started/') || url.startsWith('guides/') || url.startsWith('reference/')) {
            // Handle relative paths
            let docId = url.replace(/\.md$/, '').replace(/.*\//, '');
            return `<a href="#" onclick="showDoc('${docId}'); return false;">${text}</a>`;
        } else {
            // External links
            return `<a href="${url}" target="_blank">${text}</a>`;
        }
    });
    
    // Lists
    html = html.replace(/^\* (.+)$/gm, '<li>$1</li>');
    html = html.replace(/^\- (.+)$/gm, '<li>$1</li>');
    html = html.replace(/^\d+\. (.+)$/gm, '<li>$1</li>');
    
    // Wrap consecutive list items
    html = html.replace(/(<li>.*<\/li>\n)+/g, (match) => {
        return '<ul>\n' + match + '</ul>\n';
    });
    
    // Blockquotes
    html = html.replace(/^> (.+)$/gm, '<blockquote>$1</blockquote>');
    
    // Paragraphs
    html = html.split('\n\n').map(para => {
        if (para.trim() && !para.startsWith('<')) {
            return '<p>' + para + '</p>';
        }
        return para;
    }).join('\n\n');
    
    // Tables (simple support)
    html = html.replace(/\|(.+)\|/g, (match, content) => {
        const cells = content.split('|').map(cell => `<td>${cell.trim()}</td>`).join('');
        return `<tr>${cells}</tr>`;
    });
    
    return html;
}

// Read all documentation files
const docsDir = path.join(__dirname);
const docs = {
    index: fs.readFileSync(path.join(docsDir, 'index.md'), 'utf8'),
    installation: fs.readFileSync(path.join(docsDir, 'getting-started/installation.md'), 'utf8'),
    quickstart: fs.readFileSync(path.join(docsDir, 'getting-started/quickstart.md'), 'utf8'),
    concepts: fs.readFileSync(path.join(docsDir, 'getting-started/concepts.md'), 'utf8'),
    converting: fs.readFileSync(path.join(docsDir, 'guides/converting.md'), 'utf8'),
    tracking: fs.readFileSync(path.join(docsDir, 'guides/tracking.md'), 'utf8'),
    collaboration: fs.readFileSync(path.join(docsDir, 'guides/collaboration.md'), 'utf8'),
    conflicts: fs.readFileSync(path.join(docsDir, 'guides/conflicts.md'), 'utf8'),
    'auto-sync': fs.readFileSync(path.join(docsDir, 'guides/auto-sync.md'), 'utf8'),
    'use-cases': fs.readFileSync(path.join(docsDir, 'guides/use-cases.md'), 'utf8'),
    commands: fs.readFileSync(path.join(docsDir, 'reference/commands.md'), 'utf8'),
    configuration: fs.readFileSync(path.join(docsDir, 'reference/configuration.md'), 'utf8'),
    'json-format': fs.readFileSync(path.join(docsDir, 'reference/json-format.md'), 'utf8'),
    troubleshooting: fs.readFileSync(path.join(docsDir, 'reference/troubleshooting.md'), 'utf8'),
};

// Convert all to HTML
const htmlDocs = {};
for (const [key, content] of Object.entries(docs)) {
    htmlDocs[key] = convertMarkdownToHtml(content);
}

// Generate the HTML file
const html = `<!DOCTYPE html>
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
        h4 { font-size: 1.2em; color: #34495e; margin: 20px 0 10px; }
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
        mark {
            background: #ffeb3b;
            padding: 2px;
        }
        @media (max-width: 768px) {
            .container { flex-direction: column; }
            .sidebar { width: 100%; height: auto; position: relative; }
            .content { padding: 20px; }
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
            ${Object.entries(htmlDocs).map(([id, content]) => 
                `<div id="${id}" class="doc-section ${id === 'index' ? 'active' : ''}">${content}</div>`
            ).join('\n')}
        </main>
    </div>
    
    <script>
        const docs = ${JSON.stringify(Object.keys(htmlDocs))};
        
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
            if (event && event.target) {
                event.target.classList.add('active');
            }
            
            // Scroll to top
            window.scrollTo(0, 0);
            
            // Update URL hash
            window.location.hash = docId;
        }
        
        function searchDocs(event) {
            const searchTerm = event.target.value.toLowerCase();
            
            if (searchTerm.length < 2) {
                // Reset highlighting
                document.querySelectorAll('.doc-section').forEach(section => {
                    section.innerHTML = section.innerHTML.replace(/<mark>/g, '').replace(/<\\/mark>/g, '');
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
                        document.getElementById(section.id).classList.add('active');
                        found = true;
                    }
                }
            });
        }
        
        // Handle hash navigation
        window.addEventListener('load', function() {
            const hash = window.location.hash.substring(1);
            if (hash && document.getElementById(hash)) {
                showDoc(hash);
                // Find and activate the corresponding link
                document.querySelectorAll('.sidebar a').forEach(link => {
                    if (link.getAttribute('onclick') === \`showDoc('\${hash}')\`) {
                        link.classList.add('active');
                    }
                });
            } else {
                // Set first link as active
                document.querySelector('.sidebar a').classList.add('active');
            }
        });
    </script>
</body>
</html>`;

// Write the output file
const outputPath = path.join(__dirname, '../dist/gitcells-docs.html');
fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, html);

console.log('✅ Documentation built successfully!');
console.log(`📄 Open in your browser: file://${path.resolve(outputPath)}`);
console.log(`💾 File size: ${(fs.statSync(outputPath).size / 1024).toFixed(2)} KB`);