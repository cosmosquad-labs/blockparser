package cmd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tendermint/tendermint/libs/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	crecmd "github.com/crescent-network/crescent/v2/cmd/crescentd/cmd"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
	"github.com/spf13/cobra"

	"github.com/crescent-network/crescent/v2/app"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewBlockParserCmd() *cobra.Command {
	crecmd.GetConfig()
	cmd := &cobra.Command{
		Use:  "blockparser [dir] [start-height] [end-height] [address]",
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			addr := args[3]
			startHeight, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("parse start-Height: %w", err)
			}

			endHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("parse end-Height: %w", err)
			}
			return Main(dir, startHeight, endHeight, addr)
		},
	}
	return cmd
}

func Main(dir string, startHeight, endHeight int64, addrInput string) error {

	u := url.URL{Scheme: "ws", Host: "127.0.0.1:26657", Path: "/websocket"}
	fmt.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("dial:", err)
	}
	defer c.Close()

	//done := make(chan struct{})
	//
	//go func() {
	//	defer close(done)
	//	for {
	//		_, message, err := c.ReadMessage()
	//		if err != nil {
	//			log.Println("read:", err)
	//			return
	//		}
	//		log.Printf("recv: %s", message)
	//	}
	//}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	//query := `{"jsonrpc":"2.0","method":"subscribe","id":0,"params":{"query":"tm.event='NewBlock'"}}`
	query := `{"jsonrpc":"2.0","method":"consensus_params","id":0,"params":{"height":"402001"}}`
	err = c.WriteMessage(websocket.TextMessage, []byte(query))
	if err != nil {
		log.Println("write:", err)
	}

	cons := ConsensusParamsResponse{}
	_, res, err := c.ReadMessage()
	err = json.Unmarshal(res, &cons)
	fmt.Println(cons, err)
	fmt.Println(string(res))

	// =============================================================
	//addr, err := sdk.AccAddressFromBech32(addrInput)
	//if err != nil {
	//	return err
	//}

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		"127.0.0.1:9090",    // your gRPC server address.
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
		// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
		// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
		//grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	liquidityClient := liquiditytypes.NewQueryClient(grpcConn)

	fmt.Println("version :", "v0.1.0")
	fmt.Println("Loaded : ", dir+"/data/")
	fmt.Println("Input Start Height :", startHeight)
	fmt.Println("Input End Height :", endHeight)

	// Init app
	//encCfg := app.MakeEncodingConfig()
	//app := app.NewApp(log.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, encCfg, app.EmptyAppOptions{})
	//// should be load last height from v2 (sdk 0.45.*)
	//if err := app.LoadHeight(endHeight); err != nil {
	//	panic(err)
	//}

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
	//OrderMapByAddr := map[uint64]map[string]map[int64][]liquiditytypes.Order{}

	// TODO: convert to mmOrder and indexing by mm address
	mmOrderMap := map[int64]map[uint64][]liquiditytypes.MsgLimitOrder{}
	//mmOrderCancelMap := map[int64]map[uint64]liquiditytypes.MsgLimitOrder{}

	// TODO: OrderData list
	//OrderDataList := []OrderData{}

	// TODO: GET params, considering update height
	//app.MarketMakerKeeper.GetParams(ctx)

	// iterate from startHeight to endHeight
	for i := startHeight; i < endHeight; i++ {
		_, err := QueryParamsGRPC(liquidityClient, i)
		if err != nil {
			fmt.Println("prunned height", i)
			continue
		}

		if i%10000 == 0 {
			fmt.Println(i, time.Now().Sub(startTime), orderCount, orderTxCount)
		}
		orderMap[i] = map[uint64][]liquiditytypes.Order{}
		mmOrderMap[i] = map[uint64][]liquiditytypes.MsgLimitOrder{}
		// Address -> pair -> height -> orders

		// Query paris
		//pairs, err := QueryPairsGRPC(liquidityClient, i)
		//if err != nil {
		//	return err
		//}
		//for _, pair := range pairs {
		orders, err := QueryOrdersByOrdererGRPC(liquidityClient, addrInput, i)
		if err != nil {
			return err
		}
		if len(orders) != 0 {
			fmt.Println(i, len(orders), orders)
		}
		//}
		//	// TODO: check pruning
		//
		//	orders, err := QueryOrders(*app, pair.Id, i)
		//	if err != nil {
		//		return err
		//	}
		//	for _, order := range orders {
		//		// TODO: query pools, refactoring to function
		//		pools, err := QueryPools(*app, order.PairId, i)
		//		if err != nil {
		//			return err
		//		}
		//		OrderDataList = append(OrderDataList, OrderData{
		//			Order:     order,
		//			Pools:     pools,
		//			Height:    i,
		//			BlockTime: block.Time,
		//		})
		//
		//		orderCount++
		//		orderMap[i][pair.Id] = append(orderMap[i][pair.Id], order)
		//
		//		// indexing order.PairId, address
		//		// TODO: filtering only mm address, mm order
		//		if _, ok := OrderMapByAddr[pair.Id]; !ok {
		//			OrderMapByAddr[pair.Id] = map[string]map[int64][]liquiditytypes.Order{}
		//		}
		//		if _, ok := OrderMapByAddr[pair.Id][order.Orderer]; !ok {
		//			OrderMapByAddr[pair.Id][order.Orderer] = map[int64][]liquiditytypes.Order{}
		//		}
		//		OrderMapByAddr[pair.Id][order.Orderer][i] = append(OrderMapByAddr[pair.Id][order.Orderer][i], order)
		//	}
		//}
	}
	//jsonString, err := json.Marshal(OrderDataList)
	//fmt.Println(string(jsonString))
	//fmt.Println("finish", orderCount)
	// TODO: analysis logic
	return nil
}

func QueryPairs(app app.App, height int64) (poolsRes []liquiditytypes.Pair, err error) {
	pairsReq := liquiditytypes.QueryPairsRequest{
		Pagination: &query.PageRequest{
			Limit: 100,
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

func QueryPairsGRPC(cli liquiditytypes.QueryClient, height int64) (poolsRes []liquiditytypes.Pair, err error) {
	pairsReq := liquiditytypes.QueryPairsRequest{
		Pagination: &query.PageRequest{
			Limit: 100,
		},
	}

	var header metadata.MD

	pairs, err := cli.Pairs(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10)),
		&pairsReq,
		grpc.Header(&header),
	)
	if err != nil {
		return nil, err
	}

	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	resHeight, err := strconv.ParseInt(blockHeight[0], 10, 64)
	if err != nil {
		return nil, err
	}
	if resHeight != height {
		fmt.Println(fmt.Errorf("pairs height error %d, %d", resHeight, height))
	}

	return pairs.Pairs, nil
}

func QueryParamsGRPC(cli liquiditytypes.QueryClient, height int64) (resParams *liquiditytypes.QueryParamsResponse, err error) {
	req := liquiditytypes.QueryParamsRequest{}

	var header metadata.MD

	res, err := cli.Params(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10)),
		&req,
		grpc.Header(&header),
	)
	if err != nil {
		return nil, err
	}

	//blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	//resHeight, err := strconv.ParseInt(blockHeight[0], 10, 64)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if resHeight != height {
	//	fmt.Println(fmt.Errorf("pairs height error %d, %d", resHeight, height))
	//}

	return res, nil
}

func QueryOrdersByOrdererGRPC(cli liquiditytypes.QueryClient, orderer string, height int64) (poolsRes []liquiditytypes.Order, err error) {
	req := liquiditytypes.QueryOrdersByOrdererRequest{
		Orderer: orderer,
		Pagination: &query.PageRequest{
			Limit: 1000,
		},
	}

	var header metadata.MD

	res, err := cli.OrdersByOrderer(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10)),
		&req,
		grpc.Header(&header),
	)
	if err != nil {
		return nil, err
	}

	//blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	//resHeight, err := strconv.ParseInt(blockHeight[0], 10, 64)
	//if err != nil {
	//	return nil, err
	//}

	//// TODO: didn't works
	//if resHeight != height {
	//	fmt.Println(fmt.Errorf("pairs height error %d, %d", resHeight, height))
	//}

	return res.Orders, nil
}

func QueryOrdersByOrdererByPairGRPC(cli liquiditytypes.QueryClient, orderer string, pairId uint64, height int64) (poolsRes []liquiditytypes.Order, err error) {
	req := liquiditytypes.QueryOrdersByOrdererRequest{
		Orderer: orderer,
		PairId:  pairId,
		Pagination: &query.PageRequest{
			Limit: 1000,
		},
	}

	var header metadata.MD

	res, err := cli.OrdersByOrderer(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10)),
		&req,
		grpc.Header(&header),
	)
	if err != nil {
		return nil, err
	}

	//blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	//resHeight, err := strconv.ParseInt(blockHeight[0], 10, 64)
	//if err != nil {
	//	return nil, err
	//}
	//if resHeight != height {
	//	fmt.Println(fmt.Errorf("pairs height error %d, %d", resHeight, height))
	//}

	return res.Orders, nil
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

type ConsensusParamsResponse struct {
	Jsonrpc string                       `json:"jsonrpc"`
	ID      int                          `json:"id"`
	Result  ctypes.ResultConsensusParams `json:"result"`
}
