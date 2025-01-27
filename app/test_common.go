package app

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmdb "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/issuance"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/x/swap"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

var (
	emptyTime    time.Time
	emptyChainID string
)

// TestApp is a simple wrapper around an App. It exposes internal keepers for use in integration tests.
// This file also contains test helpers. Ideally they would be in separate package.
// Basic Usage:
// 	Create a test app with NewTestApp, then all keepers and their methods can be accessed for test setup and execution.
// Advanced Usage:
// 	Some tests call for an app to be initialized with some state. This can be achieved through keeper method calls (ie keeper.SetParams(...)).
// 	However this leads to a lot of duplicated logic similar to InitGenesis methods.
// 	So TestApp.InitializeFromGenesisStates() will call InitGenesis with the default genesis state.
//	and TestApp.InitializeFromGenesisStates(authState, cdpState) will do the same but overwrite the auth and cdp sections of the default genesis state
// 	Creating the genesis states can be combersome, but helper methods can make it easier such as NewAuthGenStateFromAccounts below.
type TestApp struct {
	App
}

func NewTestApp() TestApp {
	config := sdk.GetConfig()
	SetBech32AddressPrefixes(config)
	SetBip44CoinType(config)

	db := tmdb.NewMemDB()
	app := NewApp(log.NewNopLogger(), db, nil, AppOptions{})
	return TestApp{App: *app}
}

// nolint
func (tApp TestApp) GetAccountKeeper() auth.AccountKeeper { return tApp.accountKeeper }
func (tApp TestApp) GetBankKeeper() bank.Keeper           { return tApp.bankKeeper }
func (tApp TestApp) GetSupplyKeeper() supply.Keeper       { return tApp.supplyKeeper }
func (tApp TestApp) GetStakingKeeper() staking.Keeper     { return tApp.stakingKeeper }
func (tApp TestApp) GetSlashingKeeper() slashing.Keeper   { return tApp.slashingKeeper }
func (tApp TestApp) GetMintKeeper() mint.Keeper           { return tApp.mintKeeper }
func (tApp TestApp) GetDistrKeeper() distribution.Keeper  { return tApp.distrKeeper }
func (tApp TestApp) GetGovKeeper() gov.Keeper             { return tApp.govKeeper }
func (tApp TestApp) GetCrisisKeeper() crisis.Keeper       { return tApp.crisisKeeper }
func (tApp TestApp) GetUpgradeKeeper() upgrade.Keeper     { return tApp.upgradeKeeper }
func (tApp TestApp) GetParamsKeeper() params.Keeper       { return tApp.paramsKeeper }
func (tApp TestApp) GetVVKeeper() validatorvesting.Keeper { return tApp.vvKeeper }
func (tApp TestApp) GetAuctionKeeper() auction.Keeper     { return tApp.auctionKeeper }
func (tApp TestApp) GetCDPKeeper() cdp.Keeper             { return tApp.cdpKeeper }
func (tApp TestApp) GetPriceFeedKeeper() pricefeed.Keeper { return tApp.pricefeedKeeper }
func (tApp TestApp) GetBep3Keeper() bep3.Keeper           { return tApp.bep3Keeper }
func (tApp TestApp) GetKavadistKeeper() kavadist.Keeper   { return tApp.kavadistKeeper }
func (tApp TestApp) GetIncentiveKeeper() incentive.Keeper { return tApp.incentiveKeeper }
func (tApp TestApp) GetHardKeeper() hard.Keeper           { return tApp.hardKeeper }
func (tApp TestApp) GetCommitteeKeeper() committee.Keeper { return tApp.committeeKeeper }
func (tApp TestApp) GetIssuanceKeeper() issuance.Keeper   { return tApp.issuanceKeeper }
func (tApp TestApp) GetSwapKeeper() swap.Keeper           { return tApp.swapKeeper }

// InitializeFromGenesisStates calls InitChain on the app using the default genesis state, overwitten with any passed in genesis states
func (tApp TestApp) InitializeFromGenesisStates(genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainID(emptyTime, emptyChainID, genesisStates...)
}

// InitializeFromGenesisStatesWithTime calls InitChain on the app using the default genesis state, overwitten with any passed in genesis states and genesis Time
func (tApp TestApp) InitializeFromGenesisStatesWithTime(genTime time.Time, genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainID(genTime, emptyChainID, genesisStates...)
}

// InitializeFromGenesisStatesWithTimeAndChainID calls InitChain on the app using the default genesis state, overwitten with any passed in genesis states and genesis Time
func (tApp TestApp) InitializeFromGenesisStatesWithTimeAndChainID(genTime time.Time, chainID string, genesisStates ...GenesisState) TestApp {
	// Create a default genesis state and overwrite with provided values
	genesisState := NewDefaultGenesisState()
	for _, state := range genesisStates {
		for k, v := range state {
			genesisState[k] = v
		}
	}

	// Initialize the chain
	stateBytes, err := codec.MarshalJSONIndent(tApp.cdc, genesisState)
	if err != nil {
		panic(err)
	}
	tApp.InitChain(
		abci.RequestInitChain{
			Time:          genTime,
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
			ChainId:       chainID,
		},
	)
	tApp.Commit()
	tApp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: tApp.LastBlockHeight() + 1, Time: genTime}})
	return tApp
}

func (tApp TestApp) CheckBalance(t *testing.T, ctx sdk.Context, owner sdk.AccAddress, expectedCoins sdk.Coins) {
	acc := tApp.GetAccountKeeper().GetAccount(ctx, owner)
	require.NotNilf(t, acc, "account with address '%s' doesn't exist", owner)
	require.Equal(t, expectedCoins, acc.GetCoins())
}

// GeneratePrivKeyAddressPairsFromRand generates (deterministically) a total of n secp256k1 private keys and addresses.
func GeneratePrivKeyAddressPairs(n int) (keys []crypto.PrivKey, addrs []sdk.AccAddress) {
	r := rand.New(rand.NewSource(12345)) // make the generation deterministic
	keys = make([]crypto.PrivKey, n)
	addrs = make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		secret := make([]byte, 32)
		_, err := r.Read(secret)
		if err != nil {
			panic("Could not read randomness")
		}
		keys[i] = secp256k1.GenPrivKeySecp256k1(secret)
		addrs[i] = sdk.AccAddress(keys[i].PubKey().Address())
	}
	return
}

// Create a new auth genesis state from some addresses and coins. The state is returned marshalled into a map.
func NewAuthGenState(addresses []sdk.AccAddress, coins []sdk.Coins) GenesisState {
	// Create GenAccounts
	accounts := authexported.GenesisAccounts{}
	for i := range addresses {
		accounts = append(accounts, auth.NewBaseAccount(addresses[i], coins[i], nil, 0, 0))
	}
	// Create the auth genesis state
	authGenesis := auth.NewGenesisState(auth.DefaultParams(), accounts)
	return GenesisState{auth.ModuleName: auth.ModuleCdc.MustMarshalJSON(authGenesis)}
}

// AuthGenesisBuilder is a tool for creating an auth genesis state.
// Helper methods create basic accounts types and add them to a default genesis state.
// All methods are immutable and return updated copies of the builder.
// The builder inherits from auth.GenesisState, so fields can be accessed directly if a helper method doesn't exist.
//
// Example:
//     // create a single account genesis state
//     builder := NewAuthGenesisBuilder().WithSimpleAccount(testUserAddress, testCoins)
//     genesisState := builder.Build()
//
type AuthGenesisBuilder struct {
	auth.GenesisState
}

// NewAuthGenesisBuilder creates a AuthGenesisBuilder containing a default genesis state.
func NewAuthGenesisBuilder() AuthGenesisBuilder {
	return AuthGenesisBuilder{
		GenesisState: auth.DefaultGenesisState(),
	}
}

// Build assembles and returns the final GenesisState
func (builder AuthGenesisBuilder) Build() auth.GenesisState {
	return builder.GenesisState
}

// BuildMarshalled assembles the final GenesisState and json encodes it into a generic genesis type.
func (builder AuthGenesisBuilder) BuildMarshalled() GenesisState {
	return GenesisState{
		auth.ModuleName: auth.ModuleCdc.MustMarshalJSON(builder.Build()),
	}
}

// WithAccounts adds accounts of any type to the genesis state.
func (builder AuthGenesisBuilder) WithAccounts(account ...authexported.GenesisAccount) AuthGenesisBuilder {
	builder.Accounts = append(builder.Accounts, account...)
	return builder
}

// WithSimpleAccount adds a standard account to the genesis state.
func (builder AuthGenesisBuilder) WithSimpleAccount(address sdk.AccAddress, balance sdk.Coins) AuthGenesisBuilder {
	return builder.WithAccounts(auth.NewBaseAccount(address, balance, nil, 0, 0))
}

// WithSimpleModuleAccount adds a module account to the genesis state.
func (builder AuthGenesisBuilder) WithSimpleModuleAccount(moduleName string, balance sdk.Coins, permissions ...string) AuthGenesisBuilder {
	account := supply.NewEmptyModuleAccount(moduleName, permissions...)
	account.SetCoins(balance)
	return builder.WithAccounts(account)
}

// WithSimplePeriodicVestingAccount adds a periodic vesting account to the genesis state.
func (builder AuthGenesisBuilder) WithSimplePeriodicVestingAccount(address sdk.AccAddress, balance sdk.Coins, periods vesting.Periods, firstPeriodStartTimestamp int64) AuthGenesisBuilder {
	baseAccount := auth.NewBaseAccount(address, balance, nil, 0, 0)

	originalVesting := sdk.NewCoins()
	for _, p := range periods {
		originalVesting = originalVesting.Add(p.Amount...)
	}

	var totalPeriods int64
	for _, p := range periods {
		totalPeriods += p.Length
	}
	endTime := firstPeriodStartTimestamp + totalPeriods

	baseVestingAccount, err := vesting.NewBaseVestingAccount(baseAccount, originalVesting, endTime)
	if err != nil {
		panic(err.Error())
	}
	periodicVestingAccount := vesting.NewPeriodicVestingAccountRaw(baseVestingAccount, firstPeriodStartTimestamp, periods)

	return builder.WithAccounts(periodicVestingAccount)
}

// WithEmptyValidatorVestingAccount adds a stub validator vesting account to the genesis state.
func (builder AuthGenesisBuilder) WithEmptyValidatorVestingAccount(address sdk.AccAddress) AuthGenesisBuilder {
	// TODO create a validator vesting account builder and remove this method
	bacc := auth.NewBaseAccount(address, nil, nil, 0, 0)
	bva, err := vesting.NewBaseVestingAccount(bacc, nil, 1)
	if err != nil {
		panic(err.Error())
	}
	account := validatorvesting.NewValidatorVestingAccountRaw(bva, 0, nil, sdk.ConsAddress{}, nil, 90)
	return builder.WithAccounts(account)
}
