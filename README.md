# KerenScope

**Autonomous financial research agent for the Indonesian stock market — powered by [Sectors](https://sectors.app).**

KerenScope turns a research question into an investigation. Give it something like
*"bank dengan fundamental terbaik di Indonesia?"* and it plans the research,
queries the Sectors API, cross-checks every claim against the evidence, and
writes a cited report — all from the terminal.

Built for [Sectors Hackathon 2026](https://hackathon.sectors.app) — Track 01: AI Agents & Assistants.

## Quickstart

```bash
go install github.com/autokeren/kerenscope@latest   # or: make build
export SECTORS_API_KEY="..."   # create one at sectors.app/api → API Key Management
```

```bash
keren company BBCA     # one-shot company report
keren credits          # local estimate of API credit spend
```

## Roadmap

- `keren compare BBCA BBRI BMRI` — peer head-to-head
- `keren "question"` — autonomous multi-step research with plan → act → verify → synthesize
- Interactive REPL mode

See [docs/SPEC.md](docs/SPEC.md) for the full technical spec and [docs/PLAN.md](docs/PLAN.md) for the build plan.

## How it works

```
User question
   ↓
Planner (research plan as explicit steps)
   ↓
Executor (purpose-built tools over the Sectors REST API, disk-cached)
   ↓
Verifier (every claim cross-checked against the retrieved data)
   ↓
Synthesizer (evidence-backed report + confidence + limitations)
```

Sectors is the core data layer: company reports, quarterly financials, subsector
reports, price history, foreign flow, broker activity, and insider filings —
all from the [Sectors Financial API](https://docs.sectors.app). Remove Sectors
and KerenScope has no eyes.

## Engineering notes

- **Disk cache with per-endpoint TTL** — development iterations don't burn API credits.
- **Local credit ledger** — `keren credits` shows exactly what you've spent.
- **Retry with backoff** — 429s and 5xxs are handled with `Retry-After` respect.

## Disclaimer

KerenScope is an information and analysis tool. It is **not** investment advice,
and it does not execute trades. Data provided by the Sectors API.

## License

MIT — see [LICENSE](LICENSE).
