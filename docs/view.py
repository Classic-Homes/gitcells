#!/usr/bin/env python3
"""
GitCells Documentation Viewer
Simple, zero-dependency documentation server
"""

import http.server
import socketserver
import os
import webbrowser
from urllib.parse import unquote

PORT = 8080

class DocHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        # Serve the pre-built HTML file
        if self.path == '/' or self.path == '/index.html':
            html_path = os.path.join(os.path.dirname(__file__), '../dist/gitcells-docs.html')
            if os.path.exists(html_path):
                self.send_response(200)
                self.send_header('Content-type', 'text/html')
                self.end_headers()
                with open(html_path, 'rb') as f:
                    self.wfile.write(f.read())
                return
        
        # Default handler for other files
        super().do_GET()

if __name__ == "__main__":
    os.chdir(os.path.dirname(__file__))
    
    print("📚 GitCells Documentation Viewer")
    print(f"🌐 Starting server at http://localhost:{PORT}")
    print("Press Ctrl+C to stop\n")
    
    # Check if HTML file exists
    html_path = os.path.join(os.path.dirname(__file__), '../dist/gitcells-docs.html')
    if not os.path.exists(html_path):
        print("⚠️  Documentation not built yet!")
        print("Run one of these commands first:")
        print("  - node docs/build.js")
        print("  - ./scripts/build-docs.sh")
        exit(1)
    
    # Open browser
    webbrowser.open(f'http://localhost:{PORT}')
    
    # Start server
    with socketserver.TCPServer(("", PORT), DocHandler) as httpd:
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n👋 Shutting down...")