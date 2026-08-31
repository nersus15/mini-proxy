package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/webcore-go/webcore/infra/logger"
)

const DefaultIdempotencyTTLMinutes = 5

func RequestFingerprint(client string, env string, method string, url string, body []byte) string {
	h := sha256.New()

	// Pemisah \x00 mencegah client "a"+env "bc" berkunci sama dengan "ab"+"c".
	for _, part := range []string{client, env, strings.ToUpper(method), url} {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	h.Write(body)

	return hex.EncodeToString(h.Sum(nil))
}

// IdempotencyMemKey memberi prefiks supaya tidak bentrok dengan key cache lain.
func IdempotencyMemKey(fingerprint string) string {
	return "idem_" + fingerprint
}

// IdempotencyTTL jatuh ke default kalau config tidak diisi.
func IdempotencyTTL(minutes int) time.Duration {
	if minutes <= 0 {
		minutes = DefaultIdempotencyTTLMinutes
	}
	return time.Duration(minutes) * time.Minute
}

func ClientIdFromAuth(repo *repository.ProxyRepository, auth string) string {
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if token == "" {
		return ""
	}

	if tokens := repo.GetToken(&token); len(tokens) == 1 {
		return tokens[0].ClientId
	}

	return token
}

// SetIdempotencyMemory menyimpan respons di cache memory.
func SetIdempotencyMemory(repo *repository.ProxyRepository, fingerprint string, body json.RawMessage, ttl time.Duration) {
	mem := repo.Memory
	if mem == nil || ttl <= 0 {
		return
	}

	if err := mem.Set(IdempotencyMemKey(fingerprint), body, ttl); err != nil {
		logger.Debug("Gagal menyimpan cache idempotency di memory", "err", err.Error())
	}
}

func LookupIdempotency(repo *repository.ProxyRepository, fingerprint string) (any, bool) {
	if fingerprint == "" {
		return nil, false
	}

	if mem := repo.Memory; mem != nil {
		var cached json.RawMessage
		if mem.Get(IdempotencyMemKey(fingerprint), &cached) && len(cached) > 0 {
			return cached, true
		}
	}

	record, err := repo.FindIdempotency(fingerprint)
	if err != nil {
		logger.Error("Gagal membaca cache idempotency", "err", err.Error())
		return nil, false
	}
	if record == nil || len(record.ResponseBody) == 0 {
		return nil, false
	}

	SetIdempotencyMemory(repo, fingerprint, record.ResponseBody, time.Until(record.ExpiredAt))

	return record.ResponseBody, true
}

// IdempotencyRecord adalah data untuk menyimpan satu entri.
type IdempotencyRecord struct {
	Fingerprint  string
	Client       string
	Env          string
	Method       string
	ResourceType string
	Url          string
	ResourceID   string
	ResponseBody []byte
	TTL          time.Duration
}

func StoreIdempotency(repo *repository.ProxyRepository, rec IdempotencyRecord) {
	if rec.Fingerprint == "" || len(rec.ResponseBody) == 0 {
		return
	}

	SetIdempotencyMemory(repo, rec.Fingerprint, rec.ResponseBody, rec.TTL)

	now := time.Now()
	record := &entity.RequestIdempotency{
		Fingerprint:  rec.Fingerprint,
		Client:       rec.Client,
		Env:          rec.Env,
		Method:       rec.Method,
		ResourceType: rec.ResourceType,
		Url:          rec.Url,
		ResourceID:   rec.ResourceID,
		ResponseBody: rec.ResponseBody,
		CreatedAt:    now,
		ExpiredAt:    now.Add(rec.TTL),
	}

	RunBackground("SaveIdempotency", func() {
		if err := repo.SaveIdempotency(record); err != nil {
			logger.Error("Gagal menyimpan cache idempotency", "err", err.Error(), "fingerprint", rec.Fingerprint)
		}
	})
}

func CleanupIdempotency(repo *repository.ProxyRepository) {
	deleted, err := repo.DeleteExpiredIdempotency()
	if err != nil {
		logger.Error("Gagal membersihkan cache idempotency", "err", err.Error())
		return
	}
	if deleted > 0 {
		logger.Info("Cache idempotency dibersihkan", "dihapus", deleted)
	}
}
