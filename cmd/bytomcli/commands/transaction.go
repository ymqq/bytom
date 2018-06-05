package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"

	"github.com/bytom/api"
	"github.com/bytom/blockchain/txbuilder"
	chainjson "github.com/bytom/encoding/json"
	"github.com/bytom/protocol/bc/types"
	"github.com/bytom/util"
)

func init() {
	buildTransactionCmd.PersistentFlags().StringVarP(&buildType, "type", "t", "", "transaction type, valid types: 'issue', 'spend'")
	buildTransactionCmd.PersistentFlags().StringVarP(&receiverProgram, "receiver", "r", "", "program of receiver")
	buildTransactionCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "address of receiver")
	buildTransactionCmd.PersistentFlags().StringVarP(&btmGas, "gas", "g", "20000000", "gas of this transaction")
	buildTransactionCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty print json result")
	buildTransactionCmd.PersistentFlags().BoolVar(&alias, "alias", false, "use alias build transaction")

	lockContractTransactionCmd.PersistentFlags().StringVarP(&btmGas, "gas", "g", "20000000", "program of receiver")
	lockContractTransactionCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty print json result")
	lockContractTransactionCmd.PersistentFlags().BoolVar(&alias, "alias", false, "use alias build transaction")

	unlockContractTransactionCmd.PersistentFlags().StringVarP(&btmGas, "gas", "g", "20000000", "program of receiver")
	unlockContractTransactionCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty print json result")
	unlockContractTransactionCmd.PersistentFlags().BoolVar(&alias, "alias", false, "use alias build transaction")
	unlockContractTransactionCmd.PersistentFlags().StringVarP(&contractName, "contract-name", "c", "",
		"name of template contract, currently supported: 'LockWithPublicKey', 'LockWithMultiSig', 'LockWithPublicKeyHash',"+
			"\n\t\t\t       'RevealPreimage', 'TradeOffer', 'Escrow', 'CallOption', 'LoanCollateral'")

	signTransactionCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password of the account which sign these transaction(s)")
	signTransactionCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty print json result")

	listTransactionsCmd.PersistentFlags().StringVar(&txID, "id", "", "transaction id")
	listTransactionsCmd.PersistentFlags().StringVar(&account, "account_id", "", "account id")
	listTransactionsCmd.PersistentFlags().BoolVar(&detail, "detail", false, "list transactions details")
}

var (
	buildType       = ""
	btmGas          = ""
	receiverProgram = ""
	address         = ""
	password        = ""
	pretty          = false
	alias           = false
	txID            = ""
	account         = ""
	detail          = false
	contractName    = ""
)

var buildIssueReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "issue", "asset_id": "%s", "amount": %s},
		{"type": "control_address", "asset_id": "%s", "amount": %s, "address": "%s"}
	]}`

var buildIssueReqFmtByAlias = `
	{"actions": [
		{"type": "spend_account", "asset_alias": "BTM", "amount":%s, "account_alias": "%s"},
		{"type": "issue", "asset_alias": "%s", "amount": %s},
		{"type": "control_address", "asset_alias": "%s", "amount": %s, "address": "%s"}
	]}`

var buildSpendReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "spend_account", "asset_id": "%s","amount": %s,"account_id": "%s"},
		{"type": "control_receiver", "asset_id": "%s", "amount": %s, "receiver":{"control_program": "%s"}}
	]}`

var buildSpendReqFmtByAlias = `
	{"actions": [
		{"type": "spend_account", "asset_alias": "BTM", "amount":%s, "account_alias": "%s"},
		{"type": "spend_account", "asset_alias": "%s","amount": %s,"account_alias": "%s"},
		{"type": "control_receiver", "asset_alias": "%s", "amount": %s, "receiver":{"control_program": "%s"}}
	]}`

var buildRetireReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "spend_account", "asset_id": "%s","amount": %s,"account_id": "%s"},
		{"type": "retire", "asset_id": "%s","amount": %s,"account_id": "%s"}
	]}`

var buildRetireReqFmtByAlias = `
	{"actions": [
		{"type": "spend_account", "asset_alias": "BTM", "amount":%s, "account_alias": "%s"},
		{"type": "spend_account", "asset_alias": "%s","amount": %s,"account_alias": "%s"},
		{"type": "retire", "asset_alias": "%s","amount": %s,"account_alias": "%s"}
	]}`

var buildControlAddressReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "spend_account", "asset_id": "%s","amount": %s,"account_id": "%s"},
		{"type": "control_address", "asset_id": "%s", "amount": %s,"address": "%s"}
	]}`

var buildControlAddressReqFmtByAlias = `
	{"actions": [
		{"type": "spend_account", "asset_alias": "BTM", "amount":%s, "account_alias": "%s"},
		{"type": "spend_account", "asset_alias": "%s","amount": %s, "account_alias": "%s"},
		{"type": "control_address", "asset_alias": "%s", "amount": %s,"address": "%s"}
	]}`

var buildContractReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "spend_account", "asset_id": "%s","amount": %s,"account_id": "%s"},
		{"type": "control_program", "asset_id": "%s", "amount": %s, "control_program": "%s", "reference_data": {}}
	]}`

var buildContractReqFmtByAlias = `
	{"actions": [
		{"type": "spend_account", "asset_alias": "btm", "amount":%s, "account_alias": "%s"},
		{"type": "spend_account", "asset_alias": "%s","amount": %s,"account_alias": "%s"},
		{"type": "control_program", "asset_alias": "%s", "amount": %s, "control_program": "%s", "reference_data": {}}
	]}`

var buildTransactionCmd = &cobra.Command{
	Use:   "build-transaction <accountID|alias> <assetID|alias> <amount>",
	Short: "Build one transaction template,default use account id and asset id",
	Args:  cobra.RangeArgs(3, 4),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("type")
		if buildType == "spend" {
			cmd.MarkFlagRequired("receiver")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var buildReqStr string
		accountInfo := args[0]
		assetInfo := args[1]
		amount := args[2]
		switch buildType {
		case "issue":
			if alias {
				buildReqStr = fmt.Sprintf(buildIssueReqFmtByAlias, btmGas, accountInfo, assetInfo, amount, assetInfo, amount, address)
				break
			}
			buildReqStr = fmt.Sprintf(buildIssueReqFmt, btmGas, accountInfo, assetInfo, amount, assetInfo, amount, address)
		case "spend":
			if alias {
				buildReqStr = fmt.Sprintf(buildSpendReqFmtByAlias, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, receiverProgram)
				break
			}
			buildReqStr = fmt.Sprintf(buildSpendReqFmt, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, receiverProgram)
		case "retire":
			if alias {
				buildReqStr = fmt.Sprintf(buildRetireReqFmtByAlias, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, accountInfo)
				break
			}
			buildReqStr = fmt.Sprintf(buildRetireReqFmt, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, accountInfo)
		case "address":
			if alias {
				buildReqStr = fmt.Sprintf(buildControlAddressReqFmtByAlias, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, address)
				break
			}
			buildReqStr = fmt.Sprintf(buildControlAddressReqFmt, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, address)
		default:
			jww.ERROR.Println("Invalid transaction template type")
			os.Exit(util.ErrLocalExe)
		}

		var buildReq api.BuildRequest
		if err := json.Unmarshal([]byte(buildReqStr), &buildReq); err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		data, exitCode := util.ClientCall("/build-transaction", &buildReq)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		if pretty {
			printJSON(data)
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if ok != true {
			jww.ERROR.Println("invalid type assertion")
			os.Exit(util.ErrLocalParse)
		}

		rawTemplate, err := json.Marshal(dataMap)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalParse)
		}

		jww.FEEDBACK.Printf("Template Type: %s\n%s\n", buildType, string(rawTemplate))
	},
}

var lockContractTransactionCmd = &cobra.Command{
	Use:   "lock-contract-transaction <accountID|alias> <assetID|alias> <amount> <contractProgram>",
	Short: "Build one transaction template,default use account id and asset id",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		var buildReqStr string
		accountInfo := args[0]
		assetInfo := args[1]
		amount := args[2]
		contractProgram := args[3]

		if alias {
			buildReqStr = fmt.Sprintf(buildContractReqFmtByAlias, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, contractProgram)
		} else {
			buildReqStr = fmt.Sprintf(buildContractReqFmt, btmGas, accountInfo, assetInfo, amount, accountInfo, assetInfo, amount, contractProgram)
		}

		var buildReq api.BuildRequest
		if err := json.Unmarshal([]byte(buildReqStr), &buildReq); err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		data, exitCode := util.ClientCall("/lock-contract-transaction", &buildReq)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		if pretty {
			printJSON(data)
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if ok != true {
			jww.ERROR.Println("invalid type assertion")
			os.Exit(util.ErrLocalParse)
		}

		rawTemplate, err := json.Marshal(dataMap)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalParse)
		}

		jww.FEEDBACK.Printf("Template Type: %s\n%s\n", buildType, string(rawTemplate))
	},
}

var unlockContractTransactionCmd = &cobra.Command{
	Use:   "unlock-contract-transaction <outputID> <accountID|alias> <assetID|alias> <amount> -c <contractName> <contractArgs>",
	Short: "Build transaction for template contract, default use account id and asset id",
	Args:  cobra.RangeArgs(4, 20),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("contract-name")
	},
	Run: func(cmd *cobra.Command, args []string) {
		minArgsCount := 4
		usage := "Usage:\n  bytomcli unlock-contract-transaction <outputID> <accountID|alias> <assetID|alias> <amount> -c <contractName>"
		if err := CheckContractArgs(contractName, args, minArgsCount, usage); err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		req, err := BuildReq(contractName, args, alias, btmGas)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		data, exitCode := util.ClientCall("/unlock-contract-transaction", req)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		if pretty {
			printJSON(data)
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if ok != true {
			jww.ERROR.Println("invalid type assertion")
			os.Exit(util.ErrLocalParse)
		}

		rawTemplate, err := json.Marshal(dataMap)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalParse)
		}

		jww.FEEDBACK.Printf("Template Type: %s\n%s\n", buildType, string(rawTemplate))
	},
}

var signTransactionCmd = &cobra.Command{
	Use:   "sign-transaction  <json templates>",
	Short: "Sign transaction templates with account password",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("password")
	},
	Run: func(cmd *cobra.Command, args []string) {
		template := txbuilder.Template{}

		err := json.Unmarshal([]byte(args[0]), &template)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		var req = struct {
			Password string             `json:"password"`
			Txs      txbuilder.Template `json:"transaction"`
		}{Password: password, Txs: template}

		jww.FEEDBACK.Printf("\n\n")
		data, exitCode := util.ClientCall("/sign-transaction", &req)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		if pretty {
			printJSON(data)
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if ok != true {
			jww.ERROR.Println("invalid type assertion")
			os.Exit(util.ErrLocalParse)
		}

		rawSign, err := json.Marshal(dataMap)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalParse)
		}
		jww.FEEDBACK.Printf("\nSign Template:\n%s\n", string(rawSign))
	},
}

var submitTransactionCmd = &cobra.Command{
	Use:   "submit-transaction  <signed json raw_transaction>",
	Short: "Submit signed transaction",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ins = struct {
			Tx types.Tx `json:"raw_transaction"`
		}{}

		err := json.Unmarshal([]byte(args[0]), &ins)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		data, exitCode := util.ClientCall("/submit-transaction", &ins)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var estimateTransactionGasCmd = &cobra.Command{
	Use:   "estimate-transaction-gas  <json templates>",
	Short: "estimate gas for build transaction",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		template := txbuilder.Template{}

		err := json.Unmarshal([]byte(args[0]), &template)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		var req = struct {
			TxTemplate txbuilder.Template `json:"transaction_template"`
		}{TxTemplate: template}

		data, exitCode := util.ClientCall("/estimate-transaction-gas", &req)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var decodeRawTransactionCmd = &cobra.Command{
	Use:   "decode-raw-transaction <raw_transaction>",
	Short: "decode the raw transaction",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ins = struct {
			Tx types.Tx `json:"raw_transaction"`
		}{}

		err := ins.Tx.UnmarshalText([]byte(args[0]))
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		data, exitCode := util.ClientCall("/decode-raw-transaction", &ins)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var getTransactionCmd = &cobra.Command{
	Use:   "get-transaction <hash>",
	Short: "get the transaction by matching the given transaction hash",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		txInfo := &struct {
			TxID string `json:"tx_id"`
		}{TxID: args[0]}

		data, exitCode := util.ClientCall("/get-transaction", txInfo)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var listTransactionsCmd = &cobra.Command{
	Use:   "list-transactions",
	Short: "List the transactions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		filter := struct {
			ID        string `json:"id"`
			AccountID string `json:"account_id"`
			Detail    bool   `json:"detail"`
		}{ID: txID, AccountID: account, Detail: detail}

		data, exitCode := util.ClientCall("/list-transactions", &filter)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSONList(data)
	},
}

var getUnconfirmedTransactionCmd = &cobra.Command{
	Use:   "get-unconfirmed-transaction <hash>",
	Short: "get unconfirmed transaction by matching the given transaction hash",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		txID, err := hex.DecodeString(args[0])
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(util.ErrLocalExe)
		}

		txInfo := &struct {
			TxID chainjson.HexBytes `json:"tx_id"`
		}{TxID: txID}

		data, exitCode := util.ClientCall("/get-unconfirmed-transaction", txInfo)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var listUnconfirmedTransactionsCmd = &cobra.Command{
	Use:   "list-unconfirmed-transactions",
	Short: "list unconfirmed transactions hashes",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		data, exitCode := util.ClientCall("/list-unconfirmed-transactions")
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var gasRateCmd = &cobra.Command{
	Use:   "gas-rate",
	Short: "Print the current gas rate",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		data, exitCode := util.ClientCall("/gas-rate")
		if exitCode != util.Success {
			os.Exit(exitCode)
		}
		printJSON(data)
	},
}
