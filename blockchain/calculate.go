package blockchain

import (
	"context"
	"sync"
	"github.com/bytom/blockchain/txbuilder"
	"github.com/bytom/errors"
	"github.com/bytom/net/http/reqid"
	"fmt"
)

// POST /submit-transaction
func (a *BlockchainReactor) calculateGas(ctx context.Context, x SubmitArg) (interface{}, error) {
	responses := make([]interface{}, len(x.Transactions))
	var wg sync.WaitGroup
	wg.Add(len(responses))
	for i := range responses {
		go func(i int) {
			subctx := reqid.NewSubContext(ctx, reqid.New())
			defer wg.Done()
			defer batchRecover(subctx, &responses[i])

			gas, err := a.calculateSingle(subctx, &x.Transactions[i], x.WaitUntil)
			if err != nil {
				responses[i] = err
			} else {
				responses[i] = gas
			}
		}(i)
	}

	wg.Wait()
	return responses, nil
}

func (a *BlockchainReactor) calculateSingle(ctx context.Context, tpl *txbuilder.Template, waitUntil string) (interface{}, error) {
	if tpl.Transaction == nil {
		return nil, errors.Wrap(txbuilder.ErrMissingRawTx)
	}

	usedgas, err := a.chain.GetUsedGas(tpl.Transaction)
	if err != nil {
		return nil, err
	}
	fmt.Println("after run the trasaction control program, the used gas:", usedgas)

	return usedgas, nil
}
