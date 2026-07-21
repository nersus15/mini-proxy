package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

// backgroundWorkerCountExpected harus sama dengan konstanta backgroundWorkerCount
// di modules/proxy/helper/utils/background.go. Di-duplikasi di sini karena
// unexported dan test ini sengaja black-box (package tests, bukan package utils),
// mengikuti pola test lain di folder ini.
const backgroundWorkerCountExpected = 20

// TestHttpClientConnectionReuse memverifikasi bahwa CreateHttpClient() melakukan
// reuse koneksi TCP+TLS di bawah traffic concurrent tinggi ke host yang sama,
// alih-alih membuka koneksi baru di tiap request (perilaku default Go sebelum
// MaxIdleConnsPerHost/MaxConnsPerHost diset).
func TestHttpClientConnectionReuse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond) // simulasi latency jaringan ke ILDKI/SatuSehat
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := utils.CreateHttpClient("")

	const totalRequests = 500
	const concurrency = 100

	var newConns int64
	var reusedConns int64
	var failedRequests int64

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	start := time.Now()
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			trace := &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) {
					if info.Reused {
						atomic.AddInt64(&reusedConns, 1)
					} else {
						atomic.AddInt64(&newConns, 1)
					}
				},
			}
			ctx := httptrace.WithClientTrace(context.Background(), trace)
			req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			if err != nil {
				atomic.AddInt64(&failedRequests, 1)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt64(&failedRequests, 1)
				return
			}
			// PENTING: body harus di-drain (dibaca sampai EOF) sebelum Close(),
			// kalau tidak Go's http.Transport tidak bisa mengembalikan koneksi
			// ke idle pool untuk di-reuse — akan selalu dianggap 0% reuse walau
			// connection pooling di httpclient.go sudah benar.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	reuseRate := float64(reusedConns) / float64(totalRequests) * 100

	t.Logf("=== HTTP Client Connection Reuse ===")
	t.Logf("Total request      : %d", totalRequests)
	t.Logf("Concurrency        : %d", concurrency)
	t.Logf("Waktu total        : %s", elapsed)
	t.Logf("Throughput         : %.1f req/s", float64(totalRequests)/elapsed.Seconds())
	t.Logf("Koneksi baru       : %d", newConns)
	t.Logf("Koneksi di-reuse   : %d", reusedConns)
	t.Logf("Reuse rate         : %.1f%%", reuseRate)
	t.Logf("Request gagal      : %d", failedRequests)

	if failedRequests > 0 {
		t.Errorf("Ada %d request gagal, seharusnya 0 untuk dummy server lokal", failedRequests)
	}

	// Toleransi kecil di atas concurrency untuk warm-up awal pool.
	if newConns > int64(concurrency)+10 {
		t.Errorf(
			"Koneksi baru (%d) jauh melebihi batas wajar (~%d) — connection pooling kemungkinan tidak bekerja sesuai harapan (cek MaxIdleConnsPerHost/MaxConnsPerHost di httpclient.go)",
			newConns, concurrency,
		)
	}

	if reuseRate < 70 {
		t.Errorf("Reuse rate cuma %.1f%%, diharapkan jauh lebih tinggi (>70%%) dengan connection pooling aktif", reuseRate)
	}
}

// TestBackgroundWorkerPoolBounded memverifikasi bahwa RunBackground() tidak
// membuat goroutine tumbuh tanpa batas saat menerima banyak job sekaligus
// (perilaku lama: goroutine baru di tiap panggilan, tanpa batas atas).
func TestBackgroundWorkerPoolBounded(t *testing.T) {
	const totalJobs = 2000
	const jobDuration = 10 * time.Millisecond

	// Beri jeda sebentar & GC supaya baseline goroutine stabil sebelum diukur,
	// termasuk goroutine worker pool yang sudah diinisialisasi dari test sebelumnya.
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	var completed int64
	var wg sync.WaitGroup
	wg.Add(totalJobs)

	var peakMu sync.Mutex

	stopMonitor := make(chan struct{})
	monitorStarted := make(chan struct{})
	monitorDone := make(chan struct{})
	var peak int
	go func() {
		defer close(monitorDone)
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()
		close(monitorStarted) // sinyal goroutine monitor sudah hidup, aman untuk ambil baseline sekarang
		for {
			select {
			case <-stopMonitor:
				return
			case <-ticker.C:
				cur := runtime.NumGoroutine()
				peakMu.Lock()
				if cur > peak {
					peak = cur
				}
				peakMu.Unlock()
			}
		}
	}()
	<-monitorStarted // pastikan goroutine monitor sudah dihitung SEBELUM baseline diambil, supaya tidak jadi bias +1 di delta

	baseline := runtime.NumGoroutine()
	peakMu.Lock()
	peak = baseline
	peakMu.Unlock()

	start := time.Now()
	for i := 0; i < totalJobs; i++ {
		jobName := fmt.Sprintf("perf-test-job-%d", i)
		utils.RunBackground(jobName, func() {
			time.Sleep(jobDuration) // simulasi kerja: mis. panggilan HTTP keluar ke ILDKI
			atomic.AddInt64(&completed, 1)
			wg.Done()
		})
	}
	wg.Wait()
	elapsed := time.Since(start)

	close(stopMonitor)
	<-monitorDone

	peakMu.Lock()
	finalPeak := peak
	peakMu.Unlock()

	goroutineDelta := finalPeak - baseline

	t.Logf("=== Background Worker Pool ===")
	t.Logf("Total job          : %d", totalJobs)
	t.Logf("Waktu total        : %s", elapsed)
	t.Logf("Throughput         : %.1f job/s", float64(totalJobs)/elapsed.Seconds())
	t.Logf("Goroutine baseline : %d", baseline)
	t.Logf("Goroutine peak     : %d", finalPeak)
	t.Logf("Selisih (delta)    : %d", goroutineDelta)
	t.Logf("Job selesai        : %d", completed)

	if completed != totalJobs {
		t.Errorf("Job selesai = %d, want %d (ada job yang hilang/tidak jalan)", completed, totalJobs)
	}

	// Toleransi: worker pool tetap (backgroundWorkerCountExpected) + overflow
	// terbatas (lihat backgroundOverflowMax di background.go, saat ini 100) +
	// sedikit kelonggaran untuk goroutine test itu sendiri (mis. runtime
	// scheduler, GC assist goroutine sesaat).
	maxAcceptableDelta := backgroundWorkerCountExpected + 100 + 5
	if goroutineDelta > maxAcceptableDelta {
		t.Errorf(
			"Goroutine delta (%d) jauh melebihi ekspektasi bounded pool (worker tetap=%d, toleransi total=%d) — kemungkinan overflow semaphore di background.go tidak bekerja sesuai harapan",
			goroutineDelta, backgroundWorkerCountExpected, maxAcceptableDelta,
		)
	}
}

// TestExecuteBackgroundRequestRateLimited memverifikasi bahwa rate limiter
// BACKUP di ExecuteBackgroundRequest() benar-benar membatasi laju request
// keluar ke ILDKI, walau burst yang masuk jauh melebihi limit. Limiter
// backup dikalibrasi ~12 request/detik dengan burst 15 (dari total budget
// 18/detik yang sekarang dibagi: 6/detik untuk credential, 12/detik untuk
// backup), berdasarkan WAF rule ILDKI:
//
//	stick-table type ip size 100k expire 60s store http_req_rate(10s)
//	http-request deny if { sc_http_req_rate(0) gt 240 }
//
// yaitu 240 request per 10 detik = 24 request/detik gabungan.
func TestExecuteBackgroundRequestRateLimited(t *testing.T) {
	const limiterRate = 12.0
	const limiterBurst = 15.0

	var mu sync.Mutex
	var arrivalTimes []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		arrivalTimes = append(arrivalTimes, time.Now())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 20 * time.Second}

	const totalRequests = 200 // jauh melebihi burst supaya limiter benar-benar diuji

	var wg sync.WaitGroup
	wg.Add(totalRequests)

	start := time.Now()
	for i := 0; i < totalRequests; i++ {
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Errorf("gagal membuat request: %v", err)
				return
			}
			utils.ExecuteBackgroundRequest(client, req, func(resp *http.Response) {
				_, _ = io.Copy(io.Discard, resp.Body)
			}, func(err error) {
				t.Errorf("request gagal (tidak diharapkan dalam batas waktu test 30s timeout limiter): %v", err)
			})
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	mu.Lock()
	times := append([]time.Time(nil), arrivalTimes...)
	mu.Unlock()

	t.Logf("=== Rate Limiter Backup ke ILDKI (ExecuteBackgroundRequest) ===")
	t.Logf("Total request dikirim          : %d", totalRequests)
	t.Logf("Total request diterima server  : %d", len(times))
	t.Logf("Waktu total                    : %s", elapsed)

	if len(times) != totalRequests {
		t.Errorf("Request diterima server = %d, want %d", len(times), totalRequests)
	}

	// Hitung request per jendela 1 detik (bucket berdasarkan detik ke berapa
	// sejak start), cari bucket dengan jumlah tertinggi.
	buckets := map[int]int{}
	for _, ts := range times {
		sec := int(ts.Sub(start).Seconds())
		buckets[sec]++
	}
	maxInWindow := 0
	for _, count := range buckets {
		if count > maxInWindow {
			maxInWindow = count
		}
	}
	t.Logf("Request tertinggi dalam 1 jendela detik : %d", maxInWindow)

	// Toleransi: rate + burst + kelonggaran timing jitter. Sengaja lebih
	// tinggi dari rate murni karena burst awal boleh "meledak" sesaat, tapi
	// TIDAK BOLEH mendekati totalRequests (200) yang berarti limiter tidak
	// bekerja sama sekali.
	maxAcceptablePerSecond := limiterRate + limiterBurst + 10
	if float64(maxInWindow) > maxAcceptablePerSecond {
		t.Errorf(
			"Request per detik tertinggi (%d) melebihi batas wajar (~%.0f) — rate limiter backup kemungkinan tidak bekerja",
			maxInWindow, maxAcceptablePerSecond,
		)
	}

	// Waktu total MINIMAL yang diharapkan sekitar (total-burst)/rate. Kalau
	// waktu total jauh lebih cepat dari itu, berarti limiter tidak benar-benar
	// menahan laju request.
	minExpectedSeconds := (float64(totalRequests) - limiterBurst) / limiterRate
	if elapsed.Seconds() < minExpectedSeconds*0.5 { // toleransi 50% untuk timing jitter
		t.Errorf(
			"Waktu total (%s) jauh lebih cepat dari ekspektasi minimal (~%.1fs) — rate limiter kemungkinan tidak menahan laju request",
			elapsed, minExpectedSeconds,
		)
	}
}

// BenchmarkRunBackground mengukur overhead per-panggilan RunBackground di bawah
// beban tinggi. Jalankan dengan: go test -bench=BenchmarkRunBackground -benchtime=3s ./modules/proxy/tests/
func BenchmarkRunBackground(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.RunBackground("bench-job", func() {
			wg.Done()
		})
	}
	wg.Wait()
}
