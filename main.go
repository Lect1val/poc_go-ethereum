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

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to retrieve latest block header: %v", err)
	}

	fmt.Printf("The latest block number is: %d\n", header.Number.Int64())

	contractAddress := common.HexToAddress("0xdf0527bDe17EBd936Ff0BC5082930769022C5c91")
	mintedEventSignature := common.HexToHash("0xf9f4d81453ddebad232670c9402c17d910e88b4e805cf759f854b4d387fc7f61")

	// Define your starting and ending block here
	startBlock := big.NewInt(21262077) // Example start block
	endBlock := big.NewInt(0)          // Use zero to signify the latest block

	// If endBlock is 0, query the latest block number from the blockchain
	if endBlock.Cmp(big.NewInt(0)) == 0 {
		latestBlock, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			log.Fatal(err)
		}
		endBlock = latestBlock.Number()
	}

	blockRange := big.NewInt(1000) // Define the size of each block range

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
	// now := time.Now()
	// firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// startLastMonth := firstOfMonth.AddDate(0, -1, 0)
	// endLastMonth := firstOfMonth.Add(-time.Nanosecond)

	// // Find blocks for the start and end of the last month
	// startBlock, err := findBlockByTimestamp(client, startLastMonth.Unix())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// endBlock, err := findBlockByTimestamp(client, endLastMonth.Unix())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("Blocks for last month range from %d to %d\n", startBlock, endBlock)
}

func handleLog(client *ethclient.Client, vLog types.Log) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
	if err != nil {
		log.Fatalf("Failed to fetch block: %v", err)
		return
	}

	blockTime := time.Unix(int64(block.Time()), 0)
	if blockTime.Month() == time.January { // Filtering for January
		fmt.Printf("Found event in January: Block %v, Data %x\n", vLog.BlockNumber, vLog.Data)
	}
}

// func findBlockByTimestamp(client *ethclient.Client, targetTimestamp int64) (int64, error) {
// 	var low, high *big.Int

// 	header, err := client.HeaderByNumber(context.Background(), nil)
// 	if err != nil {
// 		return 0, err
// 	}
// 	high = header.Number
// 	low = big.NewInt(0)

// 	for low.Cmp(high) <= 0 {
// 		mid := new(big.Int).Add(low, high)
// 		mid.Div(mid, big.NewInt(2))

// 		header, err := client.HeaderByNumber(context.Background(), mid)
// 		if err != nil {
// 			return 0, err
// 		}

// 		blockTime := int64(header.Time)
// 		if blockTime < targetTimestamp {
// 			low = mid.Add(mid, big.NewInt(1))
// 		} else if blockTime > targetTimestamp {
// 			high = mid.Sub(mid, big.NewInt(1))
// 		} else {
// 			return mid.Int64(), nil
// 		}
// 	}

// 	return high.Int64(), nil
// }
