package main

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytom/blockchain"
	"github.com/bytom/blockchain/rpc"
	"github.com/bytom/blockchain/txbuilder"
	"github.com/bytom/crypto/ed25519/chainkd"
	"github.com/bytom/encoding/json"
	"github.com/bytom/env"
	"github.com/bytom/errors"
	chainjson "github.com/bytom/encoding/json"
	"github.com/bytom/protocol/vm"
	"encoding/hex"
	"encoding/binary"
	"strconv"
)

// config vars
var (
	home    = blockchain.HomeDirFromEnvironment()
	coreURL = env.String("BYTOM_URL", "http://localhost:1999")

	// build vars; initialized by the linker
	buildTag    = "?"
	buildCommit = "?"
	buildDate   = "?"
)

type command struct {
	f func(*rpc.Client, []string)
}

type grantReq struct {
	Policy    string      `json:"policy"`
	GuardType string      `json:"guard_type"`
	GuardData interface{} `json:"guard_data"`
}

var commands = map[string]*command{
	"lock":        				{buildTransaction},
	"unlock":       	 		{unLockBuildTransaction},
	"unlockPubKey":       		{unlockPublicKey},
	"unlockMultiSig":       	{unlockMultiSig},
	"unlockPubKeyHash":      	{unlockPublicKeyHash},
	"unlockRevPreimage":     	{unlockRevealPreimage},
	"unlockTradeOffer":     	{unlockTradeOffer},
	"unlockEscrow":     		{unlockEscrow},
	"unlockCallOption":     	{unlockCallOption},
	"unlockLoanCollateral":     {unlockLoanCollateral},
	"unlockPriceChanger":       {unlockPriceChanger},
	"pubkey":   				{createPubKey},
	"contract":   				{createContract},
	"create-control-program":   {createControlProgram},
	"create-account-receiver":  {createAccountReceiver},
	"sign-transactions":        {signTransactions},
	"list-transactions":        {listTransactions},
	"get-gas":        			{GetLockGas},
}

func main() {
	env.Parse()

	if len(os.Args) >= 2 && os.Args[1] == "-version" {
		var version string
		if buildTag != "?" {
			// build tag with bytom- prefix indicates official release
			version = strings.TrimPrefix(buildTag, "bytom-")
		} else {
			// version of the form rev123 indicates non-release build
			//version = rev.ID
		}
		fmt.Printf("bytomcli %s\n", version)
		return
	}

	if len(os.Args) < 2 {
		help(os.Stdout)
		os.Exit(0)
	}
	cmd := commands[os.Args[1]]
	if cmd == nil {
		fmt.Fprintln(os.Stderr, "unknown command:", os.Args[1])
		help(os.Stderr)
		os.Exit(1)
	}
	cmd.f(mustRPCClient(), os.Args[2:])
}

func mustRPCClient() *rpc.Client {
	// TODO(kr): refactor some of this cert-loading logic into bytom/blockchain
	// and use it from cored as well.
	// Note that this function, unlike maybeUseTLS in cored,
	// does not load the cert and key from env vars,
	// only from the filesystem.
	certFile := filepath.Join(home, "tls.crt")
	keyFile := filepath.Join(home, "tls.key")
	config, err := blockchain.TLSConfig(certFile, keyFile, "")
	if err == blockchain.ErrNoTLS {
		return &rpc.Client{BaseURL: *coreURL}
	} else if err != nil {
		fatalln("error: loading TLS cert:", err)
	}

	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSClientConfig:       config,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	url := *coreURL
	if strings.HasPrefix(url, "http:") {
		url = "https:" + url[5:]
	}

	return &rpc.Client{
		BaseURL: url,
		Client:  &http.Client{Transport: t},
	}
}

func fatalln(v ...interface{}) {
	fmt.Printf("%v", v)
	os.Exit(2)
}

func dieOnRPCError(err error, prefixes ...interface{}) {
	if err == nil {
		return
	}

	if len(prefixes) > 0 {
		fmt.Fprintln(os.Stderr, prefixes...)
	}

	if msgErr, ok := errors.Root(err).(rpc.ErrStatusCode); ok && msgErr.ErrorData != nil {
		fmt.Fprintln(os.Stderr, "RPC error:", msgErr.ErrorData.ChainCode, msgErr.ErrorData.Message)
		if msgErr.ErrorData.Detail != "" {
			fmt.Fprintln(os.Stderr, "Detail:", msgErr.ErrorData.Detail)
		}
	} else {
		fmt.Fprintln(os.Stderr, "RPC error:", err)
	}

	os.Exit(2)
}

func help(w io.Writer) {
	fmt.Fprintln(w, "usage: corectl [-version] [command] [arguments]")
	fmt.Fprint(w, "\nThe commands are:\n\n")
	for name := range commands {
		fmt.Fprintln(w, "\t", name)
	}
	fmt.Fprint(w, "\nFlags:\n")
	fmt.Fprintln(w, "\t-version   print version information")
	fmt.Fprintln(w)
}

func sign(client *rpc.Client, tpl []txbuilder.Template, password string) []txbuilder.Template {
	type param struct {
		Auth string
		Txs  []txbuilder.Template `json:"transactions"`
	}

	in := param{Txs: tpl, Auth: password}
	response := make([]txbuilder.Template, len(tpl))
	client.Call(context.Background(), "/sign-transactions", &in, &response)
	fmt.Println(response)
	return response
}

func buildTransaction(client *rpc.Client, args []string) {
	if len(args) != 5 {
		fatalln("error: need args: [account id] [asset id] [password] [spend amount] [control_program] \n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")


	buildReqFmt := `
		{"actions": [
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"},
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}}
		]}`

	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[1], args[3], args[0], args[1], args[3], args[4])

	/*
	//expires_at := time.Time()
	buildReqFmt := `
		{"actions": [
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "control_receiver", "asset_id": "%s", "amount": %s, "receiver":{"control_program": "%v"}, "reference_data": {}}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[1], args[3], args[4], args[1], args[3], args[5])
	*/

	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
		os.Exit(1)
	}
	fmt.Println("buildReq:", buildReq)

	//generate the txbuilder template
	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	// sign transaction
	signResp := sign(client, tpl, args[2])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// submit-transaction-Spend_account
	var submitResponse interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
	fmt.Printf("submit transaction:%v\n", submitResponse)
}

func unLockBuildTransaction(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [account id] [asset id] [password] [amount] [client_token] [args1] [args2]...\n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")

	buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[1])
	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
	}

	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	//set arguments
	if len(args) > 6 {		//judge the args whether include args1 args2 ...
		var data [][]byte
		var tmp []byte
		i := 7
		for i <= len(args) {
			tmp, _ = hex.DecodeString(args[i-1])
			data = append(data, tmp)
			i = i + 1
		}
		fmt.Println("data:", data)

		err = txbuilder.SetWitnessArguments(&tpl[0], data)
		if err != nil {
			fmt.Printf("SetWitnessArguments error. err:%v\n", err)
			os.Exit(1)
		}
	}

	// sign transaction
	signResp := sign(client, tpl, args[3])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// submit-transaction-Spend_account
	var submitResponse interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
	fmt.Printf("submit transaction:%v\n", submitResponse)

}

//unlock contract of LockWithPublicKey
func unlockPublicKey(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [account id] [asset id] [password] [amount] [client_token] [root_pubkey] [path1] [path2]...\n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")

	buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[1])
	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
		os.Exit(1)
	}

	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	//set clause arguments
	var root chainkd.XPub
	pub , _:= hex.DecodeString(args[6])
	copy(root[:], pub[:])

	var path []chainjson.HexBytes
	path1, _ := hex.DecodeString(args[7])
	path2, _ := hex.DecodeString(args[8])
    path = append(path, path1)
	path = append(path, path2)

	fmt.Printf("createPubKey.Root:%v\n", root)
	fmt.Printf("createPubKey.Path:%v\n", path)

	var totalroot []chainkd.XPub
	var totalpath [][]chainjson.HexBytes
	totalroot = append(totalroot, root)
	totalpath = append(totalpath, path)

	var si txbuilder.SigningInstruction
	err = si.AddRawTxSigWitness(totalroot, totalpath, 1)
	if err != nil {
		fmt.Printf("AddRawTxSigWitness return error.")
		os.Exit(1)
	}

	for i, inp := range tpl[0].Transaction.InputIDs{
		fmt.Printf("tpl[0].Transaction.InputIDs[%d]:%v\n", i, inp)
	}

	length := len(tpl[0].SigningInstructions)
	if length <= 0 {
		length = 1
		tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
		tpl[0].SigningInstructions[length - 1].Position = 0
	} else {
		tpl[0].SigningInstructions[0] = &si
	}
	fmt.Println("length of tpl[0].SigningInstructions:", length)
	for i, _ := range tpl[0].SigningInstructions{
		fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
	}

	// sign transaction
	signResp := sign(client, tpl, args[3])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// submit-transaction-Spend_account
	var submitResponse interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
	fmt.Printf("submit transaction:%v\n", submitResponse)

}

//unlock contract of LockWithMultiSig
func unlockMultiSig(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [account id] [asset id] [password] [amount] [client_token] [root_pubkey1] [path11] [path12]" +
			" [root_pubkey2] [path21] [path22]...\n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")

	buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[1])
	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
		os.Exit(1)
	}

	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	//set clause arguments
	var root chainkd.XPub
	var path11 []chainjson.HexBytes
	var path22 []chainjson.HexBytes
	var totalroot []chainkd.XPub
	var totalpath [][]chainjson.HexBytes

	//add the arguments1
	pub , _:= hex.DecodeString(args[6])
	copy(root[:], pub[:])

	path1, _ := hex.DecodeString(args[7])
	path2, _ := hex.DecodeString(args[8])
	path11 = append(path11, path1)
	path11 = append(path11, path2)

	fmt.Printf("root_pubkey1:%v\n", root)
	fmt.Printf("pubkey_path1:%v\n", path11)

	totalroot = append(totalroot, root)
	totalpath = append(totalpath, path11)

	//add the arguments2
	pub , _= hex.DecodeString(args[9])
	copy(root[:], pub[:])

	path1, _ = hex.DecodeString(args[10])
	path2, _ = hex.DecodeString(args[11])
	path22 = append(path22, path1)
	path22 = append(path22, path2)

	fmt.Printf("root_pubkey2:%v\n", root)
	fmt.Printf("pubkey_path2:%v\n", path22)

	totalroot = append(totalroot, root)
	totalpath = append(totalpath, path22)

	//Add args into RawTxSigWitness
	var si txbuilder.SigningInstruction
	err = si.AddRawTxSigWitness(totalroot, totalpath, 2)
	if err != nil {
		fmt.Printf("AddRawTxSigWitness return error.")
		os.Exit(1)
	}

	for i, inp := range tpl[0].Transaction.InputIDs{
		fmt.Printf("tpl[0].Transaction.InputIDs[%d]:%v\n", i, inp)
	}

	length := len(tpl[0].SigningInstructions)
	if length <= 0 {
		length = 1
		tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
		tpl[0].SigningInstructions[length - 1].Position = 0
	} else {
		tpl[0].SigningInstructions[0] = &si
	}
	fmt.Println("length of tpl[0].SigningInstructions:", length)
	for i, _ := range tpl[0].SigningInstructions{
		fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
	}

	// sign transaction
	signResp := sign(client, tpl, args[3])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// submit-transaction-Spend_account
	var submitResponse interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
	fmt.Printf("submit transaction:%v\n", submitResponse)

}

//unlock contract of LockWithPublicKeyHash
func unlockPublicKeyHash(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [account id] [asset id] [password] [amount] [client_token] [pubkey] [root_pubkey] [path1] [path2]...\n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")

	buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[1])
	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
		os.Exit(1)
	}

	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	//set clause arguments
	var si txbuilder.SigningInstruction
	var data []chainjson.HexBytes
	pubkey , _ := hex.DecodeString(args[6])
	data = append(data, pubkey)
	si.AddDataWitness(data)

	var root chainkd.XPub
	pub , _:= hex.DecodeString(args[7])
	copy(root[:], pub[:])

	var path []chainjson.HexBytes
	path1, _ := hex.DecodeString(args[8])
	path2, _ := hex.DecodeString(args[9])
	path = append(path, path1)
	path = append(path, path2)

	fmt.Printf("createPubKey.Root:%v\n", root)
	fmt.Printf("createPubKey.Path:%v\n", path)

	var totalroot []chainkd.XPub
	var totalpath [][]chainjson.HexBytes
	totalroot = append(totalroot, root)
	totalpath = append(totalpath, path)

	err = si.AddRawTxSigWitness(totalroot, totalpath, 1)
	if err != nil {
		fmt.Printf("AddRawTxSigWitness return error.")
		os.Exit(1)
	}

	for i, inp := range tpl[0].Transaction.InputIDs{
		fmt.Printf("tpl[0].Transaction.InputIDs[%d]:%v\n", i, inp)
	}

	length := len(tpl[0].SigningInstructions)
	if length <= 0 {
		length = 1
		tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
		tpl[0].SigningInstructions[length - 1].Position = 0
	} else {
		tpl[0].SigningInstructions[0] = &si
	}
	fmt.Println("length of tpl[0].SigningInstructions:", length)
	for i, _ := range tpl[0].SigningInstructions{
		fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
	}

	// sign transaction
	signResp := sign(client, tpl, args[3])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// submit-transaction-Spend_account
	var submitResponse interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
	fmt.Printf("submit transaction:%v\n", submitResponse)
}

//unlock contract of RevealPreimage
func unlockRevealPreimage(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [account id] [asset id] [password] [amount] [client_token] [args1] [args2]...\n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")

	buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[1])
	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
		os.Exit(1)
	}

	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	//Add args into dataWitness
	if len(args) > 6 {		//judge the args whether include args1 args2 ...
		var data []chainjson.HexBytes
		var tmp chainjson.HexBytes
		i := 7
		for i <= len(args) {
			tmp, _ = hex.DecodeString(args[i-1])
			data = append(data, tmp)
			i = i + 1
		}
		fmt.Println("data:", data)

		var si txbuilder.SigningInstruction
		si.AddDataWitness(data)

		length := len(tpl[0].SigningInstructions)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length - 1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}

		fmt.Println("length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}
	}

	// sign transaction
	signResp := sign(client, tpl, args[3])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// submit-transaction-Spend_account
	var submitResponse interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
	fmt.Printf("submit transaction:%v\n", submitResponse)
}

//unlock contract of TradeOffer
func unlockTradeOffer(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [out account id] [asset id] [password] [amount] [client_token]" +
			" [clause selector] ([inner asset id] [inner amount] [inner account id] [recv control_program]) " +
			" | ([root_pubkey] [path1] [path2])\n")
	}

	//the clause of TradeOffer
	trade := "00000000"
	cancel := "13000000"
	ending := "1a000000"

	fmt.Println("clause selector:", args[6])
	if args[6] == trade {	//select clasue trade
		// Build Transaction.
		// notice the action order: out_spend - inner_ctl - inner_spend - out_ctl
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type": "spend_account_unspent_output", "output_id": "%s", "reference_data": {}, "client_token": "%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}},
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt,
			args[0], args[5],
			args[7], args[8], args[10],
			args[7], args[8], args[9],
			args[1],
			args[2], args[4], args[1])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Printf("json Unmarshal error.\n")
			os.Exit(1)
		}
		fmt.Println("buildReq:", buildReq)

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//Add args into dataWitness
		var data []chainjson.HexBytes
		var tmp chainjson.HexBytes

		tmp, _ = hex.DecodeString(args[6])
		data = append(data, tmp)
		fmt.Println("data:", data)

		var si txbuilder.SigningInstruction
		si.AddDataWitness(data)

		length := len(tpl[0].SigningInstructions)
		fmt.Println("before length of tpl[0].SigningInstructions:", length)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length-1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}

		fmt.Println("after length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}

		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)

	} else if args[6] == cancel {	//select clasue cancel
		// Build Transaction.
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[1])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Printf("json Unmarshal error.")
			os.Exit(1)
		}

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//get clause paramenter
		var root chainkd.XPub
		pub , _:= hex.DecodeString(args[7])
		copy(root[:], pub[:])

		var path []chainjson.HexBytes
		path1, _ := hex.DecodeString(args[8])
		path2, _ := hex.DecodeString(args[9])
		path = append(path, path1)
		path = append(path, path2)

		fmt.Printf("createPubKey.Root:%v\n", root)
		fmt.Printf("createPubKey.Path:%v\n", path)

		var totalroot []chainkd.XPub
		var totalpath [][]chainjson.HexBytes
		totalroot = append(totalroot, root)
		totalpath = append(totalpath, path)

		//add rootpubkey and path
		var si txbuilder.SigningInstruction
		err = si.AddRawTxSigWitness(totalroot, totalpath, 1)
		if err != nil {
			fmt.Printf("AddRawTxSigWitness return error.")
			os.Exit(1)
		}

		//add clause selector
		var data []chainjson.HexBytes
		tmp, _ := hex.DecodeString(args[6])
		data = append(data, tmp)
		si.AddDataWitness(data)

		for i, inp := range tpl[0].Transaction.InputIDs{
			fmt.Printf("tpl[0].Transaction.InputIDs[%d]:%v\n", i, inp)
		}

		length := len(tpl[0].SigningInstructions)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length - 1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}
		fmt.Println("length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}
		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)
	} else if args[6] == ending {	//no clause selected, ending exit
		fmt.Printf("no clause was selected in this program, ending exit!!!\n")
		os.Exit(0)
	} else {
		fmt.Printf("selected clause [%v] error\n", args[6])
		fmt.Printf("clause must in set:[%v, %v, %v]\n", trade, cancel, ending)
		os.Exit(1)
	}
}

//unlock contract of Escrow
func unlockEscrow(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [out control_program] [asset id] [password] [amount] [client_token]" +
			" [clause selector] [root_pubkey] [path1] [path2] [account id]\n")
	}

	//the clause of TradeOffer
	approve := "00000000"
	reject := "1b000000"
	ending := "2a000000"

	fmt.Println("clause selector:", args[6])
	if args[6] == approve || args[6] == reject {	//select clasue approve or reject
		// Build Transaction.
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}},
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1], args[10])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Printf("json Unmarshal error.")
			os.Exit(1)
		}

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//get clause paramenter
		var root chainkd.XPub
		pub , _:= hex.DecodeString(args[7])
		copy(root[:], pub[:])

		var path []chainjson.HexBytes
		path1, _ := hex.DecodeString(args[8])
		path2, _ := hex.DecodeString(args[9])
		path = append(path, path1)
		path = append(path, path2)

		fmt.Printf("createPubKey.Root:%v\n", root)
		fmt.Printf("createPubKey.Path:%v\n", path)

		var totalroot []chainkd.XPub
		var totalpath [][]chainjson.HexBytes
		totalroot = append(totalroot, root)
		totalpath = append(totalpath, path)

		//add rootpubkey and path
		var si txbuilder.SigningInstruction
		err = si.AddRawTxSigWitness(totalroot, totalpath, 1)
		if err != nil {
			fmt.Printf("AddRawTxSigWitness return error.")
			os.Exit(1)
		}

		//add clause selector
		var data []chainjson.HexBytes
		tmp, _ := hex.DecodeString(args[6])
		data = append(data, tmp)
		si.AddDataWitness(data)

		length := len(tpl[0].SigningInstructions)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length - 1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}
		fmt.Println("length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}
		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)
	} else if args[6] == ending {	//no clause selected, ending exit
		fmt.Printf("no clause was selected in this program, ending exit!!!\n")
		os.Exit(0)
	} else {
		fmt.Printf("selected clause [%v] error\n", args[6])
		fmt.Printf("clause must in set:[%v, %v, %v]\n", approve, reject, ending)
		os.Exit(1)
	}
}

//unlock contract of CallOption
func unlockCallOption(client *rpc.Client, args []string) {

}

//unlock contract of LoanCollateral
func unlockLoanCollateral(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [out control_program] [asset id] [password] [amount] [client_token]" +
			" [clause selector] ([inner asset id] [inner amount] [inner account id] [recv control_program]) \n")
	}

	//the clause of TradeOffer
	repay := "00000000"
	funcdefault := "1c000000"
	ending := "28000000"

	fmt.Println("clause selector:", args[6])
	if args[6] == repay { //select clasue repay
		// Build Transaction.
		// notice the action order: out_spend - inner_ctl - inner_spend - out_ctl
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type": "spend_account_unspent_output", "output_id": "%s", "reference_data": {}, "client_token": "%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data":{}},
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt,
			args[0], args[5],
			args[7], args[8], args[10],
			args[2], args[4], args[1],
			args[7], args[8], args[9])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Printf("json Unmarshal error.\n")
			os.Exit(1)
		}
		fmt.Println("buildReq:", buildReq)

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//Add args into dataWitness
		var data []chainjson.HexBytes
		var tmp chainjson.HexBytes

		tmp, _ = hex.DecodeString(args[6])
		data = append(data, tmp)
		fmt.Println("data:", data)

		var si txbuilder.SigningInstruction
		si.AddDataWitness(data)

		length := len(tpl[0].SigningInstructions)
		fmt.Println("before length of tpl[0].SigningInstructions:", length)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length-1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}

		fmt.Println("after length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}

		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)

	} else if args[6] == funcdefault {	//select clasue default
		// Build Transaction.
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Printf("json Unmarshal error.")
			os.Exit(1)
		}

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//add clause selector
		var si txbuilder.SigningInstruction
		var data []chainjson.HexBytes
		tmp, _ := hex.DecodeString(args[6])
		data = append(data, tmp)
		si.AddDataWitness(data)

		length := len(tpl[0].SigningInstructions)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length - 1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}
		fmt.Println("length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}
		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)

	} else if args[6] == ending {	//no clause selected, ending exit
		fmt.Printf("no clause was selected in this program, ending exit!!!\n")
		os.Exit(0)
	} else {
		fmt.Printf("selected clause [%v] error\n", args[6])
		fmt.Printf("clause must in set:[%v, %v, %v]\n", repay, funcdefault, ending)
		os.Exit(1)
	}
}

func createPubKey(client *rpc.Client, args []string) {
	if len(args) != 1 {
		fatalln("error:createContract need args: [account_id]\n")
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
		Pubkey chainjson.HexBytes   `json:"pubkey"`
		Path   []chainjson.HexBytes `json:"pubkey_derivation_path"`
	}

	var acc createAccountPubkeyRequest

	acc.AccountID = account_id
	acc.AccountAlias = ""
	fmt.Println("accReq.AccountID:", acc.AccountID)

	pubresponse := createAccountPubkeyResponse{}
	client.Call(context.Background(), "/create-account-pubkey", &acc, &pubresponse)
	fmt.Printf("createAccountPubkeyResponse.Root:%v\n", pubresponse.Root)
	fmt.Printf("createAccountPubkeyResponse.Pubkey:%v\n", hex.EncodeToString(pubresponse.Pubkey))
	fmt.Printf("createAccountPubkeyResponse.Path:%v\n", pubresponse.Path)

	var result []string
	for i, p := range pubresponse.Path{
		fmt.Printf("path[%d]:%v\n", i, hex.EncodeToString(p))
		result = append(result, hex.EncodeToString(p))
	}
	fmt.Println("result:", result)

	//convert path idx from LittleEndian to BigEndian Uint64
	var pathdata []byte
	pathdata = pubresponse.Path[1]
	idx := binary.LittleEndian.Uint64(pathdata)
	fmt.Println("path idx:", idx)
}

func createContract(client *rpc.Client, args []string){
	if len(args) < 2 || len(args) > 3 {
		fatalln("error:createContract need args: [account_id] [control_program] | [idx]\n")
	}

	account_id := args[0]
	control_program := args[1]
	account_alias := ""

	//when the contract not contain publickey, the idx will be not related to publickey, the args idx can be no-exist
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

	fmt.Println("Parsed:", parse)
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
	client.Call(context.Background(), "/create-control-program", &[]Ins{ins}, &responses)
	fmt.Printf("create-control-program responses:%v\n", responses)

}

func createControlProgram(client *rpc.Client, args []string) {
	var account_alias string
	var account_id string

	if len(args) == 1 {
		//alias is ALI:+[name], create account alias is name, the command can't be used
		account_alias = ""
		account_id = args[0]
	}else if len(args) == 2 {
		account_alias = args[1]
		account_id = args[0]
	}else{
		fatalln("error:createControlProgram need args: [account_id] | [account_alias]\n")
	}

	fmt.Println("account_alias:", account_alias)
	fmt.Println("account_id:", account_id)

	type Parsed struct {
		AccountAlias string `json:"account_alias"`
		AccountID    string `json:"account_id"`
	}
	parse := Parsed {
		AccountAlias: account_alias,
		AccountID: account_id,
	}

	params , _:= stdjson.Marshal(parse)
	//fmt.Println("params:", params)

	type Ins struct {
		Type   string
		Params stdjson.RawMessage
	}
	var ins Ins
	//TODO:undefined arguments to ins
	ins = Ins {
		Type:"account",
		Params: params,
		}

	responses := make([]interface{}, 50)
	client.Call(context.Background(), "/create-control-program", &[]Ins{ins}, &responses)
	fmt.Printf("responses:%v\n", responses)
}

func createAccountReceiver(client *rpc.Client, args []string) {
	var account_alias string
	var account_id string
	var expires_at time.Time

	if len(args) == 1 {
		//alias is ALI:+[name], create account alias is name, the command can't be used
		account_alias = ""
		account_id = args[0]
		expires_at = time.Time{}
	}else{
		fatalln("error:createControlProgram need args: [account_id] \n")
	}

	type Ins struct {
		AccountID    string    `json:"account_id"`
		AccountAlias string    `json:"account_alias"`
		ExpiresAt    time.Time `json:"expires_at"`
	}
	var ins Ins
	//TODO:undefined argument to ExpiresAt
	ins.AccountID = account_id
	ins.AccountAlias = account_alias
	ins.ExpiresAt = expires_at
	fmt.Println("Ins:", ins)

	responses := make([]interface{}, 50)
	client.Call(context.Background(), "/create-account-receiver", &[]Ins{ins}, &responses)
	fmt.Printf("responses:%v\n", responses)
}

func listTransactions(client *rpc.Client, args []string) {
	if len(args) != 0 {
		fatalln("error:listTransactions not use args")
	}
	type requestQuery struct {
		Filter       string        `json:"filter,omitempty"`
		FilterParams []interface{} `json:"filter_params,omitempty"`
		SumBy        []string      `json:"sum_by,omitempty"`
		PageSize     int           `json:"page_size"`
		AscLongPoll  bool          `json:"ascending_with_long_poll,omitempty"`
		Timeout      json.Duration `json:"timeout"`
		After        string        `json:"after"`
		StartTimeMS  uint64        `json:"start_time,omitempty"`
		EndTimeMS    uint64        `json:"end_time,omitempty"`
		TimestampMS  uint64        `json:"timestamp,omitempty"`
		Type         string        `json:"type"`
		Aliases      []string      `json:"aliases,omitempty"`
	}
	var in requestQuery
	after := in.After
	out := in
	out.After = after
	client.Call(context.Background(), "/list-transactions", &[]requestQuery{in}, nil)
}



func signTransactions(client *rpc.Client, args []string) {

	// sign-transaction
	type param struct {
		Auth  string
		Txs   []*txbuilder.Template `json:"transactions"`
		XPubs chainkd.XPub          `json:"xpubs"`
		XPrv  chainkd.XPrv          `json:"xprv"`
	}

	var in param
	var xprv chainkd.XPrv
	var xpub chainkd.XPub
	var err error

	if len(args) == 3 {
		err = xpub.UnmarshalText([]byte(args[1]))
		if err == nil {
			fmt.Printf("xpub:%v\n", xpub)
		} else {
			fmt.Printf("xpub unmarshal error:%v\n", xpub)
		}
		in.XPubs = xpub
		in.Auth = args[2]

	} else if len(args) == 2 {
		err = xprv.UnmarshalText([]byte(args[1]))
		if err == nil {
			fmt.Printf("xprv:%v\n", xprv)
		} else {
			fmt.Printf("xprv unmarshal error:%v\n", xprv)
		}
		in.XPrv = xprv

	} else {
		fatalln("error: signTransaction need args: [tpl file name] [xPub] [password], 3 args not equal"+
			"or [tpl file name] [xPrv], 2 args not equal ", len(args))
	}

	var tpl txbuilder.Template
	file, _ := os.Open(args[0])
	tpl_byte := make([]byte, 10000)
	file.Read(tpl_byte)
	fmt.Printf("tpl_byte:%v\n", string(tpl_byte))
	err = stdjson.Unmarshal(bytes.Trim(tpl_byte, "\x00"), &tpl)
	fmt.Printf("tpl:%v, err:%v\n", tpl, err)
	in.Txs = []*txbuilder.Template{&tpl}

	var response []interface{} = make([]interface{}, 1)
	client.Call(context.Background(), "/sign-transactions", &in, &response)
	fmt.Printf("sign response:%v\n", response)
}

//unlock contract of PriceChanger
func unlockPriceChanger(client *rpc.Client, args []string) {
	if len(args) < 6 {
		fatalln("error: need args: [output id] [(control_program) | (out account id)] [asset id] [password] [amount] [client_token]" +
			" [clause selector] ([new amount] [new asset id] [root_pubkey] [path1] [path2] " +
			" | ([inner asset id] [inner amount] [inner account id] [recv control_program] )\n")
	}

	//the clause of PriceChanger
	changePrice := "00000000"
	redeem := "33000000"
	ending := "3d000000"

	fmt.Println("clause selector:", args[6])
	if args[6] == changePrice {	//select clasue changePrice
		// Build Transaction.
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type":"spend_account_unspent_output", "output_id":"%s", "reference_data":{}, "client_token":"%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data":{}}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[5], args[2], args[4], args[1])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Println("json Unmarshal error. ", err)
			os.Exit(1)
		}

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//get clause paramenter
		var root chainkd.XPub
		pub , _:= hex.DecodeString(args[9])
		copy(root[:], pub[:])

		var path []chainjson.HexBytes
		path1, _ := hex.DecodeString(args[10])
		path2, _ := hex.DecodeString(args[11])
		path = append(path, path1)
		path = append(path, path2)

		fmt.Printf("createPubKey.Root:%v\n", root)
		fmt.Printf("createPubKey.Path:%v\n", path)

		var totalroot []chainkd.XPub
		var totalpath [][]chainjson.HexBytes
		totalroot = append(totalroot, root)
		totalpath = append(totalpath, path)

		//add clause paramenter into program
		var si txbuilder.SigningInstruction

		//add newAmount and newAsset (DataWitness)
		var newdata []chainjson.HexBytes
		amount, err := strconv.ParseInt(args[7], 10, 64)
		newAmount := vm.Int64Bytes(amount)
		newdata = append(newdata, newAmount)
		newAsset, _ := hex.DecodeString(args[8])
		newdata = append(newdata, newAsset)
		si.AddDataWitness(newdata)

		//add rootpubkey and path (RawTxSigWitness)
		err = si.AddRawTxSigWitness(totalroot, totalpath, 1)
		if err != nil {
			fmt.Printf("AddRawTxSigWitness return error.")
			os.Exit(1)
		}

		//add clause selector (DataWitness)
		var data []chainjson.HexBytes
		tmp, _ := hex.DecodeString(args[6])
		data = append(data, tmp)
		si.AddDataWitness(data)

		for i, inp := range tpl[0].Transaction.InputIDs{
			fmt.Printf("tpl[0].Transaction.InputIDs[%d]:%v\n", i, inp)
		}

		length := len(tpl[0].SigningInstructions)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length - 1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}
		fmt.Println("length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}
		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)

	} else if args[6] == redeem {	//select clasue redeem
		// Build Transaction.
		// notice the action order: out_spend - inner_ctl - inner_spend - out_ctl
		fmt.Printf("To build transaction:\n")
		buildReqFmt := `
		{"actions": [
			{"type": "spend_account_unspent_output", "output_id": "%s", "reference_data": {}, "client_token": "%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}},
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s", "reference_data":{}}
		]}`
		buildReqStr := fmt.Sprintf(buildReqFmt,
			args[0], args[5],
			args[7], args[8], args[10],
			args[7], args[8], args[9],
			args[2], args[4], args[1])
		var buildReq blockchain.BuildRequest
		err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
		if err != nil {
			fmt.Printf("json Unmarshal error.\n")
			os.Exit(1)
		}
		fmt.Println("buildReq:", buildReq)

		tpl := make([]txbuilder.Template, 1)
		client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
		marshalTpl, _ := stdjson.Marshal(tpl[0])
		if tpl[0].Transaction == nil {
			fmt.Printf("tpl:%v\n", string(marshalTpl))
			fmt.Printf("build transaction error.\n")
			os.Exit(1)
		}
		fmt.Printf("tpl:%v\n", string(marshalTpl))

		//Add args into dataWitness
		var data []chainjson.HexBytes
		var tmp chainjson.HexBytes

		tmp, _ = hex.DecodeString(args[6])
		data = append(data, tmp)
		fmt.Println("data:", data)

		var si txbuilder.SigningInstruction
		si.AddDataWitness(data)

		length := len(tpl[0].SigningInstructions)
		fmt.Println("before length of tpl[0].SigningInstructions:", length)
		if length <= 0 {
			length = 1
			tpl[0].SigningInstructions = append(tpl[0].SigningInstructions, &si)
			tpl[0].SigningInstructions[length-1].Position = 0
		} else {
			tpl[0].SigningInstructions[0] = &si
		}

		fmt.Println("after length of tpl[0].SigningInstructions:", length)
		for i, _ := range tpl[0].SigningInstructions{
			fmt.Printf("tpl[0].SigningInstructions[%d].postion:%v\n", i, tpl[0].SigningInstructions[i].Position)
		}

		// sign transaction
		signResp := sign(client, tpl, args[3])
		fmt.Printf("sign tpl:%v\n", tpl[0])

		// submit-transaction-Spend_account
		var submitResponse interface{}
		submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
		client.Call(context.Background(), "/submit-transaction", submitArg, &submitResponse)
		fmt.Printf("submit transaction:%v\n", submitResponse)

	} else if args[6] == ending {	//no clause selected, ending exit
		fmt.Printf("no clause was selected in this program, ending exit!!!\n")
		os.Exit(0)
	} else {
		fmt.Printf("selected clause [%v] error\n", args[6])
		fmt.Printf("clause must in set:[%v, %v, %v]\n", changePrice, redeem, ending)
		os.Exit(1)
	}
}

func GetLockGas(client *rpc.Client, args []string) {
	if len(args) != 5 {
		fatalln("error: need args: [account id] [asset id] [password] [spend amount] [control_program] \n")
	}

	// Build Transaction.
	fmt.Printf("To build transaction:\n")


	buildReqFmt := `
		{"actions": [
			{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":20000000, "account_id": "%s"},
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%v", "reference_data": {}}
		]}`

	buildReqStr := fmt.Sprintf(buildReqFmt, args[0], args[1], args[3], args[0], args[1], args[3], args[4])

	/*
	//expires_at := time.Time()
	buildReqFmt := `
		{"actions": [
			{"type": "spend_account", "asset_id": "%s", "amount": %s, "account_id": "%s"},
			{"type": "control_receiver", "asset_id": "%s", "amount": %s, "receiver":{"control_program": "%v"}, "reference_data": {}}
		]}`
	buildReqStr := fmt.Sprintf(buildReqFmt, args[1], args[3], args[4], args[1], args[3], args[5])
	*/

	var buildReq blockchain.BuildRequest
	err := stdjson.Unmarshal([]byte(buildReqStr), &buildReq)
	if err != nil {
		fmt.Printf("json Unmarshal error.")
		os.Exit(1)
	}
	fmt.Println("buildReq:", buildReq)

	//generate the txbuilder template
	tpl := make([]txbuilder.Template, 1)
	client.Call(context.Background(), "/build-transaction", []*blockchain.BuildRequest{&buildReq}, &tpl)
	marshalTpl, _ := stdjson.Marshal(tpl[0])
	fmt.Printf("tpl:%v\n", string(marshalTpl))

	// sign transaction
	signResp := sign(client, tpl, args[2])
	fmt.Printf("sign tpl:%v\n", tpl[0])

	// calculate gas
	var Response interface{}
	submitArg := blockchain.SubmitArg{Transactions: signResp, Wait: json.Duration{Duration: time.Duration(1000000)}, WaitUntil: "none"}
	client.Call(context.Background(), "/calculate-gas", submitArg, &Response)
	fmt.Printf("calculated gas:%v\n", Response)
}