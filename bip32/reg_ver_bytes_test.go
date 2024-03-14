package bip32

import (
	"github.com/tyler-smith/assert"
	"testing"
)

func TestRegisteredVersionedBytes(t *testing.T) {
	for _, hdBytes := range HDVersionBytesSlice {
		assert.True(t, hdBytes.IsRegistered())
		_, err := hdBytes.PurposeVal()
		assert.NoError(t, err)
		_, err = hdBytes.PurposeValFull()
		assert.NoError(t, err)
	}
	for val, hdBytesArr := range PurposeFullToHDBytesSlice {
		for _, hdBytes := range hdBytesArr {
			purposeFullVal, err := hdBytes.PurposeValFull()
			assert.NoError(t, err)
			assert.True(t, val == purposeFullVal)
		}
	}
	for val, hdBytesArr := range PurposeToHDBytesSlice {
		for _, hdBytes := range hdBytesArr {
			purposeVal, err := hdBytes.PurposeVal()
			assert.NoError(t, err)
			assert.True(t, val == purposeVal)
		}
	}
}
