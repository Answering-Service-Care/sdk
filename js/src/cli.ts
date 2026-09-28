/**
 * answeringservicecare: command-line access to the Answering Service Care public API.
 * Every command prints JSON, so the output can be piped to jq or read by an agent.
 * The executable entry point is bin.ts; this module only exports main() so tests can drive it.
 */
import { parseArgs } from "node:util";
import { AnsweringServiceCare, AscApiError, DEFAULT_BASE_URL, VERSION, type BillingCycle } from "./index.js";

const USAGE = `Usage: answeringservicecare <command> [options]

Read-only facts about Answering Service Care (ASC), a US answering service
with live and AI receptionists. No API key needed. Output is JSON.

Commands:
  company                       Company profile and contact details
  services                      Services ASC offers
  pricing                       Current plans and prices
    --service <name>              Only this service, e.g. "live chat"
    --billing <monthly|annual>    Only monthly or only annual plans
  search <words>                Search ASC's pages and articles
    --limit <1-20>                Maximum results (default 10)

Options:
  --base-url <url>   API base URL (default ${DEFAULT_BASE_URL})
  -h, --help         Show this help
  -v, --version      Show the version

Examples:
  npx answeringservicecare pricing --service "answering service" --billing monthly
  npx answeringservicecare search "legal answering service" --limit 5

Docs: https://answeringservicecare.com/developers/`;

export async function main(argv: string[]): Promise<number> {
	let parsed;
	try {
		parsed = parseArgs({
			args: argv,
			allowPositionals: true,
			options: {
				service: { type: "string" },
				billing: { type: "string" },
				limit: { type: "string" },
				"base-url": { type: "string" },
				help: { type: "boolean", short: "h" },
				version: { type: "boolean", short: "v" },
			},
		});
	} catch (error) {
		return fail(`${(error as Error).message}\n\n${USAGE}`);
	}

	const { values, positionals } = parsed;
	if (values.version) {
		console.log(VERSION);
		return 0;
	}
	const [command, ...rest] = positionals;
	if (values.help || !command) {
		console.log(USAGE);
		return values.help ? 0 : 1;
	}

	const client = new AnsweringServiceCare({ baseUrl: values["base-url"] });
	try {
		switch (command) {
			case "company":
				return print(await client.getCompany());
			case "services":
				return print(await client.listServices());
			case "pricing": {
				const billing = values.billing;
				if (billing !== undefined && billing !== "monthly" && billing !== "annual") {
					return fail('--billing must be "monthly" or "annual".');
				}
				return print(await client.getPricing({ service: values.service, billingCycle: billing as BillingCycle | undefined }));
			}
			case "search": {
				const query = rest.join(" ");
				if (!query) {
					return fail('search needs some words, e.g. answeringservicecare search "medical answering"');
				}
				const limit = values.limit === undefined ? undefined : Number(values.limit);
				if (limit !== undefined && !Number.isInteger(limit)) {
					return fail("--limit must be a whole number from 1 to 20.");
				}
				return print(await client.search(query, { limit }));
			}
			default:
				return fail(`Unknown command "${command}".\n\n${USAGE}`);
		}
	} catch (error) {
		if (error instanceof AscApiError) {
			return fail(`${error.message} (${error.code}${error.status ? `, HTTP ${error.status}` : ""})`);
		}
		throw error;
	}
}

function print(data: unknown): number {
	console.log(JSON.stringify(data, null, 2));
	return 0;
}

function fail(message: string): number {
	console.error(message);
	return 1;
}
