package util

import "errors"

var (
	ErrInputValidationFailed       = errors.New("input validation failed")
	ErrInvalidMnemonicWordsCount   = errors.New("provided mnemonic words count is not supported")
	ErrInvalidBIP39Seed            = errors.New("provided seed is not valid")
	ErrInvalidRootKey              = errors.New("root key is invalid")
	ErrInvalidRangeProvided        = errors.New("range provided is invalid")
	ErrUnSupportedPurposePath      = errors.New("unsupported purpose path")
	ErrCoinTypeIsRequired          = errors.New("coin type is required")
	ErrAccountPathShouldBeHardened = errors.New("account path should be hardened")
	ErrChangeIsRequired            = errors.New("change is required")
	ErrUnsupportedCoinType         = errors.New("coin type is not supported")
	ErrUnSupportedOrInvalidPath    = errors.New("path is unsupported and/or invalid")
	ErrPathDepthNeedGreaterThanOne = errors.New("path depth must be greater than or equal to one")
)
