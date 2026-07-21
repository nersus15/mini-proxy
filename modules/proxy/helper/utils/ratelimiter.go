package utils

import (
	"context"
	"math"
	"sync"
	"time"
)

// tokenBucket adalah implementasi rate limiter sederhana (token bucket) tanpa
// dependency eksternal (golang.org/x/time/rate sengaja tidak dipakai supaya
// tidak perlu ubah go.mod/go.sum + go get di environment production).
//
// Cara kerja: "ember" berisi token yang terisi ulang dengan kecepatan tetap
// (refillRate token/detik). Tiap mau kirim request, harus ambil 1 token dulu;
// kalau ember kosong, menunggu sampai ada token baru (atau context timeout).
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // token per detik
	lastRefill time.Time
}

func newTokenBucket(ratePerSecond float64, burst float64) *tokenBucket {
	return &tokenBucket{
		tokens:     burst,
		maxTokens:  burst,
		refillRate: ratePerSecond,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) refillLocked() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.maxTokens, b.tokens+elapsed*b.refillRate)
		b.lastRefill = now
	}
}

// tryTake mencoba ambil 1 token tanpa menunggu. Return true kalau berhasil.
func (b *tokenBucket) tryTake() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refillLocked()
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// wait menunggu sampai dapat 1 token, atau context selesai (timeout/cancel).
// Dicek tiap 20ms supaya tidak busy-loop tapi tetap responsif.
func (b *tokenBucket) wait(ctx context.Context) error {
	if b.tryTake() {
		return nil
	}

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if b.tryTake() {
				return nil
			}
		}
	}
}

// --- Dua limiter TERPISAH untuk ILDKI, bukan satu limiter gabungan ---
//
// Sebelumnya credential-sync dan backup-forward berbagi SATU limiter+antrean
// yang sama, tanpa prioritas — kalau backup-forward (volume biasanya jauh
// lebih besar) menumpuk duluan, credential-sync ikut tertahan di
// belakangnya. Padahal credential-sync itu kritis: ILDKI butuh token faskes
// segera untuk resolve dependency resource (mis. Patient) saat clone
// Encounter. Dipisah supaya credential PUNYA JATAH SENDIRI yang tidak bisa
// direbut oleh trafik backup, apapun besar volumenya.
//
// Total gabungan (6+12=18/detik) tetap sama dengan sebelumnya, di bawah
// batas WAF ILDKI (24/detik dari 240 request/10 detik), cuma dibagi supaya
// tidak saling menyandera.

// ildkiCredentialRateLimiter reserved KHUSUS untuk sync credential ke ILDKI.
// Volume credential jauh lebih kecil dari backup, jadi jatah kecil ini
// biasanya lebih dari cukup — yang penting selalu TERSEDIA, tidak digerus
// backup.
var ildkiCredentialRateLimiter = newTokenBucket(6, 8)

// ildkiBackupRateLimiter dipakai untuk forward/backup resource ke ILDKI —
// volume lebih besar dan lebih toleran terhadap delay (bukan jalur kritis
// dependency resolution seperti credential).
var ildkiBackupRateLimiter = newTokenBucket(12, 15)

// ildkiRateLimitWaitTimeout adalah batas maksimal request background boleh
// "mengantre" menunggu jatah rate limit (berlaku untuk kedua limiter di
// atas). Kalau lewat dari ini, request dilepas ke mekanisme retry yang sudah
// ada (disimpan sebagai transaksi PENDING) alih-alih terus menahan
// goroutine/slot worker pool selamanya.
//
// Bukan const supaya bisa disuntik nilai lain lewat SetIldki*RateLimiterForTest
// saat unit test.
var ildkiRateLimitWaitTimeout = 30 * time.Second

// WaitForIldkiSlot menunggu jatah rate limit ke ILDKI sebelum request boleh
// dikirim, memilih limiter berdasarkan resourceType ("credential" pakai
// limiter credential yang reserved, selain itu pakai limiter backup).
//
// Ini dipakai oleh cronjob retry (TaskService.sendRequest di task.go) supaya
// jalur retry ikut terlindungi rate limit yang SAMA dengan jalur real-time-
// triggered background (ExecuteBackgroundRequest / ExecuteCredentialBackgroundRequest)
// — sebelumnya cron punya jalur HTTP sendiri yang sama sekali tidak melewati
// limiter manapun, jadi bisa saja tetap kena deny 403 WAF ILDKI walau
// limiter di jalur lain sudah benar.
//
// Return error (biasanya context.DeadlineExceeded) kalau sudah menunggu
// sampai ildkiRateLimitWaitTimeout tapi jatah belum juga longgar — pemanggil
// sebaiknya menandai task tsb gagal untuk siklus ini (RetryCount naik),
// bukan memaksa kirim tanpa jatah, karena tetap akan ditolak WAF.
func WaitForIldkiSlot(resourceType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ildkiRateLimitWaitTimeout)
	defer cancel()

	limiter := ildkiBackupRateLimiter
	if resourceType == "credential" {
		limiter = ildkiCredentialRateLimiter
	}
	return limiter.wait(ctx)
}

// SetIldkiCredentialRateLimiterForTest & SetIldkiBackupRateLimiterForTest
// mengganti limiter & timeout terkait secara SEMENTARA, khusus untuk
// kebutuhan unit test black-box di modules/proxy/tests. WAJIB panggil fungsi
// restore yang dikembalikan (idealnya lewat defer).
func SetIldkiCredentialRateLimiterForTest(ratePerSecond, burst float64, waitTimeout time.Duration) (restore func()) {
	origLimiter := ildkiCredentialRateLimiter
	origTimeout := ildkiRateLimitWaitTimeout

	ildkiCredentialRateLimiter = newTokenBucket(ratePerSecond, burst)
	ildkiRateLimitWaitTimeout = waitTimeout

	return func() {
		ildkiCredentialRateLimiter = origLimiter
		ildkiRateLimitWaitTimeout = origTimeout
	}
}

func SetIldkiBackupRateLimiterForTest(ratePerSecond, burst float64, waitTimeout time.Duration) (restore func()) {
	origLimiter := ildkiBackupRateLimiter
	origTimeout := ildkiRateLimitWaitTimeout

	ildkiBackupRateLimiter = newTokenBucket(ratePerSecond, burst)
	ildkiRateLimitWaitTimeout = waitTimeout

	return func() {
		ildkiBackupRateLimiter = origLimiter
		ildkiRateLimitWaitTimeout = origTimeout
	}
}
