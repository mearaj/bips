package main

import (
	"flag"
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/markdown"
	"gioui.org/x/outlay"
	"gioui.org/x/richtext"
	"github.com/mearaj/bips"
	"github.com/mearaj/bips/bip32"
	"github.com/mearaj/bips/util"
	"image"
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
type Flex = layout.Flex

type KeyPathTab struct {
	util.KeyPath
	widget.Clickable
}
type KeyPathTabs struct {
	layout.List
	tabs     []KeyPathTab
	selected int
}

var keyPathTabs KeyPathTabs

var Rigid = layout.Rigid

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

var viewLayout = layout.List{Axis: layout.Vertical}
var tabsSlider Slider

var markdownRd = markdown.NewRenderer()
var richTextState richtext.InteractiveText

var onKeyPathChange = func() {
	str := "m"
	if derivationPath.IsValid() {
		str = derivationPath.String()
	}
	keyPaths, _ := bps.DeriveBIP32Result(util.Path(str))
	keyPathTabs.tabs = make([]KeyPathTab, 0)
	for _, keyPath := range keyPaths {
		keyPathTabs.tabs = append(keyPathTabs.tabs,
			KeyPathTab{
				KeyPath: keyPath,
			},
		)
	}
	if keyPathTabs.selected >= len(keyPathTabs.tabs) {
		keyPathTabs.selected = len(keyPathTabs.tabs) - 1
	}
}

var onAutoCreateMnemonicClicked = func() {
	val := radioGroupMnemonicWords.Value
	wordsCount, _ := strconv.Atoi(val)
	mnemonic, _ = util.GenerateMnemonic(byte(wordsCount))
	mnemonicField.SetText(mnemonic)
	seed, _ = util.DeriveSeedFromMnemonic(mnemonic, mnemonicPassPhrase)
	bip39SeedField.SetText(seed)
	rootKey, _ := util.RootKeyFromSeed(seed)
	bps.SetRootKey(*rootKey)
	onKeyPathChange()
	rootKey = bps.RootKey()
	rootKeyStr = rootKey.String()
	bip32RootKeyField.SetText(rootKeyStr)
}
var onMnemonicChange = func() {
	mnemonic = mnemonicField.Text()
	seed, _ = util.DeriveSeedFromMnemonic(mnemonic, mnemonicPassPhrase)
	bip39SeedField.SetText(seed)
	rootKey, _ := util.RootKeyFromSeed(seed)
	bps.SetRootKey(*rootKey)
	onKeyPathChange()
	rootKey = bps.RootKey()
	rootKeyStr = rootKey.String()
	bip32RootKeyField.SetText(rootKeyStr)
}

var onMnemonicPassphraseChange = func() {
	mnemonicPassPhrase = mnemonicPassphraseField.Text()
	seed, _ = util.DeriveSeedFromMnemonic(mnemonic, mnemonicPassPhrase)
	bip39SeedField.SetText(seed)
	rootKey, _ := util.RootKeyFromSeed(seed)
	bps.SetRootKey(*rootKey)
	onKeyPathChange()
	rootKey = bps.RootKey()
	rootKeyStr = rootKey.String()
	bip32RootKeyField.SetText(rootKeyStr)
}

var onSeedChange = func() {
	seed = bip39SeedField.Text()
	mnemonic = ""
	mnemonicField.SetText(mnemonic)
	rootKey, _ := util.RootKeyFromSeed(seed)
	bps.SetRootKey(*rootKey)
	onKeyPathChange()
	rootKey = bps.RootKey()
	rootKeyStr = rootKey.String()
	bip32RootKeyField.SetText(rootKeyStr)
}

var onRootKeyStrChange = func() {
	rootKeyStr = bip32RootKeyField.Text()
	mnemonic = ""
	mnemonicField.SetText(mnemonic)
	seed = ""
	bip39SeedField.SetText(seed)
	rootKey, err := bip32.B58Deserialize(rootKeyStr)
	if err == nil {
		bps.SetRootKey(rootKey)
		onKeyPathChange()
		rootKey = *bps.RootKey()
		rootKeyStr = rootKey.String()
		bip32RootKeyField.SetText(rootKeyStr)
	} else {
		bps.SetRootKey(bip32.Key{})
	}
}
var OnDerivationPathChange = func() {
	derivationStr = derivationField.Text()
	derivationPath = util.Path(derivationStr).Formatted()
	onKeyPathChange()
	rootKey := *bps.RootKey()
	rootKeyStr = rootKey.String()
	bip32RootKeyField.SetText(rootKeyStr)
}

func loop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	var ops op.Ops
	radioGroupMnemonicWords.Value = "12"
	for {
		switch e := w.NextEvent().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			inset := layout.UniformInset(32)
			derivationPath = util.Path(derivationStr).Formatted()
			if btnGenerateMnemonic.Clicked(gtx) {
				onAutoCreateMnemonicClicked()
			}
			if mnemonicField.Text() != mnemonic {
				onMnemonicChange()
			}
			if mnemonicPassphraseField.Text() != mnemonicPassPhrase {
				onMnemonicPassphraseChange()
			}
			if bip39SeedField.Text() != seed {
				onSeedChange()
			}
			if rootKeyStr != bip32RootKeyField.Text() {
				onRootKeyStrChange()
			}
			if util.Path(derivationField.Text()).String() !=
				util.Path(derivationStr).String() {
				OnDerivationPathChange()
			}
			masterKey := bps.RootKey()
			if !masterKey.IsValid() {

			}
			gtx.Constraints.Min = gtx.Constraints.Max
			inset.Layout(gtx, func(gtx Gtx) Dim {
				return viewLayout.Layout(gtx, 1, func(gtx Gtx, index int) Dim {
					flex := Flex{Axis: layout.Vertical}
					return flex.Layout(gtx,
						Rigid(func(gtx Gtx) Dim {
							return layout.Center.Layout(gtx, func(gtx Gtx) Dim {
								return material.H3(th, "INPUT").Layout(gtx)
							})
						}),
						Rigid(func(gtx Gtx) Dim {
							flex := Flex{Axis: layout.Vertical, Alignment: layout.Start}
							spc := layout.Spacer{Height: 8}
							return flex.Layout(gtx,
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
						}),
						Rigid(layout.Spacer{Height: 16}.Layout),
						Rigid(func(gtx Gtx) Dim {
							return mnemonicField.Layout(gtx, th, "Enter your mnemonic")
						}),
						Rigid(func(gtx Gtx) Dim {
							return mnemonicPassphraseField.Layout(gtx, th,
								"Enter your mnemonic passphrase(optional)")
						}),
						Rigid(func(gtx Gtx) Dim {
							return bip39SeedField.Layout(gtx, th, "Enter your BIP39 Seed")
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
									Rigid(func(gtx Gtx) layout.Dimensions {
										txt := material.H5(th, "Enter Derivation Path")
										txt.Alignment = text.Middle
										return txt.Layout(gtx)
									}),
									Rigid(layout.Spacer{Height: 2}.Layout),
									Rigid(func(gtx Gtx) layout.Dimensions {
										txt := material.Label(th, 16, "Ex m/44'/60'/0'/0/0 (Bip44 for Ethereum Coin(60'))")
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
						Rigid(func(gtx Gtx) Dim {
							return keyPathTabs.List.Layout(gtx, len(keyPathTabs.tabs), func(gtx Gtx, tabIdx int) Dim {
								t := &keyPathTabs.tabs[tabIdx]
								if t.Clicked(gtx) {
									if keyPathTabs.selected < tabIdx {
										tabsSlider.PushLeft()
									} else if keyPathTabs.selected > tabIdx {
										tabsSlider.PushRight()
									}
									keyPathTabs.selected = tabIdx
								}
								var tabWidth int
								return layout.Stack{Alignment: layout.S}.Layout(gtx,
									layout.Stacked(func(gtx Gtx) Dim {
										dims := material.Clickable(gtx, &t.Clickable, func(gtx Gtx) Dim {
											return layout.UniformInset(unit.Dp(12)).Layout(gtx,
												material.H6(th, t.KeyPath.Path.String()).Layout,
											)
										})
										tabWidth = dims.Size.X
										return dims
									}),
									layout.Stacked(func(gtx Gtx) Dim {
										if keyPathTabs.selected != tabIdx {
											return layout.Dimensions{}
										}
										tabHeight := gtx.Dp(unit.Dp(4))
										tabRect := image.Rect(0, 0, tabWidth, tabHeight)
										paint.FillShape(gtx.Ops, th.Palette.ContrastBg, clip.Rect(tabRect).Op())
										return layout.Dimensions{
											Size: image.Point{X: tabWidth, Y: tabHeight},
										}
									}),
								)
							})
						}),
						Rigid(func(gtx Gtx) layout.Dimensions {
							return tabsSlider.Layout(gtx, func(gtx Gtx) layout.Dimensions {
								isValid := keyPathTabs.selected < len(keyPathTabs.tabs) &&
									len(keyPathTabs.tabs) != 0 &&
									keyPathTabs.selected >= 0
								if !isValid {
									return layout.Dimensions{}
								}
								keyPath := keyPathTabs.tabs[keyPathTabs.selected]
								flex := layout.Flex{Axis: layout.Vertical}
								return flex.Layout(gtx,
									Rigid(func(gtx Gtx) layout.Dimensions {
										return material.Label(th, 16, keyPath.Path.String()).Layout(gtx)
									}),
									Rigid(func(gtx Gtx) layout.Dimensions {
										return material.Label(th, 16, keyPath.Key.String()).Layout(gtx)
									}),
									Rigid(func(gtx Gtx) layout.Dimensions {
										publicKeyExtended := keyPath.Key.PublicKeyExtended()
										return material.Label(th, 16, publicKeyExtended.String()).Layout(gtx)
									}),
									Rigid(func(gtx Gtx) layout.Dimensions {
										return material.Label(th, 16, keyPath.Key.PrivateKeyHex()).Layout(gtx)
									}),
									Rigid(func(gtx Gtx) layout.Dimensions {
										publicKeyExtended := keyPath.Key.PublicKeyExtended()
										return material.Label(th, 16, publicKeyExtended.PublicKeyHex()).Layout(gtx)
									}),
									Rigid(func(gtx Gtx) layout.Dimensions {
										return material.Label(th, 16, keyPath.AddrHex()).Layout(gtx)
									}),
								)
							})
						}),
						Rigid(func(gtx layout.Context) layout.Dimensions {
							ly, err := markdownRd.Render(bips.ReadMeMarkdown)
							if err != nil {
								return material.Label(th, 16, err.Error()).Layout(gtx)
							}
							return richtext.Text(&richTextState, th.Shaper, ly...).Layout(gtx)
						}),
					)
				})
			})
			e.Frame(gtx.Ops)
		}
	}
}
