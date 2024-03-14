package bip32

import (
	"errors"
	"fmt"
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
// than or equal to FirstHardenedChild or if it's value is lesser but contains apostrophe(')
type Path string

var PathRegex = regexp.MustCompile(`^[mM](/\d+')*(/\d+)*$`)

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
//	0    1        2        3      4
//
// m/purpose'/coin_type'/account'/change/address_index
// Results in []uint32{Purpose,CoinType,Account,Change,AddIndex}
// Returns real values  (i.e. 40 is 40 and not 40 + FirstHardenedChild)
func (p Path) ValuesAtDepth() ([]uint32, error) {
	if !PathRegex.MatchString(p.String()) {
		return nil, ErrUnSupportedOrInvalidPath
	}
	// there's no value associated with m at depth 0 hence
	// in order to avoid confusion and map to bip32 path derivation,
	// depth level starts at m as zero depth
	pathArr := []uint32{0}
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

// ValueAtDepth
// Example
// 0   1         2         3         4     5
// m/purpose'/coin_type'/account'/change/address_index
func (p Path) ValueAtDepth(b byte) (uint32, error) {
	vals, err := p.ValuesAtDepth()
	if err != nil {
		return 0, err
	}
	if !(b < byte(len(vals))) {
		return 0, ErrUnSupportedOrInvalidPath
	}
	return vals[b], nil
}
func (p Path) ReplaceValueAtDepth(b byte, val uint32) (Path, error) {
	newPath := p
	vals, err := newPath.ValuesAtDepth()
	if err != nil {
		return newPath, err
	}
	if !(b < byte(len(vals))) {
		return newPath, ErrUnSupportedOrInvalidPath
	}
	valsArr := strings.Split(newPath.String(), "/")
	if val >= FirstHardenedChild {
		valsArr[b] = fmt.Sprintf("%d'", val%FirstHardenedChild)
	} else {
		valsArr[b] = fmt.Sprintf("%d", val)
	}
	return Path(strings.Join(valsArr, "/")), nil
}
