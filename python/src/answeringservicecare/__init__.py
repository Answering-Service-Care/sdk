"""Official client for the Answering Service Care public API.

The API is read-only and needs no key: company details, services, current plan
pricing, and site search. It is described by
https://answeringservicecare.com/openapi.json.

    >>> from answeringservicecare import AnsweringServiceCare
    >>> asc = AnsweringServiceCare()
    >>> pricing = asc.get_pricing(service="answering service", billing_cycle="monthly")
"""

from .client import DEFAULT_BASE_URL, AnsweringServiceCare, AscApiError, __version__
from .types import BillingCycle, Company, Plan, PostalAddress, Pricing, SearchResult, Service, ServicePlans

__all__ = [
    "DEFAULT_BASE_URL",
    "AnsweringServiceCare",
    "AscApiError",
    "BillingCycle",
    "Company",
    "Plan",
    "PostalAddress",
    "Pricing",
    "SearchResult",
    "Service",
    "ServicePlans",
    "__version__",
]
