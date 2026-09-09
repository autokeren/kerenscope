# KerenScope — Technical Spec

Sectors Hackathon 2026 · Track 01: AI Agents & Assistants
Status: DRAFT v1 — menunggu onboarding tim selesai sebelum coding.

---

## 1. Identitas Produk

| | |
|---|---|
| **Nama (working title)** | KerenScope — mudah di-rename, jangan terpaku |
| **Tagline** | *Your Autonomous Financial Research Agent* |
| **Track** | 01 — AI Agents & Assistants |
| **One-sentence problem statement (untuk submission)** | "Investor retail dan analis muda Indonesia butuh berhari-hari untuk meriset satu pertanyaan pasar; KerenScope adalah agent riset keuangan otonom yang menyelidiki pertanyaan riset apa pun tentang saham Indonesia secara end-to-end dan menghasilkan laporan berbasis bukti dalam hitungan menit." |

**Posisi produk**: bukan chatbot saham. CLI-first research agent yang **berpikir seperti junior analyst**: merencanakan investigasi → memilih sendiri data apa yang perlu diambil → mengumpulkan bukti → cross-check → laporan. Sectors = mata & data layer, KerenScope = otak & orchestrator.

**Kenapa ini lolos "qualifying test" Track 01** (kata mereka sendiri): agent logic/orchestration-nya custom-built; produk TIDAK hilang kalau prompt-nya dipindah ke client orang lain — loop, planner, tool selection, verify, dan report engine semuanya milik kita. Contoh arah resmi mereka ("a research agent that plans and executes a multi-step company comparison using Sectors data") = persis killer workflow kita.

**Batas produk (wajib)**: information & analysis tool. BUKAN rekomendasi investasi. Tidak ada eksekusi transaksi. Disclaimer selalu tampil di report.

---

## 2. High-Level Architecture

```
                ┌────────────────────────────────────────────┐
                │              keren (CLI)                   │
                │  cobra: research / company / compare /     │
                │         interactive (default)              │
                └───────────────────┬────────────────────────┘
                                    │
                        ┌───────────▼───────────┐
                        │   Agent Orchestrator   │
                        │  (custom loop, Go)     │
                        │                        │
                        │  PLAN ─► ACT ─► VERIFY │
                        │            │           │
                        │        SYNTHESIZE      │
                        └───┬───────────────┬───┘
                            │               │
                  ┌─────────▼───────┐  ┌────▼─────────────┐
                  │  LLM Provider    │  │  Tool Registry    │
                  │  (interface,     │  │  (typed tools,    │
                  │  OpenAI-compat)  │  │  JSON schema args) │
                  └──────────────────┘  └────┬──────────────┘
                                            │
                              ┌─────────────▼──────────────┐
                              │  Sectors Client (REST v2)  │
                              │  + disk cache + credit log │
                              └─────────────┬──────────────┘
                                            │
                                   api.sectors.app
```

**Bahasa**: Go. Alasan: kekuatan kita, single static binary (install 1 baris = poin real-world usability), CLI-first cocok untuk storytelling video.

**Distribusi**: `go install github.com/<org>/kerenscope@latest` + GitHub Release binary (linux/mac/windows). README 2 bahasa (EN + ID).

---

## 3. Agent Loop (inti yang dinilai 30% technical depth)

Empat fase, semua custom-built:

### 3.1 PLAN
- Input: pertanyaan user (natural language, ID/EN).
- LLM menghasilkan **research plan** terstruktur:

```json
{
  "objective": "Identifikasi bank dengan fundamental terkuat di antara BBCA, BBRI, BMRI",
  "steps": [
    {"id": 1, "intent": "Ambil ringkasan fundamental 3 bank", "tool": "company_report", "args": {...}},
    {"id": 2, "intent": "Ambil quarterly financials untuk tren", "tool": "quarterly_financials", "args": {...}},
    {"id": 3, "intent": "Bandingkan valuasi", "tool": "subsector_report", "args": {...}},
    {"id": 4, "intent": "Cek sinyal smart money", "tool": "foreign_flow", "args": {...}}
  ],
  "hypotheses": ["NIM BBCA lebih stabil dibanding peers"]
}
```

- Plan dirender ke terminal sebagai checklist → cinematic untuk video.

### 3.2 ACT (executor)
- Jalankan steps berurutan (atau parallel kalau independen — nice-to-have).
- Tiap step: render `[2/5] Mengambil quarterly financials BBCA... ✓`.
- **Dynamic re-planning**: kalau data gak ada (mis. segment revenue kosong), executor boleh minta LLM revisi sisa plan (bounded, max 2 re-plan).
- Tool hasil → masuk evidence store (bukan dump mentah ke context — di-ringkas per tool supaya context hemat).

### 3.3 VERIFY (differentiator)
- LLM diberi laporan draft + data mentah pendukung → tugasnya **cross-check tiap klaim** di draft terhadap angka di evidence.
- Klaim yang gak punya bukti → di-flag "⚠ unverified" atau dibuang.
- Output: skor confidence + daftar keterbatasan data. Ini yang bikin produk keliatan engineered, bukan prompt sederhana.

### 3.4 SYNTHESIZE
- LLM menghasilkan final report (struktur tetap, lihat §7.3).
- Report di-render ke terminal + disimpan ke `reports/<slug>-<date>.md`.

**Context management**: plan + ringkasan evidence (bukan JSON mentah) + report draft. Target < 20k token per turn untuk cost LLM murah.

**Anti-pattern yang kita hindari** (catatan desain): while-loop `prompt → tool → prompt` tanpa struktur plan/verify — itu yang bikin produk keliatan "ChatGPT + API".

---

## 4. Tool Catalog (di-map ke endpoint Sectors v2 asli)

Semua endpoint diverifikasi ada di docs.sectors.app (v2, Indonesia/IDX).

| Tool agent | Endpoint Sectors v2 | Dipakai untuk |
|---|---|---|
| `search_companies` | `/companies-screener` (where/order_by/`q`) | "cari saham consumer dengan revenue growth konsisten" |
| `list_subsectors` | helper-list subsectors/industries | validasi sektor, pemilihan peer |
| `company_report` | Company Report (`sections` param) | fundamental + valuation satu perusahaan |
| `quarterly_financials` | Quarterly Financials | tren NIM/deposit/loan bank, revenue quarter |
| `subsector_report` | Subsector Report | peer comparison sekaligus |
| `price_history` | Daily Transaction Data (≤90 hari) | momentum, drawdown |
| `top_movers` | Top Company Movers | screening berita pasar |
| `revenue_segments` | Company Revenue Segments | breakdown bisnis (Sankey data) |
| `insider_filings` | Company Filings | aktivitas insider buy/sell |
| `news` | News Articles | konteks peristiwa |
| `foreign_flow` | Daily Net Foreign Inflow | sentimen asing per saham |
| `broker_summary` | Top Buyers/Sellers Per Symbol + Broker Registry | akumulasi/distribusi broker (data yang jarang dipakai kompetitor — differentiator!) |
| `index_history` | Index Daily Transaction | konteks pasar (IHSG) |

Prinsip: tool **tipis & typed** (args JSON schema), semua intelligence ada di orchestrator. 13 tool cukup untuk MVP — jangan tambah sebelum lomba selesai.

**MCP vs REST**: pakai **REST langsung** (typed, cacheable, kontrol penuh di Go). MCP server Sectors boleh jadi appendix di README ("bisa juga dipakai lewat MCP"), tapi core product = REST. (Rules: salah satu cukup.)

---

## 5. Sectors Client & Cache Layer

- Auth: `SECTORS_API_KEY` (header `Authorization`), base `https://api.sectors.app/v2`.
- **Disk cache WAJIB dari hari 1** (credits cuma 1.000):
  - `~/.cache/kerenscope/` (atau `XDG_CACHE_HOME`), key = hash(endpoint + params).
  - TTL per jenis data: company report 24 jam · quarterly 24 jam · news 1 jam · price/foreign flow: EOD (harga demo gak berubah).
  - `--no-cache` flag khusus debugging.
- **Credit ledger**: append-only log tiap request (`~/.local/state/kerenscope/credits.log`) → perintah `keren credits` nunjukin sisa estimasi. (Verifikasi di portal: 1 request = 1 credit? Tanyakan di Slack #discussion kalau gak jelas.)
- Rate limit: respect 429 dengan exponential backoff; concurrency tool = 2 (jangan banjir).

## 6. LLM Layer

- Interface `Provider.Complete(messages, tools?) (stream, error)` — **OpenAI-compatible chat completions** (works with: GLM via CF Workers AI, OpenAI, OpenRouter, Ollama local, dsb).
- Config (`~/.config/kerenscope/config.yaml` + env):
  ```yaml
  llm:
    base_url: https://api.openai.com/v1   # atau proxy CF kita, atau apa pun
    api_key_env: KERENSCOPE_LLM_KEY
    model: gpt-...  # bebas, apa pun yang OpenAI-compatible
  ```
- Fallback model kedua (opsional, nice-to-have minggu 3): kalau model utama timeout → model cadangan.
- Streaming ke terminal per chunk (UX terasa hidup).
- **Tidak ada provider lock-in** = judge/user tinggal colok key-nya.

## 7. CLI UX

### 7.1 Perintah
```
keren "bank dengan fundamental terbaik di Indonesia?"   # shortcut → research
keren research "<pertanyaan bebas>"                     # killer workflow
keren company BBCA                                     # one-shot: profile lengkap
keren compare BBCA BBRI BMRI                           # one-shot: head-to-head
keren credits                                           # estimasi sisa credits
keren                                                   # interactive mode
```
Nama binari: `keren` (pendek, memorable; binary `kerenscope`, alias `keren`).

### 7.2 Output langkah (estetika "agent berpikir")
```
╭─ KerenScope ─ Autonomous Research ─────────────────╮

  Research objective
  ─────────────────
  Identifikasi bank big-4 dengan fundamental terkuat & valuasi wajar

  Plan
  ────
  [1/6] Ambil company report BBCA · BBRI · BMRI      ✓
  [2/6] Ambil quarterly financials 3 bank              ✓
  [3/6] Peer context via subsector report (perbankan) ✓
  [4/6] Cek foreign flow & broker activity 30 hari    ✓
  [5/6] Cross-check klaim vs bukti                    ✓
  [6/6] Susun laporan                                  ✓
```

### 7.3 Format Report (tetap, deterministik)
```
RESEARCH REPORT — <judul>
Ringkasan (3 kalimat)
Perbandingan / temuan utama      ← tabel bila relevan
Bukti per klaim                  ← tiap klaim sitasi: [endpoint, tanggal data]
⚠ Risiko & keterbatasan
Confidence: 87% — beserta alasannya
──────────────────────────────
Disclaimer: KerenScope adalah alat informasi & analisis.
Bukan rekomendasi investasi. DYOR.
```
Tersimpan juga sebagai Markdown di `reports/`.

### 7.4 Interactive mode
REPL sederhana: prompt `You >`, riwayat percakapan in-memory, perintah `/report`, `/plan`, `/exit`. Cukup — jangan bangun full TUI untuk hackathon (risiko waktu). Pretty output pakai ANSI styling; kalau sempat, lipgloss untuk warna.

## 8. Repo Structure

```
kerenscope/
├── main.go
├── cmd/                      # cobra: root, research, company, compare, interactive
├── internal/
│   ├── agent/                # orchestrator: plan.go, act.go, verify.go, synthesize.go, evidence.go
│   ├── sectors/              # REST client, cache, credit ledger
│   ├── tools/                # registry + 13 tool (typed args)
│   ├── llm/                  # provider interface + openai-compat impl
│   ├── report/               # renderer terminal + markdown writer
│   └── ui/                   # step rendering, ANSI styling
├── reports/                  # output report (gitignored)
├── docs/                     # (dipindah dari folder planning ini saat init)
├── .github/workflows/ci.yml  # build + test (badge README, kesan engineering)
├── README.md / README.id.md
├── LICENSE (MIT)
└── go.mod
```

**Aturan clean-room**: referensi arsitektur dari pengalaman pribadi boleh; copy-paste file dari repo lama TIDAK. Semua file ditulis fresh di repo baru. Commit history harus keliatan organik (bukan satu mega-commit hasil import).

## 9. Compliance & Safety (baked-in, bukan tempelan)

| Rule | Implementasi |
|---|---|
| Sectors = core data source (bukan dekoratif) | 13 tool semuanya Sectors; hapus Sectors → produk mati total. Tidak ada sumber data pasar lain. |
| Custom agent logic (Track 01) | PLAN→ACT→VERIFY→SYNTHESIZE custom loop, bukan client orang lain + prompt |
| No financial advice | System prompt constraint + template report selalu append disclaimer + review output sebelum render |
| No automated trading | Tidak ada fitur order/eksekusi sama sekali; tidak akan ditambahkan |
| No prior-project code | Repo fresh, clean-room, commit organik |
| Repo public ≥90 hari, keys bersih | `.env` di gitignore, linter cek pattern key sebelum release (`keren check` internal) |
| Freeze setelah submit | Target submit internal 28 Sep; setelah itu tangan di atas meja |

## 10. Non-Goals (jangan dibangun untuk hackathon)

- ❌ TUI penuh ala autokeren (bubbletea) — REPL cukup
- ❌ Multi-agent spawn/background — satu orchestrator saja
- ❌ Watchlist/alerting/scheduler — masuk "Roadmap" README
- ❌ Dashboard web — CLI-first adalah positioning, bukan keterbatasan
- ❌ SGX/KLSE/Mining extension — IDX saja (fokus = cerita kuat)
