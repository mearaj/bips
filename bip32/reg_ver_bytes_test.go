package bip32

import (
	"github.com/tyler-smith/assert"
	"testing"
)

func TestRegisteredVersionedBytes(t *testing.T) {
	for _, hdBytes := range HDVersionBytesSlice {
		assert.True(t, hdBytes.IsRegistered())
	}
}
