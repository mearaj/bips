package bip32

import "errors"

var (
	ErrUnSupportedHDVersionBytes = errors.New("unsupported hd version bytes")
	ErrInvalidPurpose            = errors.New("invalid purpose")
	ErrInvalidCoin               = errors.New("invalid coin")
)
