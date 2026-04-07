package cutil

import (
	"os"
	"strconv"
)

// GetEnv 封装os.Getenv(),可以指定默认值
func GetEnv(key, defaultV string) string {
	v := os.Getenv(key)
	if v == "" {
		v = defaultV
	}

	return v
}

func GetEnvMustInt(key string, defaultV int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultV
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}

	return i
}
