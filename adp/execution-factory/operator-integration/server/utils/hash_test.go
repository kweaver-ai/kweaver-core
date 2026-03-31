package utils

import (
	"fmt"
	"testing"
)

type TestHash struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func TestObjectUUIDHash(t *testing.T) {
	a := &TestHash{
		Name: "test",
		Desc: "test",
	}
	hash1, err := ObjectUUIDHash(a)
	if err != nil {
		t.Errorf("ObjectUUIDHash hash1 err: %+v", err)
	}
	hash2, err := ObjectUUIDHash(a)
	if err != nil {
		t.Errorf("ObjectUUIDHash hash2 err: %+v", err)
	}
	fmt.Println(hash1, hash2, hash1 == hash2)
}
