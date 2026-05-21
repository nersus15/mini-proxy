package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("gagal me-parse PEM private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func AddSignatureToRequest(req *http.Request) {
	timestampStr := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	dataToSign := fmt.Sprintf("%s.%s", timestampStr, string(bodyBytes))

	// 4. Lakukan Hashing SHA256 terhadap data tersebut
	hashed := sha256.Sum256([]byte(dataToSign))

	// 5. Load Private Key milik Faskes ini
	privateKey, err := loadPrivateKey("./certs/faskes-tanahabang.key")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membaca key faskes internal",
		})
	}
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membuat digital signature",
		})
	}

	// Add
}
