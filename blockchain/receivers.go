package blockchain

import (
	"context"
	"sync"
	"time"

	"github.com/bytom/crypto/ed25519/chainkd"
	chainjson "github.com/bytom/encoding/json"
	"github.com/bytom/net/http/reqid"
	"fmt"
	"strconv"
)

type (
	createAccountPubkeyRequest struct {
		AccountID    string `json:"account_id"`
		AccountAlias string `json:"account_alias"`
	}
	createAccountPubkeyResponse struct {
		Root   chainkd.XPub         `json:"root_xpub"`
		Pubkey chainjson.HexBytes   `json:"pubkey"`
		Path   []chainjson.HexBytes `json:"pubkey_derivation_path"`
		idx string 	   		   		`json:"idx"`
	}
)

// POST /create-account-pubkey
func (a *BlockchainReactor) createAccountPubkey(ctx context.Context, req createAccountPubkeyRequest) (createAccountPubkeyResponse) {
	fmt.Println("===========================================================")
	fmt.Println("createAccountPubkeyRequest:", req.AccountID)
	root, pubkey, path, index, err := a.accounts.CreatePubkey(ctx, req.AccountID, req.AccountAlias)
	if err != nil {
		return createAccountPubkeyResponse{}
	}

	idxkey := strconv.FormatUint(index, 10)
	fmt.Println("res.idx", idxkey)
	res := createAccountPubkeyResponse{
		Root:   root,
		Pubkey: chainjson.HexBytes(pubkey),
		idx: idxkey,
	}
	for _, p := range path {
		res.Path = append(res.Path, chainjson.HexBytes(p))
	}

	return res
}

// POST /create-account-receiver
func (a *BlockchainReactor) createAccountReceiver(ctx context.Context, ins []struct {
	AccountID    string    `json:"account_id"`
	AccountAlias string    `json:"account_alias"`
	ExpiresAt    time.Time `json:"expires_at"`
}) []interface{} {
	responses := make([]interface{}, len(ins))
	var wg sync.WaitGroup
	wg.Add(len(responses))

	for i := 0; i < len(responses); i++ {
		go func(i int) {
			subctx := reqid.NewSubContext(ctx, reqid.New())
			defer wg.Done()
			defer batchRecover(subctx, &responses[i])

			receiver, err := a.accounts.CreateReceiver(subctx, ins[i].AccountID, ins[i].AccountAlias, ins[i].ExpiresAt)
			if err != nil {
				responses[i] = err
			} else {
				responses[i] = receiver
			}
		}(i)
	}

	wg.Wait()
	return responses
}
