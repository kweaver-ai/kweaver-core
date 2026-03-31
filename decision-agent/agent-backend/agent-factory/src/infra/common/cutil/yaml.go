package cutil

import (
	"os"

	"gopkg.in/yaml.v3"
)

func YamlParse(file string, obj interface{}) (err error) {
	ioStream, err := os.ReadFile(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(ioStream, obj)

	return
}

func YamlParseFromStr(str string, obj interface{}) (err error) {
	err = yaml.Unmarshal([]byte(str), obj)
	return
}
