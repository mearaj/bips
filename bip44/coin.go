package bip44

import (
	"fmt"
	"strings"
)

type Coin struct {
	Type uint32 `json:"type,omitempty"`
	// PathComponent(bip32.FirstHardenedChild + Type) in hex
	PathComponent uint32 `json:"pathComponent,omitempty"`
	Symbol        string `json:"symbol,omitempty"`
	Name          string `json:"name,omitempty"`
}

func (c Coin) StringForMD() string {
	return fmt.Sprintf(
		`{Type: %d, PathComponent: 0x%x, Symbol: "%s", Name: "%s"}`,
		c.Type, c.PathComponent, c.Symbol, c.Name)
}

func (c Coin) String() string {
	var name string
	if c.Symbol == "" {
		name = c.Name
	} else {
		name = c.Symbol + " - " + c.Name
	}
	name = fmt.Sprintf("%s (Type %d)", name, c.Type)
	return name
}

func (c Coin) IsValid() bool {
	return strings.TrimSpace(c.Name)+strings.TrimSpace(c.Symbol) != "" &&
		c.Type < 0x80000000 && c.PathComponent >= 0x80000000 &&
		c.Type+0x80000000 == c.PathComponent
}
