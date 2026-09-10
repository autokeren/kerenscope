# KerenScope — Plan, Compliance & Aksi

## 1. Daftar Aksi User (URUTAN WAJIB — jangan dibalik)

Rules: **semua anggota tim wajib selesai onboarding Sectors SEBELUM tim menulis kode project.** Repo code baru dibuat setelah itu. (Menulis spec/dokumen perencanaan = boleh, sudah kita lakukan di folder ini.)

1. **Daftar tim** di portal hackathon (hackathon.sectors.app → Register). Solo boleh; kalau tim, 2–4 orang, satu orang jadi representative (penerima credits & hadiah).
2. **Onboarding Sectors** — semua anggota bikin akun di sectors.app dan **completen onboarding-nya sampai tuntas** (eligible check memverifikasi ini).
3. **Klaim 1.000 API credits** lewat team page (setelah semua onboard). Roster tim terkunci saat klaim.
4. **Generate Sectors API key** (API page) → simpan aman, jangan pernah commit.
5. Join **Slack** official (channel #discussion) — buat tanya: (a) apakah 1 request = 1 credit, (b) apakah participants dapat akses API tanpa Insider plan (kemungkinan credits hackathon yang meng-unlock).
6. **Konfirmasi ke gue**: nama final (KerenScope / alternatif), org/repo GitHub (cek availability), komposisi tim.
7. Setelah itu **gue yang gas**: `git init` repo baru di GitHub (public), scaffold structure SPEC §8, terus coding sesuai milestone di bawah.

## 2. Timeline (10 Sep → 30 Sep 2026)

| Minggu | Fokus | Deliverable |
|---|---|---|
| **W1 (10–16 Sep)** | Fondasi | Onboarding+credits selesai · repo init · Sectors client + cache + credit ledger · tools core (company_report, quarterly, screener, subsector, price_history) · provider LLM OpenAI-compat · `keren company BBCA` jalan end-to-end tanpa LLM pun (data dump mode) |
| **W2 (17–23 Sep)** | Killer workflow | Agent loop penuh (PLAN→ACT→VERIFY→SYNTHESIZE) · `keren research` + `keren compare` · interactive REPL · tools smart-money (foreign_flow, broker_summary, insider_filings) · report renderer |
| **W3 (24–28 Sep)** | Polish + submit | README 2 bahasa + CI badge · install 1-baris (go install + release binary) · hardening (retry, error message jelas) · **video teaser + video judging** · posting sosmed + thumbnail template · **submit 28 Sep** (buffer 2 hari, JANGAN nunggu 30) |
| W3+ (29–30 Sep) | Hanya kalau darurat | Freeze total setelah submit. Aturan: setelah submit tidak boleh ada commit apa pun. |

Prinsip W1: **data path dulu, cerdas belakangan**. `keren company BBCA` harus jalan di hari ke-3 — itu sanity check bahwa integrasi Sectors solid.

## 3. Budget API Credits (1.000 total)

Asumsi: 1 request = 1 credit *(verifikasi — kalau ternyata per-endpoint beda, sesuaikan)*.

| Kebutuhan | Estimasi |
|---|---|
| Dev W1 (eksplorasi response tiap endpoint, with cache) | ~150 |
| Dev W2 (test loop, banyak re-run — cache bikin murah) | ~250 |
| Demo recordings + dry-run judging video (cache warm) | ~100 |
| Buffer tak terduga / demo live minggu jaga-jaga | ~200 |
| Sisa (jangan dibakar) | ~300 |

Aturan teknis: **cache selalu on** (default), `--no-cache` hanya untuk satu-off debugging. Credit ledger (`keren credits`) dicek tiap akhir hari coding.

## 4. Budget LLM

Bukan bagian dari 1.000 credits Sectors — pakai provider sendiri:
- Dev: pakai stack yang sudah kita punya (proxy CF → GLM-5.3) — biaya ~nol dari kantong.
- Product: user colok API key OpenAI-compatible apa pun (OpenAI/OpenRouter/GLM/Ollama lokal gratis).
- Prompt tokens: plan + evidence ringkas + draft → target < 20k token per riset.

## 5. Compliance Checklist (submit gate)

- [ ] Sectors API core di 13 tool — produk mati tanpa Sectors ✔ (by design)
- [ ] Disclaimer "bukan rekomendasi investasi" muncul di: README, setiap report, output CLI help
- [ ] Tidak ada fitur order/eksekusi transaksi
- [ ] Repo public, dibuat dalam build period, commit history organik, tanpa kode dari project lama
- [ ] `.env`/API keys tidak ada di repo (gitignore + scan sebelum release)
- [ ] Repo tetap public ≥ 90 hari setelah winners announced
- [ ] Video teaser 60 dtk (screen record produk jalan, publik di YouTube/sosmed)
- [ ] Video judging ≤ 3 mnt (problem + audience + core workflow end-to-end)
- [ ] One-sentence problem statement (draft ada di SPEC §1, finalisasi sebelum submit)
- [ ] Track selection: 01 — AI Agents & Assistants + daftar nama anggota tim
- [ ] Posting sosmed (IG/LinkedIn/Threads/TikTok) tag akun Sectors resmi + pakai thumbnail template yang disediakan
- [ ] Submit sebelum 30 Sep 23:59 WIB (target 28 Sep) — setelah submit: ZERO commit

## 6. Risiko & Mitigasi

| Risiko | Mitigasi |
|---|---|
| Onboarding/credit terlambat (reg tutup 22 Sep!) | Daftar HARI INI. Ini satu-satunya blocker keras. |
| Kehabisan credits pas recording video | Cache warm + dry-run dulu, record terakhir |
| Data bank yang tidak lengkap (mis. segment kosong) | Tool wajib graceful ("data tidak tersedia"), executor re-plan |
| LLM bikin klaim halusinasi | Fase VERIFY wajib jalan; klaim tanpa bukti di-flag |
| Terlalu banyak fitur = semua setengah jadi | Non-goals di SPEC §10. 1 killer workflow jalan > 5 fitur mati |
| Video asal-asalan (30% skor!) | Storyboard DEMO.md, tulis script, dry-run 2x sebelum record |
| Salah track | Sudah diverifikasi: Track 01 requirement = custom orchestration ✔. Kalau ragu, tanya #discussion di Slack (rules bilang aman). |
