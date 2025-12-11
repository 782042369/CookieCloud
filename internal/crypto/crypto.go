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
	"io"
	"strings"
)

const (
	pkcs5SaltLen = 8
	aes256KeyLen = 32
)

// Decrypt 解密Cookie数据
// 使用UUID和密码生成密钥，解密加密的数据
func Decrypt(uuid, encrypted, password string) []byte {
	// 生成密钥：使用MD5哈希(UUID + "-" + 密码)的前16个字符
	theKey := md5String(uuid+"-"+password)[:16]
	
	// 解密数据
	decrypted, err := decryptCryptoJsAesMsg(theKey, encrypted)
	if err != nil {
		// 如果解密失败，返回空的JSON对象
		return []byte("{}")
	}
	return decrypted
}

// decryptCryptoJsAesMsg 解密使用CryptoJS.AES.encrypt(msg, password)加密的消息
// ciphertext是CryptoJS.AES.encrypt()的结果，它是以下内容的Base64字符串：
// "Salted__" + [8字节随机盐值] + [实际密文]
// 实际密文使用Pkcs7填充（使其长度与块长度对齐）
// CryptoJS使用OpenSSL兼容的EVP_BytesToKey从(password,salt)派生出(key,iv)
// 使用md5作为哈希类型，32/16作为key/block的长度
func decryptCryptoJsAesMsg(password string, ciphertext string) ([]byte, error) {
	const keylen = 32
	const blocklen = 16
	
	// Base64解码密文
	rawEncrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode Encrypted: %v", err)
	}
	
	// 验证密文格式
	if len(rawEncrypted) < 17 || len(rawEncrypted)%blocklen != 0 || string(rawEncrypted[:8]) != "Salted__" {
		return nil, fmt.Errorf("invalid ciphertext")
	}
	
	// 提取盐值和密文
	salt := rawEncrypted[8:16]
	encrypted := rawEncrypted[16:]
	
	// 从密码和盐值派生密钥和初始化向量
	key, iv := bytesToKey(salt, []byte(password), md5.New(), keylen, blocklen)
	
	// 创建AES解密器
	newCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %v", err)
	}
	
	// 使用CBC模式解密
	cfbdec := cipher.NewCBCDecrypter(newCipher, iv)
	decrypted := make([]byte, len(encrypted))
	cfbdec.CryptBlocks(decrypted, encrypted)
	
	// 移除PKCS7填充
	decrypted, err = pkcs7strip(decrypted, blocklen)
	if err != nil {
		return nil, fmt.Errorf("failed to strip pkcs7 paddings (password may be incorrect): %v", err)
	}
	
	return decrypted, nil
}

// bytesToKey 实现OpenSSL EVP_BytesToKey逻辑
// 它接受盐值、数据、哈希类型以及该类型使用的密钥/块长度
// 因此它与C语言中的openssl方法有很大不同
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
		// 连接lastHash、data和salt并将它们写入哈希
		h.Write(append(lastHash, append(data, salt...)...))
		// 传递nil给Sum()将返回当前哈希值
		lastHash = h.Sum(nil)
		// 将lastHash附加到运行总计字节中
		concat = append(concat, lastHash...)
	}
	
	return concat[:keyLen], concat[keyLen:totalLen]
}

// md5String 返回输入字符串的MD5十六进制哈希字符串（小写）
func md5String(inputs ...string) string {
	keyHash := md5.New()
	for _, str := range inputs {
		io.WriteString(keyHash, str)
	}
	return hex.EncodeToString(keyHash.Sum(nil))
}

// pkcs7strip 移除pkcs7填充
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