package main

import (
	"flag"
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/outlay"
	"github.com/mearaj/bips/bip32"
	"github.com/mearaj/bips/util"
	"log"
	"os"
	"strconv"
)

func main() {
	flag.Parse()
	go func() {
		w := app.NewWindow(app.Title("Bip39 Tool"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type Gtx = layout.Context
type Dim = layout.Dimensions

var Rigid = layout.Rigid

type Flex = layout.Flex

func loop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	var ops op.Ops
	var bps = util.Generator{}
	var btnGenerateMnemonic widget.Clickable
	var radioGroupMnemonicWords widget.Enum
	var mnemonicField component.TextField
	var mnemonicPassphraseField component.TextField
	var bip39SeedField component.TextField
	var bip32RootKeyField component.TextField
	var derivationStr string
	var derivationPath util.Path
	var derivationField component.TextField
	var rootKeyStr string
	var mnemonic string
	var mnemonicPassPhrase string
	var seed string
	var keyPaths []util.KeyPath
	var derivedResultsList layout.List
	derivedResultsList.Axis = layout.Vertical
	viewLayout := layout.List{Axis: layout.Vertical}
	radioGroupMnemonicWords.Value = "12"
	// generateAddressesRange := uint32(20)
	var consumedHeight int
	for {
		switch e := w.NextEvent().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			inset := layout.UniformInset(32)
			derivationPath = util.Path(derivationStr).Formatted()
			if btnGenerateMnemonic.Clicked(gtx) {
				val := radioGroupMnemonicWords.Value
				wordsCount, _ := strconv.Atoi(val)
				mnemonic, _ = util.GenerateMnemonic(byte(wordsCount))
				mnemonicField.SetText(mnemonic)
				seed, _ = util.DeriveSeedFromMnemonic(mnemonic, mnemonicPassPhrase)
				bip39SeedField.SetText(seed)
				rootKey, _ := util.RootKeyFromSeed(seed)
				str := "m"
				if derivationPath.IsValid() {
					str = derivationPath.String()
				}
				keyPaths, _ = bps.DeriveBIP32Result(util.Path(str))
				rootKey = bps.RootKey()
				rootKeyStr = rootKey.String()
				bip32RootKeyField.SetText(rootKeyStr)
			}
			if mnemonicField.Text() != mnemonic {
				mnemonic = mnemonicField.Text()
				seed, _ = util.DeriveSeedFromMnemonic(mnemonic, mnemonicPassPhrase)
				bip39SeedField.SetText(seed)
				rootKey, _ := util.RootKeyFromSeed(seed)
				bps.SetRootKey(*rootKey)
				str := "m"
				if derivationPath.IsValid() {
					str = derivationPath.String()
				}
				keyPaths, _ = bps.DeriveBIP32Result(util.Path(str))
				rootKey = bps.RootKey()
				rootKeyStr = rootKey.String()
				bip32RootKeyField.SetText(rootKeyStr)
			}
			if mnemonicPassphraseField.Text() != mnemonicPassPhrase {
				mnemonicPassPhrase = mnemonicPassphraseField.Text()
				seed, _ = util.DeriveSeedFromMnemonic(mnemonic, mnemonicPassPhrase)
				bip39SeedField.SetText(seed)
				rootKey, _ := util.RootKeyFromSeed(seed)
				bps.SetRootKey(*rootKey)
				str := "m"
				if derivationPath.IsValid() {
					str = derivationPath.String()
				}
				keyPaths, _ = bps.DeriveBIP32Result(util.Path(str))
				rootKey = bps.RootKey()
				rootKeyStr = rootKey.String()
				bip32RootKeyField.SetText(rootKeyStr)
			}
			if bip39SeedField.Text() != seed {
				seed = bip39SeedField.Text()
				mnemonic = ""
				mnemonicField.SetText(mnemonic)
				rootKey, _ := util.RootKeyFromSeed(seed)
				bps.SetRootKey(*rootKey)
				str := "m"
				if derivationPath.IsValid() {
					str = derivationPath.String()
				}
				keyPaths, _ = bps.DeriveBIP32Result(util.Path(str))
				rootKey = bps.RootKey()
				rootKeyStr = rootKey.String()
				bip32RootKeyField.SetText(rootKeyStr)
			}
			if rootKeyStr != bip32RootKeyField.Text() {
				rootKeyStr = bip32RootKeyField.Text()
				mnemonic = ""
				mnemonicField.SetText(mnemonic)
				seed = ""
				bip39SeedField.SetText(seed)
				rootKey, err := bip32.B58Deserialize(rootKeyStr)
				if err == nil {
					bps.SetRootKey(rootKey)
					str := "m"
					if derivationPath.IsValid() {
						str = derivationPath.String()
					}
					keyPaths, _ = bps.DeriveBIP32Result(util.Path(str))
					rootKey = *bps.RootKey()
					rootKeyStr = rootKey.String()
					bip32RootKeyField.SetText(rootKeyStr)
				} else {
					bps.SetRootKey(bip32.Key{})
				}
			}
			if util.Path(derivationField.Text()).String() !=
				util.Path(derivationStr).String() {
				derivationStr = derivationField.Text()
				derivationPath = util.Path(derivationStr).Formatted()
				str := "m"
				if derivationPath.IsValid() {
					str = derivationPath.String()
				}
				keyPaths, _ = bps.DeriveBIP32Result(util.Path(str))
				rootKey := *bps.RootKey()
				rootKeyStr = rootKey.String()
				bip32RootKeyField.SetText(rootKeyStr)
			}
			masterKey := bps.RootKey()
			if !masterKey.IsValid() {

			}
			gtx.Constraints.Min = gtx.Constraints.Max
			consumedHeight = 0
			inset.Layout(gtx, func(gtx Gtx) Dim {
				return viewLayout.Layout(gtx, 1, func(gtx Gtx, index int) Dim {
					flex := Flex{Axis: layout.Vertical}
					return flex.Layout(gtx,
						Rigid(func(gtx Gtx) Dim {
							d := layout.Center.Layout(gtx, func(gtx Gtx) Dim {
								return material.H3(th, "INPUT").Layout(gtx)
							})
							consumedHeight += d.Size.Y
							return d
						}),
						Rigid(func(gtx Gtx) Dim {
							flex := Flex{Axis: layout.Vertical, Alignment: layout.Start}
							spc := layout.Spacer{Height: 8}
							d := flex.Layout(gtx,
								Rigid(func(gtx Gtx) Dim {
									flex := Flex{Alignment: layout.Middle}
									return flex.Layout(gtx, Rigid(func(gtx Gtx) Dim {
										return material.Button(th, &btnGenerateMnemonic, "Auto Generate Mnemonic").Layout(gtx)
									}))
								}),
								Rigid(spc.Layout),
								Rigid(func(gtx Gtx) Dim {
									flowWrap := outlay.FlowWrap{}
									return flowWrap.Layout(gtx, 5, func(gtx Gtx, i int) Dim {
										var k, lbl string
										switch i {
										case 0:
											k = "12"
											lbl = "12 words"
										case 1:
											k = "15"
											lbl = "15 words"
										case 2:
											k = "18"
											lbl = "18 words"
										case 3:
											k = "21"
											lbl = "21 words"
										case 4:
											k = "24"
											lbl = "24 words"
										}
										inset := layout.Inset{Right: 16}
										return inset.Layout(gtx, func(gtx Gtx) Dim {
											return material.RadioButton(th,
												&radioGroupMnemonicWords, k, lbl).Layout(gtx)
										})
									})
								}),
							)
							consumedHeight += d.Size.Y
							return d
						}),
						Rigid(layout.Spacer{Height: 16}.Layout),
						Rigid(func(gtx Gtx) Dim {
							d := mnemonicField.Layout(gtx, th, "Enter your mnemonic")
							consumedHeight += d.Size.Y + gtx.Dp(8)
							return d
						}),
						Rigid(func(gtx Gtx) Dim {
							d := mnemonicPassphraseField.Layout(gtx, th,
								"Enter your mnemonic passphrase(optional)")
							consumedHeight += d.Size.Y
							return d
						}),
						Rigid(func(gtx Gtx) Dim {
							d := bip39SeedField.Layout(gtx, th, "Enter your BIP39 Seed")
							consumedHeight += d.Size.Y
							return d
						}),
						Rigid(func(gtx Gtx) Dim {
							return bip32RootKeyField.Layout(gtx, th, "Enter your BIP32 Root Key")
						}),
						Rigid(layout.Spacer{Height: 16}.Layout),
						Rigid(func(gtx Gtx) Dim {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return layout.Center.Layout(gtx, func(gtx Gtx) Dim {
								flex := layout.Flex{Alignment: layout.Middle, Axis: layout.Vertical}
								return flex.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										txt := material.H5(th, "Enter Derivation Path")
										txt.Alignment = text.Middle
										return txt.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Height: 2}.Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										txt := material.Label(th, 16, "Ex m/44'/60' (Bip44 for Ethereum Coin)")
										txt.Alignment = text.Middle
										return txt.Layout(gtx)
									}),
								)
							})
						}),
						Rigid(func(gtx Gtx) Dim {
							hintText := "Ex m/44'/60'"
							return derivationField.Layout(gtx, th, hintText)
						}),
						Rigid(layout.Spacer{Height: 32}.Layout),
						Rigid(func(gtx Gtx) Dim {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return layout.Center.Layout(gtx, func(gtx Gtx) Dim {
								return material.H3(th, "OUTPUT").Layout(gtx)
							})
						}),
						Rigid(layout.Spacer{Height: 32}.Layout),
						Rigid(func(gtx layout.Context) layout.Dimensions {
							return derivedResultsList.Layout(gtx, len(keyPaths), func(gtx layout.Context, index int) layout.Dimensions {
								keyPath := &keyPaths[index]
								flex := layout.Flex{Axis: layout.Vertical}
								return flex.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return material.Label(th, 16, keyPath.Path.String()).Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										insets := layout.Inset{Left: 16}
										return insets.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											flex := layout.Flex{Axis: layout.Vertical}
											return flex.Layout(gtx,
												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
													return material.Label(th, 16, keyPath.Key.String()).Layout(gtx)
												}),
												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
													publicKeyExtended := keyPath.Key.PublicKeyExtended()
													return material.Label(th, 16, publicKeyExtended.String()).Layout(gtx)
												}),
											)
										})
									}),
								)
							})
						}),
					)
				})
			})
			e.Frame(gtx.Ops)
		}
	}
}
