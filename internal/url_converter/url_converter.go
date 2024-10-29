package url_converter

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var shuffledChars string

func InitBase62Array(shuffleKey string) {
	shuffledChars = shuffleBase62Chars(shuffleKey)
}

func shuffleBase62Chars(key string) string {
	hash := sha256.Sum256([]byte(key))
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	randSource := rand.NewSource(seed)
	randGen := rand.New(randSource)

	chars := []byte(base62Chars)
	randGen.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	return string(chars)
}

func base62Encode(num int64) string {
	if num == 0 {
		return string(shuffledChars[0])
	}

	base := int64(62)
	var result []byte

	for num > 0 {
		remainder := num % base
		num /= base
		result = append([]byte{shuffledChars[remainder]}, result...)
	}
	return string(result)
}

func base62Decode(str string) int64 {
	base := int64(62)
	var result int64
	for _, c := range []byte(str) {
		index := int64(bytes.IndexByte([]byte(shuffledChars), c))
		result = result*base + index
	}
	return result
}

func EncodeID(id int64, xorSecretKey int64) string {
	obfuscatedID := id ^ xorSecretKey
	shortCode := base62Encode(obfuscatedID)
	return shortCode
}

func DecodeShortCode(shortCode string, xorSecretKey int64) int64 {
	obfuscatedID := base62Decode(shortCode)
	originalID := obfuscatedID ^ xorSecretKey
	return originalID
}
