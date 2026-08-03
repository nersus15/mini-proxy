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
- librdkafka-dev
- libsqlcipher-dev

### macOS

- Xcode Command Line Tools
- clang
- librdkafka
- sqlcipher

### Windows

- MSYS2
- MinGW GCC
- librdkafka
- SQLCipher

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

1. Doctor
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

# Makefile

Build host

```bash
make build
```

Build semua platform

```bash
make build-all
```

Release host

```bash
make release
```

Release semua platform

```bash
make release-all
```

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


# Known Issues

## Windows

> **Status:** Experimental

Saat ini build native Windows **belum diverifikasi ulang**, meski salah satu penyebab utamanya sudah hilang.

Sebelumnya ada dua dependency native yang bermasalah di toolchain Windows (MinGW/MSYS2 maupun MSVC):

- ~~librdkafka (confluent-kafka-go)~~ — **sudah tidak dipakai lagi.** `github.com/webcore-go/lib-kafka` dan `github.com/confluentinc/confluent-kafka-go/v2` sudah dinonaktifkan dari `webcore/go.mod` dan `webcore/proxy/libraries.go`. Ini adalah dependency yang paling sering gagal link di Windows, jadi blocker utamanya sudah hilang.
- SQLCipher (`github.com/mutecomm/go-sqlcipher` via `lib-sqlchiper`) — **masih dipakai aktif** (CGO, `database:sqlite` loader). Library ini sudah punya implementasi khusus Windows (`sqlite3_windows.go`) dan workflow CI sudah menyiapkan MSYS2 + `mingw-w64-x86_64-sqlcipher`, jadi kemungkinan build Windows sekarang bisa berhasil — tapi belum ada run yang berhasil untuk memastikan.

Status saat ini:

- ✅ Blocker librdkafka sudah teratasi (dependency dihapus).
- ⚠️ Blocker SQLCipher kemungkinan besar juga sudah aman (dukungan Windows sudah ada di library & CI), tapi belum ada histori run yang berhasil.
- 🔄 Job Windows di GitHub Actions **sudah diaktifkan kembali** (matrix `windows-latest` di `release.yml`) untuk verifikasi — status akan diketahui pada run berikutnya (`workflow_dispatch` atau push tag `v*`).

### Rekomendasi

Untuk pengguna Windows, jalankan Mini Proxy menggunakan **Docker**.

Unduh package **docker** (`mini-proxy-{version}-docker-amd64.tar.gz`) — bukan package
windows — karena `Dockerfile` dan `docker-compose.yml` hanya disertakan di sana.

```bash
docker compose up -d
```

atau

```bash
docker compose -f docker-compose.yml up -d
```

Docker menggunakan image Linux sehingga seluruh dependency native telah tersedia dan lebih stabil dibandingkan menjalankan binary Windows secara langsung.

### Status Platform

| Platform | Binary | Docker | Status |
|----------|:------:|:------:|:------:|
| Linux | ✅ | ✅ | Supported |
| macOS (Apple Silicon) | ✅ | ✅ | Supported |
| Windows Native | ⚠️ | - | Unverified (librdkafka blocker gone, SQLCipher belum dites ulang) |
| Windows (Docker) | ✅ | ✅ | Recommended |
