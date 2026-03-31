package cutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

func Rsa2048Decrypt(encrypted string) (decrypted string, err error) {
	privateKey := `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA4E+eiWRwffhRIPQYvlXUjf0b3HqCmosiCxbFCYI/gdfDBhrT
Uzbt3fL3o/gRQQBEPf69vhJMFH2ZMtaJM6ohE3yQef331liPVM0YvqMOgvoID+zD
a1NIZFObSsjOKhvZtv9esO0REeiVEPKNc+Dp6il3x7TV9VKGEv0+iriNjqv7TGAe
xo2jVtLm50iVKTju2qmCDG83SnVHzsiNj70MiviqiLpgz72IxjF+xN4bRw8I5dD0
GwwO8kDoJUGWgTds+VckCwdtZA65oui9Osk5t1a4pg6Xu9+HFcEuqwJTDxATvGAz
1/YW0oUisjM0ObKTRDVSfnTYeaBsN6L+M+8gCwIDAQABAoIBAG2Tl3/ImAeBmag+
diPs6+PdBJJFKq3yT9QY8HI/tWRpkXTW/+sDx1mISp9IHK2jQrMCUZCbgZz06jTi
hq29a2EIlc9yWHLWWlZzxqXCI+Gp4Oxenew9B/0ytobm54e8iTOTNp+5f4A/HSrl
QmKcOcjRLxlY5rhr8uEt4zKDC2vo/69NuQFdAAjrpdk/SKFNwTc+OarkvxW1lpSy
B8InzhKOFL7b+uqZ3HUnSAIUlxkd8rx2Qt/6wK+AQAYESff6lkjNs8ZcvLXMlPir
lU4gwYFsEGxi84+gqYJ1e4HFX9ohuYa9EoUx9jTV5p9o/GwYOiXk91NKCKctNjcm
qlUmWTECgYEA9Sgbo+DCs+SDflBARXr+ZA5eAImcRt+0CtWlU8XBpumrDh99gjXD
h/BdlNfMccG/TTXIpoEl7Mts2wB7hJcbH/G78VLLdnW8GCKQ5ADNFc1q42Pi88Po
6ac2HxnqLlmu2N8AjfVowBb7+YotN48Ku41mNGs2JEH4fMyQeBb2NZMCgYEA6jt6
NFR8GugjMvnXUbcFf7cZ8iMlQLIbpkP8UgIWU5h9dUkw5JgRN0YMnSi9gA31gzsW
V4O8f+XPsnriAk29SY5kI6mGaLq0Ywpk3dvQBRhDUDJ+rItApKbF3ZJ6hjgAgIQx
09A0cPY+T+twXMn2uUV6LfYxEiG3Igokiyb8dqkCgYEAzaLC7IdPSg3XrlAqWR19
3PegKdtD1r82ChCDCO3MLfG6pbIMWPg39wLLvFn3B0R47o66q8+QvDs2J80Tznfh
LL5b42SLfeXrzGLSHi392NfhXLMgX1BpQfQcFaJrKE3Zt9f2Yx0CrH2bBgm9O+kk
G4XTwQxc8bTUdfoxBEpeYzkCgYApIyUFR8k8GIUGEOcGDPTER24hHpcOU7mTa+FG
reMp72ApVx9lJmfvozfX6i3N7aWu1JPJ7vMOK1hc6kQDT4/s+TsRIFbg0dmYg1zP
silIm8hGr3eb6iECSd/6WB14sSE1cQInRyvOoxCyjJEBWt8gDtm0dMaNfqphKhLc
9Y3lcQKBgQCAD+KTHioiJx3uHGC0oK32SYXp9iWYKLZVhozpX/e+fy4AWSDLCI9x
AYTqOuwl0BHDgutwpA2AQy7BsxAC3CPu6F630lv4O644W1aIbaqoS4ty7EeH+tYB
c1LRBByqBLGnV+QtydYIgiwAM7vng9NxSxvUpQ9I/lr8Myu/GeS4dg==
-----END RSA PRIVATE KEY-----
	`

	// pem 解码
	block, _ := pem.Decode([]byte(privateKey))
	// X509解码
	var priv *rsa.PrivateKey

	priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}

	// base64 解密 go base64包可以处理\r\n 无需额外处理
	var tempData []byte

	tempData, err = base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return
	}

	// rsa解码
	temp, err := rsa.DecryptPKCS1v15(rand.Reader, priv, tempData)
	if err != nil {
		return
	}

	return string(temp), nil
}
