package settings

import (
	"fmt"
	"strconv"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil/v4"
	"gitlab.com/raedah/cryptopower/app"
	"gitlab.com/raedah/cryptopower/libwallet/wallets/dcr"
	"gitlab.com/raedah/cryptopower/ui/cryptomaterial"
	"gitlab.com/raedah/cryptopower/ui/load"
	"gitlab.com/raedah/cryptopower/ui/modal"
	"gitlab.com/raedah/cryptopower/ui/page/components"
	"gitlab.com/raedah/cryptopower/ui/utils"
	"gitlab.com/raedah/cryptopower/ui/values"
)

const AccountDetailsPageID = "AccountDetails"

type AcctDetailsPage struct {
	*load.Load
	// GenericPageModal defines methods such as ID() and OnAttachedToNavigator()
	// that helps this Page satisfy the app.Page interface. It also defines
	// helper methods for accessing the PageNavigator that displayed this page
	// and the root WindowNavigator.
	*app.GenericPageModal

	wallet  *dcr.Wallet
	account *dcr.Account

	theme                    *cryptomaterial.Theme
	acctDetailsPageContainer layout.List
	list                     *widget.List
	backButton               cryptomaterial.IconButton
	renameAccount            *cryptomaterial.Clickable
	extendedKeyClickable     *cryptomaterial.Clickable
	showExtendedKeyButton    *cryptomaterial.Clickable
	infoButton               cryptomaterial.IconButton

	stakingBalance   int64
	totalBalance     string
	spendable        string
	immatureRewards  string
	lockedByTickets  string
	votingAuthority  string
	immatureStakeGen string
	hdPath           string
	keys             string
	extendedKey      string

	isHiddenExtendedxPubkey bool
}

func NewAcctDetailsPage(l *load.Load, account *dcr.Account) *AcctDetailsPage {
	pg := &AcctDetailsPage{
		Load:             l,
		GenericPageModal: app.NewGenericPageModal(AccountDetailsPageID),
		wallet:           l.WL.SelectedWallet.Wallet,
		account:          account,

		theme:                    l.Theme,
		acctDetailsPageContainer: layout.List{Axis: layout.Vertical},
		list: &widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		backButton:              l.Theme.IconButton(l.Theme.Icons.NavigationArrowBack),
		renameAccount:           l.Theme.NewClickable(false),
		extendedKeyClickable:    l.Theme.NewClickable(true),
		showExtendedKeyButton:   l.Theme.NewClickable(false),
		isHiddenExtendedxPubkey: true,
	}

	pg.backButton, _ = components.SubpageHeaderButtons(l)

	return pg
}

// OnNavigatedTo is called when the page is about to be displayed and
// may be used to initialize page features that are only relevant when
// the page is displayed.
// Part of the load.Page interface.
func (pg *AcctDetailsPage) OnNavigatedTo() {

	balance := pg.account.Balance

	pg.stakingBalance = balance.ImmatureReward + balance.LockedByTickets + balance.VotingAuthority +
		balance.ImmatureStakeGeneration

	pg.totalBalance = dcrutil.Amount(balance.Total).String()
	pg.spendable = dcrutil.Amount(balance.Spendable).String()
	pg.immatureRewards = dcrutil.Amount(balance.ImmatureReward).String()
	pg.lockedByTickets = dcrutil.Amount(balance.LockedByTickets).String()
	pg.votingAuthority = dcrutil.Amount(balance.VotingAuthority).String()
	pg.immatureStakeGen = dcrutil.Amount(balance.ImmatureStakeGeneration).String()

	pg.hdPath = pg.WL.HDPrefix() + strconv.Itoa(int(pg.account.Number)) + "'"

	ext := pg.account.ExternalKeyCount
	internal := pg.account.InternalKeyCount
	imp := pg.account.ImportedKeyCount
	pg.keys = values.StringF(values.StrAcctDetailsKey, ext, internal, imp)
	_, pg.infoButton = components.SubpageHeaderButtons(pg.Load)
	pg.loadExtendedPubKey()
}

// Layout draws the page UI components into the provided C
// to be eventually drawn on screen.
// Part of the load.Page interface.
func (pg *AcctDetailsPage) Layout(gtx C) D {
	m := values.MarginPadding10
	widgets := []func(gtx C) D{
		func(gtx C) D {
			return pg.accountBalanceLayout(gtx)
		},
		func(gtx C) D {
			return layout.Inset{Top: m, Bottom: m}.Layout(gtx, pg.theme.Separator().Layout)
		},
		pg.accountInfoLayout,
		func(gtx C) D {
			return layout.Inset{Top: m, Bottom: m}.Layout(gtx, pg.theme.Separator().Layout)
		},
		func(gtx C) D {
			return layout.Inset{Bottom: m}.Layout(gtx, pg.extendedPubkey)
		},
	}
	if pg.Load.GetCurrentAppWidth() <= gtx.Dp(values.StartMobileView) {
		return pg.layoutMobile(gtx, widgets)
	}
	return pg.layoutDesktop(gtx, widgets)
}

func (pg *AcctDetailsPage) layoutDesktop(gtx layout.Context, widgets []func(gtx C) D) layout.Dimensions {
	body := func(gtx C) D {
		sp := components.SubPage{
			Load:       pg.Load,
			Title:      pg.account.Name,
			BackButton: pg.backButton,
			Back: func() {
				pg.ParentNavigator().CloseCurrentPage()
			},
			Body: func(gtx C) D {
				return pg.Theme.List(pg.list).Layout(gtx, 1, func(gtx C, i int) D {
					return layout.Inset{
						Bottom: values.MarginPadding7,
						Right:  values.MarginPadding2,
					}.Layout(gtx, func(gtx C) D {
						return pg.theme.Card().Layout(gtx, func(gtx C) D {
							return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
								return pg.acctDetailsPageContainer.Layout(gtx, len(widgets), func(gtx C, i int) D {
									return layout.Inset{}.Layout(gtx, widgets[i])
								})
							})
						})
					})
				})
			},
			ExtraItem: pg.renameAccount,
			Extra: func(gtx C) D {
				return layout.Inset{}.Layout(gtx, func(gtx C) D {
					edit := pg.Theme.Icons.EditIcon
					return layout.E.Layout(gtx, edit.Layout24dp)
				})
			},
		}
		return sp.Layout(pg.ParentWindow(), gtx)
	}
	return components.UniformPadding(gtx, body)
}

func (pg *AcctDetailsPage) layoutMobile(gtx layout.Context, widgets []func(gtx C) D) layout.Dimensions {
	body := func(gtx C) D {
		sp := components.SubPage{
			Load:       pg.Load,
			Title:      pg.account.Name,
			BackButton: pg.backButton,
			Back: func() {
				pg.ParentNavigator().CloseCurrentPage()
			},
			Body: func(gtx C) D {
				return pg.Theme.List(pg.list).Layout(gtx, 1, func(gtx C, i int) D {
					return layout.Inset{
						Bottom: values.MarginPadding7,
						Right:  values.MarginPadding2,
					}.Layout(gtx, func(gtx C) D {
						return pg.theme.Card().Layout(gtx, func(gtx C) D {
							return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
								return pg.acctDetailsPageContainer.Layout(gtx, len(widgets), func(gtx C, i int) D {
									return layout.Inset{}.Layout(gtx, widgets[i])
								})
							})
						})
					})
				})
			},
			ExtraItem: pg.renameAccount,
			Extra: func(gtx C) D {
				return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
					edit := pg.Theme.Icons.EditIcon
					return layout.E.Layout(gtx, edit.Layout24dp)
				})
			},
		}
		return sp.Layout(pg.ParentWindow(), gtx)
	}
	return components.UniformMobile(gtx, false, true, body)
}

func (pg *AcctDetailsPage) accountBalanceLayout(gtx C) D {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {

						accountIcon := pg.Theme.Icons.AccountIcon
						if pg.account.Number == load.MaxInt32 {
							accountIcon = pg.Theme.Icons.ImportedAccountIcon
						}

						m := values.MarginPadding10
						return layout.Inset{
							Right: m,
							Top:   m,
						}.Layout(gtx, accountIcon.Layout24dp)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, values.String(values.StrTotalBalance), pg.totalBalance, true)
					}),
				)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, values.String(values.StrLabelSpendable), pg.spendable, false)
			}),
			layout.Rigid(func(gtx C) D {
				if pg.stakingBalance == 0 {
					return D{}
				}

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, values.String(values.StrImmatureRewards), pg.immatureRewards, false)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, values.String(values.StrLockedByTickets), pg.lockedByTickets, false)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, values.String(values.StrVotingAuthority), pg.votingAuthority, false)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, values.String(values.StrImmatureStakeGen), pg.immatureStakeGen, false)
					}),
				)
			}),
		)
	})
}

func (pg *AcctDetailsPage) acctBalLayout(gtx C, balType string, balance string, isTotalBalance bool) D {

	marginTop := values.MarginPadding16
	marginLeft := values.MarginPadding35

	if isTotalBalance {
		marginTop = values.MarginPadding0
		marginLeft = values.MarginPadding0
	}
	return layout.Inset{
		Right: values.MarginPadding10,
		Top:   marginTop,
		Left:  marginLeft,
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				if isTotalBalance {
					return components.LayoutBalanceSize(gtx, pg.Load, balance, values.TextSize34)
				}

				return components.LayoutBalanceWithUnit(gtx, pg.Load, balance)
			}),
			layout.Rigid(func(gtx C) D {
				txt := pg.theme.Body2(balType)
				txt.Color = pg.theme.Color.GrayText2
				return txt.Layout(gtx)
			}),
		)
	})
}

func (pg *AcctDetailsPage) accountInfoLayout(gtx C) D {
	return pg.pageSections(gtx, func(gtx C) D {
		m := values.MarginPadding10
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.acctInfoLayout(gtx, values.String(values.StrAcctNum), fmt.Sprint(pg.account.Number))
			}),
			layout.Rigid(func(gtx C) D {
				inset := layout.Inset{
					Top:    m,
					Bottom: m,
				}
				return inset.Layout(gtx, func(gtx C) D {
					return pg.acctInfoLayout(gtx, values.String(values.StrHDPath), pg.hdPath)
				})
			}),
			layout.Rigid(func(gtx C) D {
				inset := layout.Inset{
					Bottom: m,
				}
				return inset.Layout(gtx, func(gtx C) D {
					return pg.acctInfoLayout(gtx, values.String(values.StrKey), pg.keys)
				})
			}),
		)
	})
}

func (pg *AcctDetailsPage) acctInfoLayout(gtx C, leftText, rightText string) D {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					leftTextLabel := pg.theme.Label(values.TextSize14, leftText)
					leftTextLabel.Color = pg.theme.Color.GrayText2
					return leftTextLabel.Layout(gtx)
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, pg.theme.Body1(rightText).Layout)
		}),
	)
}

func (pg *AcctDetailsPage) extendedPubkey(gtx C) D {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				leftTextLabel := pg.theme.Label(values.TextSize14, values.String(values.StrExtendedKey))
				leftTextLabel.Color = pg.theme.Color.GrayText2
				return leftTextLabel.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				pg.infoButton.Inset = layout.UniformInset(values.MarginPadding0)
				pg.infoButton.Size = values.MarginPadding16
				return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, pg.infoButton.Layout)
			}),
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								icon := pg.Theme.Icons.RevealIcon
								if pg.isHiddenExtendedxPubkey {
									icon = pg.Theme.Icons.ConcealIcon
								}
								return layout.Inset{
									Right: values.MarginPadding9,
								}.Layout(gtx, func(gtx C) D {
									return pg.showExtendedKeyButton.Layout(gtx, icon.Layout16dp)
								})
							}),
							layout.Rigid(func(gtx C) D {
								return layout.E.Layout(gtx, func(gtx C) D {
									lbl := pg.Theme.Label(values.TextSize14, "********")
									lbl.Color = pg.Theme.Color.GrayText1
									if !pg.isHiddenExtendedxPubkey {
										if pg.extendedKeyClickable.Clicked() {
											clipboard.WriteOp{Text: pg.extendedKey}.Add(gtx.Ops)
											pg.Toast.Notify(values.String(values.StrExtendedCopied))
										}
										lbl.Text = utils.AutoSplitSingleString(*pg.Theme, gtx, pg.extendedKey, lbl.Font, values.TextSize14)
										lbl.Color = pg.Theme.Color.Primary
										return pg.extendedKeyClickable.Layout(gtx, lbl.Layout)
									}
									return lbl.Layout(gtx)
								})
							}),
						)
					})
				})
			}),
		)
	})
}

func (pg *AcctDetailsPage) pageSections(gtx C, body layout.Widget) D {
	m := values.MarginPadding20
	mtb := values.MarginPadding5
	return layout.Inset{Left: m, Right: m, Top: mtb, Bottom: mtb}.Layout(gtx, body)
}

// HandleUserInteractions is called just before Layout() to determine
// if any user interaction recently occurred on the page and may be
// used to update the page's UI components shortly before they are
// displayed.
// Part of the load.Page interface.
func (pg *AcctDetailsPage) HandleUserInteractions() {
	if pg.renameAccount.Clicked() {
		textModal := modal.NewTextInputModal(pg.Load).
			Hint(values.String(values.StrAcctName)).
			PositiveButtonStyle(pg.Load.Theme.Color.Primary, pg.Load.Theme.Color.InvText).
			SetPositiveButtonCallback(func(newName string, tim *modal.TextInputModal) bool {
				err := pg.wallet.RenameAccount(pg.account.Number, newName)
				if err != nil {
					tim.SetError(err.Error())
					tim.SetLoading(false)
					return false
				}
				pg.account.Name = newName
				successModal := modal.NewSuccessModal(pg.Load, values.String(values.StrAcctRenamed), modal.DefaultClickFunc())
				pg.ParentWindow().ShowModal(successModal)
				return true
			})
		textModal.Title(values.String(values.StrRenameAcct)).
			SetPositiveButtonText(values.String(values.StrRename))

		pg.ParentWindow().ShowModal(textModal)
	}

	for pg.showExtendedKeyButton.Clicked() {
		if pg.extendedKey != "" {
			pg.isHiddenExtendedxPubkey = !pg.isHiddenExtendedxPubkey
		}
	}

	if pg.infoButton.Button.Clicked() {
		info := modal.NewCustomModal(pg.Load).
			Title(values.String(values.StrExtendedKey)).
			Body(values.String(values.StrExtendedInfo)).
			SetContentAlignment(layout.NW, layout.Center)
		pg.ParentWindow().ShowModal(info)
	}
}

func (pg *AcctDetailsPage) loadExtendedPubKey() {
	xpub, err := pg.WL.SelectedWallet.Wallet.GetExtendedPubKey(pg.account.Number)
	if err != nil {
		pg.Toast.NotifyError(err.Error())
	}
	pg.extendedKey = xpub
}

// OnNavigatedFrom is called when the page is about to be removed from
// the displayed window. This method should ideally be used to disable
// features that are irrelevant when the page is NOT displayed.
// NOTE: The page may be re-displayed on the app's window, in which case
// OnNavigatedTo() will be called again. This method should not destroy UI
// components unless they'll be recreated in the OnNavigatedTo() method.
// Part of the load.Page interface.
func (pg *AcctDetailsPage) OnNavigatedFrom() {}
