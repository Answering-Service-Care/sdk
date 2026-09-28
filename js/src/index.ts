/**
 * Official client for the Answering Service Care public API.
 *
 * The API is read-only and needs no key: company details, services, current
 * plan pricing, and site search. It is described by
 * https://answeringservicecare.com/openapi.json.
 */

export const VERSION = "1.0.0";
export const DEFAULT_BASE_URL = "https://answeringservicecare.com/api/v1";
const DOCS_URL = "https://answeringservicecare.com/developers/";

// --- Response shapes (components.schemas in /openapi.json) ---

export interface PostalAddress {
	streetAddress?: string;
	addressLocality?: string;
	addressRegion?: string;
	postalCode?: string;
	addressCountry?: string;
}

export interface Company {
	name: string;
	url: string;
	legalName?: string;
	alternateName?: string;
	description?: string;
	logo?: string;
	foundingDate?: string;
	numberOfEmployees?: string;
	telephone?: string;
	email?: string;
	address?: PostalAddress;
	areaServed?: string;
	sameAs?: string[];
	links?: Record<string, string>;
}

export interface Service {
	name: string;
	url: string;
	slug?: string;
	markdownUrl?: string;
	description?: string;
}

export type BillingCycle = "monthly" | "annual";

export interface Plan {
	name: string;
	billingCycle: BillingCycle;
	/** Price as displayed, e.g. "$169". */
	price?: string;
	priceValue?: number;
	priceCurrency?: string;
	term?: string;
	/** Usage included, e.g. "100 Minutes". */
	included?: string;
	billing?: string;
	/** Rate beyond the included usage. */
	overage?: string;
	features?: string[];
	mostPopular?: boolean;
	url?: string;
}

export interface ServicePlans {
	service: string;
	plans: Plan[];
}

export interface Pricing {
	heading?: string;
	summary?: string;
	notes?: string[];
	currency?: string;
	url?: string;
	termsUrl?: string;
	lastModified?: string;
	services: ServicePlans[];
}

export interface SearchResult {
	title?: string;
	url?: string;
	markdownUrl?: string;
	type?: string;
	description?: string;
	lastModified?: string;
}

export interface PricingFilter {
	/** Only this service, matched case-insensitively, e.g. "answering service" or "live chat". */
	service?: string;
	/** Only monthly or only annual plans. */
	billingCycle?: BillingCycle;
}

export interface ClientOptions {
	/** Defaults to https://answeringservicecare.com/api/v1. */
	baseUrl?: string;
	/** Custom fetch, e.g. for testing or proxies. Defaults to the global fetch. */
	fetch?: typeof fetch;
	/** Per-request timeout in milliseconds. Defaults to 10000. */
	timeoutMs?: number;
}

/** An error response from the API, or a response that wasn't the API's JSON. */
export class AscApiError extends Error {
	override readonly name = "AscApiError";

	constructor(
		message: string,
		/** HTTP status, or 0 when the request never got a response. */
		readonly status: number,
		/** Machine-readable code, e.g. "rest_no_route" or "rest_invalid_param". */
		readonly code: string,
		/** What to change to make the request succeed. */
		readonly hint?: string,
		readonly documentationUrl?: string,
	) {
		super(hint ? `${message} ${hint}` : message);
	}
}

interface ErrorBody {
	code?: string;
	message?: string;
	hint?: string;
	documentation_url?: string;
	/** WordPress puts per-parameter validation messages here, e.g. { q: "q must be at least 2 characters long." }. */
	data?: { params?: Record<string, string> };
}

export class AnsweringServiceCare {
	readonly baseUrl: string;
	readonly #fetch: typeof fetch;
	readonly #timeoutMs: number;

	constructor(options: ClientOptions = {}) {
		this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/+$/, "");
		this.#fetch = options.fetch ?? globalThis.fetch.bind(globalThis);
		this.#timeoutMs = options.timeoutMs ?? 10_000;
	}

	/** Company profile: description, phone, email, postal address, founding date, social profiles, key links. */
	getCompany(): Promise<Company> {
		return this.#get<Company>("/company");
	}

	/** Every service ASC offers, with a description, page URL and markdown URL. */
	async listServices(): Promise<Service[]> {
		return (await this.#get<{ services?: Service[] }>("/services")).services ?? [];
	}

	/** Current plans and prices in USD, grouped by service. Filters are applied to the full price list. */
	async getPricing(filter: PricingFilter = {}): Promise<Pricing> {
		const pricing = await this.#get<Pricing>("/pricing");
		let services = pricing.services ?? [];

		if (filter.service) {
			const needle = filter.service.toLowerCase();
			services = services.filter((s) => s.service.toLowerCase().includes(needle));
		}
		if (filter.billingCycle) {
			services = services
				.map((s) => ({ ...s, plans: s.plans.filter((p) => p.billingCycle === filter.billingCycle) }))
				.filter((s) => s.plans.length > 0);
		}
		return { ...pricing, services };
	}

	/** Full-text search over ASC's service, industry, location and blog pages (2-200 characters, up to 20 results). */
	async search(query: string, options: { limit?: number } = {}): Promise<SearchResult[]> {
		const params = new URLSearchParams({ q: query.trim() });
		if (options.limit !== undefined) {
			params.set("limit", String(options.limit));
		}
		return (await this.#get<{ results?: SearchResult[] }>(`/search?${params}`)).results ?? [];
	}

	async #get<T>(path: string): Promise<T> {
		const url = this.baseUrl + path;
		const headers: Record<string, string> = { Accept: "application/json" };
		// Browsers set their own User-Agent, and a custom one would force a CORS preflight.
		if (typeof window === "undefined") {
			headers["User-Agent"] = `answeringservicecare-js/${VERSION}`;
		}

		let response: Response;
		try {
			response = await this.#fetch(url, { headers, signal: AbortSignal.timeout(this.#timeoutMs) });
		} catch (error) {
			throw new AscApiError(
				`Could not reach ${url}: ${error instanceof Error ? error.message : String(error)}.`,
				0,
				"network_error",
				"Check the connection and try again.",
				DOCS_URL,
			);
		}

		const isJson = (response.headers.get("content-type") ?? "").includes("json");
		const body: unknown = isJson ? await response.json().catch(() => undefined) : undefined;

		if (!response.ok || body === undefined) {
			const error = (body ?? {}) as ErrorBody;
			const details = Object.values(error.data?.params ?? {}).join(" ");
			const message = error.message ?? `Unexpected ${response.status} response from ${url}.`;
			throw new AscApiError(
				details ? `${message} (${details})` : message,
				response.status,
				error.code ?? "unexpected_response",
				error.hint ?? (isJson ? undefined : "The response wasn't JSON; the request may have been blocked. Try again shortly."),
				error.documentation_url ?? DOCS_URL,
			);
		}
		return body as T;
	}
}

export default AnsweringServiceCare;
