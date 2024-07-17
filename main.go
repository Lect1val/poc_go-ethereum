package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial(os.Getenv("RPC_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// Set the start time to one month before the current time and end time to the current time
	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0)

	startBlock, err := findBlockByTimestamp(client, startTime.Unix())
	if err != nil {
		log.Fatalf("Failed to find start block: %v", err)
	}
	endBlock, err := findBlockByTimestamp(client, endTime.Unix())
	if err != nil {
		log.Fatalf("Failed to find end block: %v", err)
	}

	blockRange := big.NewInt(1000) // Define the size of each block range
	contractAddress := common.HexToAddress("0xdf0527bDe17EBd936Ff0BC5082930769022C5c91")
	mintedEventSignature := common.HexToHash("0xf9f4d81453ddebad232670c9402c17d910e88b4e805cf759f854b4d387fc7f61")

	// Loop through blocks in chunks
	for current := new(big.Int).Set(startBlock); current.Cmp(endBlock) <= 0; current.Add(current, blockRange) {
		toBlock := new(big.Int).Add(current, blockRange)
		if toBlock.Cmp(endBlock) > 0 {
			toBlock.Set(endBlock)
		}

		query := ethereum.FilterQuery{
			FromBlock: current,
			ToBlock:   toBlock,
			Addresses: []common.Address{contractAddress},
			Topics:    [][]common.Hash{{mintedEventSignature}},
		}

		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to fetch logs: %v", err)
		}

		for _, vLog := range logs {
			handleLog(client, vLog)
		}
	}
}

func handleLog(client *ethclient.Client, vLog types.Log) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
	if err != nil {
		log.Fatalf("Failed to fetch block: %v", err)
		return
	}

	blockTime := time.Unix(int64(block.Time()), 0)
	fmt.Printf("Found event: Block %v, Timestamp %v, Data %x\n", vLog.BlockNumber, blockTime, vLog.Data)
}

func findBlockByTimestamp(client *ethclient.Client, timestamp int64) (*big.Int, error) {
	head, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	var low, high *big.Int = big.NewInt(0), head.Number
	for low.Cmp(high) < 0 {
		mid := new(big.Int).Add(low, high)
		mid.Div(mid, big.NewInt(2))
		header, err := client.HeaderByNumber(context.Background(), mid)
		if err != nil {
			return nil, err
		}
		blockTime := header.Time
		if blockTime < uint64(timestamp) {
			low = mid.Add(mid, big.NewInt(1))
		} else {
			high = mid
		}
	}
	return high, nil
}
