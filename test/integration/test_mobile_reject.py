"""Run against a disposable server and database; never point at production.

SMS_TEST_URL=http://127.0.0.1:38082 SMS_TEST_REGISTRATION_KEY=... python3 -m unittest discover -s test/integration
"""
import base64
from concurrent.futures import ThreadPoolExecutor
import json
import os
import unittest
import urllib.error
import urllib.request
import uuid

BASE = os.environ.get("SMS_TEST_URL", "http://127.0.0.1:38082") + "/api"
KEY = os.environ.get("SMS_TEST_REGISTRATION_KEY", "prueba-local")


def request(method, path, payload=None, auth=None):
    headers = {}
    if auth:
        headers["Authorization"] = auth
    if payload is not None:
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(BASE + path, method=method, headers=headers,
        data=json.dumps(payload).encode() if payload is not None else None)
    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as exc:
        return exc.code, None


class RejectTest(unittest.TestCase):
    def test_rejection_contract(self):
        code, owner = request("POST", "/mobile/v1/device", {"name":"reject-test", "pushToken":None}, "Bearer " + KEY)
        self.assertEqual(code, 201)
        code, other = request("POST", "/mobile/v1/device", {"name":"other-test", "pushToken":None}, "Bearer " + KEY)
        self.assertEqual(code, 201)
        bearer = "Bearer " + owner["token"]
        other_bearer = "Bearer " + other["token"]
        basic = "Basic " + base64.b64encode((owner["login"] + ":" + owner["password"]).encode()).decode()
        message_id = str(uuid.uuid4())
        code, _ = request("POST", "/3rdparty/v1/messages", {"id":message_id, "deviceId":owner["id"], "phoneNumbers":["+34600000000"], "textMessage":{"text":"test"}}, basic)
        self.assertEqual(code, 202)
        route = "/mobile/v1/message/" + message_id + "/reject"
        body = {"reason":"wrong_recipient"}
        self.assertEqual(request("POST", route, body)[0], 401)
        self.assertEqual(request("POST", route, body, other_bearer)[0], 404)
        self.assertEqual(request("POST", "/mobile/v1/message/missing/reject", body, bearer)[0], 404)
        self.assertEqual(request("POST", route, {"reason":"invalid"}, bearer)[0], 400)
        code, pending = request("GET", "/mobile/v1/messages", auth=bearer)
        self.assertEqual(code, 200)
        self.assertIn(message_id, [item["id"] for item in pending])
        expected = {"id":message_id, "state":"Failed", "reason":"wrong_recipient"}
        with ThreadPoolExecutor(max_workers=4) as pool:
            results = list(pool.map(lambda _: request("POST", route, body, bearer), range(4)))
        self.assertEqual(results, [(200, expected)] * 4)
        self.assertEqual(request("POST", route, body, bearer), (200, expected))
        self.assertEqual(request("POST", route, {"reason":"unsupported_format"}, bearer)[0], 409)
        code, pending = request("GET", "/mobile/v1/messages", auth=bearer)
        self.assertEqual(code, 200)
        self.assertNotIn(message_id, [item["id"] for item in pending])
        code, result = request("GET", "/3rdparty/v1/messages/" + message_id, auth=basic)
        self.assertEqual(code, 200)
        self.assertEqual(result["state"], "Failed")
        self.assertIn("Failed", result["states"])
        self.assertEqual(result["recipients"][0]["error"], "wrong_recipient")


if __name__ == "__main__":
    unittest.main()
