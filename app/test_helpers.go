package app

import (
	// "bytes"
	// "encoding/hex"
	"encoding/json"
	// "fmt"
	// "os"
	// "strconv"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	// cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	// pruningtypes "cosmossdk.io/store/pruning/types"

	// bam "github.com/cosmos/cosmos-sdk/baseapp"
	// "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	// "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

	// "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	// cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	// servertypes "github.com/cosmos/cosmos-sdk/server/types"
	// "github.com/cosmos/cosmos-sdk/testutil/mock"
	// "github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	// minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// App testing.
var DefaultConsensusParams = &cmtproto.ConsensusParams{
	Block: &cmtproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &cmtproto.ValidatorParams{
		PubKeyTypes: []string{
			cmttypes.ABCIPubKeyTypeEd25519,
		},
	},
}

func setup(withGenesis bool, invCheckPeriod uint) (*App, GenesisState) {
	db := dbm.NewMemDB()
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = invCheckPeriod

	app, _ := NewApp(log.NewNopLogger(), db, nil, true, appOptions)
	if withGenesis {
		return app, app.DefaultGenesis()
	}
	return app, GenesisState{}
}

// Setup initializes a new App. A Nop logger is set in App.
func Setup(isCheckTx bool) *App {
	app, genesisState := setup(!isCheckTx, 5)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		req := &abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		}
		app.InitChain(req)
	}

	return app
}

// SetupWithGenesisValSet initializes a new App with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in App.
func SetupWithGenesisValSet(t *testing.T, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *App {
	app, genesisState := setup(true, 5)
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdkmath.NewInt(1000000)

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(
			genAccs[0].GetAddress().String(),
			sdk.ValAddress(val.Address).String(),
			sdkmath.LegacyOneDec(),
		))
	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))...)
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	req := &abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	}
	app.InitChain(req)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})

	return app
}

// // SetupWithGenesisAccounts initializes a new App with the provided genesis
// // accounts and possible balances.
// func SetupWithGenesisAccounts(genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *App {
// 	app, genesisState := setup(true, 0)
// 	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
// 	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

// 	totalSupply := sdk.NewCoins()
// 	for _, b := range balances {
// 		totalSupply = totalSupply.Add(b.Coins...)
// 	}

// 	bankGenesis := banktypes.NewGenesisState(
// 		banktypes.DefaultGenesisState().Params,
// 		balances,
// 		totalSupply,
// 		[]banktypes.Metadata{},
// 		[]banktypes.SendEnabled{},
// 	)
// 	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

// 	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
// 	if err != nil {
// 		panic(err)
// 	}

// 	req := &abci.RequestInitChain{
// 		Validators:      []abci.ValidatorUpdate{},
// 		ConsensusParams: DefaultConsensusParams,
// 		AppStateBytes:   stateBytes,
// 	}
// 	app.InitChain(req)

// 	app.FinalizeBlock(&abci.RequestFinalizeBlock{
// 		Height:             app.LastBlockHeight() + 1,
// 		Hash:            app.LastCommitID().Hash,
// 		NextValidatorsHash: valSet.Hash(),
// 	})

// 	return app
// }

// type GenerateAccountStrategy func(int) []sdk.AccAddress

// // createRandomAccounts is a strategy used by addTestAddrs() in order to generated addresses in random order.
// func createRandomAccounts(accNum int) []sdk.AccAddress {
// 	testAddrs := make([]sdk.AccAddress, accNum)
// 	for i := 0; i < accNum; i++ {
// 		pk := ed25519.GenPrivKey().PubKey()
// 		testAddrs[i] = sdk.AccAddress(pk.Address())
// 	}

// 	return testAddrs
// }

// // createIncrementalAccounts is a strategy used by addTestAddrs() in order to generated addresses in ascending order.
// func createIncrementalAccounts(accNum int) []sdk.AccAddress {
// 	var addresses []sdk.AccAddress
// 	var buffer bytes.Buffer

// 	// start at 100 so we can make up to 999 test addresses with valid test addresses
// 	for i := 100; i < (accNum + 100); i++ {
// 		numString := strconv.Itoa(i)
// 		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") // base address string
// 		buffer.WriteString(numString)                                 // adding on final two digits to make addresses unique
// 		res, _ := sdk.AccAddressFromHex(buffer.String())
// 		bech := res.String()
// 		addr, _ := TestAddr(buffer.String(), bech)

// 		addresses = append(addresses, addr)
// 		buffer.Reset()
// 	}

// 	return addresses
// }

// // AddTestAddrsFromPubKeys adds the addresses into the App providing only the public keys.
// func AddTestAddrsFromPubKeys(app *App, ctx sdk.Context, pubKeys []cryptotypes.PubKey, accAmt sdkmath.Int) {
// 	initCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), accAmt))

// 	for _, pk := range pubKeys {
// 		initAccountWithCoins(app, ctx, sdk.AccAddress(pk.Address()), initCoins)
// 	}
// }

// // AddTestAddrs constructs and returns accNum amount of accounts with an
// // initial balance of accAmt in random order
// func AddTestAddrs(app *App, ctx sdk.Context, accNum int, accAmt sdkmath.Int) []sdk.AccAddress {
// 	return addTestAddrs(app, ctx, accNum, accAmt, createRandomAccounts)
// }

// // AddTestAddrsIncremental constructs and returns accNum amount of accounts with an
// // initial balance of accAmt in incremental order
// func AddTestAddrsIncremental(app *App, ctx sdk.Context, accNum int, accAmt sdkmath.Int) []sdk.AccAddress {
// 	return addTestAddrs(app, ctx, accNum, accAmt, createIncrementalAccounts)
// }

// func addTestAddrs(app *App, ctx sdk.Context, accNum int, accAmt sdkmath.Int, strategy GenerateAccountStrategy) []sdk.AccAddress {
// 	testAddrs := strategy(accNum)

// 	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	initCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, accAmt))

// 	for _, addr := range testAddrs {
// 		initAccountWithCoins(app, ctx, addr, initCoins)
// 	}

// 	return testAddrs
// }

// func initAccountWithCoins(app *App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
// 	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// // ConvertAddrsToValAddrs converts the provided addresses to ValAddress.
// func ConvertAddrsToValAddrs(addrs []sdk.AccAddress) []sdk.ValAddress {
// 	valAddrs := make([]sdk.ValAddress, len(addrs))

// 	for i, addr := range addrs {
// 		valAddrs[i] = sdk.ValAddress(addr)
// 	}

// 	return valAddrs
// }

// func TestAddr(addr string, bech string) (sdk.AccAddress, error) {
// 	res, err := sdk.AccAddressFromBech32(addr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	bechexpected := res.String()
// 	if bech != bechexpected {
// 		return nil, fmt.Errorf("bech encoding doesn't match reference")
// 	}

// 	bechres, err := sdk.AccAddressFromBech32(bech)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !bytes.Equal(bechres, res) {
// 		return nil, err
// 	}

// 	return res, nil
// }

// // CheckBalance checks the balance of an account.
// func CheckBalance(t *testing.T, app *App, addr sdk.AccAddress, balances sdk.Coins) {
// 	ctxCheck := app.BaseApp.NewContext(true, cmtproto.Header{})
// 	require.True(t, balances.IsEqual(app.BankKeeper.GetAllBalances(ctxCheck, addr)))
// }

// // SignCheckDeliver checks a generated signed transaction and simulates a
// // block commitment with the given transaction. A test assertion is made using
// // the parameter 'expPass' against the result. A corresponding result is
// // returned.
// func SignCheckDeliver(
// 	t *testing.T, txCfg client.TxConfig, app *bam.BaseApp, header cmtproto.Header, msgs []sdk.Msg,
// 	chainID string, accNums, accSeqs []uint64, expSimPass, expPass bool, priv ...cryptotypes.PrivKey,
// ) (sdk.GasInfo, *sdk.Result, error) {

// 	tx, err := staking.GenTx(
// 		txCfg,
// 		msgs,
// 		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
// 		staking.DefaultGenTxGas,
// 		chainID,
// 		accNums,
// 		accSeqs,
// 		priv...,
// 	)
// 	require.NoError(t, err)
// 	txBytes, err := txCfg.TxEncoder()(tx)
// 	require.Nil(t, err)

// 	// Must simulate now as CheckTx doesn't run Msgs anymore
// 	_, res, err := app.Simulate(txBytes)

// 	if expSimPass {
// 		require.NoError(t, err)
// 		require.NotNil(t, res)
// 	} else {
// 		require.Error(t, err)
// 		require.Nil(t, res)
// 	}

// 	// Simulate a sending a transaction and committing a block
// 	app.BeginBlock(&abci.RequestBeginBlock{Header: header})
// 	gInfo, res, err := app.Deliver(txCfg.TxEncoder(), tx)

// 	if expPass {
// 		require.NoError(t, err)
// 		require.NotNil(t, res)
// 	} else {
// 		require.Error(t, err)
// 		require.Nil(t, res)
// 	}

// 	app.EndBlock(&abci.RequestEndBlock{})
// 	app.Commit()

// 	return gInfo, res, err
// }

// // GenSequenceOfTxs generates a set of signed transactions of messages, such
// // that they differ only by having the sequence numbers incremented between
// // every transaction.
// func GenSequenceOfTxs(txGen client.TxConfig, msgs []sdk.Msg, accNums []uint64, initSeqNums []uint64, numToGenerate int, priv ...cryptotypes.PrivKey) ([]sdk.Tx, error) {
// 	txs := make([]sdk.Tx, numToGenerate)
// 	var err error
// 	for i := 0; i < numToGenerate; i++ {
// 		txs[i], err = staking.GenTx(
// 			txGen,
// 			msgs,
// 			sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
// 			staking.DefaultGenTxGas,
// 			"",
// 			accNums,
// 			initSeqNums,
// 			priv...,
// 		)
// 		if err != nil {
// 			break
// 		}
// 		incrementAllSequenceNumbers(initSeqNums)
// 	}

// 	return txs, err
// }

// func incrementAllSequenceNumbers(initSeqNums []uint64) {
// 	for i := 0; i < len(initSeqNums); i++ {
// 		initSeqNums[i]++
// 	}
// }

// // CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
// func CreateTestPubKeys(numPubKeys int) []cryptotypes.PubKey {
// 	var publicKeys []cryptotypes.PubKey
// 	var buffer bytes.Buffer

// 	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
// 	for i := 100; i < (numPubKeys + 100); i++ {
// 		numString := strconv.Itoa(i)
// 		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
// 		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
// 		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
// 		buffer.Reset()
// 	}

// 	return publicKeys
// }

// // NewPubKeyFromHex returns a PubKey from a hex string.
// func NewPubKeyFromHex(pk string) (res cryptotypes.PubKey) {
// 	pkBytes, err := hex.DecodeString(pk)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if len(pkBytes) != ed25519.PubKeySize {
// 		panic(fmt.Errorf("invalid pubkey size"))
// 	}
// 	return &ed25519.PubKey{Key: pkBytes}
// }

// // EmptyAppOptions is a stub implementing AppOptions
// type EmptyAppOptions struct{}

// // Get implements AppOptions
// func (ao EmptyAppOptions) Get(o string) interface{} {
// 	return nil
// }

// // FundAccount is a utility function that funds an account by minting and
// // sending the coins to the address. This should be used for testing purposes
// // only!
// //
// // TODO: Instead of using the mint module account, which has the
// // permission of minting, create a "faucet" account. (@fdymylja)
// func FundAccount(bankKeeper bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
// 	if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
// 		return err
// 	}

// 	return bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
// }

// // FundModuleAccount is a utility function that funds a module account by
// // minting and sending the coins to the address. This should be used for testing
// // purposes only!
// //
// // TODO: Instead of using the mint module account, which has the
// // permission of minting, create a "faucet" account. (@fdymylja)
// func FundModuleAccount(bankKeeper bankkeeper.Keeper, ctx sdk.Context, recipientMod string, amounts sdk.Coins) error {
// 	if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
// 		return err
// 	}

// 	return bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, recipientMod, amounts)
// }
