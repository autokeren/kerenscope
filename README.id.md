# KerenScope

**Agen riset keuangan otonom untuk pasar saham Indonesia — didukung [Sectors](https://sectors.app).**

[![CI](https://github.com/autokeren/kerenscope/actions/workflows/ci.yml/badge.svg)](https://github.com/autokeren/kerenscope/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.26-00ADD8)](https://go.dev)
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
    ✓ 14/14 klaim didukung bukti · confidence 93/100
    ✓ 108 angka ter-match deterministik · 29 diadjudikasi

  ✓ Report saved to reports/….md
  ✓ HTML report saved to reports/….html
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

**Satu baris (Linux / macOS):**
```bash
curl -fsSL https://raw.githubusercontent.com/autokeren/kerenscope/main/install.sh | bash
```

**Satu baris (Windows PowerShell):**
```powershell
irm https://raw.githubusercontent.com/autokeren/kerenscope/main/install.ps1 | iex
```

**Dari source (butuh Go):**
```bash
go install github.com/autokeren/kerenscope@latest
```

**Binary manual:** ambil di [Releases](https://github.com/autokeren/kerenscope/releases) —
linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64/arm64). Satu file statis, tanpa runtime.
Di macOS, hapus karantina sekali: `xattr -d com.apple.quarantine keren`.

## Mulai cepat (30 detik)

```bash
keren company BBCA               # belum punya key? Jalankan saja —
                                  # keren bertanya sekali, memvalidasi live,
                                  # dan menyimpan ke ~/.kerenscope/config
keren doctor                     # periksa seluruh setup
```

Key juga bisa diatur lewat environment variable (`SECTORS_API_KEY` dulu;
`KERENSCOPE_LLM_*` untuk riset). Environment variable lebih diutamakan
daripada file config.

## Riset otonom butuh LLM juga

`keren research` / `keren compare` memanggil model chat-completions.
KerenScope jalan dengan provider OpenAI-compatible mana pun —
lihat tabel lengkapnya di [README](README.md#autonomous-research-needs-an-llm-too).

## Penggunaan

```bash
keren "pertanyaan"          # pintasan → research
keren research "pertanyaan"  # investigasi otonom multi-langkah
keren compare BBCA BBRI BMRI
keren company BBCA          # laporan perusahaan sekali jalan
keren credits              # estimasi pemakaian kredit
keren doctor                # periksa setup: key, endpoint, cache
```

Saat riset berjalan, rencana yang diusulkan ditampilkan dulu dan — di terminal
interaktif — kamu memilih **approve / regenerate / quit** sebelum eksekusi.
Laporan akhir muncul baris demi baris, bukan diloncat sekali jalan
(`--reveal-slow` untuk tempo yang lebih dramatis saat perekaman).

## Catatan rekayasa

- **Cache disk + ledger kredit** — pertanyaan riset yang berulang = 0 kredit
- **Eksekusi tool paralel** — batch tool jalan bersamaan, urutan protokol tetap
- **Digest deterministik** — 90 hari deret harga/arus TIDAK di-dump mentah ke
  context; engine menghitung ringkasannya

## Ketahanan — agen yang memperbaiki dirinya sendiri

- **Guard berbasis skema** — setiap panggilan divalidasi di sisi client dulu
  terhadap OpenAPI spec resmi Sectors: whitelist 70 endpoint (209 field
  screener, parameter yang sah per rute). Panggilan yang pasti gagal ditolak
  *sebelum* menghabiskan kredit, lengkap dengan opsi yang benar di pesan
  errornya agar model bisa retry dengan tepat.
- **Screener yang self-healing** — saat Sectors menolak filter dengan
  *"Field X requires bracket notation with a year. Example: X[2024]"*, tool
  mem-parsing bentuk yang benar langsung dari pesan error API-nya, menulis
  ulang clause, dan me-retry — tanpa keterlibatan LLM, tanpa percobaan sia-sia.
- **Fallback model dua lapis** — panggilan yang gagal di-retry pada effort
  rendah di model yang sama, lalu berpindah ke
  `KERENSCOPE_LLM_FALLBACK_MODEL` (mis. GLM-5.3 → GLM-5.3-flash).
  429/5xx menghormati `Retry-After`; tiap panggilan punya deadline keras.
- **Degradasi yang anggun** — anggaran tool yang habis tetap protokol-correct;
  error yang tak terpulihkan keluar sebagai pesan bersih yang actionable,
  bukan crash.

## Roadmap

CLI-nya adalah engine; laporannya produk. Arah selanjutnya:

**Dari riset sekali-jalan menjadi analis harian**
- *Watchlist & monitoring* — `keren watch BBCA` men-riset ulang sesuai jadwal
  dan hanya melaporkan **apa yang berubah**: foreign flow berbalik arah,
  insider mulai menjual, valuasi keluar dari band historisnya.
- *Memori antar sesi* — riset baru membawa temuan sebelumnya ("kemarin kamu
  menandai risiko X — masih valid?").
- *REPL interaktif* — pertanyaan lanjutan tanpa memulai ulang investigasi.

**Kedalaman**
- *Draft streaming* — analisis muncul live saat disusun.
- *Riset multi-agent* — sub-investigasi paralel per sektor, lalu disintesis.
- *Digest terjadwal* — laporan pagi untuk watchlist-mu via email/Telegram.
- *Data lebih luas* — SGX/KLSE, mining extension, corporate actions & buyback.

**Interfacing**
- *Front-end desktop (Tauri) & web* — shell tipis di atas engine yang sama.
- *Share link laporan* — publikasikan laporan HTML ke URL, tanpa lampiran.

**Ketelitian jangka panjang**
- *Verifikasi level 3* — backtest klaim historis terhadap realisasi;
  agent-nya menjaga track record yang jujur.
- *Plugin tools* — bawa endpoint dan kalkulasi milikmu sendiri.

## Disclaimer

KerenScope adalah alat informasi dan analisis. **Bukan rekomendasi investasi**
dan tidak mengeksekusi transaksi. Data dari Sectors API.

## Lisensi

MIT — lihat [LICENSE](LICENSE). Dokumentasi: [SPEC](docs/SPEC.md) · [Plan](docs/PLAN.md) · [Storyboard demo](docs/DEMO.md).
