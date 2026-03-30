import os
import json
from http.server import HTTPServer, BaseHTTPRequestHandler


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        body = json.dumps({"text": os.environ.get("SERVER_TEXT", "Hello from backend")}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass


port = int(os.environ.get("SERVER_PORT", 8080))
HTTPServer(("0.0.0.0", port), Handler).serve_forever()
