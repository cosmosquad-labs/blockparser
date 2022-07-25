package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	crecmd "github.com/crescent-network/crescent/v2/cmd/crescentd/cmd"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb/opt"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"

	"github.com/crescent-network/crescent/v2/app"
)

func NewBlockParserCmd() *cobra.Command {
	crecmd.GetConfig()
	cmd := &cobra.Command{
		Use:  "blockparser [chain-dir] [start-height] [end-height]",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			startHeight, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("parse start-Height: %w", err)
			}

			endHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("parse end-Height: %w", err)
			}
			return Main(dir, startHeight, endHeight)
		},
	}
	return cmd
}

func Main(dir string, startHeight, endHeight int64) error {
	blockDB, err := sdk.NewLevelDB("data/blockstore", dir)
	if err != nil {
		panic(err)
	}
	defer blockDB.Close()

	stateDB, err := sdk.NewLevelDB("data/state", dir)
	if err != nil {
		panic(err)
	}
	defer stateDB.Close()

	db, err := sdk.NewLevelDB("data/application", dir)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	blockStore := store.NewBlockStore(blockDB)
	//stateStore := state.NewStore(stateDB)
	//txStore := kv.NewTxIndex(txDB)

	fmt.Println("version :", "v0.1.0")
	fmt.Println("Loaded : ", dir+"/data/")
	fmt.Println("Input Start Height :", startHeight)
	fmt.Println("Input End Height :", endHeight)
	fmt.Println("Latest Height :", blockStore.Height())

	// checking start height
	block := blockStore.LoadBlock(startHeight)
	if block == nil {
		fmt.Println(startHeight, "is not available on this data")
		for i := 0; i < 1000000000000; i++ {
			block := blockStore.LoadBlock(int64(i))
			if block != nil {
				fmt.Println("available starting Height : ", i)
				break
			}
		}
		return nil
	}

	// checking end height
	if endHeight > blockStore.Height() {
		fmt.Println(endHeight, "is not available, Latest Height : ", blockStore.Height())
		return nil
	}

	// Init app
	encCfg := app.MakeEncodingConfig()
	app := app.NewApp(log.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, encCfg, app.EmptyAppOptions{})
	if err := app.LoadHeight(startHeight); err != nil {
		panic(err)
	}

	// Set tx index
	store, err := tmdb.NewGoLevelDBWithOpts("data/tx_index", dir, &opt.Options{
		ErrorIfMissing: true,
		ReadOnly:       true,
	})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer store.Close()
	// ============================ tx parsing logics ========================
	//txi := txkv.NewTxIndex(store)
	//txDecoder := encCfg.TxConfig.TxDecoder()
	// ============================ tx parsing logics ========================

	//ctx := app.BaseApp.NewContext(true, tmproto.Header{})
	//pairs := app.LiquidityKeeper.GetAllPairs(ctx)
	startTime := time.Now()
	orderCount := 0
	orderTxCount := 0

	// Height => PairId => Order
	orderMap := map[int64]map[uint64][]liquiditytypes.Order{}
	// pair => Address => height => orders
	// TODO: // pair => height => Address => orders
	OrderMapByAddr := map[uint64]map[string]map[int64][]liquiditytypes.Order{}

	// TODO: convert to mmOrder and indexing by mm address
	mmOrderMap := map[int64]map[uint64][]liquiditytypes.MsgLimitOrder{}
	//mmOrderCancelMap := map[int64]map[uint64]liquiditytypes.MsgLimitOrder{}

	// TODO: OrderData list
	OrderDataList := []OrderData{}

	// TODO: GET params, considering update height
	//app.MarketMakerKeeper.GetParams(ctx)

	// iterate from startHeight to endHeight
	for i := startHeight; i < endHeight; i++ {
		block = blockStore.LoadBlock(i)
		if i%10000 == 0 {
			fmt.Println(i, time.Now().Sub(startTime), orderCount, orderTxCount)
		}
		orderMap[i] = map[uint64][]liquiditytypes.Order{}
		mmOrderMap[i] = map[uint64][]liquiditytypes.MsgLimitOrder{}
		// Address -> pair -> height -> orders

		// Query paris
		pairs, err := QueryPairs(*app, i)
		if err != nil {
			return err
		}
		fmt.Println("-------------------")
		fmt.Println("pair num: ", len(pairs), i)
		for _, pair := range pairs {
			// TODO: check pruning

			orders, err := QueryOrders(*app, pair.Id, i)
			fmt.Println("order num: ", len(orders), i)
			if err != nil {
				return err
			}
			for _, order := range orders {
				// TODO: query pools, refactoring to function
				pools, err := QueryPools(*app, order.PairId, i)
				if err != nil {
					return err
				}
				fmt.Println("pools num: ", len(pools), i)
				OrderDataList = append(OrderDataList, OrderData{
					Order:     order,
					Pools:     pools,
					Height:    i,
					BlockTime: block.Time,
				})

				orderCount++
				orderMap[i][pair.Id] = append(orderMap[i][pair.Id], order)

				// indexing order.PairId, address
				// TODO: filtering only mm address, mm order
				if _, ok := OrderMapByAddr[pair.Id]; !ok {
					OrderMapByAddr[pair.Id] = map[string]map[int64][]liquiditytypes.Order{}
				}
				if _, ok := OrderMapByAddr[pair.Id][order.Orderer]; !ok {
					OrderMapByAddr[pair.Id][order.Orderer] = map[int64][]liquiditytypes.Order{}
				}
				OrderMapByAddr[pair.Id][order.Orderer][i] = append(OrderMapByAddr[pair.Id][order.Orderer][i], order)
			}
		}

		// ============================ tx parsing logics ========================
		//// get block result
		//block := blockStore.LoadBlock(i)
		//
		//// iterate and parse txs of this block, ordered by txResult.Index 0 -> n
		//for _, tx := range block.Txs {
		//	txResult, err := txi.Get(tx.Hash())
		//	if err != nil {
		//		return fmt.Errorf("get tx index: %w", err)
		//	}
		//
		//	// pass if not succeeded tx
		//	if txResult.Result.Code != 0 {
		//		continue
		//	}
		//
		//	sdkTx, err := txDecoder(txResult.Tx)
		//	if err != nil {
		//		return fmt.Errorf("decode tx: %w", err)
		//	}
		//
		//	// indexing only targeted msg types
		//	for _, msg := range sdkTx.GetMsgs() {
		//		switch msg := msg.(type) {
		//		// TODO: filter only MM order type MMOrder, MMOrderCancel
		//		case *liquiditytypes.MsgLimitOrder:
		//			orderTxCount++
		//			mmOrderMap[i][msg.PairId] = append(mmOrderMap[i][msg.PairId], *msg)
		//			//fmt.Println(i, msg.Orderer, msg.Price, txResult.Result.Code, hex.EncodeToString(tx.Hash()))
		//		}
		//	}
		//}
		// ============================ tx parsing logics ========================
	}
	jsonString, err := json.Marshal(OrderDataList)
	fmt.Println(string(jsonString))
	fmt.Println("finish", orderCount)
	// TODO: analysis logic

	//for _, addrMap := range OrderMapByAddr {
	//	for _, heightMap := range addrMap {
	//		for _, orders := range heightMap {
	//			if len(orders) > 1 {
	//				fmt.Println("=====================")
	//				fmt.Printf("%#v\n", orders)
	//			}
	//		}
	//	}
	//}
	return nil
}

func QueryPairs(app app.App, height int64) (poolsRes []liquiditytypes.Pair, err error) {
	pairsReq := liquiditytypes.QueryPairsRequest{
		Pagination: &query.PageRequest{
			Limit: 1000000,
		},
	}

	pairsRawData, err := pairsReq.Marshal()
	if err != nil {
		return nil, err
	}

	pairsRes := app.Query(abci.RequestQuery{
		Path:   "/crescent.liquidity.v1beta1.Query/Pairs",
		Data:   pairsRawData,
		Height: height,
		Prove:  false,
	})
	if pairsRes.Height != height {
		fmt.Println(fmt.Errorf("pairs height error %d, %d", pairsRes.Height, height))
	}
	var pairsLive liquiditytypes.QueryPairsResponse
	pairsLive.Unmarshal(pairsRes.Value)
	return pairsLive.Pairs, nil
}

func QueryPools(app app.App, pairId uint64, height int64) (poolsRes []liquiditytypes.PoolResponse, err error) {
	// Query Pool
	poolsReq := liquiditytypes.QueryPoolsRequest{
		PairId:   pairId,
		Disabled: "false",
		Pagination: &query.PageRequest{
			Limit: 50,
		},
	}
	dataPools, err := poolsReq.Marshal()
	if err != nil {
		return nil, err
	}

	resPool := app.Query(abci.RequestQuery{
		Path:   "/crescent.liquidity.v1beta1.Query/Pools",
		Data:   dataPools,
		Height: height,
		Prove:  false,
	})
	if resPool.Height != height {
		fmt.Println(fmt.Errorf("pools height error %d, %d", resPool.Height, height))
	}
	var pools liquiditytypes.QueryPoolsResponse
	pools.Unmarshal(resPool.Value)
	return pools.Pools, nil
}

func QueryOrders(app app.App, pairId uint64, height int64) (poolsRes []liquiditytypes.Order, err error) {
	a := liquiditytypes.QueryOrdersRequest{
		PairId: pairId,
		Pagination: &query.PageRequest{
			Limit: 1000000,
		},
	}
	data, err := a.Marshal()
	if err != nil {
		return nil, err
	}

	// Query Orders
	res := app.Query(abci.RequestQuery{
		Path:   "/crescent.liquidity.v1beta1.Query/Orders",
		Data:   data,
		Height: height,
		Prove:  false,
	})
	if res.Height != height {
		fmt.Println(fmt.Errorf("orders height error %d, %d", res.Height, height))
	}
	var orders liquiditytypes.QueryOrdersResponse
	orders.Unmarshal(res.Value)
	return orders.Orders, nil
}
