package util

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/mearaj/bips/bip32"
	"github.com/mearaj/bips/bip39"
	"golang.org/x/crypto/sha3"
)

const (
	Words12 = 12
	Words15 = 15
	Words18 = 18
	Words21 = 21
	Words24 = 24
)

// GenerateMnemonic
// https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki#user-content-Wordlists

func GenerateMnemonic(wordsCount byte) (string, error) {
	var bitSize int
	switch wordsCount {
	case Words12:
		bitSize = 128
	case Words15:
		bitSize = 160
	case Words18:
		bitSize = 192
	case Words21:
		bitSize = 224
	case Words24:
		bitSize = 256
	default:
		return "", ErrInvalidMnemonicWordsCount
	}
	bs, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(bs)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

func DeriveSeedFromMnemonic(mnemonic string, mnemonicPassphrase string) (string, error) {
	_, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bip39.NewSeed(mnemonic, mnemonicPassphrase)), nil
}

// RootKeyFromSeed seed usually derived from bip39.NewSeed
func RootKeyFromSeed(seed string) (*bip32.Key, error) {
	var rootKey = &bip32.Key{}
	seedBs, err := hex.DecodeString(seed)
	if err != nil {
		return rootKey, err
	}
	if len(seedBs) < 16 {
		return rootKey, ErrInvalidBIP39Seed
	}
	rootKey, err = bip32.NewMasterKey(seedBs)
	ser, err := rootKey.Serialize()
	if err != nil {
		return rootKey, err
	}
	_, err = bip32.Deserialize(ser)
	if err != nil {
		return rootKey, err
	}
	return rootKey, nil
}

type Path = bip32.Path

type KeyPath struct {
	Path Path
	bip32.Key
}

type KeyPathRange struct {
	StartIndex uint32
	EndIndex   uint32
	KeyPath
	KeyPaths []KeyPath
}

type Generator struct {
	rootKey bip32.Key
}

func (g *Generator) SetRootKey(k bip32.Key) {
	g.rootKey = k
}

func (g *Generator) RootKey() *bip32.Key {
	return &g.rootKey
}

func (g *Generator) DeriveBIP32Result(p bip32.Path) ([]KeyPath, error) {
	rootKey := g.RootKey()
	if !rootKey.IsValid() ||
		!rootKey.IsPrivate() ||
		rootKey[bip32.DepthStartIndex] != 0 {
		return nil, ErrInvalidRootKey
	}
	if !p.IsValid() {
		return nil, ErrUnSupportedOrInvalidPath
	}
	pathItems, err := p.ValuesAtDepth()
	if err != nil {
		return nil, err
	}
	derivedPath := "m"
	keyPath := KeyPath{
		Path: Path(derivedPath),
		Key:  *rootKey,
	}
	keyPaths := make([]KeyPath, 1)
	keyPaths[0] = keyPath
	currentKey := *rootKey
	for _, val := range pathItems {
		derivedPath = fmt.Sprintf("%s/%d", derivedPath, val%bip32.FirstHardenedChild)
		if val >= bip32.FirstHardenedChild {
			derivedPath += "'"
		}
		currentKey, _ = currentKey.NewChildKey(val)
		keyPaths = append(keyPaths, KeyPath{
			Path: Path(derivedPath).Formatted(),
			Key:  currentKey,
		})
	}
	return keyPaths, nil
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

func (b KeyPath) AddrP2SH(verPrefix byte) string {
	pbs, err := hex.DecodeString(b.Key.PublicKeyHex())
	if err != nil {
		return ""
	}
	pubKeyHash, err := bip32.HashDblRipeMD160onSha256(pbs)
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

func (b KeyPath) AddrHex() string {
	pbs, err := hex.DecodeString(b.Key.PublicKeyHex())
	if err != nil {
		return ""
	}
	x, y := bip32.ExpandPublicKey(pbs)
	addrBs := append(x.Bytes(), y.Bytes()...)
	kckHash := sha3.NewLegacyKeccak256()
	return hex.EncodeToString(kckHash.Sum((addrBs)[12:]))
}

func (b KeyPath) PvtKeyInWIF(compress bool, netPrefix byte) string {
	hexToConvert := fmt.Sprintf("%x%s", netPrefix, b.Key.PrivateKeyHex())
	if compress {
		hexToConvert += "01"
	}
	pvtKeyBs, err := hex.DecodeString(hexToConvert)
	if err != nil {
		return ""
	}
	pvtKeyDblChkSum, err := bip32.ChecksumDblSha256(pvtKeyBs)
	if err != nil {
		return ""
	}
	pvtKeyDblChkSumHex := hex.EncodeToString(pvtKeyDblChkSum)
	pvtKeyHex := hexToConvert + pvtKeyDblChkSumHex
	pvtKeyHexBs, err := hex.DecodeString(pvtKeyHex)
	if err != nil {
		return ""
	}
	pvtKey58 := base58.Encode(pvtKeyHexBs)
	return pvtKey58
}

//AddressesKey difference between startIndex and endIndex range should
//be less than or equal to 100
//func (g *Generator) AddressesKey(startIndex, endIndex uint32) ([]DerivedAddress, error) {
//	err := g.Validate()
//	if err != nil {
//		return nil, err
//	}
//	diff := endIndex - startIndex
//	if diff < 0 {
//		return nil, ErrInvalidRangeProvided
//	}
//	if diff > 100 {
//		return nil, ErrInvalidRangeProvided
//	}
//	pathKeys := make([]DerivedAddress, 0)
//	for i := startIndex; i <= endIndex; i++ {
//		addressKey, err := g.changeKey.NewChildKey(i)
//		if err != nil {
//			return nil, err
//		}
//		pvtKeyHex := hex.EncodeToString(addressKey.Key)
//		if err != nil {
//			return nil, err
//		}
//		pubKey := addressKey.PublicKeyExtended()
//		pubKeyHex := hex.EncodeToString(pubKey.Key)
//		deCompPubKey, err := crypto.DecompressPubkey(pubKey.Key)
//		if err != nil {
//			return nil, err
//		}
//		addKeyHex := hex.EncodeToString(crypto.Keccak256(append(deCompPubKey.X.Bytes(), deCompPubKey.Y.Bytes()...))[12:])
//		pubKeySha256 := sha256.Sum256(pubKey.Key)
//		ripEmd160 := ripemd160.New()
//		_, err = ripEmd160.Write(pubKeySha256[:])
//		if err != nil {
//			return nil, err
//		}
//		pubKeyHash := ripEmd160.Sum(nil)
//		verPubKeyHash := append([]byte{0x00}, pubKeyHash...)
//		firstPub256 := sha256.Sum256(verPubKeyHash)
//		secondPub256 := sha256.Sum256(firstPub256[:])
//		chksum := secondPub256[:4]
//		addHex := append(verPubKeyHash, chksum...)
//		addKey58 := base58.Encode(addHex)
//		pathKey := DerivedAddress{
//			BIP32Path:       g.bIP32Path,
//			PubKeyHex:       pubKeyHex,
//			PvtKeyHex:       pvtKeyHex,
//			AddrHex:         addKeyHex,
//			AddressKeyEnc58: addKey58,
//		}
//		pathKeys = append(pathKeys, pathKey)
//	}
//	return pathKeys, nil
//}
