# Mini Proxy

Mini Proxy adalah reverse proxy berbasis Go yang dirancang untuk memproses, memodifikasi, dan meneruskan request HTTP dengan dukungan berbagai backend dan integrasi.

## Features

- High Performance HTTP Proxy
- Fiber Framework
- SQLCipher Support
- Kafka Integration
- Docker Ready
- Multi-platform Build System
- GitHub Actions CI/CD
- Build Metadata
- SHA256 Checksum Verification

---

# Platform Support

| Platform | Native Binary | Docker | Status |
|----------|:-------------:|:------:|:------:|
| Linux amd64 | ✅ | ✅ | Supported |
| Linux arm64 | ✅ | ✅ | Supported |
| macOS Apple Silicon | ✅ | ✅ | Supported |
| macOS Intel | ❌ | ✅ | Docker Recommended |
| Windows Native | ❌ | ✅ | Docker Recommended |

> **Note**
>
> Docker merupakan metode deployment yang direkomendasikan untuk seluruh platform karena memberikan lingkungan runtime yang konsisten.

---

# Running Mini Proxy

## Linux (Native)

### 1. Download

Unduh release terbaru dari halaman **GitHub Releases**.

Contoh:

```
mini-proxy-v1.0.0-linux-amd64.tar.gz
```

atau

```
mini-proxy-v1.0.0-linux-arm64.tar.gz
```

### 2. Extract

```bash
tar -xzf mini-proxy-v1.0.0-linux-amd64.tar.gz
```

### 3. Masuk ke Folder

```bash
cd mini-proxy-v1.0.0-linux-amd64
```

### 4. Konfigurasi

```yml
 config.yaml
```

Edit `config.yaml` sesuai kebutuhan.

### 5. Jalankan

```bash
./webcore proxy
```

---

## Linux (Docker)

### 1. Download

```
mini-proxy-v1.0.0-linux-amd64.tar.gz
```

### 2. Extract

```bash
tar -xzf mini-proxy-v1.0.0-linux-amd64.tar.gz
```

### 3. Masuk ke Folder

```bash
cd mini-proxy-v1.0.0-linux-amd64
```

### 4. Konfigurasi

```yml
 config.yaml
```

Edit konfigurasi sesuai kebutuhan.

### 5. Jalankan

```bash
docker compose up -d
```

atau

```bash
docker compose -f docker-compose.yml up -d
```

### Melihat Log

```bash
docker compose logs -f
```

### Menghentikan Service

```bash
docker compose down
```

---

---

## macOS (Native)

> **Supported:** Apple Silicon (M1/M2/M3)

### 1. Download

Unduh release terbaru dari halaman **GitHub Releases**.

Contoh:

```
mini-proxy-v1.0.0-darwin-arm64.tar.gz
```

### 2. Extract

```bash
tar -xzf mini-proxy-v1.0.0-darwin-arm64.tar.gz
```

### 3. Masuk ke Folder

```bash
cd mini-proxy-v1.0.0-darwin-arm64
```

### 4. Konfigurasi

```yml
 config.yaml
```

Edit `config.yaml` sesuai kebutuhan.

### 5. Jalankan

Jika pertama kali dijalankan, macOS mungkin akan memblokir binary karena belum ditandatangani (unsigned).

Hilangkan atribut karantina:

```bash
xattr -dr com.apple.quarantine ./webcore
```

Berikan izin eksekusi apabila diperlukan:

```bash
chmod +x ./webcore
```

Kemudian jalankan:

```bash
./webcore proxy
```

---

## macOS (Docker)

### 1. Download

```
mini-proxy-v1.0.0-darwin-arm64.tar.gz
```

### 2. Extract

```bash
tar -xzf mini-proxy-v1.0.0-darwin-arm64.tar.gz
```

### 3. Masuk ke Folder

```bash
cd mini-proxy-v1.0.0-darwin-arm64
```

### 4. Konfigurasi

```yml
 config.yaml
```

Edit konfigurasi sesuai kebutuhan.

### 5. Jalankan

```bash
docker compose up -d
```

atau

```bash
docker compose -f docker-compose.yml up -d
```

### Melihat Log

```bash
docker compose logs -f
```

### Menghentikan Service

```bash
docker compose down
```
---

## Windows (Docker)

> **Windows Native Binary belum didukung.**
>
> Pengguna Windows disarankan menggunakan Docker Desktop.

### 1. Download

```
mini-proxy-v1.0.0-windows-amd64.zip
```

### 2. Extract

Extract menggunakan Windows Explorer, 7-Zip, atau WinRAR.

### 3. Masuk ke Folder

```
mini-proxy-v1.0.0-windows-amd64
```

### 4. Konfigurasi

Salyml` ```

menjadi

```
config.yaml
```

Lalu sesuaikan konfigurasi.

### 5. Jalankan

Buka PowerShell atau Windows Terminal pada folder tersebut.

```powershell
docker compose up -d
```

atau

```powershell
docker compose -f docker-compose.yml up -d
```

### Melihat Log

```powershell
docker compose logs -f
```

### Menghentikan Service

```powershell
docker compose down
```

---

# Windows Users

## Current Status

Windows native build masih **belum didukung**.

Mini Proxy menggunakan beberapa dependency native seperti:

- SQLCipher
- librdkafka (Confluent Kafka)

yang saat ini masih mengalami kendala kompatibilitas saat proses linking menggunakan toolchain Windows.

Akibatnya:

- Native binary Windows belum tersedia.
- Build GitHub Actions untuk Windows masih dalam tahap pengembangan.

## Recommendation

Gunakan Docker untuk menjalankan Mini Proxy pada Windows.

```bash
docker compose up -d
```

Docker menyediakan lingkungan runtime Linux yang telah memiliki seluruh dependency native sehingga lebih stabil dibandingkan menjalankan binary Windows secara langsung.

---

# Releases

Setiap release pada halaman **GitHub Releases** berisi:

- Native Binary (platform yang didukung)
- Docker Compose Files
- Configuration Template
- Build Metadata (`build.json`)
- SHA256SUMS

---

# Building from Source

Panduan build lengkap tersedia pada:

```
scripts/README.md
```

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

Sinkronkan workspace:

```bash
go work sync
```

Lihat panduan build:

```
scripts/README.md
```

---

# License

This project is licensed under the **Apache License**.

See the [LICENSE](LICENSE) file for the complete license text.
