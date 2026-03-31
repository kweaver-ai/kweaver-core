package cutil

import "testing"

func TestHash256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"空数据", []byte{}, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"简单字符串", []byte("hello"), "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"复杂字符串", []byte("hello world"), "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
		{"二进制数据", []byte{0x00, 0x01, 0x02, 0x03}, "054edec1d0211f624fed0cbca9d4f9400b0e491c43742af2c5b0abebf0c990d8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Hash256(tt.data); got != tt.want {
				t.Errorf("Hash256(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestMD5(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"空数据", []byte{}, "d41d8cd98f00b204e9800998ecf8427e"},
		{"简单字符串", []byte("hello"), "5d41402abc4b2a76b9719d911017c592"},
		{"复杂字符串", []byte("hello world"), "5eb63bbbe01eeed093cb22bb8f5acdc3"},
		{"二进制数据", []byte{0x00, 0x01, 0x02}, "b95f67f61ebb03619622d798f45fc2d3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := MD5(tt.data); got != tt.want {
				t.Errorf("MD5(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}
