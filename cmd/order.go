package cmd

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
)

// TODO: order checking status, expired as function on this?
// TODO: Calc Spread from order list of the height of a pair
// TODO: Distance, midPrice, both side, min
// testcode with input orderlist
// GET C_mt, summation of C_mt for share
// Spread, Distance, Width
// SET params as json file

type Result struct {
	MidPrice sdk.Dec
	Spread   sdk.Dec

	AskWidth    sdk.Dec
	BidWidth    sdk.Dec
	AskQuantity sdk.Int
	BidQuantity sdk.Int

	AskMaxPrice sdk.Dec
	AskMinPrice sdk.Dec
	BidMaxPrice sdk.Dec
	BidMinPrice sdk.Dec

	BidCount           int
	AskCount           int
	RemCount           int
	InvalidStatusCount int
	TotalCount         int
}

func (r Result) String() (str string) {
	s := reflect.ValueOf(&r).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		lineStr := fmt.Sprintf("%s %s = %v\n",
			typeOfT.Field(i).Name, f.Type(), f.Interface())
		str = str + lineStr
	}
	return
}

func GetResult(orders []liquiditytypes.Order) (r Result) {
	r = Result{
		MidPrice:    sdk.ZeroDec(),
		AskWidth:    sdk.ZeroDec(),
		BidWidth:    sdk.ZeroDec(),
		AskQuantity: sdk.ZeroInt(),
		BidQuantity: sdk.ZeroInt(),
		AskMaxPrice: sdk.ZeroDec(),
		AskMinPrice: sdk.ZeroDec(),
		BidMaxPrice: sdk.ZeroDec(),
		BidMinPrice: sdk.ZeroDec(),

		// info field
		Spread:             sdk.ZeroDec(),
		BidCount:           0,
		AskCount:           0,
		RemCount:           0,
		InvalidStatusCount: 0,
		TotalCount:         0,
	}
	for _, order := range orders {
		r.TotalCount += 1

		// skip orders which has not available status
		if order.Status != liquiditytypes.OrderStatusNotExecuted &&
			order.Status != liquiditytypes.OrderStatusNotMatched &&
			order.Status != liquiditytypes.OrderStatusPartiallyMatched {
			r.InvalidStatusCount += 1
			continue
		}
		// skip orders which is not over RemnantThreshold
		if order.RemainingOfferCoin.Amount.ToDec().Quo(order.OfferCoin.Amount.ToDec()).LTE(RemnantThresholdTmp) {
			r.RemCount += 1
			continue
		}
		if order.Direction == liquiditytypes.OrderDirectionBuy { // BID
			r.BidCount += 1
			r.BidQuantity = r.BidQuantity.Add(order.RemainingOfferCoin.Amount)
			if order.Price.GTE(r.BidMaxPrice) {
				r.BidMaxPrice = order.Price
			}
			if r.BidMinPrice.IsZero() || order.Price.LTE(r.BidMinPrice) {
				r.BidMinPrice = order.Price
			}
		} else if order.Direction == liquiditytypes.OrderDirectionSell { // ASK
			r.AskCount += 1
			r.AskQuantity = r.AskQuantity.Add(order.RemainingOfferCoin.Amount)
			if order.Price.GTE(r.AskMaxPrice) {
				r.AskMaxPrice = order.Price
			}
			if r.BidMinPrice.IsZero() || order.Price.LTE(r.AskMinPrice) {
				r.AskMinPrice = order.Price
			}
		}
	}
	// calc mid price, (BidMaxPrice + AskMinPrice)/2
	r.MidPrice = r.BidMaxPrice.Add(r.AskMinPrice).Quo(sdk.NewDec(2))
	r.Spread = r.AskMinPrice.Sub(r.BidMaxPrice).Quo(r.MidPrice)
	r.AskWidth = r.AskMaxPrice.Sub(r.AskMinPrice).Quo(r.MidPrice)
	r.BidWidth = r.BidMaxPrice.Sub(r.BidMinPrice).Quo(r.MidPrice)
	return
	// TODO: checking order tick cap validity
}
