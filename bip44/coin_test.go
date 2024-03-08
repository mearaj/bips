package bip44

import (
	"github.com/tyler-smith/assert"
	"testing"
)

func TestRegBip44CoinsIsValid(t *testing.T) {
	for _, eachCoin := range RegCoins {
		t.Log(eachCoin.Type, eachCoin.PathComponent, eachCoin.Type+0x80000000)
		assert.True(t, eachCoin.IsValid())
	}
}
