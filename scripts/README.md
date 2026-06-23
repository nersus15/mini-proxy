# Mini-Proxy Build and Release Scripts

Kumpulan script untuk membantu build dan release mini-proxy.

## Scripts

### 1. `build-release.sh`
Build release package untuk platform lokal Anda.

**Usage:**
```bash
# Build dengan versi default
./scripts/build-release.sh

# Build dengan versi spesifik
./scripts/build-release.sh v1.0.0
```

**Output:**
- Folder: `mini-proxy-{os}-{arch}/`
- Archive: `mini-proxy-{os}-{arch}.tar.gz`

### 2. `build-all-platforms.sh`
Build release package untuk SEMUA platform yang didukung (Linux, macOS, Windows) dengan arsitektur (amd64, arm64).

**Usage:**
```bash
# Build semua platform dengan versi default
./scripts/build-all-platforms.sh

# Build semua platform dengan versi spesifik
./scripts/build-all-platforms.sh v1.0.0
```

**Output:**
- Multiple folders: `mini-proxy-{os}-{arch}/`
- Multiple archives: `mini-proxy-{os}-{arch}.tar.gz`

**Platforms:**
- Linux x86_64
- Linux ARM64
- macOS x86_64
- macOS ARM64
- Windows x86_64

### 3. `pre-release-checklist.sh`
Interactive checklist untuk memastikan semua persiapan release sudah selesai.

**Usage:**
```bash
./scripts/pre-release-checklist.sh
```

**Checks:**
- Code changes complete and tested
- All tests passing
- Dependencies updated
- Version bumped
- Changelog updated
- README updated
- All commits pushed
- (Interactive prompt untuk create tag)

## Workflow

### Option 1: Local Build Testing
Gunakan untuk test build lokal sebelum push ke GitHub:

```bash
# Build untuk platform lokal
make release-local

# Extract dan test
tar -xzf mini-proxy-linux-amd64.tar.gz
cd mini-proxy-linux-amd64
./app/main proxy
```

### Option 2: Cross-Platform Local Build
Gunakan untuk build dan test untuk semua platform lokal:

```bash
# Build semua platform
make release-all

# Atau dengan versi spesifik
make release-version VERSION=v1.0.0
```

### Option 3: GitHub Automatic Release (Recommended)
Gunakan untuk production release melalui GitHub Actions:

```bash
# 1. Prepare
./scripts/pre-release-checklist.sh

# 2. Create and push tag
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions akan otomatis:
#    - Build untuk semua platform
#    - Create release package
#    - Upload ke GitHub Releases
```

## Makefile Targets

```bash
# Build local release
make release-local

# Build semua platform lokal
make release-all

# Build semua platform dengan versi
make release-version VERSION=v1.0.0
```

## File Structure dalam Release

```
mini-proxy-linux-amd64/
├── app/
│   ├── main                    # Binary Go (sudah executable)
│   ├── config.yaml.example     # Configuration template
│   └── go.work                 # Go workspace file
├── Dockerfile                  # For Docker image build
├── docker-compose.yml          # For Docker Compose
└── webcore/                    # Source code reference
    ├── main.go
    ├── go.mod
    ├── go.sum
    └── ...
```

## Menggunakan Release Package

```bash
# 1. Download dan extract
tar -xzf mini-proxy-linux-amd64.tar.gz
cd mini-proxy-linux-amd64

# 2. Setup konfigurasi
cp app/config.yaml.example app/config.yaml
# Edit app/config.yaml sesuai kebutuhan

# 3. Jalankan aplikasi
./app/main proxy

# 4. Atau gunakan Docker
docker-compose up -d
```

## Troubleshooting

### Script permission denied
```bash
chmod +x ./scripts/*.sh
```

### Build gagal karena Go tidak terinstall
```bash
# Install Go 1.25+
go version  # Check version
```

### tar.gz extraction di Windows
Gunakan: 7-Zip, WinRAR, atau WSL untuk extract

### Binary not executable di Linux/macOS
```bash
chmod +x mini-proxy-*/app/main
```

## Notes

- Semua script otomatis `go work sync` sebelum build
- Binary pre-built, user tidak perlu compile ulang
- Cross-platform compatible (Linux, macOS, Windows)
- Archive format tar.gz untuk universal compatibility
