package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
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

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing private key dengan format PKCS1 maupun PKCS8: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("kunci bukan bertipe RSA")
	}

	return rsaKey, nil
}

func AddSignatureToRequest(clientId string, req *http.Request, body []byte) error {

	timestampStr := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	hasher := sha256.New()

	hasher.Write([]byte(timestampStr))
	hasher.Write([]byte(":"))
	hasher.Write(body)

	hashed := hasher.Sum(nil)

	privateKey, err := LoadPrivateKey("modules/proxy/certs/private.key")
	if err != nil {
		return fmt.Errorf("Gagal membuat digital signature: %s", err.Error())
	}
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return fmt.Errorf("Gagal membuat digital signature: %s", err.Error())
	}

	signatureBase64 := base64.StdEncoding.EncodeToString(signatureBytes)

	req.Header.Set("X-Faskes-ID", clientId)
	req.Header.Set("X-Faskes-Signature", signatureBase64)
	req.Header.Set("X-Faskes-Timestamp", timestampStr)

	return nil
}
