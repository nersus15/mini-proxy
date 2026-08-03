# Mini Proxy Build System

Build system untuk **mini-proxy** yang mendukung build lokal, packaging, checksum, dan GitHub Actions.

## Requirements

- Go 1.25+
- Git
- Bash
- pkg-config
- C Compiler (sesuai platform)

### Linux

- gcc
- libsqlcipher-dev

### macOS

- Xcode Command Line Tools
- clang
- sqlcipher

### Windows

- MSYS2 (MINGW64)
- MinGW GCC
- SQLCipher

> `librdkafka` sudah tidak dibutuhkan lagi sejak dependency kafka
> (`lib-kafka` / `confluent-kafka-go`) dinonaktifkan.

---

# Scripts

```
scripts/
├── build.sh      # Entry point: env check, workspace sync, build, package, checksum
├── release.sh    # build.sh + verify checksum + changelog + gh release
├── lib.sh        # Shared logic (logger, metadata, platform/compiler, packaging, checksum)
└── build.conf    # Configuration (app name, platforms, output name)
```
---

## File Structure di Release

Ada dua jenis paket: **native** (jalankan binary langsung) dan **docker**.

### Native — `mini-proxy-{version}-{os}-{arch}`

```
mini-proxy-{version}-{os}-{arch}/
├── db/                   # Folder yang disiapkan untuk tempat database sqlite (opsional)
├── main                  # File binary (main.exe di Windows)
├── config.yaml           # File config bawaan (dari config.yaml.example)
├── build.json            # Build metadata
└── webcore/              # File migrations
```

### Docker — `mini-proxy-{version}-docker-{arch}`

Berisi binary **linux** ditambah file Docker.

```
mini-proxy-{version}-docker-{arch}/
├── db/                   # Folder yang disiapkan untuk tempat database sqlite (opsional)
├── main                  # File binary linux
├── config.yaml           # File config bawaan (dari config.yaml.example)
├── build.json            # Build metadata
├── Dockerfile            # Untuk build Docker image
├── docker-compose.yml    # Untuk run dengan Docker Compose
└── webcore/              # File migrations
```

## Notes

- Binary sudah pre-built, user tidak perlu build ulang
- `config.yaml.example` dari root project di-*rename* menjadi `config.yaml` saat packaging, sehingga package siap pakai
- Edit file `config.yaml` sebelum jalankan
- Semua platform didukung: Linux, macOS, Windows
- Paket native **tidak** menyertakan `Dockerfile`/`docker-compose.yml` — gunakan paket docker untuk itu
- Paket docker dibuat otomatis dari setiap hasil build linux
- Archive dalam format tar.gz untuk cross-platform compatibility

---

# Environment Check

`build.sh` dan `release.sh` otomatis memeriksa apakah environment build sudah siap (bagian dari `lib.sh`, fungsi `check_environment`) sebelum mulai build — tidak perlu dijalankan terpisah lagi.

Contoh output:

```
Go               OK
Git              OK
pkg-config       OK
GCC              OK
Tar              OK
Zip              OK
```

---

# build.sh

Melakukan build binary.

Secara default hanya membangun **platform host**.

```bash
./scripts/build.sh
```

Build dengan versi tertentu.

```bash
./scripts/build.sh v1.0.0
```

Build semua target yang didefinisikan di `build.conf`.

```bash
./scripts/build.sh --all
```

Build semua target dengan versi tertentu.

```bash
./scripts/build.sh v1.0.0 --all
```

Output:

```
build/
├── linux-amd64/
├── linux-arm64/
├── darwin-arm64/
└── windows-amd64/
```

---

# release.sh

Melakukan proses release lengkap.

Tahapan yang dijalankan:

1. Environment Check
2. Workspace Sync
3. Build
4. Package
5. SHA256 Checksum
6. Verify Checksum

Default hanya membangun platform host.

```bash
./scripts/release.sh
```

Versi tertentu.

```bash
./scripts/release.sh v1.0.0
```

Semua platform.

```bash
./scripts/release.sh --all
```

Semua platform dengan versi tertentu.

```bash
./scripts/release.sh v1.0.0 --all
```

Output:

```
dist/

mini-proxy-v1.0.0-linux-amd64.tar.gz
mini-proxy-v1.0.0-linux-arm64.tar.gz
mini-proxy-v1.0.0-darwin-arm64.tar.gz
mini-proxy-v1.0.0-windows-amd64.zip
mini-proxy-v1.0.0-docker-amd64.tar.gz
mini-proxy-v1.0.0-docker-arm64.tar.gz
SHA256SUMS
```

---

# Package Format

Linux/macOS/Docker

```
mini-proxy-v1.0.0-linux-amd64.tar.gz
mini-proxy-v1.0.0-docker-amd64.tar.gz
```

Windows

```
mini-proxy-v1.0.0-windows-amd64.zip
```

Isi masing-masing package dijelaskan pada bagian
[File Structure di Release](#file-structure-di-release).

---

# Build Metadata

Binary menyimpan metadata berikut:

- Version
- Commit
- Branch
- Tag
- Build Date
- Builder
- Host
- Go Version
- Platform
- Architecture

---

# SHA256

Release otomatis membuat checksum.

```
dist/
├── mini-proxy-v1.0.0-linux-amd64.tar.gz
├── mini-proxy-v1.0.0-windows-amd64.zip
├── mini-proxy-v1.0.0-docker-amd64.tar.gz
└── SHA256SUMS
```

Verifikasi:

Linux

```bash
sha256sum -c SHA256SUMS
```

macOS

```bash
shasum -a 256 -c SHA256SUMS
```

---

# GitHub Actions

Workflow GitHub Actions menggunakan `build.sh`.

Setiap runner hanya membangun platform miliknya sendiri.

| Runner | Output |
|---------|--------|
| Ubuntu | Linux |
| macOS | macOS |
| Windows | Windows |

Selanjutnya GitHub Actions akan:

- Upload artifact
- Menggabungkan semua artifact
- Membuat GitHub Release
- Mengunggah seluruh package release

---

# Directory Structure

```
build/
```

Berisi hasil build sementara.

```
dist/
```

Berisi package release final.

---

# Troubleshooting

### Permission denied

```bash
chmod +x scripts/*.sh
```

---

### Go Workspace

Sinkronisasi manual.

```bash
go work sync
```

---

### Bersihkan hasil build

```bash
rm -rf build dist
```

---

# Build Flow

```
check_environment
      │
      ▼
go work sync
      │
      ▼
go build
      │
      ▼
package_all
      │
      ▼
SHA256SUMS
      │
      ▼
Verify Checksum (release.sh only)
```

---

# Notes

- Build lokal hanya membangun platform host.
- Gunakan `--all` untuk membangun seluruh platform yang didefinisikan pada `build.conf`.
- GitHub Actions membangun setiap platform pada runner masing-masing sehingga tidak memerlukan cross-compiler.
- Package release selalu disertai `build.json` dan `SHA256SUMS`.
- Folder sementara di `build/` akan dihapus setelah package berhasil dibuat.
- Nama file binary bisa ditentukan pada `build.conf`

---

# Windows

> **Status:** Supported

Build native Windows sudah didukung. Sebelumnya build Windows gagal karena
`librdkafka` (dari `confluent-kafka-go`) tidak bisa di-link dengan toolchain
MinGW. Dependency kafka sudah dinonaktifkan, sehingga yang tersisa hanya
SQLCipher yang memang sudah punya dukungan Windows.

Job Windows di GitHub Actions berjalan pada `windows-latest` dengan MSYS2
(MINGW64) dan `shell: msys2` beserta `path-type: inherit`, supaya Go dari
`actions/setup-go` tetap terbaca di dalam shell MSYS2.

## Menjalankan di Background

Package Windows berisi `main.exe`. Perintah menjalankannya:

```
main.exe proxy
```

Kalau dijalankan langsung, jendela Command Prompt harus tetap terbuka.
Untuk menjalankan di latar belakang, gunakan **NSSM**.

### NSSM (rekomendasi)

NSSM membungkus `main.exe` menjadi Windows Service, sehingga otomatis jalan
saat boot, restart sendiri kalau crash, dan tidak butuh Command Prompt terbuka.

```
nssm install mini-proxy "C:\mini-proxy\main.exe" proxy
nssm set mini-proxy AppDirectory C:\mini-proxy
nssm set mini-proxy AppStdout C:\mini-proxy\logs\out.log
nssm set mini-proxy AppStderr C:\mini-proxy\logs\err.log
nssm start mini-proxy
```

Perintah lain:

```
nssm restart mini-proxy
nssm stop mini-proxy
nssm remove mini-proxy confirm
```

Karena `logging.output` pada `config.yaml` bernilai `stdout`, `AppStdout` dan
`AppStderr` perlu diisi supaya log tetap tersimpan.

Overhead NSSM kecil: satu proses wrapper beberapa MB dan praktis 0% CPU saat
idle.

> `sc.exe create` bawaan Windows tidak bisa dipakai langsung karena binary Go
> tidak mengimplementasikan Service Control Handler (error 1053), jadi tetap
> perlu wrapper seperti NSSM.

### Alternatif tanpa install tool

Task Scheduler bisa dipakai kalau tidak mau menambah tool. Tidak ada proses
wrapper sama sekali, tapi kontrol service dan log rotation harus diurus sendiri.

```
schtasks /create /tn "mini-proxy" /tr "cmd /c C:\mini-proxy\main.exe proxy >> C:\mini-proxy\logs\out.log 2>&1" /sc onstart /ru SYSTEM /rl HIGHEST
```

---

# Docker

Gunakan package **docker** (`mini-proxy-{version}-docker-{arch}.tar.gz`), karena
`Dockerfile` dan `docker-compose.yml` hanya disertakan di sana.

```bash
docker compose up -d
```

---

# Status Platform

| Platform | Binary | Docker | Status |
|----------|:------:|:------:|:------:|
| Linux | ✅ | ✅ | Supported |
| macOS (Apple Silicon) | ✅ | ✅ | Supported |
| Windows | ✅ | ✅ | Supported |
