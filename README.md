# KerenScope

**Autonomous financial research agent for the Indonesian stock market — powered by [Sectors](https://sectors.app).**

[![CI](https://github.com/autokeren/kerenscope/actions/workflows/ci.yml/badge.svg)](https://github.com/autokeren/kerenscope/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.26-00ADD8)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

KerenScope turns a research question into an investigation. Give it something like
*"bandingkan BBCA, BBRI, dan BMRI — mana yang fundamentalnya paling kuat?"* and it
plans the research, pulls live data from the Sectors API, computes the metrics
deterministically, cross-checks every claim against the evidence, and writes a
cited report — all from the terminal.

Built for [Sectors Hackathon 2026](https://hackathon.sectors.app) — Track 01: AI Agents & Assistants.
[English](README.md) · [Bahasa Indonesia](README.id.md)

## What it looks like

```
$ keren compare BBCA BBRI BMRI

╭───────────────────────────────────────────────────╮
│  KerenScope — Autonomous Financial Research       │
│  Indonesian market intelligence · Powered by Sectors │
╰───────────────────────────────────────────────────╯
  Question
  Bandingkan BBCA, BBRI, BMRI secara menyeluruh ...

  ⠸ planning investigation… 3s

  ◆ Research objective
    Membandingkan fundamental BBCA, BBRI, BMRI ...
  ◆ Plan · 8 steps
     1. company_report      Laporan lengkap BBCA: valuasi, konsensus
     2. company_report      Laporan lengkap BBRI ...
    (planned in 3s)

    ✓ company_report    BBCA.JK · PT Bank Central Asia Tbk. · market cap Rp796.3T (#1 IDX)
    ✓ quarterly_financials BBRI · 8 quarters, latest 2026-06-30
  ✓ Analysis drafted (42s)

  ◆ Verification
    ✓ 14/14 claims supported · confidence 93/100
    ✓ 108 numbers matched deterministically · 29 adjudicated

  ✓ Report saved to reports/….md
  ✓ HTML report saved to reports/….html
```

Reports are saved as **Markdown and standalone HTML** — the HTML page opens in any
browser, so the research output is readable by anyone, terminal or not.

## How it works

```
User question
   │
   ▼
Planner ─────── proposes an explicit research plan (steps = tools + args)
   │              you approve it first (approve / regenerate / quit)
   ▼
Executor ────── runs tools in parallel over the Sectors REST API
   │              tool results are post-processed by a
   ▼              deterministic compute engine (see below)
Verifier ────── two layers:
   │              1. numeric claims matched mechanically against evidence
   │                 (locale-aware: 513,1 T / 8,23x / 19,1%)
   │              2. an LLM adjudicates unmatched numbers + semantic claims
   ▼
Report ──────── analysis + claim-by-claim verification + confidence score
                + honest limitations + disclaimer, saved as .md and .html
```

**The LLM never does arithmetic.** A Go compute engine derives ROE, ROA, NIM,
CASA, loan-to-deposit, cost-to-income, growth rates, valuation premiums vs peers,
price momentum (change %, max drawdown, volatility), and net foreign flow — all
from raw Sectors data, with provenance attached. The model narrates numbers the
engine computed; the verifier then checks the narration against the same facts.

Remove Sectors and KerenScope has no eyes: every tool, digest, and computed
metric is built on the [Sectors Financial API](https://docs.sectors.app).

## Tools

13 purpose-built tools over Sectors v2: company report, quarterly financials,
companies screener (SQL-like filters), subsector reports, price history, foreign
flow, broker accumulation/distribution, insider filings, news, top movers, index
history, plus a generic `sectors_api` escape hatch for the rest of the endpoint
catalog (corporate actions, shareholders composition, mining data, ...).

Every call is validated, disk-cached with per-endpoint TTL, and recorded in a
local credit ledger — run `keren credits` to see the spend.

## Install

**One-liner (Linux / macOS):**
```bash
curl -fsSL https://raw.githubusercontent.com/autokeren/kerenscope/main/install.sh | bash
```

**One-liner (Windows PowerShell):**
```powershell
irm https://raw.githubusercontent.com/autokeren/kerenscope/main/install.ps1 | iex
```

**From source (any platform with Go):**
```bash
go install github.com/autokeren/kerenscope@latest
```

**Manual binaries:** grab one from [Releases](https://github.com/autokeren/kerenscope/releases) —
linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64/arm64). No runtime needed;
it is a single static binary. On macOS, clear the quarantine once: `xattr -d com.apple.quarantine keren`.

## Quickstart (30 seconds)

```bash
keren company BBCA               # no keys yet? Just run it —
                                  # keren asks once, validates live,
                                  # and saves to ~/.kerenscope/config
keren doctor                     # verify the whole setup
```

Keys can also be set as environment variables (`SECTORS_API_KEY` first;
`KERENSCOPE_LLM_*` for research). Environment variables take precedence
over the config file.

## Autonomous research needs an LLM too

`keren research` / `keren compare` call a chat-completions model.
KerenScope works with any OpenAI-compatible provider:

| Variable | Default | Meaning |
|---|---|---|
| `SECTORS_API_KEY` | — | Sectors API key (required) |
| `KERENSCOPE_LLM_BASE_URL` | `https://api.openai.com/v1` | any OpenAI-compatible endpoint |
| `KERENSCOPE_LLM_API_KEY` | `OPENAI_API_KEY` fallback | LLM provider key |
| `KERENSCOPE_LLM_MODEL` | `gpt-4o-mini` | model id |
| `KERENSCOPE_LLM_REASONING` | auto (`high` for GLM) | reasoning effort hint |
| `KERENSCOPE_LLM_FALLBACK_MODEL` | — | secondary model used automatically when the primary fails (e.g. `@cf/zai-org/glm-5.3-flash`) |
| `KERENSCOPE_LLM_MAX_TOKENS` | `32768` | output token budget |
| `KERENSCOPE_LLM_CALL_TIMEOUT_SECS` | `240` | per-call hard deadline (auto-retries at low effort once) |
| `KEREN_DEBUG` | off | per-turn latency tracing |

## Usage

```bash
keren "question"            # shorthand → research
keren research "question"   # autonomous multi-step investigation
keren compare BBCA BBRI BMRI
keren company BBCA          # one-shot company report
keren credits               # local credit spend estimate
keren doctor                # diagnose setup: keys, endpoints, cache
```

During research, the proposed plan is shown first and — on an interactive
terminal — you choose to **approve / regenerate / quit** before execution.
The final report reveals line by line instead of dumping a wall of text
(`--reveal-slow` gives a more dramatic pace for recordings).
Flags: `--json` (raw output), `--no-cache` (debugging; spends credits).

## Engineering notes

- **Disk cache + credit ledger** — repeated research questions cost 0 credits; we
  tracked every request since day one (`keren credits`).
- **Parallel tool execution** — batched tool calls run concurrently, protocol
  order preserved.
- **Deterministic digests** — 90-day price/flow series are never dumped raw into
  context; the engine computes momentum/flow summaries instead.

## Reliability — the agent that repairs itself

- **Schema-driven guards** — every call is validated client-side first against
  the official Sectors OpenAPI spec: 70 endpoint whitelists (209 screener
  fields, allowed query params per route). Doomed calls are rejected *before*
  they cost a credit, with the correct options in the message so the model
  retries properly.
- **Self-healing screener** — when Sectors rejects a filter with
  *"Field X requires bracket notation with a year. Example: X[2024]"*, the tool
  parses the corrected form from the API's own error, rewrites the clause and
  retries — no LLM involvement, no wasted attempts.
- **Two-layer model fallback** — a failed call retries at low reasoning effort
  on the same model, then switches to `KERENSCOPE_LLM_FALLBACK_MODEL`
  (e.g. GLM-5.3 → GLM-5.3-flash). 429/5xx honor `Retry-After`; every call has
  a hard per-call deadline.
- **Graceful degradation** — tool budget exhaustion stays protocol-correct;
  unrecoverable errors exit with clean, actionable messages instead of a crash.

## Roadmap

The CLI is the engine; the report is the product. Where it goes next:

**From one-shot research to a daily analyst**
- *Watchlists & monitoring* — `keren watch BBCA` re-researches on a schedule and
  reports only **what changed**: foreign flow reversing, insiders selling,
  valuation leaving its historical band.
- *Memory across sessions* — new research carries prior findings forward
  ("yesterday you flagged risk X — is it still valid?").
- *Interactive REPL* — follow-up questions without restarting the investigation.

**Depth**
- *Streaming draft* — the analysis types out live.
- *Multi-agent research* — parallel per-sector sub-investigations, synthesized.
- *Scheduled digest* — a morning report for your watchlist via email/Telegram.
- *Broader data* — SGX/KLSE, the mining extension, corporate actions & buybacks.

**Interfacing**
- *Tauri desktop & web front-end* — thin shells over the same engine.
- *Report share links* — publish an HTML report to a URL, no files attached.

**Long-term rigor**
- *Verification level 3* — backtest historical claims against what actually
  happened; the agent keeps an honest track record.
- *Plugin tools* — bring your own endpoints and computations.

## Disclaimer

KerenScope is an information and analysis tool. It is **not investment advice**
and it does not execute trades. Data from the Sectors API.

## License

MIT — see [LICENSE](LICENSE). Docs: [SPEC](docs/SPEC.md) · [Plan](docs/PLAN.md) · [Demo storyboard](docs/DEMO.md).
