package utils

import (
	"runtime/debug"
	"sync"

	"github.com/webcore-go/webcore/infra/logger"
)

// Ukuran pool & antrian background worker. Nilai ini sengaja dibuat konstanta
// dulu (bukan config) supaya perubahan awal ini minimal-risiko; bisa dipindah
// ke ModuleConfig nanti kalau perlu di-tune tanpa rebuild.
const (
	backgroundWorkerCount = 15  // worker untuk job biasa (backup/forward)
	priorityWorkerCount   = 5   // worker DEDICATED khusus job prioritas (credential) — tidak pernah ikut antre di belakang job backup
	backgroundQueueSize   = 1000
	priorityQueueSize     = 100
	backgroundOverflowMax = 100 // batas KERAS goroutine tambahan (dipakai bersama kedua jenis job) di luar worker tetap, dipakai sebagai katup pengaman saat burst
)

type backgroundJob struct {
	name string
	fn   func()
}

var (
	backgroundQueue chan backgroundJob // job biasa: backup/forward
	priorityQueue   chan backgroundJob // job prioritas: credential
	overflowSem     chan struct{}
	poolOnce        sync.Once
)

func initBackgroundPool() {
	backgroundQueue = make(chan backgroundJob, backgroundQueueSize)
	priorityQueue = make(chan backgroundJob, priorityQueueSize)
	overflowSem = make(chan struct{}, backgroundOverflowMax)

	for i := 0; i < priorityWorkerCount; i++ {
		go priorityWorker()
	}
	for i := 0; i < backgroundWorkerCount; i++ {
		go backgroundWorker()
	}

	safeLogInfo("Background worker pool diinisialisasi",
		"workers", backgroundWorkerCount, "priority_workers", priorityWorkerCount,
		"queue_size", backgroundQueueSize, "priority_queue_size", priorityQueueSize,
		"overflow_max", backgroundOverflowMax)
}

// safeLogInfo membungkus logger.Info dengan recover. Beberapa jalur startup
// (termasuk unit test standalone tanpa core.NewApp bootstrap) memanggil kode
// ini sebelum logger package siap, yang menyebabkan nil-pointer panic di
// logger itu sendiri (lihat catatan serupa di utils_url_test.go). Karena
// initBackgroundPool() dipanggil synchronous lewat sync.Once di goroutine
// PEMANGGIL RunBackground (bukan di worker goroutine yang sudah dilindungi
// RecoverBackground), panic di sini tanpa recover bisa merambat dan meng-crash
// goroutine pemanggil (mis. goroutine request HTTP yang sedang berjalan).
func safeLogInfo(msg string, args ...any) {
	defer func() {
		_ = recover()
	}()
	logger.Info(msg, args...)
}

// safeLogError sama seperti safeLogInfo tapi untuk level Error, dipakai di
// RecoverBackground(). Ini penting karena RecoverBackground() sendiri
// dipanggil dari dalam proses pemulihan panic (defer+recover) — kalau
// logger.Error() di dalamnya ikut panic (mis. logger belum siap), panic
// baru itu akan menggantikan panic asli dan tetap merambat tanpa tertahan.
func safeLogError(msg string, args ...any) {
	defer func() {
		_ = recover()
	}()
	logger.Error(msg, args...)
}

// priorityWorker HANYA memproses priorityQueue (job credential) — tidak
// pernah menyentuh backgroundQueue, supaya kapasitasnya murni dedicated dan
// tidak bisa habis dipakai job backup.
func priorityWorker() {
	for job := range priorityQueue {
		runBackgroundJob(job)
	}
}

// backgroundWorker memproses job biasa, tapi selalu CEK priorityQueue dulu
// (non-blocking) di tiap iterasi — supaya kalau priorityQueue kebetulan
// sedang penuh/backlog, worker biasa ikut membantu menguras job prioritas
// duluan sebelum lanjut ke job biasa.
func backgroundWorker() {
	for {
		select {
		case job := <-priorityQueue:
			runBackgroundJob(job)
			continue
		default:
		}

		select {
		case job := <-priorityQueue:
			runBackgroundJob(job)
		case job := <-backgroundQueue:
			runBackgroundJob(job)
		}
	}
}

func runBackgroundJob(job backgroundJob) {
	defer RecoverBackground(job.name)
	safeLogInfo("Start Job: " + job.name + " di background")
	job.fn()
}

// RunBackground mengirim job BIASA (backup/forward) — lihat RunBackgroundPriority
// untuk job kritis seperti credential yang butuh jatah dedicated.
//
// Strategi 3 tingkat, semua dengan batas keras pada jumlah goroutine yang
// bisa hidup bersamaan:
//  1. Antrian utama (backgroundQueueSize) diproses backgroundWorkerCount worker tetap.
//  2. Kalau antrian penuh (burst), pakai overflow goroutine — dibatasi keras
//     oleh overflowSem (backgroundOverflowMax), BUKAN tanpa batas.
//  3. Kalau worker tetap MAUPUN overflow sama-sama penuh (overload sungguhan,
//     sustained tinggi), RunBackground akan BLOCKING (backpressure) sampai ada
//     slot kosong. Ini sengaja: downstream (ILDKI) sendiri punya rate limit
//     terbatas, jadi menambah goroutine tanpa batas saat overload tidak ada
//     gunanya — justru memperparah tekanan ke memory/scheduler.
func RunBackground(name string, fn func()) {
	// PENTING: poolOnce.Do() WAJIB dipanggil di sini, SEBELUM backgroundQueue
	// dibaca sebagai argumen ke runBackgroundOnQueue — bukan di dalam
	// runBackgroundOnQueue itu sendiri. Argumen fungsi di Go dievaluasi SAAT
	// pemanggilan, sebelum badan fungsi yang dipanggil mulai jalan. Kalau
	// poolOnce.Do dipindah ke dalam runBackgroundOnQueue, pada panggilan
	// PERTAMA backgroundQueue masih bernilai nil saat dievaluasi sebagai
	// argumen (initBackgroundPool belum sempat mengisi variabel globalnya),
	// sehingga select di tingkat 1 tidak akan pernah "siap" (kirim ke channel
	// nil tidak pernah ready) dan SELALU salah dianggap "antrian penuh" di
	// request pertama — bug nyata yang sempat kejadian, lihat log "Background
	// overflow goroutine dipakai (antrian penuh)" yang muncul di request
	// PERTAMA server (mustahil antrian kapasitas 100 sudah penuh dengan 1 job).
	poolOnce.Do(initBackgroundPool)
	runBackgroundOnQueue(name, fn, backgroundQueue)
}

// RunBackgroundPriority sama seperti RunBackground, tapi job masuk ke
// priorityQueue yang dilayani worker DEDICATED (priorityWorkerCount) —
// dipakai untuk job kritis yang tidak boleh tertahan di belakang job biasa
// yang volumenya lebih besar (mis. sync credential ke ILDKI).
func RunBackgroundPriority(name string, fn func()) {
	poolOnce.Do(initBackgroundPool) // lihat penjelasan lengkap di RunBackground
	runBackgroundOnQueue(name, fn, priorityQueue)
}

func runBackgroundOnQueue(name string, fn func(), queue chan backgroundJob) {
	// Tingkat 1: coba antrian yang sesuai dulu (jalur tercepat & paling umum).
	select {
	case queue <- backgroundJob{name: name, fn: fn}:
		return
	default:
	}

	// Tingkat 2: antrian penuh, coba overflow (tetap dengan batas keras,
	// dipakai bersama oleh kedua jenis antrian).
	select {
	case overflowSem <- struct{}{}:
		go func() {
			defer func() { <-overflowSem }()
			defer RecoverBackground(name)
			safeLogInfo("Background overflow goroutine dipakai (antrian penuh)", "job", name)
			fn()
		}()
		return
	default:
	}

	// Tingkat 3: worker tetap DAN overflow sama-sama penuh — backpressure.
	safeLogInfo("Background pool & overflow penuh, menerapkan backpressure", "job", name)
	queue <- backgroundJob{name: name, fn: fn}
}

func RecoverBackground(name string) {
	if r := recover(); r != nil {
		safeLogError("Panic terjadi di background "+name, "err", r, "stack", string(debug.Stack()))
	}
}

// --- Hook khusus untuk unit test (black-box, package tests) ---

// ResetBackgroundPoolForTest mereset seluruh state pool (poolOnce, kedua
// antrian, overflow semaphore) ke kondisi seolah-olah server BARU SAJA
// start — supaya panggilan RunBackground/RunBackgroundPriority berikutnya
// benar-benar mengulang jalur inisialisasi "panggilan pertama", termasuk
// celah evaluasi argumen yang pernah jadi bug (lihat komentar di
// RunBackground). WAJIB hanya dipakai di test, tidak thread-safe untuk
// dipanggil bersamaan dengan trafik lain.
func ResetBackgroundPoolForTest() {
	poolOnce = sync.Once{}
	backgroundQueue = nil
	priorityQueue = nil
	overflowSem = nil
}

// OverflowSemLenForTest mengembalikan jumlah slot overflow yang SEDANG
// dipakai saat ini — dipakai test untuk membuktikan suatu job TIDAK lewat
// jalur overflow (tingkat 2), melainkan berhasil masuk antrian utama
// (tingkat 1) seperti seharusnya.
func OverflowSemLenForTest() int {
	if overflowSem == nil {
		return 0
	}
	return len(overflowSem)
}
