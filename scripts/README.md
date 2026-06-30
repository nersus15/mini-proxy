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
├── build.sh
├── release.sh
├── doctor.sh
├── package.sh
├── common.sh
├── platform.sh
└── build.conf
```
---

## File Structure di Release

```
mini-proxy-{os}-{arch}/
├── db/                   # Folder yang disiapkan untuk tempat database sqlite (opsional)
├── Dockerfile            # Untuk build Docker image
├── docker-compose.yml    # Untuk run dengan Docker Compose
├── webcore               # File binnary
├── config.yaml           # File config bawaan
└── webcore/              # File migrations

```

## Notes

- Binary sudah pre-built, user tidak perlu build ulang
- Edit file `config.yaml` sebelum jalankan
- Semua platform didukung: Linux, macOS, Windows
- Archive dalam format tar.gz untuk cross-platform compatibility

---

# doctor.sh

Memeriksa apakah environment build sudah siap.

```bash
./scripts/doctor.sh
```

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
SHA256SUMS
```

---

# Package Format

Linux/macOS

```
mini-proxy-v1.0.0-linux-amd64.tar.gz
```

Windows

```
mini-proxy-v1.0.0-windows-amd64.zip
```

Setiap package berisi:

```
mini-proxy/
├── mini-proxy
├── config.yaml.example
├── build.json
├── LICENSE
└── README.md
```

Windows:

```
mini-proxy.exe
```

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
doctor.sh
      │
      ▼
go work sync
      │
      ▼
go build
      │
      ▼
package.sh
      │
      ▼
SHA256SUMS
      │
      ▼
Verify Checksum
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

Saat ini build native Windows **belum didukung secara penuh**.

Hal ini disebabkan oleh dependency native berikut:

- SQLCipher
- librdkafka (confluent-kafka-go)

Kedua library tersebut masih mengalami kendala kompatibilitas pada proses linking menggunakan toolchain Windows (MinGW/MSYS2 maupun MSVC) sehingga:

- ❌ Build GitHub Actions untuk Windows belum berhasil.
- ❌ Build native Windows menggunakan `build.sh` atau `release.sh` belum dapat menghasilkan binary yang stabil.

### Rekomendasi

Untuk pengguna Windows, jalankan Mini Proxy menggunakan **Docker**.

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
| Windows Native | ❌ | - | Experimental |
| Windows (Docker) | ✅ | ✅ | Recommended |
