# KerenScope — Voice-Over Script (TTS-ready)

Teks narasi siap tempel ke TTS (ElevenLabs / Azure / OpenAI TTS), per seksi.
Total narasi ±2 menit — sisakan ruang di video biar terminal yang "ngomong".
Semua seksi: ** Bahasa Indonesia.**

## Catatan TTS (baca dulu)

- Generate **per seksi**, jangan satu blob — biar timing gampang disinkron ke layar
- Kalau akronim dibaca aneh, ganti di teks: `ROE` → `R O E` · `CASA` → `kasa` · `NIM` → `N I M`
- "Sectors" aman dibaca "sek-tors". "KerenScope" biasa dibaca "ke-ren-scope" ✓
- Angka biarkan angka (108, 14 dari 14, 93) — TTS Indonesia handle itu dengan baik
- Kecepatan: pilih yang natural-medium, jangan fastest — kriteria video: *engaging*

---

## S1 · Problem (target ≤ 20 detik)

Riset satu pertanyaan pasar butuh belasan tab, file spreadsheet, dan berhari-hari. Investor retail dan analis muda di Indonesia tidak punya waktu itu. KerenScope hadir untuk mengatasinya: agen riset keuangan otonom yang bekerja seperti junior analyst, langsung dari terminal.

## S2 · Intro produk (target ≤ 15 detik)

Beri pertanyaan riset apa pun tentang saham Indonesia. KerenScope merencanakan investigasinya sendiri, mengambil data dari Sectors API, menghitung metrik dengan engine deterministik, lalu memverifikasi setiap klaim sebelum menulis laporan. Semuanya jalan tanpa konfigurasi — tinggal install dan ketik.

## S3a · Demo — plan muncul (target ≤ 15 detik)

Pertama, planner menyusun rencana investigasi yang eksplisit. Laporan perusahaan untuk setiap bank, financials kuartalan, konteks sektor. Saya approve, dan eksekusi dimulai.

## S3b · Demo — tools jalan (target ≤ 20 detik)

Perhatikan. Tool berjalan paralel, dan datanya live dari Sectors. Setiap hasil melewati compute engine: ROE, CASA, NIM — semua dihitung oleh program, bukan oleh model. LLM-nya tidak pernah berhitung.

## S3c · Demo — draft (target ≤ 10 detik)

Setelah data lengkap, agent menyusun draft analisis. Spinner yang hidup ini proses nyata, bukan video editing.

## S3d · Demo — verification (target ≤ 30 detik — SEKSI PALING PENTING)

Ini bagian favorit saya. Verifikasi dua lapis. Setiap angka di draft dicocokkan mekanis ke bukti — deterministik, bukan opini LLM. Sisanya diadjudikasi. Hasilnya jujur: berapa klaim yang didukung bukti, confidence berapa, dan keterbatasannya apa.

> **Sisipkan live** (kalau mau): "108 angka ter-match deterministik. 14 dari 14 klaim didukung bukti. Confidence 93." — **ganti dengan angka dari run rekamanmu sendiri.**

## S4 · Output (target ≤ 20 detik)

Hasilnya: laporan lengkap. Markdown untuk terminal, HTML untuk siapa pun — bisa dikirim ke investor, dibuka di HP, tanpa install apa pun. Dan disclaimer-nya jelas: ini alat informasi dan analisis, bukan rekomendasi investasi.

## S5 · How it works (target ≤ 15 detik)

Semuanya open source. Planner, executor paralel, compute engine, hybrid verifier, cache hemat kredit, dan self-healing API calls. Ditulis dari nol di Go — bukan client orang lain plus prompt.

## S6 · Close (target ≤ 10 detik)

KerenScope. Autonomous financial research agent. Built with Sectors. Repo-nya di github.com/autokeren/kerenscope. Terima kasih.

---

## Tips sinkronisasi (editing)

| Seksi | Muncul di layar saat | Tips |
|---|---|---|
| S1 | Slide/teks problem | Narasi dulu, baru cut ke terminal |
| S2 | Terminal kosong + ketik command | Enter pas kata "menulis laporan" |
| S3a | Plan muncul step-by-step | Jangan buru-buru — biarkan list kebaca |
| S3b | ✓ tool lines keluar paralel | Highlight biar keliatan banyak yang jalan |
| S3c | Spinner + "Analysis drafted" | Suara selesai sebelum verification muncul |
| S3d | "◆ Verification" + limitation | Bisa pause 1-2 dtk sebelum angka disebut |
| S4 | File HTML kebuka di browser | Scroll pelan report-nya |
| S5 | README GitHub (bagian diagram) | Zoom ke bagian How it works |
| S6 | End card: nama + repo URL | Music fade out |
