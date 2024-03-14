package bip32

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcutil/base58"
)

type Version [4]byte
type Depth [1]byte
type FingerPrint [4]byte
type ChildNumber [4]byte
type ChainCode [32]byte

// KeyBytes For PrivateKey, index starts from 1 instead of 0
type KeyBytes [33]byte

type PvtKeyBytes [32]byte // If first byte of KeyBytes is zero

// Key is [ Version + Depth + FingerPrint + ChildNumber + ChainCode + KeyBytes ]Array
// m/purpose'/coin_type'/account'/change/address_index
type Key [78]byte

const (
	VersionStartIndex     = 0
	VersionEndIndex       = VersionStartIndex + 4
	DepthStartIndex       = VersionEndIndex
	DepthEndIndex         = DepthStartIndex + 1
	FingerPrintStartIndex = DepthEndIndex
	FingerPrintEndIndex   = FingerPrintStartIndex + 4
	ChildNumberStartIndex = FingerPrintEndIndex
	ChildNumberEndIndex   = ChildNumberStartIndex + 4
	ChainCodeStartIndex   = ChildNumberEndIndex
	ChainCodeEndIndex     = ChainCodeStartIndex + 32
	PubKeyStartIndex      = ChainCodeEndIndex
	PvtKeyStartIndex      = PubKeyStartIndex + 1
	PubKeyEndIndex        = PubKeyStartIndex + 33
	PvtKeyEndIndex        = PubKeyEndIndex
)

const (
	// FirstHardenedChild is the index of the first "hardened" child key as per the
	FirstHardenedChild = uint32(0x80000000)
	// PublicKeyCompressedLength is the byte count of a compressed public key
	PublicKeyCompressedLength = 33
)

var DefaultMainnetVersion = Bitcoinxprvxpub
var DefaultTestnetVersion = Bitcointprvtpub

var (
	// ErrSerializedKeyWrongSize is returned when trying to deserialize a key that
	// has an incorrect length
	ErrSerializedKeyWrongSize = errors.New("serialized keys should by exactly 82 bytes")

	// ErrHardenedChildPublicKey is returned when trying to create a hardened child
	// of the public key
	ErrHardenedChildPublicKey = errors.New("can't create hardened child for public key")

	// ErrInvalidChecksum is returned when deserializing a key with an incorrect
	// ChecksumDblSha256
	ErrInvalidChecksum = errors.New("ChecksumDblSha256 doesn't match")

	// ErrInvalidPrivateKey is returned when a derived private key is invalid
	ErrInvalidPrivateKey = errors.New("invalid private key")

	// ErrInvalidPublicKey is returned when a derived public key is invalid
	ErrInvalidPublicKey = errors.New("invalid public key")

	ErrEmptyKey = errors.New("key is empty")
)

// NewMasterKey creates a new master extended key from a seed
// VersionBytes.PvtKeyFlagBytes used is DefaultMainnetVersion.PvtKeyFlagBytes
func NewMasterKey(seed []byte) (*Key, error) {
	// Generate key and chaincode
	hm := hmac.New(sha512.New, []byte("Bitcoin seed"))
	_, err := hm.Write(seed)
	if err != nil {
		return nil, err
	}
	intermediary := hm.Sum(nil)
	// Split it into our key and chain code
	keyBytes := PvtKeyBytes(intermediary[:32])
	chainCode := ChainCode(intermediary[32:])
	// Validate key
	err = ValidatePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	var key = &Key{}
	vs := DefaultMainnetVersion.PvtKeyFlagBytes()
	key.SetVersion(vs)
	key.setChainCode(chainCode)
	key.setPvtKeyBytes(keyBytes)
	return key, nil
}

func (key *Key) IsPrivate() bool {
	return key[PubKeyStartIndex] == 0
}

// NewChildKey derives a child key from a given parent as outlined by bip32
func (key *Key) NewChildKey(childIdx uint32) (Key, error) {
	var childKey Key
	// Fail early if trying to create hardened child from public key
	if !key.IsPrivate() && childIdx >= FirstHardenedChild {
		return childKey, ErrHardenedChildPublicKey
	}
	intermediary, err := key.getIntermediary(childIdx)
	if err != nil {
		return childKey, err
	}
	vs := Version(key[VersionStartIndex:VersionEndIndex])
	vsVal := binary.BigEndian.Uint32(vs[:])
	hdBsArr, ok := PvtFlagToHDBytesSlice[vsVal]
	if key.IsPrivate() {
		if !ok || len(hdBsArr) == 0 {
			return childKey, ErrUnSupportedHDVersionBytes
		}
	} else {
		hdBsArr, ok = PubFlagToHDBytesSlice[vsVal]
		if !ok || len(hdBsArr) == 0 {
			return childKey, ErrUnSupportedHDVersionBytes
		}
		vs = hdBsArr[0].PubKeyFlagBytes()
	}

	childKey.setChildNumber(ChildNumber(uint32Bytes(childIdx)))
	childKey.setChainCode(ChainCode(intermediary[32:]))
	childKey.setDepth(key[DepthStartIndex] + 1)
	childKey.SetVersion(vs)
	// Bip32 CKDpriv
	if key.IsPrivate() {
		fingerprint, err := HashRipeMD160onSha256(
			publicKeyForPrivateKey(key[PvtKeyStartIndex:PubKeyEndIndex]),
		)
		if err != nil {
			return childKey, err
		}
		childKey.setFingerPrint(FingerPrint(fingerprint))
		kbs := addPrivateKeys(intermediary[:32], key[PvtKeyStartIndex:])
		childKey.setPvtKeyBytes(PvtKeyBytes(kbs))

		// Validate key
		err = ValidatePrivateKey(childKey.GetPvtKeyBytes())
		if err != nil {
			return childKey, err
		}
		// Bip32 CKDpub
	} else {
		vs = hdBsArr[0].PubKeyFlagBytes()
		keyBytes := publicKeyForPrivateKey(intermediary[:32])
		// Validate key
		err := validateChildPublicKey(keyBytes)
		if err != nil {
			return childKey, err
		}
		fingerprint, err := HashRipeMD160onSha256(key[PubKeyStartIndex:])
		if err != nil {
			return childKey, err
		}
		copy(childKey[FingerPrintStartIndex:FingerPrintEndIndex], fingerprint)
		kbs := addPublicKeys(keyBytes, key[PubKeyStartIndex:])
		childKey.setPubKeyBytes(KeyBytes(kbs))
	}
	return childKey, nil
}

func (key *Key) getIntermediary(childIdx uint32) ([]byte, error) {
	// Get intermediary to create key and chaincode from
	// Hardened children are based on the private key
	// NonHardened children are based on the public key
	childIndexBytes := uint32Bytes(childIdx)
	var data [33]byte
	if childIdx >= FirstHardenedChild {
		copy(data[1:], key[PvtKeyStartIndex:PubKeyEndIndex])
	} else {
		if key.IsPrivate() {
			copy(data[:], publicKeyForPrivateKey(key[PvtKeyStartIndex:PubKeyEndIndex]))
		} else {
			copy(data[:], key[PubKeyStartIndex:])
		}
	}
	dataN := append(data[:], childIndexBytes...)
	hm := hmac.New(sha512.New, key[ChainCodeStartIndex:ChainCodeEndIndex])
	_, err := hm.Write(dataN)
	if err != nil {
		return nil, err
	}
	return hm.Sum(nil), nil
}

// PublicKeyExtended returns the public version of key or return a copy
// The 'Neuter' function from the bip32 spec
// If corresponding version is not found then
// default VersionBytes.PubKeyFlagBytes (DefaultMainnetVersion.PubKeyFlagBytes) is used
func (key *Key) PublicKeyExtended() Key {
	vs := Version(key[VersionStartIndex:VersionEndIndex])
	vsVal := binary.BigEndian.Uint32(vs[:])
	hdBsArr, ok := PvtFlagToHDBytesSlice[vsVal]
	var pubKey Key
	copy(pubKey[:], key[:])
	if key.IsPrivate() {
		if !ok || len(hdBsArr) == 0 {
			vs = DefaultMainnetVersion.PubKeyFlagBytes()
		} else {
			vs = hdBsArr[0].PubKeyFlagBytes()
		}
		copy(pubKey[PubKeyStartIndex:PubKeyEndIndex],
			publicKeyForPrivateKey(key[PvtKeyStartIndex:]),
		)
	} else {
		hdBsArr, ok = PubFlagToHDBytesSlice[vsVal]
		if !ok || len(hdBsArr) == 0 {
			vs = DefaultMainnetVersion.PubKeyFlagBytes()
		}
	}
	pubKey.SetVersion(vs)
	return pubKey
}

// PrivateKeyHex returns private key in hex string without prefix 0x
func (key *Key) PrivateKeyHex() string {
	if key.IsPrivate() {
		return hex.EncodeToString(key[PvtKeyStartIndex:PvtKeyEndIndex])
	}
	return ""
}

// PublicKeyHex returns public key in hex string without prefix 0x
func (key *Key) PublicKeyHex() string {
	if key.IsPrivate() {
		pubKey := key.PublicKeyExtended()
		return hex.EncodeToString(pubKey[PubKeyStartIndex:PubKeyEndIndex])
	}
	return hex.EncodeToString(key[PubKeyStartIndex:PubKeyEndIndex])
}

// Serialize a Key to a slice of 82 byte
// (78 byte(s) + 4 byte(s) of checksum)
func (key *Key) Serialize() ([]byte, error) {
	isEmpty := true
	for _, v := range key {
		if v != 0 {
			isEmpty = false
			break
		}
	}
	if isEmpty {
		return nil, ErrEmptyKey
	}
	// Append the standard doublesha256 ChecksumDblSha256
	serializedKey, err := AddChecksumDblSha256ToBytes(key[:])
	if err != nil {
		return nil, err
	}
	return serializedKey, nil
}

// B58Serialize encodes the Key in the standard Bitcoin base58 encoding
func (key *Key) B58Serialize() string {
	serializedKey, err := key.Serialize()
	if err != nil {
		return ""
	}
	return base58.Encode(serializedKey)
}

// String encodes the Key in the standard Bitcoin base58 encoding
// Set Version before calling String()
func (key *Key) String() string {
	return key.B58Serialize()
}

func (key *Key) IsValid() bool {
	bs, err := key.Serialize()
	if err != nil {
		return false
	}
	_, err = Deserialize(bs)
	if err != nil {
		return false
	}
	return true
}

// Deserialize a byte slice into a Key
func Deserialize(data []byte) (Key, error) {
	if len(data) != 82 {
		if len(data) < 82 {
			exData := make([]byte, 82-len(data))
			data = append(data, exData...)
		}
		return Key(data[:78]), ErrSerializedKeyWrongSize
	}
	// validate ChecksumDblSha256
	cs1, err := ChecksumDblSha256(data[0 : len(data)-4])
	if err != nil {
		return Key(data[:78]), err
	}
	cs2 := data[len(data)-4:]
	for i := range cs1 {
		if cs1[i] != cs2[i] {
			return Key(data[:78]), ErrInvalidChecksum
		}
	}
	return Key(data[:78]), nil
}

func (key *Key) SetVersionUint32(v uint32) {
	vsVal := uint32Bytes(v)
	key.SetVersion(Version(vsVal))
}

func (key *Key) SetVersion(v Version) {
	copy(key[VersionStartIndex:VersionEndIndex], v[:])
}

func (key *Key) GetVersion() uint32 {
	return binary.BigEndian.Uint32(key[VersionStartIndex:VersionEndIndex])
}

func (key *Key) setChainCodeUint32(code uint32) {
	codeVal := uint32Bytes(code)
	key.setChainCode(ChainCode(codeVal))
}

func (key *Key) setChainCode(code ChainCode) {
	copy(key[ChainCodeStartIndex:ChainCodeEndIndex], code[:])
}

func (key *Key) GetChainCode() ChainCode {
	return ChainCode(key[ChainCodeStartIndex:ChainCodeEndIndex])
}

func (key *Key) setFingerPrint(f FingerPrint) {
	copy(key[FingerPrintStartIndex:FingerPrintEndIndex], f[:])
}

func (key *Key) GetFingerPrint() FingerPrint {
	return FingerPrint(key[FingerPrintStartIndex:FingerPrintEndIndex])
}

func (key *Key) setDepth(d byte) {
	key[DepthStartIndex] = d
}

func (key *Key) GetDepth() Depth {
	return Depth(key[DepthStartIndex:DepthEndIndex])
}

func (key *Key) setPvtKeyBytes(b PvtKeyBytes) {
	copy(key[PvtKeyStartIndex:PubKeyEndIndex], b[:])
}

func (key *Key) GetPvtKeyBytes() PvtKeyBytes {
	return PvtKeyBytes(key[PvtKeyStartIndex:PubKeyEndIndex])
}

func (key *Key) setPubKeyBytes(b KeyBytes) {
	copy(key[PubKeyStartIndex:PubKeyEndIndex], b[:])
}

func (key *Key) GetKeyBytes() KeyBytes {
	return KeyBytes(key[PubKeyStartIndex:PubKeyEndIndex])
}

func (key *Key) setChildNumber(c ChildNumber) {
	copy(key[ChildNumberStartIndex:ChildNumberEndIndex], c[:])
}

func (key *Key) GetChildNumber() ChildNumber {
	return ChildNumber(key[ChildNumberStartIndex:ChildNumberEndIndex])
}

// B58Deserialize deserializes a Key encoded in base58 encoding
func B58Deserialize(data string) (Key, error) {
	b := base58.Decode(data)
	return Deserialize(b)
}

// NewSeed returns a cryptographically secure seed
func NewSeed() ([]byte, error) {
	// Well that easy, just make go read 256 random bytes into a slice
	s := make([]byte, 256)
	_, err := rand.Read(s)
	return s, err
}
