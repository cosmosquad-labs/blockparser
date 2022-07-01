package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
)

func GetResult(orders []liquiditytypes.Order) (r Result) {
	r = Result{
		MidPrice:    sdk.ZeroDec(),
		Spread:      sdk.ZeroDec(),
		AskWidth:    sdk.ZeroDec(),
		BidWidth:    sdk.ZeroDec(),
		AskQuantity: sdk.ZeroInt(),
		BidQuantity: sdk.ZeroInt(),
		AskMinPrice: sdk.ZeroDec(),
		AskMaxPrice: sdk.ZeroDec(),
		BidMinPrice: sdk.ZeroDec(),
		BidMaxPrice: sdk.ZeroDec(),
	}
	for _, order := range orders {
		// TODO: need to verify
		if !order.RemainingOfferCoin.Amount.ToDec().Quo(order.OfferCoin.Amount.ToDec()).GTE(RemnantThresholdTmp) {
			continue
		}
		if order.Direction == liquiditytypes.OrderDirectionBuy { // BID
			if order.Price.GTE(r.BidMaxPrice) {
				r.BidMaxPrice = order.Price
			}
			if order.Price.LTE(r.BidMinPrice) {
				r.BidMinPrice = order.Price
			}
		} else if order.Direction == liquiditytypes.OrderDirectionSell { // ASK
			if order.Price.GTE(r.AskMaxPrice) {
				r.AskMaxPrice = order.Price
			}
			if order.Price.LTE(r.AskMinPrice) {
				r.AskMinPrice = order.Price
			}
		}

	}
	return
	// TODO: checking order tick cap validity
}
