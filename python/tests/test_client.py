"""Tests against a local HTTP server that mimics the public API."""

from __future__ import annotations

import json
import re
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse

import pytest

from answeringservicecare import AnsweringServiceCare, AscApiError, __version__
from answeringservicecare.cli import main

PRICING = {
    "heading": "Plans & Pricing",
    "currency": "USD",
    "services": [
        {
            "service": "Answering Service",
            "plans": [
                {"name": "Starter", "billingCycle": "monthly", "price": "$40", "priceValue": 40},
                {"name": "Starter", "billingCycle": "annual", "price": "$36", "priceValue": 36},
            ],
        },
        {"service": "Live Chat", "plans": [{"name": "Chat", "billingCycle": "monthly", "price": "$99"}]},
    ],
}


class _Handler(BaseHTTPRequestHandler):
    requests: list = []

    def do_GET(self) -> None:  # noqa: N802 (http.server naming)
        url = urlparse(self.path)
        query = {k: v[0] for k, v in parse_qs(url.query).items()}
        type(self).requests.append({"path": url.path, "query": query, "headers": dict(self.headers)})

        if url.path == "/api/v1/company":
            self._send(200, {"name": "Answering Service Care", "url": "https://answeringservicecare.com/"})
        elif url.path == "/api/v1/services":
            self._send(200, {"services": [{"name": "Answering Service", "url": "https://answeringservicecare.com/services/answering-service/"}]})
        elif url.path == "/api/v1/pricing":
            self._send(200, PRICING)
        elif url.path == "/api/v1/search" and len(query.get("q", "")) < 2:
            self._send(400, {
                "code": "rest_invalid_param",
                "message": "Invalid parameter(s): q",
                "data": {"status": 400, "params": {"q": "q must be at least 2 characters long."}},
                "hint": "Check the parameter constraints in the OpenAPI document.",
                "documentation_url": "https://answeringservicecare.com/openapi.json",
            })
        elif url.path == "/api/v1/search":
            self._send(200, {"query": query["q"], "results": [{"title": f"Result for {query['q']}"}]})
        elif url.path == "/blocked/company":
            self._send(403, "<!DOCTYPE html><title>Just a moment...</title>", content_type="text/html")
        else:
            self._send(404, {"code": "rest_no_route", "message": "No route was found matching the URL and request method.", "hint": "No such endpoint."})

    def _send(self, status: int, body, content_type: str = "application/json; charset=UTF-8") -> None:
        data = (body if isinstance(body, str) else json.dumps(body)).encode()
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args) -> None:
        pass


@pytest.fixture()
def server():
    _Handler.requests = []
    httpd = ThreadingHTTPServer(("127.0.0.1", 0), _Handler)
    thread = threading.Thread(target=httpd.serve_forever, daemon=True)
    thread.start()
    yield f"http://127.0.0.1:{httpd.server_port}"
    httpd.shutdown()


@pytest.fixture()
def client(server):
    return AnsweringServiceCare(base_url=server + "/api/v1/")


def test_defaults_to_production():
    assert AnsweringServiceCare().base_url == "https://answeringservicecare.com/api/v1"


def test_company_and_user_agent(client):
    assert client.get_company()["name"] == "Answering Service Care"
    request = _Handler.requests[0]
    assert request["path"] == "/api/v1/company"
    assert request["headers"]["Accept"] == "application/json"
    assert request["headers"]["User-Agent"] == f"answeringservicecare-python/{__version__}"


def test_services_returns_the_list(client):
    assert [s["name"] for s in client.list_services()] == ["Answering Service"]


def test_pricing_without_filters(client):
    assert client.get_pricing() == PRICING


def test_pricing_filters(client):
    chat = client.get_pricing(service="live CHAT")
    assert [s["service"] for s in chat["services"]] == ["Live Chat"]

    annual = client.get_pricing(billing_cycle="annual")
    assert annual["services"] == [{"service": "Answering Service", "plans": [PRICING["services"][0]["plans"][1]]}]
    assert annual["heading"] == "Plans & Pricing"


def test_search_encodes_query_and_limit(client):
    results = client.search("  legal & law ", limit=5)
    assert results == [{"title": "Result for legal & law"}]
    assert _Handler.requests[0]["query"] == {"q": "legal & law", "limit": "5"}


def test_api_errors_carry_code_hint_and_details(client):
    with pytest.raises(AscApiError) as info:
        client.search("a")
    error = info.value
    assert (error.status, error.code) == (400, "rest_invalid_param")
    assert error.hint == "Check the parameter constraints in the OpenAPI document."
    assert error.documentation_url == "https://answeringservicecare.com/openapi.json"
    assert "q must be at least 2 characters long." in str(error)


def test_non_json_response_is_reported(server):
    with pytest.raises(AscApiError) as info:
        AnsweringServiceCare(base_url=server + "/blocked").get_company()
    assert (info.value.status, info.value.code) == (403, "unexpected_response")
    assert "wasn't JSON" in str(info.value)


def test_network_error_has_status_zero():
    with pytest.raises(AscApiError) as info:
        AnsweringServiceCare(base_url="http://127.0.0.1:9/api/v1", timeout=2).get_company()
    assert (info.value.status, info.value.code) == (0, "network_error")


def test_cli_pricing_with_filters(server, capsys):
    assert main(["--base-url", server + "/api/v1", "pricing", "--service", "answering", "--billing", "monthly"]) == 0
    out = json.loads(capsys.readouterr().out)
    assert out["services"] == [{"service": "Answering Service", "plans": [PRICING["services"][0]["plans"][0]]}]


def test_cli_search_joins_words(server, capsys):
    assert main(["--base-url", server + "/api/v1", "search", "medical", "answering", "--limit", "3"]) == 0
    assert json.loads(capsys.readouterr().out) == [{"title": "Result for medical answering"}]
    assert _Handler.requests[0]["query"] == {"q": "medical answering", "limit": "3"}


def test_cli_reports_api_errors(server, capsys):
    assert main(["--base-url", server + "/api/v2", "company"]) == 1
    err = capsys.readouterr().err
    assert "rest_no_route" in err and "HTTP 404" in err


def test_cli_rejects_bad_billing(capsys):
    with pytest.raises(SystemExit) as info:
        main(["pricing", "--billing", "weekly"])
    assert info.value.code == 2


def test_cli_without_command_prints_help(capsys):
    assert main([]) == 1
    assert "usage: answeringservicecare" in capsys.readouterr().out


def test_version_matches_pyproject():
    pyproject = (Path(__file__).parents[1] / "pyproject.toml").read_text()
    assert re.search(r'^version = "([^"]+)"', pyproject, re.M).group(1) == __version__
