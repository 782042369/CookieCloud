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

// Decrypt 解密Cookie数据
// 用UUID和密码生成密钥，然后解密数据
func Decrypt(uuid, encrypted, password string) []byte {
	// 生成密钥：用MD5哈希(UUID + "-" + 密码)的前16个字符
	theKey := md5String(uuid + "-" + password)[:16]

	// 解密数据
	decrypted, err := decryptCryptoJsAesMsg(theKey, encrypted)
	if err != nil {
		// 解密失败就返回空的JSON对象
		return []byte("{}")
	}
	return decrypted
}

// decryptCryptoJsAesMsg 解密CryptoJS加密的消息
// CryptoJS.AES.encrypt()出来的东西是这种格式：
// "Salted__" + [8字节随机盐值] + [实际密文]
// 实际密文用Pkcs7填充对齐块长度
// CryptoJS用OpenSSL兼容的EVP_BytesToKey从密码和盐值派生密钥和IV
// 用MD5做哈希，密钥32字节，IV 16字节
func decryptCryptoJsAesMsg(password string, ciphertext string) ([]byte, error) {
	// Base64解码密文
	rawEncrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode Encrypted: %v", err)
	}

	// 验证密文格式对不对
	if len(rawEncrypted) < 17 || len(rawEncrypted)%aesBlockLen != 0 || string(rawEncrypted[:8]) != "Salted__" {
		return nil, fmt.Errorf("invalid ciphertext")
	}

	// 提取盐值和实际密文
	salt := rawEncrypted[8:16]
	encrypted := rawEncrypted[16:]

	// 从密码和盐值派生密钥和IV
	key, iv := bytesToKey(salt, []byte(password), md5.New(), aes256KeyLen, aesBlockLen)

	// 创建AES解密器
	newCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %v", err)
	}

	// 用CBC模式解密
	cfbdec := cipher.NewCBCDecrypter(newCipher, iv)
	decrypted := make([]byte, len(encrypted))
	cfbdec.CryptBlocks(decrypted, encrypted)

	// 去掉PKCS7填充
	decrypted, err = pkcs7strip(decrypted, aesBlockLen)
	if err != nil {
		return nil, fmt.Errorf("failed to strip pkcs7 paddings (password may be incorrect): %v", err)
	}

	return decrypted, nil
}

// bytesToKey 实现OpenSSL EVP_BytesToKey逻辑
// 接受盐值、数据、哈希类型和密钥/块长度
// 跟C语言的openssl方法一样
func bytesToKey(salt, data []byte, h hash.Hash, keyLen, blockLen int) (key, iv []byte) {
	saltLen := len(salt)
	if saltLen > 0 && saltLen != pkcs5SaltLen {
		panic(fmt.Sprintf("Salt length is %d, expected %d", saltLen, pkcs5SaltLen))
	}

	var (
		concat   []byte
		lastHash []byte
		totalLen = keyLen + blockLen
	)

	for ; len(concat) < totalLen; h.Reset() {
		// 把lastHash、data和salt拼起来写进哈希
		h.Write(append(lastHash, append(data, salt...)...))
		// 传nil给Sum()返回当前哈希值
		lastHash = h.Sum(nil)
		// 把lastHash加到concat后面
		concat = append(concat, lastHash...)
	}

	return concat[:keyLen], concat[keyLen:totalLen]
}

// md5String 返回字符串的MD5哈希值（十六进制，小写）
func md5String(inputs ...string) string {
	h := md5.New()
	for _, s := range inputs {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// pkcs7strip 去掉pkcs7填充
func pkcs7strip(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}
	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !strings.HasSuffix(string(data), string(ref)) {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padLen], nil
}
