package bip32

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrUnSupportedOrInvalidPath = errors.New("path is unsupported and/or invalid")
)

// Path is pattern -->  m/purpose'/coin_type'/account'/change/address_index
// Ref https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
// and Ref bip32.Key
// A path component is hardened if it's value is greater
// than or equal to bip32.FirstHardenedChild or if it's value is lesser,
// contains apostrophe(')
type Path string

var PathRegex = regexp.MustCompile(`^m(/\d+')*(/\d+)*$`)

func (p Path) String() string {
	reg := regexp.MustCompile(`\s`)
	return reg.ReplaceAllString(string(p), "")
}

func (p Path) Formatted() Path {
	return Path(p.String())
}

func (p Path) IsValid() bool {
	_, err := p.ValuesAtDepth()
	return err == nil
}

// ValuesAtDepth (correspondingIndexes)
//
//	0         1          2        3      4
//
// m/purpose'/coin_type'/account'/change/address_index
// Results in []uint32{Purpose,CoinType,Account,Change,AddIndex}
// Returns real hardened and non hardened values
func (p Path) ValuesAtDepth() ([]uint32, error) {
	if !PathRegex.MatchString(p.String()) {
		return nil, ErrUnSupportedOrInvalidPath
	}
	var pathArr []uint32
	pth := strings.Split(p.String(), "/")
	for _, s := range pth[1:] {
		isHardened := strings.Contains(s, "'")
		if isHardened {
			fmtStr := strings.ReplaceAll(s, "'", "")
			val, err := strconv.Atoi(fmtStr)
			if err != nil {
				return pathArr, err
			}
			if uint32(val) >= FirstHardenedChild {
				return pathArr, ErrUnSupportedOrInvalidPath
			}
			pathArr = append(pathArr, uint32(val)+FirstHardenedChild)
		} else {
			val, err := strconv.Atoi(s)
			if err != nil {
				return pathArr, err
			}
			pathArr = append(pathArr, uint32(val))
		}
	}
	return pathArr, nil
}
