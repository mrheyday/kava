package cdp_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper cdp.Keeper
}

func (suite *GenesisTestSuite) TestInvalidGenState() {
	type args struct {
		params             cdp.Params
		cdps               cdp.CDPs
		deposits           cdp.Deposits
		startingID         uint64
		debtDenom          string
		govDenom           string
		genAccumTimes      cdp.GenesisAccumulationTimes
		genTotalPrincipals cdp.GenesisTotalPrincipals
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type genesisTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			name: "empty debt denom",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          "",
				govDenom:           cdp.DefaultGovDenom,
				genAccumTimes:      cdp.DefaultGenesisState().PreviousAccumulationTimes,
				genTotalPrincipals: cdp.DefaultGenesisState().TotalPrincipals,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "debt denom invalid",
			},
		},
		{
			name: "empty gov denom",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          cdp.DefaultDebtDenom,
				govDenom:           "",
				genAccumTimes:      cdp.DefaultGenesisState().PreviousAccumulationTimes,
				genTotalPrincipals: cdp.DefaultGenesisState().TotalPrincipals,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "gov denom invalid",
			},
		},
		{
			name: "interest factor below one",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          cdp.DefaultDebtDenom,
				govDenom:           cdp.DefaultGovDenom,
				genAccumTimes:      cdp.GenesisAccumulationTimes{cdp.NewGenesisAccumulationTime("bnb-a", time.Time{}, sdk.OneDec().Sub(sdk.SmallestDec()))},
				genTotalPrincipals: cdp.DefaultGenesisState().TotalPrincipals,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "interest factor should be ≥ 1.0",
			},
		},
		{
			name: "negative total principal",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          cdp.DefaultDebtDenom,
				govDenom:           cdp.DefaultGovDenom,
				genAccumTimes:      cdp.DefaultGenesisState().PreviousAccumulationTimes,
				genTotalPrincipals: cdp.GenesisTotalPrincipals{cdp.NewGenesisTotalPrincipal("bnb-a", sdk.NewInt(-1))},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "total principal should be positive",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			gs := cdp.NewGenesisState(tc.args.params, tc.args.cdps, tc.args.deposits, tc.args.startingID,
				tc.args.debtDenom, tc.args.govDenom, tc.args.genAccumTimes, tc.args.genTotalPrincipals)
			err := gs.Validate()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *GenesisTestSuite) TestValidGenState() {
	tApp := app.NewTestApp()

	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
			NewCDPGenStateMulti(),
		)
	})

	cdpGS := NewCDPGenStateMulti()
	gs := cdp.GenesisState{}
	cdp.ModuleCdc.UnmarshalJSON(cdpGS["cdp"], &gs)
	gs.CDPs = cdps()
	gs.StartingCdpID = uint64(5)
	appGS := app.GenesisState{"cdp": cdp.ModuleCdc.MustMarshalJSON(gs)}
	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
			appGS,
		)
	})
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
