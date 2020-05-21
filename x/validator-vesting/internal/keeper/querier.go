package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/validator-vesting/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryCirculatingSupply:
			return queryGetCirculatingSupply(ctx, req, keeper)
		case types.QueryTotalSupply:
			return queryGetTotalSupply(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

func queryGetTotalSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("ukava")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := keeper.cdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryGetCirculatingSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	supplyInt := sdk.NewInt(27190672)
	bz, err := keeper.cdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}
