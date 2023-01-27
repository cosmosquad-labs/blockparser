package cmd

import (
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"encoding/json"
	"reflect"
	"strconv"
	_ "strings"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func ObjectToJsonString(object any) string {
	b, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	return string(b)

	// Unmarshal json to object
	// var blockCommit = BlockCommit{}
	// json.Unmarshal([]byte(string(b)), &blockCommit)
}

func NewBlockParserCmd() *cobra.Command {
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

			blockDB, err := tmdb.NewGoLevelDBWithOpts("data/blockstore", dir, &opt.Options{
				ErrorIfMissing: true,
				ReadOnly:       true,
			})

			if err != nil {
				panic(err)
			}
			defer blockDB.Close()

			stateDB, err := tmdb.NewGoLevelDBWithOpts("data/state", dir, &opt.Options{
				ErrorIfMissing: true,
				ReadOnly:       true,
			})

			if err != nil {
				panic(err)
			}
			defer stateDB.Close()
			stateStore := state.NewStore(stateDB, state.StoreOptions{
				DiscardABCIResponses: false,
			})

			blockStore := store.NewBlockStore(blockDB)

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

			fmt.Println("Done! check the output files on current dir")

			conn, err := sql.Open("mysql", "root:passw0rd@tcp(127.0.0.1:3306)/backend") // 1
			if err != nil {
				fmt.Println(err)
			}
			defer conn.Close() // 3

			fmt.Printf("DB 연동: %+v\n", conn.Stats()) // 2

			ibcEventTypes := map[string]bool{
				"send_packet": true,
				// "ibc_transfer":        true,

				"recv_packet": true,
				// "write_acknowledgement" : true,
				// "denomination_trace" : true,
				// "fungible_token_packet" : true,

				"acknowledge_packet": true,
				// "fungible_token_packet" : true,

				"timeout_packet": true,
				// "timeout": true,
			}

			fmt.Println(len(ibcEventTypes))

			// packets := []EventPacket{}

			for i := startHeight; i < endHeight; i++ {

				block := blockStore.LoadBlock(i)
				blockHeight := i
				blockTime := block.Time.UTC().Unix()

				results, err := stateStore.LoadABCIResponses(i)
				// https://pkg.go.dev/github.com/tendermint/tendermint@v0.34.22/proto/tendermint/state#ABCIResponses

				if err != nil {
					return err
				}

				for _, tx := range results.DeliverTxs {

					var p = EventPacket{}
					p.block_height = strconv.FormatInt(blockHeight, 10)
					p.block_time = strconv.FormatInt(blockTime, 10)

					// p.event_type
					// p.packet_timeout_height
					// p.packet_timeout_timestamp
					// p.packet_sequence
					// p.packet_src_port
					// p.packet_src_channel
					// p.packet_dst_port
					// p.packet_dst_channel
					// p.packet_channel_ordering
					// p.packet_connection

					for _, evt := range tx.Events {
						// https://pkg.go.dev/github.com/tendermint/tendermint@v0.34.22/abci/types#Event

						if ibcEventTypes[evt.Type] {

							p.event_type = evt.Type
							// if strings.Contains(evt.Type, "packet") {
							// fmt.Println(evt.Type)
							for _, attr := range evt.Attributes {
								// https://pkg.go.dev/github.com/tendermint/tendermint@v0.34.22/abci/types#EventAttribute

								// fmt.Println(BytesToString(attr.Key), BytesToString(attr.Value), attr.Index)
								key := BytesToString(attr.Key)
								value := BytesToString(attr.Value)

								switch key {
								case "packet_timeout_height":
									p.packet_timeout_height = value
								case "packet_timeout_timestamp":
									p.packet_timeout_timestamp = value
								case "packet_sequence":
									p.packet_sequence = value
								case "packet_src_port":
									p.packet_src_port = value
								case "packet_src_channel":
									p.packet_src_channel = value
								case "packet_dst_port":
									p.packet_dst_port = value
								case "packet_dst_channel":
									p.packet_dst_channel = value
								case "packet_channel_ordering":
									p.packet_channel_ordering = value
								case "packet_connection":
									p.packet_connection = value
								}

							}

							fmt.Println(p.block_height, p.block_time, p.event_type, p.packet_timeout_height, p.packet_timeout_timestamp, p.packet_sequence, p.packet_src_port, p.packet_src_channel, p.packet_dst_port, p.packet_dst_channel, p.packet_channel_ordering, p.packet_connection)
							insertStr := fmt.Sprintf("insert ignore into packets (block_height, block_time, event_type, packet_timeout_height, packet_timeout_timestamp, packet_sequence, packet_src_port, packet_src_channel, packet_dst_port, packet_dst_channel, packet_channel_ordering, packet_connection) value ('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s')", p.block_height, p.block_time, p.event_type, p.packet_timeout_height, p.packet_timeout_timestamp, p.packet_sequence, p.packet_src_port, p.packet_src_channel, p.packet_dst_port, p.packet_dst_channel, p.packet_channel_ordering, p.packet_connection)
							_, err := conn.Exec(insertStr)

							if err != nil {
								return err
							}
							// return nil
							// fmt.Println(ObjectToJsonString(tx))
							// return nil
						}
					}

				}
			}

			return nil
		},
	}
	return cmd
}
