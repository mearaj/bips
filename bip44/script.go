package bip44

import (
	"embed"
	"github.com/ethereum/go-ethereum/common/math"
	"log"
	"os"
	"strings"
)

//go:embed slip-0044.md
var Slip044FS embed.FS
var Slip044FileName = "slip-0044.md"

// Only needed during development(upgrade) to parse slip-0044.md for RegCoins []BIP44Coin
//func init() {
//	 generateFileFromSlip44MD()
//}

// GenerateFileFromSlip44MD This generates a json file from slip-0044.md file
// Ref https://github.com/satoshilabs/slips/blob/master/slip-0044.md
// Only needed during development(upgrade) to parse slip-0044.md for RegCoins []BIP44Coin
func generateFileFromSlip44MD() {
	var registeredCoins []Coin
	val, err := Slip044FS.ReadFile(Slip044FileName)
	if err != nil {
		log.Fatal(err)
	}
	st := string(val)
	multiLines := strings.Split(st, "\n")
	for _, eachLine := range multiLines {
		var fields = strings.Split(eachLine, "| ")[1:]
		for i, _ := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		canParse := len(fields) > 2 &&
			!strings.Contains(fields[len(fields)-1], "reserved") &&
			!strings.Contains(fields[len(fields)-1], "dead")
		if canParse {
			var regCoin Coin
			coinType, ok := math.ParseUint64(fields[0])
			if ok {
				regCoin.Type = uint32(coinType)
				pathComp, ok := math.ParseUint64(fields[1])
				regCoin.PathComponent = uint32(pathComp)
				if ok {
					if len(fields) == 3 {
						regCoin.Name = fields[2]
					} else {
						regCoin.Symbol = fields[2]
						regCoin.Name = strings.Join(fields[3:], " ")
					}
					if !(regCoin.Symbol == "" && regCoin.Name == "") {
						registeredCoins = append(registeredCoins, regCoin)
					}
				}
			}
		}
	}
	var contents string
	for _, regCoin := range registeredCoins {
		contents += regCoin.StringForMD() + ",\n"
	}
	// file, _ := json.MarshalIndent(bips.RegCoins, "", " ")
	_ = os.WriteFile("testing.go", []byte(contents), os.ModePerm)
}
