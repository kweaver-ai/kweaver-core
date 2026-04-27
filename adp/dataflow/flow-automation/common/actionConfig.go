package common

type ActionConfig struct {
	FullText         ActionFullTextConfig         `yaml:"fulltext"`
	DocFormatConvert ActionDocFormatConvertConfig `yaml:"docformatconvert"`
	FileParse        ActionFileParseConfig        `yaml:"fileparse"`
	OCR              ActionOCRConfig              `yaml:"ocr"`
	AudioTransfer    ActionAudioTransferConfig    `yaml:"audiotransfer"`
}

type ActionFullTextConfig struct {
	ExpireSec int64 `yaml:"expireSec" default:"86400"`
}

type ActionDocFormatConvertConfig struct {
	ExpireSec int64 `yaml:"expireSec" default:"86400"`
}

type ActionFileParseConfig struct {
	ExpireSec int64 `yaml:"expireSec" default:"86400"`
}

type ActionOCRConfig struct {
	ExpireSec int64 `yaml:"expireSec" default:"86400"`
}

type ActionAudioTransferConfig struct {
	ExpireSec int64 `yaml:"expireSec" default:"86400"`
}
