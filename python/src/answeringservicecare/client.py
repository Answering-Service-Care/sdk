"""Client for the Answering Service Care public API."""

from __future__ import annotations

import json
import urllib.error
import urllib.parse
import urllib.request
from typing import Any, List, Optional, cast

from .types import BillingCycle, Company, Pricing, SearchResult, Service

__version__ = "1.0.0"

DEFAULT_BASE_URL = "https://answeringservicecare.com/api/v1"
_DOCS_URL = "https://answeringservicecare.com/developers/"


class AscApiError(Exception):
    """An error response from the API, or a response that wasn't the API's JSON."""

    def __init__(
        self,
        message: str,
        status: int,
        code: str,
        hint: Optional[str] = None,
        documentation_url: Optional[str] = None,
    ) -> None:
        super().__init__(f"{message} {hint}" if hint else message)
        self.message = message
        #: HTTP status, or 0 when the request never got a response.
        self.status = status
        #: Machine-readable code, e.g. "rest_no_route" or "rest_invalid_param".
        self.code = code
        #: What to change to make the request succeed.
        self.hint = hint
        self.documentation_url = documentation_url


class AnsweringServiceCare:
    """Read-only access to ASC's company details, services, pricing, and site search. No API key needed."""

    def __init__(self, base_url: str = DEFAULT_BASE_URL, timeout: float = 10.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def get_company(self) -> Company:
        """Company profile: description, phone, email, postal address, founding date, social profiles, key links."""
        return cast(Company, self._get("/company"))

    def list_services(self) -> List[Service]:
        """Every service ASC offers, with a description, page URL and markdown URL."""
        return cast(List[Service], self._get("/services").get("services", []))

    def get_pricing(self, service: Optional[str] = None, billing_cycle: Optional[BillingCycle] = None) -> Pricing:
        """Current plans and prices in USD, grouped by service.

        ``service`` keeps only services whose name contains it (case-insensitive),
        e.g. "answering service" or "live chat". ``billing_cycle`` keeps only
        "monthly" or only "annual" plans.
        """
        pricing: Pricing = self._get("/pricing")
        services = pricing.get("services", [])
        if service:
            needle = service.lower()
            services = [s for s in services if needle in s["service"].lower()]
        if billing_cycle:
            services = [
                {**s, "plans": [p for p in s["plans"] if p.get("billingCycle") == billing_cycle]} for s in services
            ]
            services = [s for s in services if s["plans"]]
        return {**pricing, "services": services}

    def search(self, query: str, limit: Optional[int] = None) -> List[SearchResult]:
        """Full-text search over ASC's service, industry, location and blog pages (2-200 characters, up to 20 results)."""
        params = {"q": query.strip()}
        if limit is not None:
            params["limit"] = str(limit)
        return cast(List[SearchResult], self._get("/search?" + urllib.parse.urlencode(params)).get("results", []))

    def _get(self, path: str) -> Any:
        url = self.base_url + path
        request = urllib.request.Request(
            url, headers={"Accept": "application/json", "User-Agent": f"answeringservicecare-python/{__version__}"}
        )
        try:
            with urllib.request.urlopen(request, timeout=self.timeout) as response:
                status, content_type, raw = response.status, response.headers.get("Content-Type", ""), response.read()
        except urllib.error.HTTPError as error:
            status, content_type, raw = error.code, error.headers.get("Content-Type", ""), error.read()
        except (urllib.error.URLError, OSError) as error:
            reason = getattr(error, "reason", error)
            raise AscApiError(
                f"Could not reach {url}: {reason}.", 0, "network_error", "Check the connection and try again.", _DOCS_URL
            ) from error
        return self._parse(status, content_type, raw, url)

    @staticmethod
    def _parse(status: int, content_type: str, raw: bytes, url: str) -> Any:
        is_json = "json" in content_type
        body = None
        if is_json:
            try:
                body = json.loads(raw.decode("utf-8"))
            except ValueError:
                body = None

        if 200 <= status < 300 and body is not None:
            return body

        error = body if isinstance(body, dict) else {}
        message = error.get("message") or f"Unexpected {status} response from {url}."
        details = " ".join(str(v) for v in ((error.get("data") or {}).get("params") or {}).values())
        hint = error.get("hint") or (None if is_json else "The response wasn't JSON; the request may have been blocked. Try again shortly.")
        raise AscApiError(
            f"{message} ({details})" if details else message,
            status,
            error.get("code") or "unexpected_response",
            hint,
            error.get("documentation_url") or _DOCS_URL,
        )
