import { readFileSync } from "node:fs";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AnsweringServiceCare, AscApiError, DEFAULT_BASE_URL, VERSION, type Pricing } from "../src/index.js";
import { main } from "../src/cli.js";

const PRICING: Pricing = {
	heading: "Plans & Pricing",
	currency: "USD",
	services: [
		{
			service: "Answering Service",
			plans: [
				{ name: "Starter", billingCycle: "monthly", price: "$40", priceValue: 40 },
				{ name: "Starter", billingCycle: "annual", price: "$36", priceValue: 36 },
			],
		},
		{ service: "Live Chat", plans: [{ name: "Chat", billingCycle: "monthly", price: "$99" }] },
	],
};

type Route = { status?: number; body: unknown; contentType?: string };

/** A fetch stub that answers by path and records every request. */
function stubFetch(routes: Record<string, Route>) {
	const calls: { url: string; headers: Record<string, string> }[] = [];
	const fetch = vi.fn(async (input: string | URL | Request, init?: RequestInit) => {
		const url = String(input);
		calls.push({ url, headers: (init?.headers ?? {}) as Record<string, string> });
		const { pathname, search } = new URL(url);
		const route = routes[pathname + search] ?? routes[pathname] ?? {
			status: 404,
			body: { code: "rest_no_route", message: "No route was found matching the URL and request method.", hint: "No such endpoint." },
		};
		const contentType = route.contentType ?? "application/json; charset=UTF-8";
		const text = typeof route.body === "string" ? route.body : JSON.stringify(route.body);
		return new Response(text, { status: route.status ?? 200, headers: { "content-type": contentType } });
	});
	return { fetch: fetch as unknown as typeof globalThis.fetch, calls };
}

describe("AnsweringServiceCare", () => {
	it("defaults to the production API and identifies itself", async () => {
		const { fetch, calls } = stubFetch({ "/api/v1/company": { body: { name: "Answering Service Care", url: "https://answeringservicecare.com/" } } });
		const company = await new AnsweringServiceCare({ fetch }).getCompany();

		expect(company.name).toBe("Answering Service Care");
		expect(calls[0].url).toBe(`${DEFAULT_BASE_URL}/company`);
		expect(calls[0].headers.Accept).toBe("application/json");
		expect(calls[0].headers["User-Agent"]).toBe(`answeringservicecare-js/${VERSION}`);
	});

	it("accepts a base URL with a trailing slash", async () => {
		const { fetch, calls } = stubFetch({ "/api/v1/services": { body: { services: [] } } });
		await new AnsweringServiceCare({ fetch, baseUrl: "http://localhost:8080/api/v1/" }).listServices();
		expect(calls[0].url).toBe("http://localhost:8080/api/v1/services");
	});

	it("returns the services array", async () => {
		const services = [{ name: "Answering Service", url: "https://answeringservicecare.com/services/answering-service/" }];
		const { fetch } = stubFetch({ "/api/v1/services": { body: { services } } });
		expect(await new AnsweringServiceCare({ fetch }).listServices()).toEqual(services);
	});

	it("returns all pricing without a filter", async () => {
		const { fetch } = stubFetch({ "/api/v1/pricing": { body: PRICING } });
		expect(await new AnsweringServiceCare({ fetch }).getPricing()).toEqual(PRICING);
	});

	it("filters pricing by service and billing cycle", async () => {
		const { fetch } = stubFetch({ "/api/v1/pricing": { body: PRICING } });
		const client = new AnsweringServiceCare({ fetch });

		const chat = await client.getPricing({ service: "live CHAT" });
		expect(chat.services.map((s) => s.service)).toEqual(["Live Chat"]);

		const annual = await client.getPricing({ billingCycle: "annual" });
		expect(annual.services).toEqual([{ service: "Answering Service", plans: [PRICING.services[0].plans[1]] }]);
		expect(annual.heading).toBe("Plans & Pricing");
	});

	it("encodes search queries and passes the limit", async () => {
		const results = [{ title: "Legal answering service", url: "https://answeringservicecare.com/industries/legal/" }];
		const { fetch, calls } = stubFetch({ "/api/v1/search": { body: { query: "legal & law", results } } });

		expect(await new AnsweringServiceCare({ fetch }).search("  legal & law ", { limit: 5 })).toEqual(results);
		expect(new URL(calls[0].url).searchParams.get("q")).toBe("legal & law");
		expect(new URL(calls[0].url).searchParams.get("limit")).toBe("5");
	});

	it("turns API errors into AscApiError with the hint and parameter details", async () => {
		const { fetch } = stubFetch({
			"/api/v1/search": {
				status: 400,
				body: {
					code: "rest_invalid_param",
					message: "Invalid parameter(s): q",
					data: { status: 400, params: { q: "q must be at least 2 characters long." } },
					hint: "Check the parameter constraints in the OpenAPI document.",
					documentation_url: "https://answeringservicecare.com/openapi.json",
				},
			},
		});

		const error = await new AnsweringServiceCare({ fetch }).search("a").catch((e: unknown) => e);
		expect(error).toBeInstanceOf(AscApiError);
		expect(error).toMatchObject({
			status: 400,
			code: "rest_invalid_param",
			hint: "Check the parameter constraints in the OpenAPI document.",
			documentationUrl: "https://answeringservicecare.com/openapi.json",
		});
		expect((error as Error).message).toContain("q must be at least 2 characters long.");
	});

	it("reports non-JSON responses, such as a challenge page, clearly", async () => {
		const { fetch } = stubFetch({ "/api/v1/company": { status: 403, body: "<!DOCTYPE html><title>Just a moment...</title>", contentType: "text/html" } });

		const error = await new AnsweringServiceCare({ fetch }).getCompany().catch((e: unknown) => e);
		expect(error).toMatchObject({ status: 403, code: "unexpected_response" });
		expect((error as Error).message).toContain("wasn't JSON");
	});

	it("reports network failures with status 0", async () => {
		const fetch = vi.fn(async () => {
			throw new TypeError("fetch failed");
		}) as unknown as typeof globalThis.fetch;

		const error = await new AnsweringServiceCare({ fetch }).getCompany().catch((e: unknown) => e);
		expect(error).toMatchObject({ status: 0, code: "network_error" });
	});
});

describe("CLI", () => {
	afterEach(() => {
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
	});

	function runCli(args: string[], routes: Record<string, Route> = {}) {
		vi.stubGlobal("fetch", stubFetch(routes).fetch);
		const out: string[] = [];
		const err: string[] = [];
		vi.spyOn(console, "log").mockImplementation((line: string) => void out.push(line));
		vi.spyOn(console, "error").mockImplementation((line: string) => void err.push(line));
		return main(args).then((code) => ({ code, out: out.join("\n"), err: err.join("\n") }));
	}

	it("prints pricing as JSON with filters applied", async () => {
		const { code, out } = await runCli(["pricing", "--service", "answering", "--billing", "monthly"], { "/api/v1/pricing": { body: PRICING } });
		expect(code).toBe(0);
		expect(JSON.parse(out).services).toEqual([{ service: "Answering Service", plans: [PRICING.services[0].plans[0]] }]);
	});

	it("joins search words and passes --limit", async () => {
		const { code, out } = await runCli(["search", "medical", "answering", "--limit", "3"], {
			"/api/v1/search?q=medical+answering&limit=3": { body: { results: [{ title: "Medical" }] } },
		});
		expect(code).toBe(0);
		expect(JSON.parse(out)).toEqual([{ title: "Medical" }]);
	});

	it("rejects a bad --billing value without calling the API", async () => {
		const { code, err } = await runCli(["pricing", "--billing", "weekly"]);
		expect(code).toBe(1);
		expect(err).toContain('"monthly" or "annual"');
		expect(fetch).not.toHaveBeenCalled();
	});

	it("prints API errors and exits 1", async () => {
		const { code, err } = await runCli(["company"]);
		expect(code).toBe(1);
		expect(err).toContain("rest_no_route");
		expect(err).toContain("HTTP 404");
	});

	it("shows help and version", async () => {
		expect((await runCli(["--help"])).out).toContain("Usage: answeringservicecare");
		expect((await runCli([])).code).toBe(1);
		expect((await runCli(["--version"])).out).toBe(VERSION);
	});
});

describe("package metadata", () => {
	it("VERSION matches package.json", () => {
		const pkg = JSON.parse(readFileSync(new URL("../package.json", import.meta.url), "utf8"));
		expect(VERSION).toBe(pkg.version);
		expect(pkg.homepage).toMatch(/^https:\/\/answeringservicecare\.com\//);
	});
});
