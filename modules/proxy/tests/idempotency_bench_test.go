package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

// Benchmark biaya sinkron lapisan idempotency per POST:
//
//	go test -bench=BenchmarkIdempotency -benchmem ./tests/
//
// Penulisan database tidak diukur karena sudah dipindah ke worker background.

const benchUrl = "https://api.satusehat.example/fhir-r4/v1/Patient"

// patientPayload adalah resource tunggal, ukuran khas request harian.
func patientPayload() []byte {
	return []byte(`{"resourceType":"Patient","identifier":[{"system":"https://fhir.kemkes.go.id/id/nik","value":"3374010101900001"}],"name":[{"use":"official","text":"Budi Santoso"}],"gender":"male","birthDate":"1990-01-01","address":[{"use":"home","city":"Semarang","postalCode":"50123","country":"ID"}]}`)
}

// bundlePayload membuat transaction Bundle dengan sejumlah entry, untuk
// melihat bagaimana biaya tumbuh pada payload besar.
func bundlePayload(entries int) []byte {
	var b strings.Builder
	b.WriteString(`{"resourceType":"Bundle","type":"transaction","entry":[`)

	for i := 0; i < entries; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"fullUrl":"urn:uuid:%08d-0000-4000-8000-000000000000","resource":{"resourceType":"Observation","status":"final","code":{"coding":[{"system":"http://loinc.org","code":"8867-4","display":"Heart rate"}]},"subject":{"reference":"Patient/P%06d"},"effectiveDateTime":"2026-08-31T08:%02d:00+07:00","valueQuantity":{"value":%d,"unit":"beats/minute","system":"http://unitsofmeasure.org","code":"/min"}},"request":{"method":"POST","url":"Observation"}}`,
			i, i, i%60, 60+i%40)
	}

	b.WriteString(`]}`)
	return []byte(b.String())
}

var benchPayloads = []struct {
	name string
	body []byte
}{
	{"Patient", patientPayload()},
	{"Bundle10", bundlePayload(10)},
	{"Bundle100", bundlePayload(100)},
	{"Bundle500", bundlePayload(500)},
}

// BenchmarkIdempotencyFingerprint mengukur biaya hash sidik jari request.
func BenchmarkIdempotencyFingerprint(b *testing.B) {
	for _, p := range benchPayloads {
		b.Run(fmt.Sprintf("%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = utils.RequestFingerprint("client-bench", "prod", "POST", benchUrl, p.body)
			}
		})
	}
}

// BenchmarkIdempotencyEncode mengukur biaya menyiapkan respons sebelum
// disimpan ke cache (json.Marshal pada storeIdempotency).
func BenchmarkIdempotencyEncode(b *testing.B) {
	for _, p := range benchPayloads {
		var raw any
		if err := json.Unmarshal(p.body, &raw); err != nil {
			b.Fatalf("payload tidak valid: %v", err)
		}

		b.Run(fmt.Sprintf("%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := json.Marshal(raw); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkIdempotencyDecode mengukur biaya membaca respons tersimpan saat
// cache ditemukan — inilah yang dibayar pada request retry.
func BenchmarkIdempotencyDecode(b *testing.B) {
	for _, p := range benchPayloads {
		b.Run(fmt.Sprintf("%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var raw any
				if err := json.Unmarshal(p.body, &raw); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkIdempotencyWritePath membandingkan biaya sinkron pada POST sukses
// sebelum dan sesudah byte respons disimpan apa adanya.
//
//	Marshal = versi lama, respons di-marshal ulang dari map
//	Passthrough = versi sekarang, byte asli SatuSehat langsung dipakai
func BenchmarkIdempotencyWritePath(b *testing.B) {
	for _, p := range benchPayloads {
		var raw any
		if err := json.Unmarshal(p.body, &raw); err != nil {
			b.Fatalf("payload tidak valid: %v", err)
		}

		b.Run(fmt.Sprintf("Marshal_%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = utils.RequestFingerprint("client-bench", "prod", "POST", benchUrl, p.body)
				if _, err := json.Marshal(raw); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Passthrough_%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = utils.RequestFingerprint("client-bench", "prod", "POST", benchUrl, p.body)
			}
		})
	}
}

// BenchmarkIdempotencyReplay membandingkan biaya melayani request retry yang
// kena cache.
//
//	Remarshal = versi lama, unmarshal ke map lalu handler marshal lagi
//	RawMessage = versi sekarang, byte tersimpan diteruskan apa adanya
func BenchmarkIdempotencyReplay(b *testing.B) {
	for _, p := range benchPayloads {
		b.Run(fmt.Sprintf("Remarshal_%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var raw any
				if err := json.Unmarshal(p.body, &raw); err != nil {
					b.Fatal(err)
				}
				if _, err := json.Marshal(raw); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("RawMessage_%s_%dB", p.name, len(p.body)), func(b *testing.B) {
			b.SetBytes(int64(len(p.body)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := json.Marshal(json.RawMessage(p.body)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkIdempotencyLookupMiss mengukur jalur request yang bukan retry —
// kondisi mayoritas. Hanya hash yang dibayar sebelum diteruskan ke SatuSehat.
func BenchmarkIdempotencyLookupMiss(b *testing.B) {
	p := benchPayloads[0]

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = utils.RequestFingerprint("client-bench", "prod", "POST", benchUrl, p.body)
	}
}
