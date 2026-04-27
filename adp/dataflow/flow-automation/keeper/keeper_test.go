package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckWorkerKey(t *testing.T) {
	tests := []struct {
		giveKey string
		wantRet int
		wantErr error
	}{
		{
			giveKey: "automation-256",
			wantRet: 256,
		},
	}

	for _, tc := range tests {
		ret, err := CheckWorkerKey(tc.giveKey)
		assert.Equal(t, tc.wantErr, err)
		assert.Equal(t, tc.wantRet, ret)
	}
}
