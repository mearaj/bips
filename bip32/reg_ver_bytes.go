package bip32

import (
	"encoding/binary"
	"regexp"
	"strconv"
	"strings"
)

// Ref https://github.com/satoshilabs/slips/blob/master/slip-0132.md

// PubFlagToHDBytesSlice
// VersionBytes.PubKeyFlag To slice of VersionBytes
var PubFlagToHDBytesSlice = map[uint32][]VersionBytes{}

// PvtFlagToHDBytesSlice
// VersionBytes.PvtKeyFlag To slice of VersionBytes
var PvtFlagToHDBytesSlice = map[uint32][]VersionBytes{}

// HDVersionBytesSlice Slice of All Registered VersionBytes
var HDVersionBytesSlice []VersionBytes

// PurposeToHDBytesSlice uint32 value is VersionBytes.PurposeVal
var PurposeToHDBytesSlice = map[uint32][]VersionBytes{}

// PurposeFullToHDBytesSlice uint32 value is VersionBytes.PurposeValFull
var PurposeFullToHDBytesSlice = map[uint32][]VersionBytes{}

// CoinTypeToHDBytesSlice uint32 value is VersionBytes.CoinVal and not VersionBytes.CoinValFull
var CoinTypeToHDBytesSlice = map[uint32][]VersionBytes{}

type AddrEncoding string

const (
	P2PKH        AddrEncoding = "P2PKH"
	P2WPKH       AddrEncoding = "P2WPKH"
	P2WPKHInP2SH AddrEncoding = "P2WPKHInP2SH"
	P2SH         AddrEncoding = "P2SH"
	P2WSH        AddrEncoding = "P2WSH"
	P2WSHInP2SH  AddrEncoding = "P2WSHInP2SH"
	P2PKT        AddrEncoding = "P2PKT"
)

type VersionBytes struct {
	// Coin Name
	Coin string
	// PvtKeyFlag(4 bytes)
	PvtKeyFlag uint32
	// PvKeyPrefix is human-readable, but it's a result of serialization
	// using above Private Key Hex Bytes (the first 4 val of serialized result obtained
	// using above PvtKeyFlag should always match this PvKeyPrefix)
	PvKeyPrefix string
	// PubKeyFlag is similar to PvtKeyFlag but deals with Public Keys
	PubKeyFlag uint32
	// PubKeyPrefix is similar to PvKeyPrefix but deals with Public Keys
	PubKeyPrefix  string
	AddrEncodings []AddrEncoding
	// Path
	Path Path
}

func (h VersionBytes) PvtKeyFlagBytes() [4]byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], h.PvtKeyFlag)
	return b
}

func (h VersionBytes) PubKeyFlagBytes() [4]byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], h.PubKeyFlag)
	return b
}

func (h VersionBytes) IsValid() bool {
	regex := regexp.MustCompile(`^m/\d{0,8}'/\d{0,8}'$`)
	return regex.MatchString(h.Path.String())
}

func (h VersionBytes) IsRegistered() bool {
	_, ok := PathToHDBytesSlice[h.Path.String()]
	return ok
}

// PurposeVal returns purposeVal in (origVal % FirstHardenedChild),
// hence it's always less than FirstHardenedChild
func (h VersionBytes) PurposeVal() (uint32, error) {
	if !h.IsValid() {
		return 0, ErrUnSupportedHDVersionBytes
	}
	pathArr := strings.Split(h.Path.String(), "/")
	purposeStr := strings.ReplaceAll(pathArr[1], "'", "")
	purposeVal, err := strconv.Atoi(purposeStr)
	if err != nil {
		return 0, err
	}
	if uint32(purposeVal) >= FirstHardenedChild {
		return 0, ErrInvalidPurpose
	}
	return uint32(purposeVal), err
}

// CoinVal returns coinVal in (origVal % FirstHardenedChild),
// hence it's always less than FirstHardenedChild
func (h VersionBytes) CoinVal() (uint32, error) {
	if !h.IsValid() {
		return 0, ErrUnSupportedHDVersionBytes
	}
	pathArr := strings.Split(h.Path.String(), "/")
	coinStr := strings.ReplaceAll(pathArr[2], "'", "")
	coinVal, err := strconv.Atoi(coinStr)
	if err != nil {
		return 0, err
	}
	if uint32(coinVal) >= FirstHardenedChild {
		return 0, ErrInvalidCoin
	}
	return uint32(coinVal), err
}

func (h VersionBytes) PurposeValFull() (uint32, error) {
	if !h.IsValid() {
		return 0, ErrUnSupportedHDVersionBytes
	}
	val, err := h.PurposeVal()
	if err != nil {
		return 0, err
	}
	return val + FirstHardenedChild, nil
}
func (h VersionBytes) CoinValFull() (uint32, error) {
	if !h.IsValid() {
		return 0, ErrUnSupportedHDVersionBytes
	}
	val, err := h.CoinVal()
	if err != nil {
		return 0, err
	}
	return val + FirstHardenedChild, nil
}

var (
	Bitcoinxprvxpub = VersionBytes{
		Coin:          "Bitcoin",
		PvtKeyFlag:    0x0488ade4,
		PvKeyPrefix:   "xprv",
		PubKeyFlag:    0x0488b21e,
		PubKeyPrefix:  "xpub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/0'",
	}
	Bitcoinyprvypub = VersionBytes{
		Coin:          "Bitcoin",
		PvtKeyFlag:    0x049d7878,
		PvKeyPrefix:   "yprv",
		PubKeyFlag:    0x049d7cb2,
		PubKeyPrefix:  "ypub",
		AddrEncodings: []AddrEncoding{P2WPKHInP2SH},
		Path:          "m/49'/0'",
	}
	Bitcoinzprvzpub = VersionBytes{
		Coin:          "Bitcoin",
		PvtKeyFlag:    0x04b2430c,
		PvKeyPrefix:   "zprv",
		PubKeyFlag:    0x04b24746,
		PubKeyPrefix:  "zpub",
		AddrEncodings: []AddrEncoding{P2WPKH},
		Path:          "m/84'/0'",
	}
	BitcoinYprvYpub = VersionBytes{
		Coin:          "Bitcoin",
		PvtKeyFlag:    0x0295b005,
		PvKeyPrefix:   "Yprv",
		PubKeyFlag:    0x0295b43f,
		PubKeyPrefix:  "Ypub",
		AddrEncodings: []AddrEncoding{P2WSHInP2SH},
		Path:          "m/84'/0'",
	}
	BitcoinZprvZpub = VersionBytes{
		Coin:          "Bitcoin",
		PvtKeyFlag:    0x02aa7a99,
		PvKeyPrefix:   "Zprv",
		PubKeyFlag:    0x02aa7ed3,
		PubKeyPrefix:  "Zpub",
		AddrEncodings: []AddrEncoding{P2WSH},
		Path:          "m/84'/0'",
	}
	Bitcointprvtpub = VersionBytes{
		Coin:          "Bitcoin Testnet",
		PvtKeyFlag:    0x04358394,
		PvKeyPrefix:   "tprv",
		PubKeyFlag:    0x043587cf,
		PubKeyPrefix:  "tpub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/1'",
	}
	Bitcoinuprvupub = VersionBytes{
		Coin:          "Bitcoin Testnet",
		PvtKeyFlag:    0x044a4e28,
		PvKeyPrefix:   "uprv",
		PubKeyFlag:    0x044a5262,
		PubKeyPrefix:  "upub",
		AddrEncodings: []AddrEncoding{P2WPKHInP2SH},
		Path:          "m/49'/1'",
	}
	Bitcoinvprvvpub = VersionBytes{
		Coin:          "Bitcoin Testnet",
		PvtKeyFlag:    0x045f18bc,
		PvKeyPrefix:   "vprv",
		PubKeyFlag:    0x045f1cf6,
		PubKeyPrefix:  "vpub",
		AddrEncodings: []AddrEncoding{P2WPKH},
		Path:          "m/84'/1'",
	}
	BitcoinUprvUpub = VersionBytes{
		Coin:          "Bitcoin Testnet",
		PvtKeyFlag:    0x024285b5,
		PvKeyPrefix:   "Uprv",
		PubKeyFlag:    0x024289ef,
		PubKeyPrefix:  "Upub",
		AddrEncodings: []AddrEncoding{P2WSHInP2SH},
		Path:          "m/84'/1'",
	}
	BitcoinVprvVpub = VersionBytes{
		Coin:          "Bitcoin Testnet",
		PvtKeyFlag:    0x02575048,
		PvKeyPrefix:   "Vprv",
		PubKeyFlag:    0x02575483,
		PubKeyPrefix:  "Vpub",
		AddrEncodings: []AddrEncoding{P2WSH},
		Path:          "m/84'/1'",
	}
	Groestlcoinxprvxpub = VersionBytes{
		Coin:          "Groestlcoin",
		PvtKeyFlag:    0x0488ade4,
		PvKeyPrefix:   "xprv",
		PubKeyFlag:    0x0488b21e,
		PubKeyPrefix:  "xpub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/17'",
	}
	Groestlcoinyprvypub = VersionBytes{
		Coin:          "Groestlcoin",
		PvtKeyFlag:    0x049d7878,
		PvKeyPrefix:   "yprv",
		PubKeyFlag:    0x049d7cb2,
		PubKeyPrefix:  "ypub",
		AddrEncodings: []AddrEncoding{P2WPKHInP2SH},
		Path:          "m/49'/17'",
	}
	Groestlcoinzprvzpub = VersionBytes{
		Coin:          "Groestlcoin",
		PvtKeyFlag:    0x04b2430c,
		PvKeyPrefix:   "zprv",
		PubKeyFlag:    0x04b24746,
		PubKeyPrefix:  "zpub",
		AddrEncodings: []AddrEncoding{P2WPKH},
		Path:          "m/84'/17'",
	}
	GroestlcoinYprvYpub = VersionBytes{
		Coin:          "Groestlcoin",
		PvtKeyFlag:    0x0295b005,
		PvKeyPrefix:   "Yprv",
		PubKeyFlag:    0x0295b43f,
		PubKeyPrefix:  "Ypub",
		AddrEncodings: []AddrEncoding{P2WSHInP2SH},
		Path:          "m/84'/17'",
	}
	GroestlcoinZprvZpub = VersionBytes{
		Coin:          "Groestlcoin",
		PvtKeyFlag:    0x02aa7a99,
		PvKeyPrefix:   "Zprv",
		PubKeyFlag:    0x02aa7ed3,
		PubKeyPrefix:  "Zpub",
		AddrEncodings: []AddrEncoding{P2WSH},
		Path:          "m/84'/17'",
	}
	Groestlcointprvtpub = VersionBytes{
		Coin:          "Groestlcoin Testnet",
		PvtKeyFlag:    0x04358394,
		PvKeyPrefix:   "tprv",
		PubKeyFlag:    0x043587cf,
		PubKeyPrefix:  "tpub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/1'",
	}
	Groestlcoinuprvupub = VersionBytes{
		Coin:          "Groestlcoin Testnet",
		PvtKeyFlag:    0x044a4e28,
		PvKeyPrefix:   "uprv",
		PubKeyFlag:    0x044a5262,
		PubKeyPrefix:  "upub",
		AddrEncodings: []AddrEncoding{P2WPKHInP2SH},
		Path:          "m/49'/1'",
	}
	Groestlcoinvprvvpub = VersionBytes{
		Coin:          "Groestlcoin Testnet",
		PvtKeyFlag:    0x045f18bc,
		PvKeyPrefix:   "vprv",
		PubKeyFlag:    0x045f1cf6,
		PubKeyPrefix:  "vpub",
		AddrEncodings: []AddrEncoding{P2WPKH},
		Path:          "m/84'/1'",
	}
	GroestlcoinUprvUpub = VersionBytes{
		Coin:          "Groestlcoin Testnet",
		PvtKeyFlag:    0x024285b5,
		PvKeyPrefix:   "Uprv",
		PubKeyFlag:    0x024289ef,
		PubKeyPrefix:  "Upub",
		AddrEncodings: []AddrEncoding{P2WSHInP2SH},
		Path:          "m/84'/1'",
	}
	GroestlcoinVprvVpub = VersionBytes{
		Coin:          "Groestlcoin Testnet",
		PvtKeyFlag:    0x02575048,
		PvKeyPrefix:   "Vprv",
		PubKeyFlag:    0x02575483,
		PubKeyPrefix:  "Vpub",
		AddrEncodings: []AddrEncoding{P2WSH},
		Path:          "m/84'/1'",
	}
	LitecoinLtpvLtub = VersionBytes{
		Coin:          "Litecoin",
		PvtKeyFlag:    0x019d9cfe,
		PvKeyPrefix:   "Ltpv",
		PubKeyFlag:    0x019da462,
		PubKeyPrefix:  "Ltub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/2'",
	}
	LitecoinMtpvMtub = VersionBytes{
		Coin:          "Litecoin",
		PvtKeyFlag:    0x01b26792,
		PvKeyPrefix:   "Mtpv",
		PubKeyFlag:    0x01b26ef6,
		PubKeyPrefix:  "Mtub",
		AddrEncodings: []AddrEncoding{P2WPKHInP2SH},
		Path:          "m/49'/2'",
	}
	Litecointtpvttub = VersionBytes{
		Coin:          "Litecoin Testnet",
		PvtKeyFlag:    0x0436ef7d,
		PvKeyPrefix:   "ttpv",
		PubKeyFlag:    0x0436f6e1,
		PubKeyPrefix:  "ttub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/1'",
	}
	Nexaxprvxpub = VersionBytes{
		Coin:          "Nexa",
		PvtKeyFlag:    0x426c6b73,
		PvKeyPrefix:   "xprv",
		PubKeyFlag:    0x42696720,
		PubKeyPrefix:  "xpub",
		AddrEncodings: []AddrEncoding{P2PKT, P2PKH, P2SH},
		Path:          "m/44'/29223'",
	}
	Nexaxprvxpub2 = VersionBytes{
		Coin:          "Nexa Testnet",
		PvtKeyFlag:    0x04358394,
		PvKeyPrefix:   "xprv",
		PubKeyFlag:    0x043587cf,
		PubKeyPrefix:  "xpub",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/1'",
	}
	Polispprvppub = VersionBytes{
		Coin:          "Polis",
		PvtKeyFlag:    0x03e25945,
		PvKeyPrefix:   "pprv",
		PubKeyFlag:    0x03e25d7e,
		PubKeyPrefix:  "ppub",
		AddrEncodings: []AddrEncoding{P2PKH},
		Path:          "m/44'/1997'",
	}
	Syscoinzprvzpub = VersionBytes{
		Coin:          "Syscoin",
		PvtKeyFlag:    0x04b2430c,
		PvKeyPrefix:   "zprv",
		PubKeyFlag:    0x04b24746,
		PubKeyPrefix:  "zpub",
		AddrEncodings: []AddrEncoding{P2WPKH},
		Path:          "m/84'/57'",
	}
	SyscoinZprvZpub = VersionBytes{
		Coin:          "Syscoin",
		PvtKeyFlag:    0x02aa7a99,
		PvKeyPrefix:   "Zprv",
		PubKeyFlag:    0x02aa7ed3,
		PubKeyPrefix:  "Zpub",
		AddrEncodings: []AddrEncoding{P2WSH},
		Path:          "m/84'/57'",
	}
	Vertcoinvtcpvtcv = VersionBytes{
		Coin:          "Vertcoin",
		PvtKeyFlag:    0x0488ade4,
		PvKeyPrefix:   "vtcv",
		PubKeyFlag:    0x0488b21e,
		PubKeyPrefix:  "vtcp",
		AddrEncodings: []AddrEncoding{P2PKH, P2SH},
		Path:          "m/44'/28'",
	}
)

// PathToHDBytesSlice Registered HD Version Bytes Path (Two Level) To VersionBytes
var PathToHDBytesSlice = map[string][]VersionBytes{
	"m/44'/0'": {Bitcoinxprvxpub},
	"m/49'/0'": {Bitcoinyprvypub},
	"m/84'/0'": {Bitcoinzprvzpub, BitcoinYprvYpub, BitcoinZprvZpub},
	"m/44'/1'": {Bitcointprvtpub, Groestlcointprvtpub, Litecointtpvttub,
		Nexaxprvxpub2},
	"m/49'/1'": {Bitcoinuprvupub, Groestlcoinuprvupub},
	"m/84'/1'": {Bitcoinvprvvpub, BitcoinUprvUpub, BitcoinVprvVpub,
		Groestlcoinvprvvpub, GroestlcoinUprvUpub, GroestlcoinVprvVpub},
	"m/44'/17'":    {Groestlcoinxprvxpub},
	"m/49'/17'":    {Groestlcoinyprvypub},
	"m/84'/17'":    {Groestlcoinzprvzpub, GroestlcoinYprvYpub, GroestlcoinZprvZpub},
	"m/44'/2'":     {LitecoinLtpvLtub},
	"m/49'/2'":     {LitecoinMtpvMtub},
	"m/44'/29223'": {Nexaxprvxpub},
	"m/44'/28'":    {Vertcoinvtcpvtcv},
	"m/44'/1997'":  {Polispprvppub},
	"m/84'/57'":    {Syscoinzprvzpub, SyscoinZprvZpub},
}

func init() {
	for _, hdBytesArr := range PathToHDBytesSlice {
		HDVersionBytesSlice = append(HDVersionBytesSlice, hdBytesArr...)
		for _, hdBytes := range hdBytesArr {
			val, _ := PvtFlagToHDBytesSlice[hdBytes.PvtKeyFlag]
			PvtFlagToHDBytesSlice[hdBytes.PvtKeyFlag] = append(val, hdBytes)
			val, _ = PubFlagToHDBytesSlice[hdBytes.PubKeyFlag]
			PubFlagToHDBytesSlice[hdBytes.PubKeyFlag] = append(val, hdBytes)
			purposeVal, _ := hdBytes.PurposeVal()
			purposeFullVal, _ := hdBytes.PurposeValFull()
			PurposeToHDBytesSlice[purposeVal] = append(PurposeToHDBytesSlice[purposeVal], hdBytes)
			PurposeFullToHDBytesSlice[purposeFullVal] = append(PurposeFullToHDBytesSlice[purposeFullVal], hdBytes)
			coinVal, _ := hdBytes.CoinVal()
			CoinTypeToHDBytesSlice[coinVal] = append(CoinTypeToHDBytesSlice[coinVal], hdBytes)
		}
	}
}
