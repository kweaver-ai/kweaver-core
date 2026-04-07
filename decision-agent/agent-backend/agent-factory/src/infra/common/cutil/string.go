package cutil

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func MustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func MustParseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return i
}

func StringToBool(s string) bool {
	s = strings.TrimSpace(s)
	return !(s == "" || strings.EqualFold(s, "false"))
}

func RuneLength(s string) int {
	return len([]rune(s))
}

func GenerateRandomString(length int) string {
	if length <= 0 || length > 100 {
		panic("length must be between 1 and 100")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}

	return string(b)
}

// 字符串根据:分割，然后换行，组装成一个新的字符串
func StringSplitAndJoin(s string) string {
	return strings.Join(strings.Split(s, ":"), "\n")
}
