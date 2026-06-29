# Mini Proxy

Mini Proxy adalah reverse proxy berbasis Go yang dirancang untuk memproses, memodifikasi, dan meneruskan request HTTP dengan dukungan berbagai backend dan integrasi.

## Features

- High performance HTTP Proxy
- Fiber Framework
- SQLCipher support
- Kafka integration
- Docker ready
- Multi-platform build system
- GitHub Actions CI/CD
- Build metadata
- SHA256 checksum verification

---

# Quick Start

## Docker (Recommended)

Cara paling mudah menjalankan Mini Proxy adalah menggunakan Docker.

```bash
docker compose up -d
```

atau

```bash
docker compose -f docker-compose.yml up -d
```

---

## Native Binary

Unduh binary sesuai sistem operasi Anda dari halaman **GitHub Releases**.

Konfigurasi:

Jalankan:

```bash
./webcore proxy
```

---

# Platform Support

| Platform | Native Binary | Docker | Status |
|----------|:-------------:|:------:|:------:|
| Linux amd64 | ✅ | ✅ | Supported |
| Linux arm64 | ✅ | ✅ | Supported |
| macOS Apple Silicon | ✅ | ✅ | Supported |
| macOS Intel | ❌ | ✅ | Docker Recommended |
| Windows Native | ❌ | ✅ | Docker Recommended |

Note: Lebih disarankan menggunakan docker untuk semua host (OS)
---

# Windows Users

## Current Status

Windows native build masih **belum didukung secara penuh**.

Project ini menggunakan dependency native seperti:

- SQLCipher
- librdkafka (Confluent Kafka)

yang saat ini masih mengalami kendala kompatibilitas saat proses linking pada toolchain Windows.

Akibatnya:

- Native binary Windows belum tersedia.
- Build GitHub Actions untuk Windows masih dalam pengembangan.

## Recommendation

Untuk pengguna Windows, hanya saat ini hanya bisa menggunakan Docker.

```bash
docker compose up -d
```

Cara ini memberikan lingkungan yang konsisten dengan Linux sehingga seluruh dependency native telah tersedia dan lebih stabil.

---

# Building from Source

Lihat panduan lengkap pada:

```
scripts/README.md
```

---

# Releases

Semua release tersedia pada halaman **GitHub Releases**.

Setiap release menyertakan:

- Binary
- Docker support
- SHA256SUMS
- Build metadata

---

# Development

Clone repository:

```bash
git clone https://github.com/nersus15/mini-proxy.git
```

Masuk ke project:

```bash
cd mini-proxy
```

Jalankan:

```bash
go work sync
```

Untuk panduan build lengkap lihat:

```
scripts/README.md
```

---

# License

Mini Proxy is released under the **Apache License**.

See the [LICENSE](LICENSE) file for the complete license text.
