"""Response shapes of the public API (components.schemas in /openapi.json).

Responses are plain dicts with the API's own camelCase keys; these TypedDicts
describe them for type checkers and editors.
"""

from __future__ import annotations

from typing import Dict, List, Literal, TypedDict

BillingCycle = Literal["monthly", "annual"]


class PostalAddress(TypedDict, total=False):
    streetAddress: str
    addressLocality: str
    addressRegion: str
    postalCode: str
    addressCountry: str


class Company(TypedDict, total=False):
    name: str
    url: str
    legalName: str
    alternateName: str
    description: str
    logo: str
    foundingDate: str
    numberOfEmployees: str
    telephone: str
    email: str
    address: PostalAddress
    areaServed: str
    sameAs: List[str]
    links: Dict[str, str]


class Service(TypedDict, total=False):
    name: str
    url: str
    slug: str
    markdownUrl: str
    description: str


class Plan(TypedDict, total=False):
    name: str
    billingCycle: BillingCycle
    price: str
    """Price as displayed, e.g. "$169"."""
    priceValue: float
    priceCurrency: str
    term: str
    included: str
    """Usage included, e.g. "100 Minutes"."""
    billing: str
    overage: str
    """Rate beyond the included usage."""
    features: List[str]
    mostPopular: bool
    url: str


class ServicePlans(TypedDict):
    service: str
    plans: List[Plan]


class Pricing(TypedDict, total=False):
    heading: str
    summary: str
    notes: List[str]
    currency: str
    url: str
    termsUrl: str
    lastModified: str
    services: List[ServicePlans]


class SearchResult(TypedDict, total=False):
    title: str
    url: str
    markdownUrl: str
    type: str
    description: str
    lastModified: str
