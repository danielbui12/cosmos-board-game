package keeper_test

import (
	"testing"

	math "cosmossdk.io/math"
	checkersapp "github.com/alice/checkers/app"
	"github.com/alice/checkers/x/checkers/keeper"
	"github.com/alice/checkers/x/checkers/testutil"
	"github.com/alice/checkers/x/checkers/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
)

const (
	alice = testutil.Alice
	bob   = testutil.Bob
	carol = testutil.Carol
)
const (
	balAlice = 50000000
	balBob   = 20000000
	balCarol = 10000000
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *checkersapp.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

var (
	checkersModuleAddress string
)

func TestCheckersKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := checkersapp.Setup(false)
	ctx := app.BaseApp.NewContext(false)

	app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	checkersModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.CheckersKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.CheckersKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func makeBalance(address string, balance int64, denom string) banktypes.Balance {
	return banktypes.Balance{
		Address: address,
		Coins: sdk.Coins{
			sdk.Coin{
				Denom:  denom,
				Amount: math.NewInt(balance),
			},
		},
	}
}

func addAll(balances []banktypes.Balance) sdk.Coins {
	total := sdk.NewCoins()
	for _, balance := range balances {
		total = total.Add(balance.Coins...)
	}
	return total
}

func getBankGenesis() *banktypes.GenesisState {
	coins := []banktypes.Balance{
		makeBalance(alice, balAlice, "stake"),
		makeBalance(bob, balBob, "stake"),
		makeBalance(bob, balBob, "coin"),
		makeBalance(carol, balCarol, "stake"),
		makeBalance(carol, balCarol, "coin"),
	}
	supply := banktypes.Supply{
		Total: addAll(coins),
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		coins,
		supply.Total,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)

	return bankGenesis
}

func (suite *IntegrationTestSuite) setupSuiteWithBalances() {
	suite.app.BankKeeper.InitGenesis(suite.ctx, getBankGenesis())
}

func (suite *IntegrationTestSuite) RequireBankBalance(expected int, atAddress string) {
	suite.RequireBankBalanceWithDenom(expected, "stake", atAddress)
}

func (suite *IntegrationTestSuite) RequireBankBalanceWithDenom(expected int,  denom string, atAddress string) {
	sdkAdd, err := sdk.AccAddressFromBech32(atAddress)
	suite.Require().Nil(err, "Failed to parse address: %s", atAddress)
	suite.Require().Equal(
		int64(expected),
		suite.app.BankKeeper.GetBalance(suite.ctx, sdkAdd, denom).Amount.Int64())
}
