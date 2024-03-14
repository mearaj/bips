package util

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/mearaj/bips/bip32"
	"golang.org/x/crypto/sha3"
)

type KeyPath struct {
	Path bip32.Path
	Key  bip32.Key
}

func (b KeyPath) AddrHex() string {
	pbs, err := hex.DecodeString(b.Key.PublicKeyHex())
	if err != nil {
		return ""
	}
	x, y := bip32.ExpandPublicKey(pbs)
	addrBs := append(x.Bytes(), y.Bytes()...)
	kckHash := sha3.NewLegacyKeccak256()
	kckHash.Write(addrBs)
	return hex.EncodeToString(kckHash.Sum(nil)[12:])
}

func (b KeyPath) AddrP2SH(verPrefix byte) string {
	pbs, err := hex.DecodeString(b.Key.PublicKeyHex())
	if err != nil {
		return ""
	}
	pubKeyHash, err := bip32.HashRipeMD160onSha256(pbs)
	if err != nil {
		return ""
	}
	verPubKeyHash := append([]byte{verPrefix}, pubKeyHash...)
	chkSum, err := bip32.ChecksumDblSha256(verPubKeyHash)
	if err != nil {
		return ""
	}
	addBs := append(verPubKeyHash, chkSum...)
	return base58.Encode(addBs)
}

//func (b KeyPath) PvtKeyInWIF(compress bool, netPrefix byte) string {
//	hexToConvert := fmt.Sprintf("%x%s", netPrefix, b.Key.PrivateKeyHex())
//	if compress {
//		hexToConvert += "01"
//	}
//	pvtKeyBs, err := hex.DecodeString(hexToConvert)
//	if err != nil {
//		return ""
//	}
//	pvtKeyDblChkSum, err := bip32.ChecksumDblSha256(pvtKeyBs)
//	if err != nil {
//		return ""
//	}
//	pvtKeyDblChkSumHex := hex.EncodeToString(pvtKeyDblChkSum)
//	pvtKeyHex := hexToConvert + pvtKeyDblChkSumHex
//	pvtKeyHexBs, err := hex.DecodeString(pvtKeyHex)
//	if err != nil {
//		return ""
//	}
//	pvtKey58 := base58.Encode(pvtKeyHexBs)
//	return pvtKey58
//}

type KeyPathRange struct {
	StartIndex uint32
	EndIndex   uint32
	KeyPath
	KeyPaths []KeyPath
}

func (b *KeyPathRange) GenerateRange() error {
	startIndex := b.StartIndex
	endIndex := b.EndIndex

	if endIndex-startIndex < 1 {
		return ErrInvalidRangeProvided
	}
	if !b.Key.IsValid() {
		return ErrInvalidRootKey
	}
	if !b.Path.IsValid() {
		return ErrUnSupportedOrInvalidPath
	}

	currentKey := b.Key
	derivedPath := b.Path.String()
	keyPaths := make([]KeyPath, 0)
	for i := startIndex; i < endIndex; i++ {
		derivedPath = fmt.Sprintf("%s/%d", derivedPath, i%bip32.FirstHardenedChild)
		if i >= bip32.FirstHardenedChild {
			derivedPath += "'"
		}
		currentKey, _ = currentKey.NewChildKey(i)
		keyPaths = append(keyPaths, KeyPath{
			Path: Path(derivedPath).Formatted(),
			Key:  currentKey,
		})
	}
	b.KeyPaths = keyPaths
	return nil
}
