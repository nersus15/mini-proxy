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
| Windows amd64 | ✅ | ✅ | Supported |

> **Note**
>
> Docker merupakan metode deployment yang direkomendasikan untuk seluruh platform karena memberikan lingkungan runtime yang konsisten.

---


## Project Structure

```
mini-proxy/
├── webcore/              # Main application
│   ├── main.go
│   ├── proxy/           # Proxy implementation
│   ├── init/            # Initialization
│   └── go.mod
├── modules/             # Additional modules
│   └── proxy/
├── config.yaml.example  # Configuration template
├── Dockerfile           # Production Docker image
├── docker-compose.yml   # Docker Compose file
├── scripts/             # Build and release scripts
└── README.md            # This file
```

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
./main proxy
```

---

## Linux (Docker)

### 1. Download

Gunakan package **docker**, karena hanya package ini yang menyertakan
`Dockerfile` dan `docker-compose.yml`.

```
mini-proxy-v1.0.0-docker-amd64.tar.gz
```

### 2. Extract

```bash
tar -xzf mini-proxy-v1.0.0-docker-amd64.tar.gz
```

### 3. Masuk ke Folder

```bash
cd mini-proxy-v1.0.0-docker-amd64
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
xattr -dr com.apple.quarantine ./main
```

Berikan izin eksekusi apabila diperlukan:

```bash
chmod +x ./main
```

Kemudian jalankan:

```bash
./main proxy
```

---

## macOS (Docker)

### 1. Download

Gunakan package **docker**. Package ini berisi binary linux, karena Docker
memang menjalankan container linux.

```
mini-proxy-v1.0.0-docker-amd64.tar.gz
```

### 2. Extract

```bash
tar -xzf mini-proxy-v1.0.0-docker-amd64.tar.gz
```

### 3. Masuk ke Folder

```bash
cd mini-proxy-v1.0.0-docker-amd64
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

## Windows (Native)

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

Edit `config.yaml` sesuai kebutuhan.

### 5. Jalankan

Buka PowerShell atau Windows Terminal pada folder tersebut.

```powershell
.\main.exe proxy
```

Cara ini mengharuskan jendela terminal tetap terbuka. Untuk menjalankan di
latar belakang, lihat bagian [Menjalankan sebagai Service](#menjalankan-sebagai-service-windows).

---

## Windows (Docker)

### 1. Download

Gunakan package **docker**, bukan package windows, karena hanya package ini
yang menyertakan `Dockerfile` dan `docker-compose.yml`.

```
mini-proxy-v1.0.0-docker-amd64.tar.gz
```

### 2. Extract

Extract menggunakan Windows Explorer, 7-Zip, atau WinRAR.

### 3. Masuk ke Folder

```
mini-proxy-v1.0.0-docker-amd64
```

### 4. Konfigurasi

Edit `config.yaml` sesuai kebutuhan.

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

# Menjalankan sebagai Service (Windows)

Menjalankan `main.exe` langsung dari terminal mengharuskan jendela terminal
tetap terbuka. Untuk menjalankannya di latar belakang, gunakan **NSSM**.

## NSSM

NSSM membungkus `main.exe` menjadi Windows Service, sehingga otomatis jalan
saat boot, restart sendiri kalau crash, dan tidak butuh terminal terbuka.

Unduh NSSM dari [nssm.cc](https://nssm.cc), lalu:

```powershell
nssm install mini-proxy "C:\mini-proxy\main.exe" proxy
nssm set mini-proxy AppDirectory C:\mini-proxy
nssm set mini-proxy AppStdout C:\mini-proxy\logs\out.log
nssm set mini-proxy AppStderr C:\mini-proxy\logs\err.log
nssm start mini-proxy
```

Karena `logging.output` pada `config.yaml` bernilai `stdout`, `AppStdout` dan
`AppStderr` perlu diisi supaya log tetap tersimpan.

Mengelola service:

```powershell
nssm restart mini-proxy
nssm stop mini-proxy
nssm remove mini-proxy confirm
```

Overhead NSSM kecil: satu proses wrapper beberapa MB dan praktis 0% CPU saat idle.

> `sc.exe create` bawaan Windows tidak bisa dipakai langsung, karena binary Go
> tidak mengimplementasikan Service Control Handler sehingga gagal dengan
> error 1053.

## Alternatif: Task Scheduler

Kalau tidak ingin menambah tool, Task Scheduler bisa dipakai. Tidak ada proses
wrapper sama sekali, tapi kontrol service dan rotasi log harus diurus sendiri.

```powershell
schtasks /create /tn "mini-proxy" /tr "cmd /c C:\mini-proxy\main.exe proxy >> C:\mini-proxy\logs\out.log 2>&1" /sc onstart /ru SYSTEM /rl HIGHEST
```

---

# Releases

Setiap release pada halaman **GitHub Releases** berisi dua jenis package:

| Package | Isi |
|---------|-----|
| `mini-proxy-{version}-{os}-{arch}` | Binary native, `config.yaml`, migrations, `build.json` |
| `mini-proxy-{version}-docker-{arch}` | Binary linux + `Dockerfile` + `docker-compose.yml` |

Beserta `SHA256SUMS` untuk verifikasi.

> `Dockerfile` dan `docker-compose.yml` hanya ada pada package **docker**.

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
