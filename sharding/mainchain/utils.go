package mainchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/sharding/contracts"
	"github.com/ethereum/go-ethereum/sharding/params"
)

// dialRPC endpoint to node.
func dialRPC(endpoint string) (*rpc.Client, error) {
	if endpoint == "" {
		endpoint = node.DefaultIPCEndpoint(ClientIdentifier)
	}
	return rpc.Dial(endpoint)
}

// initSMC initializes the sharding manager contract bindings.
// If the SMC does not exist, it will be deployed.
func initSMC(s *SMCClient) (*contracts.SMC, error) {
	b, err := s.client.CodeAt(context.Background(), params.DefaultConfig.SMCAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get contract code at %s: %v", params.DefaultConfig.SMCAddress.Hex(), err)
	}

	// Deploy SMC for development only.
	// TODO: Separate contract deployment from the sharding node. It would only need to be deployed
	// once on the mainnet, so this code would not need to ship with the node.
	if len(b) == 0 {
		log.Info(fmt.Sprintf("No sharding manager contract found at %s. Deploying new contract.", params.DefaultConfig.SMCAddress.Hex()))

		txOps, err := s.CreateTXOpts(big.NewInt(0))
		if err != nil {
			return nil, fmt.Errorf("unable to initiate the transaction: %v", err)
		}

		addr, tx, contract, err := contracts.DeploySMC(txOps, s.client)
		if err != nil {
			return nil, fmt.Errorf("unable to deploy sharding manager contract: %v", err)
		}

		for pending := true; pending; _, pending, err = s.client.TransactionByHash(context.Background(), tx.Hash()) {
			if err != nil {
				return nil, fmt.Errorf("unable to get transaction by hash: %v", err)
			}
			time.Sleep(1 * time.Second)
		}

		log.Info(fmt.Sprintf("New contract deployed at %s", addr.Hex()))
		return contract, nil
	}

	return contracts.NewSMC(params.DefaultConfig.SMCAddress, s.client)
}
