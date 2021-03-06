package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lcnem/jpyx/app"
	"github.com/lcnem/jpyx/x/auction"
	"github.com/lcnem/jpyx/x/cdp/keeper"
	"github.com/lcnem/jpyx/x/cdp/types"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AuctionTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *AuctionTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	taddr := sdk.AccAddress(crypto.AddressHash([]byte("StakeTestUser1")))
	authGS := app.NewAuthGenState([]sdk.AccAddress{taddr}, []sdk.Coins{cs(c("jpyx", 21000000000))})
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = []sdk.AccAddress{taddr}
	return
}

func (suite *AuctionTestSuite) TestNetDebtSurplus() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100)))
	suite.NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 10)))
	suite.NoError(err)
	suite.NotPanics(func() { suite.keeper.NetSurplusAndDebt(suite.ctx) })
	acc := sk.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("debt", 90)), acc.GetCoins())
}

func (suite *AuctionTestSuite) TestCollateralAuction() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 21000000000), c("bnb", 190000000000)))
	suite.Require().NoError(err)
	testDeposit := types.NewDeposit(1, suite.addrs[0], c("bnb", 190000000000))
	err = suite.keeper.AuctionCollateral(suite.ctx, types.Deposits{testDeposit}, i(21000000000), "jpyx")
	suite.Require().NoError(err)
}

func (suite *AuctionTestSuite) TestSurplusAuction() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 600000000000)))
	suite.NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100000000000)))
	suite.NoError(err)
	suite.keeper.RunSurplusAndDebtAuctions(suite.ctx)
	acc := sk.GetModuleAccount(suite.ctx, auction.ModuleName)
	suite.Equal(cs(c("jpyx", 10000000000)), acc.GetCoins())
	acc = sk.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("jpyx", 490000000000)), acc.GetCoins())
}

func (suite *AuctionTestSuite) TestDebtAuction() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 100000000000)))
	suite.NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 200000000000)))
	suite.NoError(err)
	suite.keeper.RunSurplusAndDebtAuctions(suite.ctx)
	acc := sk.GetModuleAccount(suite.ctx, auction.ModuleName)
	suite.Equal(cs(c("debt", 10000000000)), acc.GetCoins())
	acc = sk.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("debt", 90000000000)), acc.GetCoins())
}

func (suite *AuctionTestSuite) TestGetTotalSurplus() {
	sk := suite.app.GetSupplyKeeper()

	// liquidator account has zero coins
	suite.Require().Equal(sdk.NewInt(0), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// mint some coins
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 100e6)))
	suite.Require().NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 200e6)))
	suite.Require().NoError(err)

	// liquidator account has 300e6 total jpyx
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// mint some debt
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 500e6)))
	suite.Require().NoError(err)

	// liquidator account still has 300e6 total jpyx -- debt balance is ignored
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// burn some jpyx
	err = sk.BurnCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 50e6)))
	suite.Require().NoError(err)

	// liquidator jpyx decreases
	suite.Require().Equal(sdk.NewInt(250e6), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))
}

func (suite *AuctionTestSuite) TestGetTotalDebt() {
	sk := suite.app.GetSupplyKeeper()

	// liquidator account has zero debt
	suite.Require().Equal(sdk.NewInt(0), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// mint some debt
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100e6)))
	suite.Require().NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 200e6)))
	suite.Require().NoError(err)

	// liquidator account has 300e6 total debt
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalDebt(suite.ctx, types.LiquidatorMacc))

	// mint some jpyx
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("jpyx", 500e6)))
	suite.Require().NoError(err)

	// liquidator account still has 300e6 total debt -- jpyx balance is ignored
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalDebt(suite.ctx, types.LiquidatorMacc))

	// burn some debt
	err = sk.BurnCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 50e6)))
	suite.Require().NoError(err)

	// liquidator debt decreases
	suite.Require().Equal(sdk.NewInt(250e6), suite.keeper.GetTotalDebt(suite.ctx, types.LiquidatorMacc))
}

func TestAuctionTestSuite(t *testing.T) {
	suite.Run(t, new(AuctionTestSuite))
}
