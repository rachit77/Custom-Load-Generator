package loadbot

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/core/types"

	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/celo-org/celo-blockchain/mycelo/env"
	"golang.org/x/sync/errgroup"
)

// 110k gas for stable token transfer is pretty reasonable. It's just under 100k in practice
const GasForTransferWithComment = 110000

var m = "alien shell toy depth share work clarify tattoo grass tank master board"

// LoadGenerator keeps track of in-flight transactions
type LoadGenerator struct {
	MaxPending uint64
	Pending    uint64
	PendingMu  sync.Mutex
}

// TxConfig contains the options for a transaction
// type txConfig struct {
// 	Acc               env.Account
// 	Nonce             uint64
// 	Recipient         common.Address
// 	Value             *big.Int
// 	Verbose           bool
// 	SkipGasEstimation bool
// 	MixFeeCurrency    bool
// }

// Config represent the load bot run configuration
type Config struct {
	ChainID               *big.Int
	Accounts              []env.Account
	Amount                *big.Int
	TransactionsPerSecond int
	Clients               []*ethclient.Client
	Verbose               bool
	MaxPending            uint64
	SkipGasEstimation     bool
	MixFeeCurrency        bool
}

// Start will start loads bots
func Start(ctx context.Context, cfg *Config) error {

	sendIdx := 1
	gasPriceReduction := big.NewInt(500000000000)

	// Fire off transactions
	period := 1 * time.Second / time.Duration(cfg.TransactionsPerSecond)
	ticker := time.NewTicker(period)
	group, ctx := errgroup.WithContext(ctx)
	lg := &LoadGenerator{
		MaxPending: cfg.MaxPending,
	}

	for {
		select {
		case <-ticker.C:
			fmt.Printf("before mu %v ", lg.Pending)
			fmt.Printf("\n")

			// lg.PendingMu.Lock()
			// if lg.MaxPending != 0 && lg.Pending > lg.MaxPending {
			// 	lg.PendingMu.Unlock()
			// 	fmt.Printf("inside if")
			// 	fmt.Printf("\n")
			// 	continue
			// } else {
			// 	lg.Pending++
			// 	lg.PendingMu.Unlock()
			// 	fmt.Printf("inside else")
			// 	fmt.Printf("\n")
			// }

			lg.PendingMu.Lock()
			lg.Pending++
			lg.PendingMu.Unlock()

			fmt.Printf("after mu")
			fmt.Printf("\n")

			senderpointer, err := env.DeriveAccount(m, 1, sendIdx)
			if err != nil {
				log.Fatal(err)
			}

			sender := *senderpointer

			fmt.Printf("sender %v", sender.Address.Hex())
			fmt.Printf("\n")

			//finding balance of each account through samrt contract
			bal, err2 := cfg.Clients[0].BalanceAt(ctx, sender.Address, nil)
			if err2 != nil {
				return fmt.Errorf("failed to retrieve balance for account")
			}

			// client, err := ethclient.Dial("https://localhost:8545")
			// if err != nil {
			// log.Fatal(err)
			// }

			fmt.Printf("after bal %v", bal)
			fmt.Printf("\n")

			nonceAdr, err5 := cfg.Clients[0].PendingNonceAt(ctx, sender.Address)
			if err5 != nil {
				return fmt.Errorf("failed to retrieve pending nonce for account  %v", err5)
			}
			k := nonceAdr

			fmt.Printf("nonce %v", k)
			fmt.Printf("\n")

			b := big.NewInt(20000000000000) //13 zeroes

			value := new(big.Int).Sub(bal, b)

			//value := big.NewInt() // in wei (1 eth)
			gasLimit := uint64(21000) // in units
			// gasPrice, err := cfg.Clients[0].SuggestGasPrice(context.Background())
			// if err != nil {
			// 	log.Fatal(err)
			// }

			//gasPriceReduction :=big.NewInt(500000000000)

			gasPriceReduction = new(big.Int).Sub(gasPriceReduction, big.NewInt(100000))
			gasPrice := gasPriceReduction

			fmt.Printf("after gas price %v", gasPrice)
			fmt.Printf("\n")

			var data []byte

			recpointer, err1 := env.DeriveAccount(m, 1, sendIdx+1)
			if err != nil {
				log.Fatal(err1)
			}
			rec := (*recpointer).Address

			tx := types.NewTransaction(k, rec, value, gasLimit, gasPrice, nil, nil, nil, data)
			//func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, feeCurrency, gatewayFeeRecipient *common.Address, gatewayFee *big.Int, data []byte) *Transaction {

			signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cfg.ChainID), sender.PrivateKey)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("after trx signed")
			fmt.Printf("\n")

			err = cfg.Clients[0].SendTransaction(context.Background(), signedTx)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("tx sent: %s", signedTx.Hash().Hex())

			sendIdx++
			fmt.Printf("increase sendidx")
			fmt.Printf("\n")

		case <-ctx.Done():
			fmt.Printf("ctx done")
			return group.Wait()

		}

	}
}
