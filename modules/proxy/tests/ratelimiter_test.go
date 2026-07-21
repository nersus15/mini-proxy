package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

// TestExecuteBackgroundRequestFallsBackToOnErrorWhenLimiterTimesOut membuktikan
// bahwa saat rate limiter BACKUP (dipakai ExecuteBackgroundRequest, untuk
// jalur forward/backup resource) tidak kunjung dapat token dalam batas waktu
// tunggu, ExecuteBackgroundRequest() BENAR-BENAR memanggil onError (bukan
// tetap mengirim request, dan bukan diam tanpa memanggil callback apapun).
func TestExecuteBackgroundRequestFallsBackToOnErrorWhenLimiterTimesOut(t *testing.T) {
	// burst=0 -> bucket selalu kosong (tokens tidak akan pernah >= 1), jadi
	// wait() PASTI timeout, tanpa perlu menunggu 30 detik asli.
	restore := utils.SetIldkiBackupRateLimiterForTest(0.0001, 0, 100*time.Millisecond)
	defer restore()

	requestReachedServer := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReachedServer = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("gagal membuat request: %v", err)
	}

	var mu sync.Mutex
	var onSuccessCalled, onErrorCalled bool
	var receivedErr error

	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		utils.ExecuteBackgroundRequest(client, req, func(resp *http.Response) {
			mu.Lock()
			onSuccessCalled = true
			mu.Unlock()
		}, func(err error) {
			mu.Lock()
			onErrorCalled = true
			receivedErr = err
			mu.Unlock()
		})
	}()
	wg.Wait()
	elapsed := time.Since(start)

	mu.Lock()
	defer mu.Unlock()

	t.Logf("Waktu tunggu sebelum fallback: %s (timeout diset: 100ms)", elapsed)

	if requestReachedServer {
		t.Error("Request sampai ke server — seharusnya TIDAK PERNAH terkirim karena limiter harus timeout duluan")
	}
	if onSuccessCalled {
		t.Error("onSuccess terpanggil — seharusnya limiter timeout dan request tidak pernah dikirim")
	}
	if !onErrorCalled {
		t.Fatal("onError TIDAK terpanggil — jalur fallback rate-limit-timeout tidak bekerja sama sekali (bug nyata, bukan cuma belum teruji)")
	}
	if receivedErr == nil {
		t.Fatal("onError terpanggil tapi error yang diterima nil")
	}
	if !errors.Is(receivedErr, context.DeadlineExceeded) {
		t.Errorf("error yang diterima (%v) seharusnya membungkus context.DeadlineExceeded (dari timeout limiter)", receivedErr)
	}

	t.Logf("onError terpanggil sesuai harapan dengan pesan: %v", receivedErr)
}

// TestExecuteBackgroundRequestSucceedsWhenTokenAvailable adalah kontrol
// pembanding: memastikan kalau token TERSEDIA, request tetap benar-benar
// terkirim dan onSuccess terpanggil seperti biasa (bukti limiter tidak salah
// menahan semua request tanpa pandang bulu).
func TestExecuteBackgroundRequestSucceedsWhenTokenAvailable(t *testing.T) {
	restore := utils.SetIldkiBackupRateLimiterForTest(10, 5, 5*time.Second) // token cukup tersedia
	defer restore()

	requestReachedServer := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReachedServer = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("gagal membuat request: %v", err)
	}

	var mu sync.Mutex
	var onSuccessCalled, onErrorCalled bool

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.ExecuteBackgroundRequest(client, req, func(resp *http.Response) {
			mu.Lock()
			onSuccessCalled = true
			mu.Unlock()
		}, func(err error) {
			mu.Lock()
			onErrorCalled = true
			mu.Unlock()
			t.Logf("onError tidak terduga: %v", err)
		})
	}()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if !requestReachedServer {
		t.Error("Request tidak sampai ke server, padahal token seharusnya tersedia")
	}
	if !onSuccessCalled {
		t.Error("onSuccess tidak terpanggil, padahal request seharusnya berhasil")
	}
	if onErrorCalled {
		t.Error("onError terpanggil, padahal token seharusnya tersedia dan request seharusnya berhasil")
	}
}

// TestExecuteCredentialBackgroundRequestNotBlockedByBackupLimiter membuktikan
// perbaikan utama: kalau limiter BACKUP sedang kosong/timeout, limiter
// CREDENTIAL yang terpisah tetap bisa mengirim request seperti biasa — tidak
// ikut tertahan. Ini adalah bukti langsung bahwa backup tidak bisa lagi
// menyandera credential.
func TestExecuteCredentialBackgroundRequestNotBlockedByBackupLimiter(t *testing.T) {
	// Backup limiter sengaja dibuat kosong total (mensimulasikan backup
	// sedang penuh/burst tinggi).
	restoreBackup := utils.SetIldkiBackupRateLimiterForTest(0.0001, 0, 200*time.Millisecond)
	defer restoreBackup()
	// Credential limiter tetap punya token tersedia.
	restoreCredential := utils.SetIldkiCredentialRateLimiterForTest(10, 5, 5*time.Second)
	defer restoreCredential()

	requestReachedServer := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReachedServer = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("gagal membuat request: %v", err)
	}

	var mu sync.Mutex
	var onSuccessCalled, onErrorCalled bool

	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		// Pakai ExecuteCredentialBackgroundRequest, BUKAN ExecuteBackgroundRequest.
		utils.ExecuteCredentialBackgroundRequest(client, req, func(resp *http.Response) {
			mu.Lock()
			onSuccessCalled = true
			mu.Unlock()
		}, func(err error) {
			mu.Lock()
			onErrorCalled = true
			mu.Unlock()
			t.Logf("onError tidak terduga: %v", err)
		})
	}()
	wg.Wait()
	elapsed := time.Since(start)

	mu.Lock()
	defer mu.Unlock()

	t.Logf("Waktu tempuh request credential (walau backup limiter kosong): %s", elapsed)

	if !requestReachedServer {
		t.Error("Request credential tidak sampai ke server — seharusnya TIDAK terpengaruh backup limiter yang kosong")
	}
	if !onSuccessCalled {
		t.Error("onSuccess tidak terpanggil untuk request credential, padahal limiter credential-nya sendiri punya token tersedia")
	}
	if onErrorCalled {
		t.Error("onError terpanggil — berarti request credential ikut tertahan/gagal gara-gara backup limiter, padahal seharusnya terpisah total")
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("Request credential makan waktu %s — terlalu lama untuk limiter yang tokennya tersedia, indikasi kemungkinan tercampur dengan limiter backup", elapsed)
	}
}

// TestWaitForIldkiSlotRoutesByResourceType membuktikan bahwa WaitForIldkiSlot
// (dipakai cronjob retry di task.go) benar-benar mengarahkan resourceType
// "credential" ke limiter credential, dan resourceType lain (mis. "Bundle",
// "Patient") ke limiter backup — bukan satu limiter gabungan seperti
// sebelum perbaikan. Ini menutup celah cronjob retry yang sebelumnya sama
// sekali tidak melewati limiter manapun.
func TestWaitForIldkiSlotRoutesByResourceType(t *testing.T) {
	// Backup limiter dikosongkan total (selalu timeout), credential limiter
	// tetap ada tokennya. Kalau routing benar: resourceType "credential"
	// harus tetap SUKSES (cepat), resourceType lain harus TIMEOUT (gagal).
	restoreBackup := utils.SetIldkiBackupRateLimiterForTest(0.0001, 0, 150*time.Millisecond)
	defer restoreBackup()
	restoreCredential := utils.SetIldkiCredentialRateLimiterForTest(10, 5, 5*time.Second)
	defer restoreCredential()

	// resourceType "credential" -> harus pakai limiter credential (tokennya ada) -> sukses cepat.
	start := time.Now()
	errCredential := utils.WaitForIldkiSlot("credential")
	elapsedCredential := time.Since(start)

	if errCredential != nil {
		t.Errorf("WaitForIldkiSlot(\"credential\") gagal (%v), padahal limiter credential seharusnya punya token tersedia", errCredential)
	}
	if elapsedCredential > 100*time.Millisecond {
		t.Errorf("WaitForIldkiSlot(\"credential\") makan waktu %s, terlalu lama untuk limiter yang tokennya tersedia — indikasi salah routing ke limiter backup yang kosong", elapsedCredential)
	}

	// resourceType selain "credential" (mis. "Bundle") -> harus pakai limiter
	// backup (kosong total) -> timeout.
	start = time.Now()
	errBackup := utils.WaitForIldkiSlot("Bundle")
	elapsedBackup := time.Since(start)

	if errBackup == nil {
		t.Error("WaitForIldkiSlot(\"Bundle\") berhasil tanpa menunggu, padahal limiter backup seharusnya kosong total dan harus timeout")
	}
	if elapsedBackup < 100*time.Millisecond {
		t.Errorf("WaitForIldkiSlot(\"Bundle\") kembali dalam %s, terlalu cepat untuk limiter backup yang kosong (timeout diset 150ms) — indikasi salah routing ke limiter credential yang masih ada token", elapsedBackup)
	}

	t.Logf("credential: %s (err=%v) | backup/Bundle: %s (err=%v)", elapsedCredential, errCredential, elapsedBackup, errBackup)
}
