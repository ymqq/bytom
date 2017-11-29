package commands

import (
	"context"
	stdjson "encoding/json"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/bytom/blockchain"
	"github.com/bytom/crypto/ed25519/chainkd"
	"github.com/bytom/encoding/json"
	"github.com/bytom/blockchain/txbuilder"
	"encoding/hex"
	"fmt"
	"time"
	"encoding/binary"
)

var createPubkeyCmd = &cobra.Command{
	Use:   "create-pubkey",
	Short: "Create an account pubkey",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			jww.ERROR.Println("create-pubkey need args: [account_id]")
			return
		}

		account_id := args[0]

		//create account publickey request
		type createAccountPubkeyRequest struct {
			AccountID    string `json:"account_id"`
			AccountAlias string `json:"account_alias"`
		}

		//create account publickey response
		type createAccountPubkeyResponse struct {
			Root   chainkd.XPub         `json:"root_xpub"`
			Pubkey json.HexBytes   `json:"pubkey"`
			Path   []json.HexBytes `json:"pubkey_derivation_path"`
		}

		var acc createAccountPubkeyRequest
		acc.AccountID = account_id
		acc.AccountAlias = ""

		pubresponse := createAccountPubkeyResponse{}
		client := mustRPCClient()
		client.Call(context.Background(), "/create-account-pubkey", &acc, &pubresponse)

		jww.FEEDBACK.Printf("createAccountPubkeyResponse.Root:%v\n", pubresponse.Root)
		jww.FEEDBACK.Printf("createAccountPubkeyResponse.Pubkey:%v\n", hex.EncodeToString(pubresponse.Pubkey))
		jww.FEEDBACK.Printf("createAccountPubkeyResponse.Path:%v\n", pubresponse.Path)

		//convert path idx from LittleEndian to BigEndian Uint64
		var pathdata []byte
		pathdata = pubresponse.Path[1]
		idx := binary.LittleEndian.Uint64(pathdata)
		jww.FEEDBACK.Printf("path idx:", idx)
	},
}

var createContractAccountCmd = &cobra.Command{
	Use:   "create-contract",
	Short: "Create an contract account",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 || len(args) > 3 {
			jww.ERROR.Println("create-contract need args: [account_id] [control_program] | [idx]")
			return
		}

		account_id := args[0]
		control_program := args[1]
		account_alias := ""

		var idx string
		if len(args) == 3 {
			idx = args[2]
		} else {
			idx = ""
		}

		//create contract control program
		type Parsed struct {
			AccountID    string    `json:"account_id"`
			AccountAlias string    `json:"account_alias"`
			ControlProgram string  `json:"control_program"`
			Index string  		   `json:"index"`
		}

		parse := Parsed {
			AccountAlias: account_alias,
			AccountID: account_id,
			ControlProgram: control_program,
			Index: idx,
		}

		params , _:= stdjson.Marshal(parse)

		type Ins struct {
			Type   string
			Params stdjson.RawMessage
		}

		ins := Ins {
			Type:"contract",
			Params: params,
		}

		responses := make([]interface{}, 50)
		client := mustRPCClient()
		client.Call(context.Background(), "/create-control-program", &[]Ins{ins}, &responses)
		jww.FEEDBACK.Printf("create-control-program responses:%v\n", responses)
	},
}

var BuildLockContractCmd = &cobra.Command{
	Use:   "lock",
	Short: "Build an lock contract",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 5 {
			jww.ERROR.Println("lock need args: [account id] [asset id] [account xprv] [spend amount] [control_program]")
			return
		}

		//parse the privatekey
		var xprvAccount chainkd.XPrv
		err := xprvAccount.UnmarshalText([]byte(args[2]))
		if err != nil {
			jww.ERROR.Printf("xprv unmarshal error:%v\n", xprvAccount)
			return
		}

		jww.FEEDBACK.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}}
		]}`

		buildReqStr := fmt.Sprintf(buildReqFmt, args[1], args[3], args[0], args[1], args[3], args[4])
		var buildReq blockchain.BuildRequest
		err = stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			jww.ERROR.Println("json Unmarshal error.")
			return
		}

		//generate the txbuilder template
		tpl := make([]txbuilder.Template, 1)
		client := mustRPCClient()
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		jww.FEEDBACK.Printf("tpl:%v\n", string(marshalTpl))

		// sign transaction
		err = txbuilder.Sign(context.Background(), &tpl[0], []chainkd.XPub{xprvAccount.XPub()}, "", func(_ context.Context, _ chainkd.XPub, path [][]byte, data [32]byte, _ string) ([]byte, error) {
			derived := xprvAccount.Derive(path)
			return derived.Sign(data[:]), nil
		})
		if err != nil {
			jww.ERROR.Printf("sign-transaction error. err:%v\n", err)
			return
		}
		jww.FEEDBACK.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: tpl, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		jww.FEEDBACK.Printf("submit transaction:%v\n", submitResponse)

	},
}

var BuildUnlockContractCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Build an unlock contract",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 5 {
			jww.ERROR.Println("unlock need args: [output id] [account id] [asset id] [account xprv] [amount]")
			return
		}

		//parse the privatekey
		var xprvAccount chainkd.XPrv
		err := xprvAccount.UnmarshalText([]byte(args[3]))
		if err != nil {
			jww.ERROR.Printf("xprv unmarshal error:%v\n", xprvAccount)
			return
		}

		// Build Transaction.
		jww.FEEDBACK.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":""},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}}
		]}`

		buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[2], args[4], args[1])
		var buildReq blockchain.BuildRequest
		err = stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			jww.ERROR.Printf("json Unmarshal error.")
			return
		}

		tpl := make([]txbuilder.Template, 1)
		client := mustRPCClient()
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		jww.FEEDBACK.Printf("tpl:%v\n", string(marshalTpl))

		// sign transaction
		err = txbuilder.Sign(context.Background(), &tpl[0], []chainkd.XPub{xprvAccount.XPub()}, "", func(_ context.Context, _ chainkd.XPub, path [][]byte, data [32]byte, _ string) ([]byte, error) {
			derived := xprvAccount.Derive(path)
			return derived.Sign(data[:]), nil
		})
		if err != nil {
			jww.ERROR.Printf("sign-transaction error. err:%v\n", err)
			return
		}
		jww.FEEDBACK.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: tpl, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		jww.FEEDBACK.Printf("submit transaction:%v\n", submitResponse)

	},
}