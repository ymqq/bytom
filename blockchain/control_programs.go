package blockchain

import (
	"context"
	stdjson "encoding/json"
	"sync"
	"time"

	"github.com/bytom/encoding/json"
	"github.com/bytom/errors"
	"github.com/bytom/net/http/httpjson"
	"github.com/bytom/net/http/reqid"
	"fmt"
	"encoding/hex"
	"strconv"
)

// POST /create-control-program
func (a *BlockchainReactor) createControlProgram(ctx context.Context, ins []struct {
	Type   string
	Params stdjson.RawMessage
}) interface{} {
	responses := make([]interface{}, len(ins))
	var wg sync.WaitGroup
	wg.Add(len(responses))

	for i := 0; i < len(responses); i++ {
		go func(i int) {
			subctx := reqid.NewSubContext(ctx, reqid.New())
			defer wg.Done()
			defer batchRecover(subctx, &responses[i])

			var (
				prog interface{}
				err  error
			)
			switch ins[i].Type {
			case "account":
				prog, err = a.createAccountControlProgram(subctx, ins[i].Params)
			case "contract":
				prog, err = a.createContractControlProgram(subctx, ins[i].Params)
			default:
				err = errors.WithDetailf(httpjson.ErrBadRequest, "unknown control program type %q", ins[i].Type)
			}
			if err != nil {
				responses[i] = err
			} else {
				responses[i] = prog
			}
		}(i)
	}

	wg.Wait()
	return responses
}

func (a *BlockchainReactor) createAccountControlProgram(ctx context.Context, input []byte) (interface{}, error) {
	var parsed struct {
		AccountAlias string `json:"account_alias"`
		AccountID    string `json:"account_id"`
	}
	err := stdjson.Unmarshal(input, &parsed)
	if err != nil {
		return nil, errors.WithDetailf(httpjson.ErrBadRequest, "bad parameters for account control program")
	}

	fmt.Println("parsed:", parsed)
	accountID := parsed.AccountID
	if accountID == "" {
		acc, err := a.accounts.FindByAlias(ctx, parsed.AccountAlias)
		if err != nil {
			return nil, err
		}
		accountID = acc.ID
	}
	fmt.Println("accountID:", accountID)

	controlProgram, err := a.accounts.CreateControlProgram(ctx, accountID, false, time.Time{})
	if err != nil {
		return nil, err
	}

	fmt.Println("controlProgram:", controlProgram)
	ret := map[string]interface{}{
		"control_program": json.HexBytes(controlProgram),
	}
	return ret, nil
}

func (a *BlockchainReactor) createContractControlProgram(ctx context.Context, input []byte) (interface{}, error) {
	var parsed struct {
		AccountAlias string 	`json:"account_alias"`
		AccountID    string 	`json:"account_id"`
		ControlProgram string 	`json:"control_program"`
		idx string 	   		    `json:"idx"`
	}
	err := stdjson.Unmarshal(input, &parsed)
	if err != nil {
		return nil, errors.WithDetailf(httpjson.ErrBadRequest, "bad parameters for account control program")
	}

	fmt.Println("parsed:", parsed)
	accountID := parsed.AccountID
	if accountID == "" {
		acc, err := a.accounts.FindByAlias(ctx, parsed.AccountAlias)
		if err != nil {
			return nil, err
		}
		accountID = acc.ID
	}
	fmt.Println("accountID:", accountID)

	control, err := hex.DecodeString(parsed.ControlProgram)
	if err != nil {
		return nil, err
	}

	index, err := strconv.ParseUint(parsed.idx, 10, 64)
	fmt.Println("strconv idx:", index)

	controlProgram, err := a.accounts.CreateContractProgram(ctx, accountID, control, false, index, time.Time{})
	if err != nil {
		return nil, err
	}
	fmt.Println("controlProgram:", hex.EncodeToString(controlProgram))

	ret := map[string]interface{}{
		"control_program": json.HexBytes(controlProgram),
	}
	return ret, nil
}