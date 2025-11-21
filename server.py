from http.server import BaseHTTPRequestHandler, HTTPServer
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("content-length", 0))
        _ = self.rfile.read(length)
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        # Always return en-US for mock
        self.wfile.write(json.dumps({"lang": "en-US"}).encode())

HTTPServer(("0.0.0.0", 8080), Handler).serve_forever()
