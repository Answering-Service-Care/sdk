"""answeringservicecare: command-line access to the Answering Service Care public API.

Every command prints JSON, so the output can be piped to jq or read by an agent.
"""

from __future__ import annotations

import argparse
import json
import sys
from typing import Any, List, Optional

from .client import DEFAULT_BASE_URL, AnsweringServiceCare, AscApiError, __version__

_DESCRIPTION = (
    "Read-only facts about Answering Service Care (ASC), a US answering service with live and AI "
    "receptionists. No API key needed. Output is JSON."
)
_EPILOG = """examples:
  answeringservicecare pricing --service "answering service" --billing monthly
  answeringservicecare search "legal answering service" --limit 5

docs: https://answeringservicecare.com/developers/"""


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="answeringservicecare",
        description=_DESCRIPTION,
        epilog=_EPILOG,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL, help=f"API base URL (default {DEFAULT_BASE_URL})")
    parser.add_argument("-v", "--version", action="version", version=__version__)
    commands = parser.add_subparsers(dest="command", metavar="<command>")

    commands.add_parser("company", help="Company profile and contact details")
    commands.add_parser("services", help="Services ASC offers")
    pricing = commands.add_parser("pricing", help="Current plans and prices")
    pricing.add_argument("--service", help='Only this service, e.g. "live chat"')
    pricing.add_argument("--billing", choices=["monthly", "annual"], help="Only monthly or only annual plans")
    search = commands.add_parser("search", help="Search ASC's pages and articles")
    search.add_argument("words", nargs="+", help="Words to search for")
    search.add_argument("--limit", type=int, help="Maximum results, 1-20 (default 10)")
    return parser


def main(argv: Optional[List[str]] = None) -> int:
    parser = _parser()
    args = parser.parse_args(argv)
    if not args.command:
        parser.print_help()
        return 1

    client = AnsweringServiceCare(base_url=args.base_url)
    try:
        if args.command == "company":
            data: Any = client.get_company()
        elif args.command == "services":
            data = client.list_services()
        elif args.command == "pricing":
            data = client.get_pricing(service=args.service, billing_cycle=args.billing)
        else:
            data = client.search(" ".join(args.words), limit=args.limit)
    except AscApiError as error:
        status = f", HTTP {error.status}" if error.status else ""
        print(f"{error} ({error.code}{status})", file=sys.stderr)
        return 1

    print(json.dumps(data, indent=2, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
