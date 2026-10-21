package hashutil

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/blake2b"
)

const (
	// 密钥最小长度安全策略（单位：字节）
	MinMACKeySize = 32
)

// Sum256 计算数据的 BLAKE2b-256 哈希值
func Sum256(data []byte) []byte {
	sum := blake2b.Sum256(data)
	return sum[:]
}

// Sum256Hex 计算哈希并返回十六进制字符串
func Sum256Hex(data []byte) string {
	return hex.EncodeToString(Sum256(data))
}

// Sum256Reader 计算数据流的 BLAKE2b-256 哈希值
func Sum256Reader(r io.Reader) ([]byte, error) {
	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// Sum256ReaderHex 计算数据流哈希并返回十六进制字符串
func Sum256ReaderHex(r io.Reader) (string, error) {
	sum, err := Sum256Reader(r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

// HashToken 生成 Token 的哈希摘要
func HashToken(token string) string {
	return Sum256Hex([]byte(token))
}

// VerifyToken 使用恒定时间比较安全地验证 Token
func VerifyToken(token, storedHashHex string) bool {
	stored, err := hex.DecodeString(storedHashHex)
	if err != nil {
		return false
	}
	actual := Sum256([]byte(token))
	return hmac.Equal(actual, stored)
}

// MAC256 使用带有密钥的 BLAKE2b 生成 MAC
func MAC256(data, key []byte) ([]byte, error) {
	if len(key) < MinMACKeySize {
		return nil, fmt.Errorf("MAC 密钥长度不足，要求至少 %d 字节，实际为 %d 字节", MinMACKeySize, len(key))
	}
	h, err := blake2b.New256(key)
	if err != nil {
		return nil, err
	}
	h.Write(data)
	return h.Sum(nil), nil
}

// MAC256Hex 生成 MAC 并返回十六进制字符串
func MAC256Hex(data, key []byte) (string, error) {
	sum, err := MAC256(data, key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

// MAC256Base64 生成 MAC 并返回 URL 安全的 Base64 字符串
func MAC256Base64(data, key []byte) (string, error) {
	sum, err := MAC256(data, key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sum), nil
}

// MAC256Reader 为数据流计算 MAC
func MAC256Reader(r io.Reader, key []byte) ([]byte, error) {
	if len(key) < MinMACKeySize {
		return nil, fmt.Errorf("MAC 密钥长度不足，要求至少 %d 字节，实际为 %d 字节", MinMACKeySize, len(key))
	}
	h, err := blake2b.New256(key)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// MAC256ReaderHex 为数据流计算 MAC 并返回十六进制字符串
func MAC256ReaderHex(r io.Reader, key []byte) (string, error) {
	sum, err := MAC256Reader(r, key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

// MAC256ReaderBase64 为数据流计算 MAC 并返回 Base64 字符串
func MAC256ReaderBase64(r io.Reader, key []byte) (string, error) {
	sum, err := MAC256Reader(r, key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sum), nil
}

// VerifyMAC256Hex 验证十六进制格式的 MAC
func VerifyMAC256Hex(data, key []byte, expectedHex string) bool {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}
	actual, err := MAC256(data, key)
	if err != nil {
		return false
	}
	// 使用 hmac.Equal 防止时序攻击
	return hmac.Equal(actual, expected)
}

// VerifyMAC256Base64 验证 Base64 格式的 MAC
func VerifyMAC256Base64(data, key []byte, expectedBase64 string) bool {
	expected, err := base64.RawURLEncoding.DecodeString(expectedBase64)
	if err != nil {
		return false
	}
	actual, err := MAC256(data, key)
	if err != nil {
		return false
	}
	// 使用 hmac.Equal 防止时序攻击
	return hmac.Equal(actual, expected)
}
