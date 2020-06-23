package misc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

var (
	EncryperKey []byte
	block       cipher.Block
)

func init() {

	fmt.Println("Encryper loading...")

	EncryperKey, _ = hex.DecodeString(Config.HashSecret)

	var err error
	block, err = aes.NewCipher(EncryperKey)
	if err != nil {
		panic(err)
	}
}

func PKCS7Padding(ciphertext []byte) []byte {
	padding := aes.BlockSize - len(ciphertext)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

func MyEncryper(payload string) string {

	payloadByte := []byte(payload)
	plaintext := PKCS7Padding(payloadByte)

	// 多申请 aes.BlockSize 长度用于存储 iv
	cipherString := make([]byte, aes.BlockSize+len(plaintext))

	iv := cipherString[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	fmt.Printf("iv:%x\n", iv)

	mode := cipher.NewCBCEncrypter(block, iv)
	// 注意偏移 aes.BlockSize
	mode.CryptBlocks(cipherString[aes.BlockSize:], plaintext)
	return hex.EncodeToString(cipherString)
}

func MyDeryper(cipherString string) string {
	cipherByte, _ := hex.DecodeString(cipherString)
	decipherString := make([]byte, len(cipherString))
	iv := cipherByte[0:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decipherString, cipherByte[aes.BlockSize:])

	decipherString = PKCS7UnPadding(decipherString)
	return string(decipherString)
}

func GetHash(payload string) string {
	hashSeed := sha256.New()
	hashSeed.Write([]byte(payload))
	return fmt.Sprintf("%x", hashSeed.Sum(nil))
}

func GetPhoneMask(payload string) string {
	bb := []byte(payload)
	bb[4], bb[5], bb[6] = 'X', 'X', 'X'
	return string(bb)
}
