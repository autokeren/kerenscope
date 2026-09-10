# KerenScope

**Agen riset keuangan otonom untuk pasar saham Indonesia — didukung [Sectors](https://sectors.app).**

[![CI](https://github.com/autokeren/kerenscope/actions/workflows/ci.yml/badge.svg)](https://github.com/autokeren/kerenscope/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.22-00ADD8)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

KerenScope mengubah pertanyaan riset menjadi investigasi. Beri pertanyaan seperti
*"bandingkan BBCA, BBRI, dan BMRI — mana yang fundamentalnya paling kuat?"* dan
ia akan menyusun rencana riset, menarik data live dari Sectors API, menghitung
metrik secara deterministik, mencocokkan setiap klaim dengan bukti, lalu menulis
laporan bersitat — semuanya dari terminal.

Dibangun untuk [Sectors Hackathon 2026](https://hackathon.sectors.app) — Track 01: AI Agents & Assistants.
[Bahasa Indonesia](README.id.md) · [English](README.md)

## Seperti apa rasanya

```
$ keren compare BBCA BBRI BMRI

  Research objective
  ─────────────────
  Membandingkan fundamental BBCA, BBRI, BMRI ...

  Plan
  ────
  [1/7] company_report       — Laporan lengkap BBCA
  [2/7] quarterly_financials — Tren kuartalan BBRI: laba, NII, kredit
  ...
  → company_report      ✓ {"symbol":"BBCA.JK", ...
  → quarterly_financials ✓ ...          (jalan paralel, hasil di-cache)
  ✓ Analysis drafted

  Verification: 13/15 klaim didukung bukti · confidence 78/100
  Numeric check: 85 angka ter-match deterministik

  Report saved to reports/bandingkan-....md
  HTML report saved to reports/bandingkan-....html
```

Laporan disimpan sebagai **Markdown dan HTML mandiri** — halaman HTML-nya
terbuka di browser mana pun, jadi hasil riset bisa dibaca siapa saja tanpa terminal.

## Cara kerjanya

```
Pertanyaan user
   │
   ▼
Planner ─────── mengusulkan rencana riset eksplisit (langkah = tool + argumen)
   │              kamu approve dulu (approve / regenerate / quit)
   ▼
Executor ────── menjalankan tool secara paralel ke Sectors REST API
   │              hasil tool diproses oleh compute engine deterministik
   ▼
Verifier ────── dua lapis:
   │              1. klaim angka dicocokkan mekanis ke bukti
   │                 (sadar locale: 513,1 T / 8,23x / 19,1%)
   │              2. LLM mengadjudikasi angka yang tak ter-match + klaim semantik
   ▼
Laporan ─────── analisis + verifikasi per klaim + skor confidence
                + keterbatasan yang jujur + disclaimer
```

**LLM tidak pernah berhitung.** Compute engine Go yang menghitung ROE, ROA,
NIM, CASA, LDR, cost-to-income, pertumbuhan, premi valuasi vs peer, momentum
harga (perubahan %, drawdown maksimum, volatilitas), dan arus dana asing —
semuanya dari data mentah Sectors dengan provenance. Model hanya menarasikan
angka yang sudah dihitung engine; verifier mencocokkan narasi itu ke fakta yang
sama.

Hilangkan Sectors, dan KerenScope kehilangan mata: semua tool, digest, dan
metrik terhitung dibangun di atas [Sectors Financial API](https://docs.sectors.app).

## Tool

13 tool purpose-built di atas Sectors v2: company report, financials kuartalan,
screener (filter SQL-like), laporan subsektor, riwayat harga, foreign flow,
akumulasi/distribusi broker, insider filings, berita, top movers, indeks,
plus tool generik `sectors_api` untuk katalog endpoint sisanya (corporate
actions, komposisi pemegang saham, data mining, ...).

Setiap panggilan divalidasi, di-cache di disk dengan TTL per endpoint, dan
dicatat di ledger kredit lokal — jalankan `keren credits` untuk melihat pemakainya.

## Instalasi

**Dari source (butuh Go):**
```bash
go install github.com/autokeren/kerenscope@latest
```

**Binary siap pakai:** ambil di [Releases](https://github.com/autokeren/kerenscope/releases) —
linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64/arm64). Satu file, tanpa runtime.

## Konfigurasi

```bash
export SECTORS_API_KEY="..."     # buat di sectors.app/api → API Key Management
```

KerenScope jalan dengan provider chat-completions OpenAI-compatible mana pun —
lihat tabel lengkapnya di [README](README.md#setup).

## Penggunaan

```bash
keren "pertanyaan"          # pintasan → research
keren research "pertanyaan"  # investigasi otonom multi-langkah
keren compare BBCA BBRI BMRI
keren company BBCA          # laporan perusahaan sekali jalan
keren credits              # estimasi pemakaian kredit
```

Saat riset berjalan, rencana yang diusulkan ditampilkan dulu dan — di terminal
interaktif — kamu memilih **approve / regenerate / quit** sebelum eksekusi.

## Catatan rekayasa

- **Cache disk + ledger kredit** — pertanyaan riset yang berulang = 0 kredit
- **Eksekusi tool paralel** — batch tool jalan bersamaan, urutan protokol tetap
- **Digest deterministik** — 90 hari deret harga/arus TIDAK di-dump mentah ke
  context; engine menghitung ringkasannya
- **Ketahanan** — backoff 429/5xx, `Retry-After` dihormati, deadline per-call
  dengan fallback effort rendah

## Roadmap

REPL interaktif, watchlist & riset berulang, memori antar sesi, cakupan
SGX/KLSE, dan front-end desktop/web (CLI-nya engine; laporannya produk).

## Disclaimer

KerenScope adalah alat informasi dan analisis. **Bukan rekomendasi investasi**
dan tidak mengeksekusi transaksi. Data dari Sectors API.

## Lisensi

MIT — lihat [LICENSE](LICENSE). Dokumentasi: [SPEC](docs/SPEC.md) · [Plan](docs/PLAN.md) · [Storyboard demo](docs/DEMO.md).
