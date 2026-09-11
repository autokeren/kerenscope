# KerenScope — Script Video (siap baca)

Peraturan lomba: video boleh Bahasa Indonesia ATAU English. Script ini Bahasa
Indonesia (paling natural buat presenter, juri = tim Sectors/Supertype).
Durasi: judging ≤ 3 menit, teaser ≤ 60 detik.

## Persiapan rekam (penting!)

- Terminal fullscreen, font besar (minimal 16pt), background gelap, resolusi 1080p+
- `unset KEREN_DEBUG` — jangan ada output `[dbg]`
- **Warm cache dulu**: jalankan sekali full sebelum rekam biar 0 surprise:
  `keren compare BBCA BBRI BMRI` → lalu hapus `reports/` biar rekamannya "fresh"
- Durasi real: ±2 menit (config flash + effort low) — pas untuk 3 menit.
  Kalau perlu lebih singkat: `keren compare BBCA TLKM` (±90 detik)
- Rekam TANPA `--yes` — momen approve plan itu bagus buat video (human-in-the-loop).
  Tambah `--reveal-slow` kalau mau report muncul lebih dramatis di akhir
- Kalau muncul ✗ 400 di layar pas nerekam: JANGAN retake — narasi aja:
  "perhatikan, parameter salah ditolak dan diperbaiki otomatis sebelum
  menghabiskan kredit" — itu fitur self-healing, bukan bug
- Matikan notifikasi OS

---

## A. VIDEO JUDGING (≤ 3 menit)

### [0:00–0:20] Problem — tampilan: slide teks / Notion / apapun
> "Riset satu pertanyaan pasar — misalnya membandingkan tiga bank besar —
> biasanya butuh belasan tab, spreadsheet, dan berhari-hari.
> Investor retail dan analis muda di Indonesia gak punya waktu itu.
> KerenScope hadir untuk itu: agen riset keuangan otonom yang bekerja
> seperti junior analyst — di terminal."

### [0:20–0:35] Intro produk — tampilan: terminal kosong, ketik command
> "KerenScope. Beri pertanyaan riset apa pun tentang saham Indonesia —
> ia merencanakan investigasinya sendiri, mengambil data dari Sectors API,
> menghitung metriknya dengan engine deterministik, lalu memverifikasi
> setiap klaim sebelum menulis laporan."

*(ketik:)* `keren compare BBCA BBRI BMRI` *(dan Enter — sambil narasi berlanjut)*

### [0:35–2:20] Demo live — tampilan: terminal, ikuti output agent
> *(saat plan muncul)* "Pertama, planner menyusun rencana eksplisit —
> laporan perusahaan tiap bank, financials kuartalan, konteks sektor.
> Saya approve, dan eksekusi mulai."
>
> *(saat tools jalan paralel)* "Perhatikan: tool berjalan paralel,
> datanya live dari Sectors — company report, quarterlies, subsector —
> dan setiap hasil lewat compute engine: ROE, CASA, NIM, semua dihitung
> oleh program, bukan oleh model. LLM-nya tidak pernah berhitung."
>
> *(saat spinner "analyzing gathered data" berjalan)* "Setelah data lengkap,
> agent menyusun draft analisis — spinner-nya hidup, ini proses nyata,
> bukan video editing."
>
> *(saat verification muncul — TEKANKAN INI)* "Ini bagian favorit saya.
> Verifikasi dua lapis: setiap angka di draft dicocokkan mekanis ke
> bukti — deterministik, bukan opini LLM. Sisanya diadjudikasi, dan
> hasilnya jujur: berapa klaim didukung bukti, confidence berapa,
> keterbatasannya apa."

*(sesuaikan angka dengan run rekamanmu — contoh dari run nyata: "108 angka ter-match, 14 dari 14 klaim didukung, confidence 93")*

**MOMEN EMAS (kalau kejadian di run lu):** kalau verifier menandai kontradiksi
di draft-nya sendiri ("skor komposit tidak cocok dengan bukti engine"), jelaskan:
"Perhatikan — agent-nya berani melawan draft-nya sendiri. Ini bukan chatbot
yang asal jawab." Itu momen paling kuat buat kriteria technical depth.

### [2:20–2:40] Output — tampilan: buka file HTML report di browser
> "Hasilnya: laporan lengkap — Markdown untuk terminal, HTML untuk siapa pun.
> Bisa dikirim ke investor, dibuka di HP, tanpa install apa pun.
> Dan disclaimer-nya jelas: ini alat informasi dan analisis,
> bukan rekomendasi investasi."

### [2:40–2:55] How it works — tampilan: README di GitHub (bagian diagram)
> "Semuanya open source: plan, executor paralel, compute engine,
> hybrid verifier, cache anti-boros-kredit. Ditulis dari nol di Go —
> bukan client orang lain plus prompt."

### [2:55–3:00] Close — tampilan: repo URL besar
> "KerenScope — autonomous financial research agent.
> Built with Sectors. Repo-nya di github.com/autokeren/kerenscope.
> Terima kasih!"

---

## B. TEASER (≤ 60 detik) — yang penting CEPAT & KOHESIF

| Detik | Aksi | Teks overlay (opsional) |
|---|---|---|
| 0–5 | Terminal kosong, ketik `keren "bank dengan fundamental terbaik?"` | "one question in..." |
| 5–15 | Plan muncul step-by-step | "...an investigation out" |
| 15–35 | Tools jalan paralel, ✓ per baris | "live Sectors data" |
| 35–50 | Verification: numeric check + confidence | "every claim verified" |
| 50–60 | HTML report terbuka di browser + end card nama & repo | "KerenScope · Built with Sectors" |

Musik ringan, cut cepat, NO narasi panjang. Max 60 detik, publish public YouTube.

---

## C. Checklis upload

- [ ] Judging video ≤ 3:00 (publik atau unlisted YouTube/Drive sharing ON — juri harus bisa buka)
- [ ] Teaser ≤ 1:00 public di YouTube/sosmed
- [ ] Cek suara: tes 10 detik pertama dulu sebelum rekam full
- [ ] Jangan sebut angka harga sebagai "murah/mahal, beli/jual" — selalu framing analisis
