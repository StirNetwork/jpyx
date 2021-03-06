package cdp_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/lcnem/jpyx/app"
	"github.com/lcnem/jpyx/x/cdp"
	"github.com/lcnem/jpyx/x/pricefeed"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func NewPricefeedGenState(asset string, price sdk.Dec) app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: asset + ":jpy", BaseAsset: asset, QuoteAsset: "jpy", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			{
				MarketID:      asset + ":jpy",
				OracleAddress: sdk.AccAddress{},
				Price:         price,
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func NewCDPGenState(asset string, liquidationRatio sdk.Dec) app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewInt64Coin("jpyx", 1000000000000),
			SurplusAuctionThreshold:      cdp.DefaultSurplusThreshold,
			SurplusAuctionLot:            cdp.DefaultSurplusLot,
			DebtAuctionThreshold:         cdp.DefaultDebtThreshold,
			DebtAuctionLot:               cdp.DefaultDebtLot,
			SavingsDistributionFrequency: cdp.DefaultSavingsDistributionFrequency,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:               asset,
					LiquidationRatio:    liquidationRatio,
					DebtLimit:           sdk.NewInt64Coin("jpyx", 1000000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(1000000000),
					Prefix:              0x20,
					ConversionFactor:    i(6),
					SpotMarketID:        asset + ":jpy",
					LiquidationMarketID: asset + ":jpy",
				},
			},
			DebtParam: cdp.DebtParam{
				Denom:            "jpyx",
				ReferenceAsset:   "jpy",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
				SavingsRate:      d("0.95"),
			},
		},
		StartingCdpID:            cdp.DefaultCdpStartingID,
		DebtDenom:                cdp.DefaultDebtDenom,
		GovDenom:                 cdp.DefaultGovDenom,
		CDPs:                     cdp.CDPs{},
		PreviousDistributionTime: cdp.DefaultPreviousDistributionTime,
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: "btc:jpy", BaseAsset: "btc", QuoteAsset: "jpy", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:jpy", BaseAsset: "xrp", QuoteAsset: "jpy", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			{
				MarketID:      "btc:jpy",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:jpy",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}
func NewCDPGenStateMulti() app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewInt64Coin("jpyx", 1000000000000),
			SurplusAuctionThreshold:      cdp.DefaultSurplusThreshold,
			SurplusAuctionLot:            cdp.DefaultSurplusLot,
			DebtAuctionThreshold:         cdp.DefaultDebtThreshold,
			DebtAuctionLot:               cdp.DefaultDebtLot,
			SavingsDistributionFrequency: cdp.DefaultSavingsDistributionFrequency,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:               "xrp",
					LiquidationRatio:    sdk.MustNewDecFromStr("2.0"),
					DebtLimit:           sdk.NewInt64Coin("jpyx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(7000000000),
					Prefix:              0x20,
					SpotMarketID:        "xrp:jpy",
					LiquidationMarketID: "xrp:jpy",
					ConversionFactor:    i(6),
				},
				{
					Denom:               "btc",
					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
					DebtLimit:           sdk.NewInt64Coin("jpyx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000000782997609"), // %2.5 apr
					LiquidationPenalty:  d("0.025"),
					AuctionSize:         i(10000000),
					Prefix:              0x21,
					SpotMarketID:        "btc:jpy",
					LiquidationMarketID: "btc:jpy",
					ConversionFactor:    i(8),
				},
			},
			DebtParam: cdp.DebtParam{
				Denom:            "jpyx",
				ReferenceAsset:   "jpy",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
				SavingsRate:      d("0.95"),
			},
		},
		StartingCdpID:            cdp.DefaultCdpStartingID,
		DebtDenom:                cdp.DefaultDebtDenom,
		GovDenom:                 cdp.DefaultGovDenom,
		CDPs:                     cdp.CDPs{},
		PreviousDistributionTime: cdp.DefaultPreviousDistributionTime,
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func cdps() (cdps cdp.CDPs) {
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	c1 := cdp.NewCDP(uint64(1), addrs[0], sdk.NewCoin("xrp", sdk.NewInt(100000000)), sdk.NewCoin("jpyx", sdk.NewInt(8000000)), tmtime.Canonical(time.Now()))
	c2 := cdp.NewCDP(uint64(2), addrs[1], sdk.NewCoin("xrp", sdk.NewInt(100000000)), sdk.NewCoin("jpyx", sdk.NewInt(10000000)), tmtime.Canonical(time.Now()))
	c3 := cdp.NewCDP(uint64(3), addrs[1], sdk.NewCoin("btc", sdk.NewInt(1000000000)), sdk.NewCoin("jpyx", sdk.NewInt(10000000)), tmtime.Canonical(time.Now()))
	c4 := cdp.NewCDP(uint64(4), addrs[2], sdk.NewCoin("xrp", sdk.NewInt(1000000000)), sdk.NewCoin("jpyx", sdk.NewInt(50000000)), tmtime.Canonical(time.Now()))
	cdps = append(cdps, c1, c2, c3, c4)
	return
}
