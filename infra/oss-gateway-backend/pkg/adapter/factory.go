package adapter

import "fmt"

func NewAdapter(config StorageConfig) (OSSAdapter, error) {
	switch config.VendorType {
	case VendorOSS, VendorOBS, VendorECEPH, VendorTOS:
		return NewMinIOAdapter(config)
	default:
		return nil, fmt.Errorf("unsupported vendor type: %s", config.VendorType)
	}
}
