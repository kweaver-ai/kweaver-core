package mod

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommAsync(t *testing.T) {
	opt := CommandOption{}
	CommSync()(&opt)
	assert.True(t, opt.isSync)
}

func TestCommTimeout(t *testing.T) {
	opt := CommandOption{}
	CommSyncTimeout(time.Second)(&opt)
	assert.Equal(t, time.Second, opt.syncTimeout)
	CommSyncTimeout(0)(&opt)
	assert.Equal(t, time.Second, opt.syncTimeout)
}

func TestCommInterval(t *testing.T) {
	opt := CommandOption{}
	CommSyncInterval(time.Second)(&opt)
	assert.Equal(t, time.Second, opt.syncInterval)
	CommSyncInterval(0)(&opt)
	assert.Equal(t, time.Second, opt.syncInterval)
}
