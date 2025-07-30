package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/docs"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - GitCells Documentation</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
        }
        .header {
            background: #2c3e50;
            color: white;
            padding: 15px 30px;
            position: fixed;
            width: 100%;
            top: 0;
            z-index: 1000;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 24px;
            display: inline-block;
            margin-right: 30px;
        }
        .header nav {
            display: inline-block;
        }
        .header nav a {
            color: white;
            text-decoration: none;
            padding: 8px 16px;
            margin: 0 5px;
            border-radius: 4px;
            transition: background 0.3s;
        }
        .header nav a:hover {
            background: #34495e;
        }
        .container {
            display: flex;
            margin-top: 60px;
            min-height: calc(100vh - 60px);
        }
        .sidebar {
            width: 280px;
            background: white;
            padding: 20px;
            overflow-y: auto;
            border-right: 1px solid #e0e0e0;
            position: fixed;
            height: calc(100vh - 60px);
        }
        .sidebar h2 {
            font-size: 16px;
            margin-top: 20px;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        .sidebar ul {
            list-style: none;
        }
        .sidebar li {
            margin-bottom: 5px;
        }
        .sidebar a {
            color: #555;
            text-decoration: none;
            display: block;
            padding: 5px 10px;
            border-radius: 3px;
            transition: all 0.3s;
        }
        .sidebar a:hover {
            background: #f0f0f0;
            color: #2c3e50;
        }
        .sidebar a.active {
            background: #3498db;
            color: white;
        }
        .content {
            flex: 1;
            margin-left: 280px;
            padding: 40px;
            max-width: 900px;
        }
        .content-wrapper {
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1, h2, h3, h4, h5, h6 {
            margin-top: 24px;
            margin-bottom: 16px;
            font-weight: 600;
        }
        h1 { font-size: 2.5em; color: #2c3e50; }
        h2 { font-size: 2em; color: #34495e; }
        h3 { font-size: 1.5em; color: #34495e; }
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
        ul, ol {
            margin-bottom: 16px;
            padding-left: 24px;
        }
        li { margin-bottom: 4px; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .search-box {
            float: right;
            margin-top: -5px;
        }
        .search-box input {
            padding: 8px 16px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
            width: 300px;
        }
        @media (max-width: 768px) {
            .sidebar {
                width: 100%;
                height: auto;
                position: relative;
            }
            .content {
                margin-left: 0;
                padding: 20px;
            }
            .container {
                flex-direction: column;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>GitCells Documentation</h1>
        <nav>
            <a href="/">Home</a>
            <a href="/getting-started/installation">Install</a>
            <a href="/getting-started/quickstart">Quick Start</a>
            <a href="/reference/commands">Commands</a>
        </nav>
        <div class="search-box">
            <input type="text" placeholder="Search documentation..." onkeyup="searchDocs(event)">
        </div>
    </div>
    <div class="container">
        <nav class="sidebar">
            {{.Navigation}}
        </nav>
        <main class="content">
            <div class="content-wrapper">
                {{.Content}}
            </div>
        </main>
    </div>
    <script>
        // Highlight active page
        const currentPath = window.location.pathname;
        document.querySelectorAll('.sidebar a').forEach(link => {
            if (link.getAttribute('href') === currentPath) {
                link.classList.add('active');
            }
        });
        
        // Simple search function
        function searchDocs(event) {
            if (event.key === 'Enter') {
                const searchTerm = event.target.value.toLowerCase();
                const content = document.querySelector('.content-wrapper');
                
                // Remove previous highlights
                content.innerHTML = content.innerHTML.replace(/<mark>/g, '').replace(/<\/mark>/g, '');
                
                if (searchTerm) {
                    // Simple highlight
                    const regex = new RegExp('(' + searchTerm + ')', 'gi');
                    content.innerHTML = content.innerHTML.replace(regex, '<mark>$1</mark>');
                    
                    // Scroll to first match
                    const firstMatch = content.querySelector('mark');
                    if (firstMatch) {
                        firstMatch.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    }
                }
            }
        }
    </script>
</body>
</html>
`

type Page struct {
	Title      string
	Content    template.HTML
	Navigation template.HTML
}

type DocServer struct {
	md       goldmark.Markdown
	template *template.Template
	port     int
}

func NewDocServer(port int) *DocServer {
	// Configure goldmark with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	tmpl, err := template.New("page").Parse(htmlTemplate)
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	return &DocServer{
		md:       md,
		template: tmpl,
		port:     port,
	}
}

func (ds *DocServer) getNavigation() string {
	nav := `
<h2>Getting Started</h2>
<ul>
    <li><a href="/">Overview</a></li>
    <li><a href="/getting-started/installation">Installation</a></li>
    <li><a href="/getting-started/quickstart">Quick Start</a></li>
    <li><a href="/getting-started/concepts">Basic Concepts</a></li>
</ul>

<h2>User Guides</h2>
<ul>
    <li><a href="/guides/converting">Converting Files</a></li>
    <li><a href="/guides/tracking">Tracking Changes</a></li>
    <li><a href="/guides/collaboration">Team Collaboration</a></li>
    <li><a href="/guides/conflicts">Resolving Conflicts</a></li>
    <li><a href="/guides/auto-sync">Auto-sync Setup</a></li>
    <li><a href="/guides/use-cases">Common Use Cases</a></li>
</ul>

<h2>Reference</h2>
<ul>
    <li><a href="/reference/commands">Commands</a></li>
    <li><a href="/reference/configuration">Configuration</a></li>
    <li><a href="/reference/json-format">JSON Format</a></li>
    <li><a href="/reference/troubleshooting">Troubleshooting</a></li>
</ul>
`
	return nav
}

func (ds *DocServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		path = "/index"
	}

	// Remove leading slash and add .md extension if not present
	docPath := strings.TrimPrefix(path, "/")
	if !strings.HasSuffix(docPath, ".md") {
		docPath += ".md"
	}

	// Get content from docs package
	content, err := docs.GetDoc(docPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := ds.md.Convert([]byte(content), &buf); err != nil {
		http.Error(w, "Failed to convert markdown", http.StatusInternalServerError)
		return
	}

	// Extract title from content
	title := "GitCells Documentation"
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Prepare page data
	page := Page{
		Title:      title,
		Content:    template.HTML(buf.String()),
		Navigation: template.HTML(ds.getNavigation()),
	}

	// Render the template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := ds.template.Execute(w, page); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

func main() {
	port := 8080
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &port)
	}

	// ASCII art banner
	fmt.Println(`
    ____  _ __  ______     ____    
   / ___\(_) /_/ ____/__  / / /____
  / / __/ / __/ /   / _ \/ / / ___/
 / /_/ / / /_/ /___/  __/ / (__  ) 
 \____/_/\__/\____/\___/_/_/____/  
                                   
       Documentation Viewer
`)

	server := NewDocServer(port)
	addr := fmt.Sprintf("localhost:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("📚 Starting GitCells Documentation Server\n")
	fmt.Printf("🌐 Server URL: %s\n", url)
	fmt.Printf("📁 Documentation embedded in binary\n")
	fmt.Printf("\n✨ Opening browser...\n")
	fmt.Printf("Press Ctrl+C to stop the server\n\n")

	// Open browser after a short delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := openBrowser(url); err != nil {
			fmt.Printf("⚠️  Could not open browser automatically\n")
			fmt.Printf("   Please open %s manually\n", url)
		}
	}()

	// Start server
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatal("Server failed:", err)
	}
}