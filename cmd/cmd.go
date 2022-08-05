package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ttypes "github.com/tendermint/tendermint/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/crescent-network/crescent/v2/app"
	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/app/params"
	crecmd "github.com/crescent-network/crescent/v2/cmd/crescentd/cmd"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"

	"github.com/cosmosquad-labs/blockparser/api/server"
)

type Context struct {
	LastHeight int64
	SyncStatus bool // TODO: delete?

	LiquidityClient liquiditytypes.QueryClient
	// TODO: bank, etc

	RpcWebsocketClient *rpchttp.HTTP
	WebsocketCtx       context.Context

	SyncDB *leveldb.DB

	Enc params.EncodingConfig
	Config
}

type Config struct {
	AccList            []string
	PairList           []uint64 // if empty, all pairs
	StartHeight        int64
	OrderbookKeepBlock int // zero == keep all?
	AllOrdersKeepBlock int
	PairPoolKeepBlock  int
	BalanceKeepBlock   int
	GrpcEndpoint       string
	RpcEndpoint        string
	Dir                string

	// TODO: orderbook argument
}

var config = Config{
	AccList: []string{
		"cre1tcgjtr03xqzjjwxslrpfaajvsk7nclv6fkgtxt",
		"cre1xzmgjjlz7kacgkpxk5gn6lqa0dvavg8rjgdvcz",
	},
	PairList:           []uint64{},
	StartHeight:        0,
	OrderbookKeepBlock: 0,
	AllOrdersKeepBlock: 0,
	PairPoolKeepBlock:  0,
	BalanceKeepBlock:   0,
	GrpcEndpoint:       "127.0.0.1:9090",
	RpcEndpoint:        "tcp://127.0.0.1:26657",
	Dir:                "", // TODO: home
	// TODO: GRPC, RPC endpoint
}

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
			return Main(dir, startHeight, endHeight, addr, config)
		},
	}
	return cmd
}

func Main(dir string, startHeight, endHeight int64, addrInput string, config Config) error {

	var ctx Context
	ctx.Config = config
	ctx.Enc = chain.MakeEncodingConfig()

	go server.Serv()

	// ================= set db ==============================
	syncdb, err := leveldb.OpenFile(ctx.Config.Dir+"mmapi/sync", nil)
	if err != nil {
		return err
	}
	defer syncdb.Close()
	ctx.SyncDB = syncdb

	// ================================================= Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		ctx.Config.GrpcEndpoint, // your gRPC server address.
		grpc.WithInsecure(),     // The Cosmos SDK doesn't support any transport security mechanism.
		// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
		// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
		//grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	ctx.LiquidityClient = liquiditytypes.NewQueryClient(grpcConn)
	// =========================================================================================

	//// TODO: start after all synced
	// TODO: go routine
	//// ======================================================================= websocket
	wc, err := rpchttp.New(ctx.Config.RpcEndpoint, "/websocket")
	if err != nil {
		log.Fatal(err)
	}
	ctx.RpcWebsocketClient = wc

	err = ctx.RpcWebsocketClient.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.RpcWebsocketClient.Stop()
	wcCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ctx.WebsocketCtx = wcCtx

	query := "tm.event = 'NewBlock'"
	txs, err := ctx.RpcWebsocketClient.Subscribe(ctx.WebsocketCtx, "test-wc", query)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for e := range txs {
			switch data := e.Data.(type) {
			case ttypes.EventDataNewBlock:
				fmt.Printf("Block %s - Height: %d \n", hex.EncodeToString(data.Block.Hash()), data.Block.Height)
				// TODO: trigger for data sync
				//go func() {
				//	pairs, err := QueryPairsGRPC(ctx.LiquidityClient, data.Block.Height)
				//	if err != nil {
				//		fmt.Println(err)
				//		return
				//	}
				//	ctx.SetPairs(pairs, data.Block.Height)
				//	ctx.PairsPrint()
				//}()
				for _, addr := range ctx.Config.AccList {
					// TODO: go routine to each acc

					orders, err := QueryOrdersByOrdererGRPC(ctx.LiquidityClient, addr, data.Block.Height)
					if err != nil {
						fmt.Println(err)
						continue
					}
					if len(orders) != 0 {
						fmt.Println(data.Block.Height, len(orders), orders)
					}
				}

				ctx.SyncLog(data.Block.Height)
				//err = ctx.SyncLogPrint()
				if err != nil {
					fmt.Println(err)
				}

				break
			}
		}
	}()

	//// ======================================================================= websocket

	//wc.ABCIQuery()
	status, err := ctx.RpcWebsocketClient.Status(ctx.WebsocketCtx)
	if err != nil {
		return err
	}
	fmt.Println(status.SyncInfo)

	////==================== abci query

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
		_, err := QueryParamsGRPC(ctx.LiquidityClient, i)
		if err != nil {
			//fmt.Println("prunned height", i)
			continue
		}

		if i%10000 == 0 {
			fmt.Println(i, time.Now().Sub(startTime), orderCount, orderTxCount)
		}
		orderMap[i] = map[uint64][]liquiditytypes.Order{}
		mmOrderMap[i] = map[uint64][]liquiditytypes.MsgLimitOrder{}
		// Address -> pair -> height -> orders

		// Query paris
		//pairs, err := QueryPairsGRPC(ctx.liquidityClient, i)
		//if err != nil {
		//	return err
		//}
		//for _, pair := range pairs {
		orders, err := QueryOrdersByOrdererGRPC(ctx.LiquidityClient, addrInput, i)
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

	return res.Orders, nil
}

func QueryOrdersGRPC(cli liquiditytypes.QueryClient, pairId uint64, height int64) (poolsRes []liquiditytypes.Order, err error) {
	req := liquiditytypes.QueryOrdersRequest{
		PairId: pairId,
		Pagination: &query.PageRequest{
			Limit: 1000,
		},
	}

	var header metadata.MD

	res, err := cli.Orders(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10)),
		&req,
		grpc.Header(&header),
	)
	if err != nil {
		return nil, err
	}

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
