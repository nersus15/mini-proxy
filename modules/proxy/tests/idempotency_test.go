package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

const (
	clientA = "client-a"
	clientB = "client-b"
	urlA    = "https://api.satusehat.example/fhir-r4/v1/Patient"
)

// Body identik dari client, env, method, dan url yang sama harus menghasilkan
// sidik jari yang sama — inilah dasar dedup-nya.
func TestFingerprintSamaUntukRequestIdentik(t *testing.T) {
	body := []byte(`{"resourceType":"Patient","name":[{"text":"Budi"}]}`)

	a := utils.RequestFingerprint(clientA, "prod", "POST", urlA, body)
	b := utils.RequestFingerprint(clientA, "prod", "POST", urlA, body)

	if a != b {
		t.Errorf("sidik jari berbeda untuk request identik:\n a = %s\n b = %s", a, b)
	}
}

// Beda satu byte pun harus dianggap request baru, supaya perubahan data tidak
// pernah tertahan cache.
func TestFingerprintBerbedaUntukBodyBerbeda(t *testing.T) {
	a := utils.RequestFingerprint(clientA, "prod", "POST", urlA,
		[]byte(`{"resourceType":"Patient","name":[{"text":"Budi"}]}`))
	b := utils.RequestFingerprint(clientA, "prod", "POST", urlA,
		[]byte(`{"resourceType":"Patient","name":[{"text":"Bud1"}]}`))

	if a == b {
		t.Error("sidik jari sama padahal body berbeda")
	}
}

// Ini pengaman terpenting: dua faskes yang mengirim payload identik tidak
// boleh saling menerima respons milik yang lain.
func TestFingerprintBerbedaAntarClient(t *testing.T) {
	body := []byte(`{"resourceType":"Patient","name":[{"text":"Budi"}]}`)

	a := utils.RequestFingerprint(clientA, "prod", "POST", urlA, body)
	b := utils.RequestFingerprint(clientB, "prod", "POST", urlA, body)

	if a == b {
		t.Error("sidik jari sama untuk client berbeda — respons bisa bocor antar faskes")
	}
}

// Data dev tidak boleh tercampur dengan prod.
func TestFingerprintBerbedaAntarEnv(t *testing.T) {
	body := []byte(`{"resourceType":"Patient"}`)

	a := utils.RequestFingerprint(clientA, "dev", "POST", urlA, body)
	b := utils.RequestFingerprint(clientA, "prod", "POST", urlA, body)

	if a == b {
		t.Error("sidik jari sama untuk env berbeda")
	}
}

// Resource type berbeda berarti url berbeda, jadi harus beda sidik jari.
func TestFingerprintBerbedaAntarUrl(t *testing.T) {
	body := []byte(`{"resourceType":"Patient"}`)

	a := utils.RequestFingerprint(clientA, "prod", "POST", urlA, body)
	b := utils.RequestFingerprint(clientA, "prod", "POST",
		"https://api.satusehat.example/fhir-r4/v1/Encounter", body)

	if a == b {
		t.Error("sidik jari sama untuk url berbeda")
	}
}

// Pemisah antar bagian kunci harus mencegah ambiguitas, supaya potongan yang
// digeser tidak menghasilkan kunci yang sama.
func TestFingerprintTidakAmbiguAntarBagian(t *testing.T) {
	body := []byte(`{}`)

	a := utils.RequestFingerprint("ab", "c", "POST", urlA, body)
	b := utils.RequestFingerprint("a", "bc", "POST", urlA, body)

	if a == b {
		t.Error("sidik jari sama padahal batas client/env berbeda")
	}
}

// Method ikut menentukan kunci, dan penulisannya dinormalisasi.
func TestFingerprintMethodTidakCaseSensitive(t *testing.T) {
	body := []byte(`{}`)

	a := utils.RequestFingerprint(clientA, "prod", "POST", urlA, body)
	b := utils.RequestFingerprint(clientA, "prod", "post", urlA, body)

	if a != b {
		t.Error("sidik jari berbeda hanya karena huruf besar/kecil pada method")
	}
}

// Body kosong tetap harus menghasilkan sidik jari yang valid, bukan panic.
func TestFingerprintBodyKosong(t *testing.T) {
	got := utils.RequestFingerprint(clientA, "prod", "POST", urlA, nil)

	if len(got) != 64 {
		t.Errorf("panjang sidik jari = %d, mau 64 (sha256 hex)", len(got))
	}
}

// Optimasi cache idempotency bertumpu pada satu asumsi: respons yang disimpan
// sebagai json.RawMessage diteruskan Fiber apa adanya, tanpa di-encode ulang
// jadi string atau base64. Test ini menjaga asumsi itu.
func TestCachedResponseDikirimUtuh(t *testing.T) {
	stored := json.RawMessage(`{"resourceType":"Patient","id":"abc-123","name":[{"text":"Budi"}]}`)

	app := fiber.New()
	app.Post("/replay", func(c *fiber.Ctx) error {
		// Meniru handler PostResource saat cache ditemukan.
		var raw any = stored
		return c.Status(fiber.StatusOK).JSON(raw)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/replay", nil))
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer resp.Body.Close()

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("gagal membaca body: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, mau 200", resp.StatusCode)
	}

	var wantObj, gotObj any
	if err := json.Unmarshal(stored, &wantObj); err != nil {
		t.Fatalf("payload uji tidak valid: %v", err)
	}
	if err := json.Unmarshal(got, &gotObj); err != nil {
		t.Fatalf("respons bukan JSON valid (kemungkinan ter-encode ganda): %s", string(got))
	}

	if !reflect.DeepEqual(wantObj, gotObj) {
		t.Errorf("respons berubah saat dikirim:\n mau  = %s\n dapat = %s", string(stored), string(got))
	}
}
