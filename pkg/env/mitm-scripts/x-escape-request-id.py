import requests
import os


class Addon:
    def __init__(self):
        port = os.getenv("HEALTH_CHECK_PORT", "8080")
        self.log_url = f"http://127.0.0.1:{port}/log"

    def request(self, flow):
        request_id = flow.request.headers.get("X-Escape-Request-Id", "")
        if request_id:
            requests.post(self.log_url, data=f'Forwarding X-Escape-Request-Id: {request_id}')

addons = [Addon()]