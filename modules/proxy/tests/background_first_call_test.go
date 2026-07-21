package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

// TestRunBackgroundFirstCallUsesQueueNotOverflow adalah regression test untuk
// bug nyata yang ditemukan lewat log produksi: panggilan PERTAMA ke
// RunBackground/RunBackgroundPriority setelah server start salah masuk ke
// jalur overflow (tingkat 2), padahal antrian utama (tingkat 1, kapasitas
// besar) masih kosong sama sekali.
//
// Root cause: argumen fungsi di Go dievaluasi SEBELUM badan fungsi yang
// dipanggil mulai jalan — jadi pada panggilan pertama, channel antrian masih
// bernilai nil saat dioper sebagai argumen (poolOnce.Do belum sempat mengisi
// variabel globalnya). Kirim ke channel nil di dalam select tidak pernah
// "siap", sehingga selalu salah dianggap "antrian penuh". Sudah diperbaiki
// dengan memindahkan poolOnce.Do() ke SEBELUM argumen queue dibaca. Lihat
// komentar lengkap di RunBackground (helper/utils/background.go).
//
// Catatan: test ini memakai ResetBackgroundPoolForTest() yang membuang
// worker pool lama begitu saja (goroutine lama jadi orphan, menunggu channel
// yang sudah tidak dipakai lagi) — dapat diterima karena ini murni
// kebutuhan test untuk mensimulasikan ulang kondisi "baru start", bukan
// pola yang dipakai di kode production.
func TestRunBackgroundFirstCallUsesQueueNotOverflow(t *testing.T) {
	utils.ResetBackgroundPoolForTest()

	started := make(chan struct{})
	release := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	utils.RunBackground("first-call-regression-test", func() {
		close(started)
		<-release
		wg.Done()
	})

	select {
	case <-started:
		// Job berhasil diambil & mulai dieksekusi oleh worker tetap.
	case <-time.After(2 * time.Second):
		t.Fatal("Job tidak pernah mulai dieksekusi dalam 2 detik — kemungkinan deadlock (regresi ke bug channel nil)")
	}

	// Kalau job berhasil masuk lewat antrian utama (tingkat 1), overflowSem
	// harus tetap 0 (tidak ada slot overflow yang terpakai untuk job ini).
	if got := utils.OverflowSemLenForTest(); got != 0 {
		t.Errorf("overflowSem terpakai (%d slot) untuk panggilan PERTAMA — seharusnya 0 karena antrian utama masih kosong total dan seharusnya cukup (ini persis gejala bug yang pernah terjadi)", got)
	}

	close(release)
	wg.Wait()
}

// TestRunBackgroundPriorityFirstCallUsesQueueNotOverflow sama seperti di
// atas, tapi untuk RunBackgroundPriority (antrian credential) — bug yang
// sama persis terjadi di sini, sesuai log produksi asli: job
// "SendCredentialToProxyIL" salah masuk overflow tepat di request pertama
// server ("Background overflow goroutine dipakai (antrian penuh)"), padahal
// priorityQueue baru saja diinisialisasi dan pasti masih kosong.
func TestRunBackgroundPriorityFirstCallUsesQueueNotOverflow(t *testing.T) {
	utils.ResetBackgroundPoolForTest()

	started := make(chan struct{})
	release := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	utils.RunBackgroundPriority("SendCredentialToProxyIL-regression-test", func() {
		close(started)
		<-release
		wg.Done()
	})

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("Job prioritas tidak pernah mulai dieksekusi dalam 2 detik — kemungkinan deadlock (regresi ke bug channel nil)")
	}

	if got := utils.OverflowSemLenForTest(); got != 0 {
		t.Errorf("overflowSem terpakai (%d slot) untuk panggilan PERTAMA RunBackgroundPriority — seharusnya 0 karena priorityQueue masih kosong total (ini persis gejala bug yang terlihat di log produksi)", got)
	}

	close(release)
	wg.Wait()
}
