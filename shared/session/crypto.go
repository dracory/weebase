package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

// encrypt encrypts data using AES-GCM
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// encodeSessionData encodes session data to a base64 string
func encodeSessionData(data map[string]interface{}, key []byte) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	encrypted, err := encrypt(jsonData, key)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(encrypted), nil
}

// decodeSessionData decodes session data from a base64 string
func decodeSessionData(encodedData string, key []byte) (map[string]interface{}, error) {
	decoded, err := base64.URLEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, err
	}

	decrypted, err := decrypt(decoded, key)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, err
	}

	return data, nil
}
