// Package crypto 提供加密解密功能
// 兼容 CryptoJS 的 AES-256-CBC 加密算法和 EVP_BytesToKey 密钥派生
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"
)

const (
	pkcs5SaltLen = 8
	aes256KeyLen = 32
	aesBlockLen  = 16
)

// Decrypt 解密 Cookie 数据
// 用 UUID 和密码生成密钥，然后解密数据
func Decrypt(uuid, encrypted, password string) []byte {
	key := md5String(uuid + "-" + password)[:16]

	decrypted, err := decryptCryptoJsAesMsg(key, encrypted)
	if err != nil {
		return []byte("{}")
	}
	return decrypted
}

// decryptCryptoJsAesMsg 解密 CryptoJS 加密的消息
//
// CryptoJS.AES.encrypt() 输出格式：
// "Salted__" + [8字节盐值] + [PKCS7填充的密文]
//
// 使用 OpenSSL 兼容的 EVP_BytesToKey 从密码和盐值派生密钥和 IV
// 哈希算法：MD5，密钥长度：32字节，IV长度：16字节
func decryptCryptoJsAesMsg(password, ciphertext string) ([]byte, error) {
	rawEncrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	if len(rawEncrypted) < 17 || len(rawEncrypted)%aesBlockLen != 0 || string(rawEncrypted[:8]) != "Salted__" {
		return nil, fmt.Errorf("invalid ciphertext format")
	}

	salt := rawEncrypted[8:16]
	encrypted := rawEncrypted[16:]

	key, iv := bytesToKey(salt, []byte(password), md5.New(), aes256KeyLen, aesBlockLen)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher creation failed: %v", err)
	}

	cbc := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	cbc.CryptBlocks(decrypted, encrypted)

	decrypted, err = pkcs7strip(decrypted, aesBlockLen)
	if err != nil {
		return nil, fmt.Errorf("pkcs7 strip failed (password may be incorrect): %v", err)
	}

	return decrypted, nil
}

// bytesToKey 实现 OpenSSL EVP_BytesToKey 密钥派生算法
func bytesToKey(salt, data []byte, h hash.Hash, keyLen, blockLen int) (key, iv []byte) {
	if len(salt) > 0 && len(salt) != pkcs5SaltLen {
		panic(fmt.Sprintf("salt length %d, expected %d", len(salt), pkcs5SaltLen))
	}

	var (
		result   []byte
		lastHash []byte
		totalLen = keyLen + blockLen
	)

	for len(result) < totalLen {
		h.Reset()
		h.Write(append(lastHash, append(data, salt...)...))
		lastHash = h.Sum(nil)
		result = append(result, lastHash...)
	}

	return result[:keyLen], result[keyLen:totalLen]
}

// md5String 计算字符串的 MD5 哈希值（十六进制格式）
func md5String(inputs ...string) string {
	h := md5.New()
	for _, s := range inputs {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// pkcs7strip 移除 PKCS7 填充
func pkcs7strip(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: data is not block-aligned")
	}

	padLen := int(data[length-1])
	if padLen > blockSize || padLen == 0 {
		return nil, errors.New("pkcs7: invalid padding length")
	}

	expected := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if !strings.HasSuffix(string(data), string(expected)) {
		return nil, errors.New("pkcs7: invalid padding")
	}

	return data[:length-padLen], nil
}
