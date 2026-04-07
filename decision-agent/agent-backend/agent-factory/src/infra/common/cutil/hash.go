package cutil

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func Hash256(data []byte) (hash string) {
	arr := sha256.Sum256(data)
	hash = hex.EncodeToString(arr[:])

	return
}

func MD5(data []byte) (hash string) {
	arr := md5.Sum(data)
	hash = hex.EncodeToString(arr[:])

	return
}
