package commands

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/bytom/encoding/json"
	"github.com/bytom/errors"
	"github.com/bytom/exp/ivy/compiler"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/util"

	"encoding/hex"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var compileCmd = &cobra.Command{
	Use:   "compile <contractPathFile> <contractArgs>",
	Short: "Compile contract of Bytomcli",
	Args:  cobra.RangeArgs(1, 20),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		inputFile, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Print(err)
			os.Exit(0)
		}
		inputStr := string(inputFile)

		contractArgs, err := BuildContractArgs(args)
		if err != nil {
			fmt.Print(err)
			os.Exit(0)
		}

		var compileReq = struct {
			Contract string                 `json:"contract"`
			Args     []compiler.ContractArg `json:"args"`
		}{Contract: inputStr, Args: contractArgs}

		jww.FEEDBACK.Printf("\n\n")
		data, exitCode := util.ClientCall("/compile", &compileReq)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

// BuildContractArgs build the ContractArg for contact
func BuildContractArgs(args []string) (contractArgs []compiler.ContractArg, err error) {
	inputFile, err := os.Open(args[0])
	if err != nil {
		return
	}
	defer inputFile.Close()

	inputReader := bufio.NewReader(inputFile)
	contracts, err := compiler.Compile(inputReader)
	if len(contracts) != 1 {
		err = errors.New("Invalid contract format, because of support only one contract!")
		return
	}

	contract := contracts[0]
	switch contract.Name {
	case "LockWithPublicKey":
		pubkeyStr := args[1]
		if len(pubkeyStr) != 64 {
			err = errors.WithDetailf(errors.New("mismatched arguments"), "the length of byte pubkey[%d] is not equal 64\n", len(pubkeyStr))
		}
		pubkey, _ := hex.DecodeString(pubkeyStr)
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey)})

	case "LockWithMultiSig":
		pubkeyStr1 := args[1]
		pubkeyStr2 := args[2]
		pubkeyStr3 := args[3]
		if len(pubkeyStr1) != 64 || len(pubkeyStr2) != 64 || len(pubkeyStr3) != 64 {
			err = errors.WithDetailf(errors.New("mismatched arguments"), "the length of byte pubkey1[%d] or pubkey2[%d] or pubkey3[%d] is not equal 64\n",
				len(pubkeyStr1), len(pubkeyStr2), len(pubkeyStr3))
		}

		pubkey1, _ := hex.DecodeString(pubkeyStr1)
		pubkey2, _ := hex.DecodeString(pubkeyStr2)
		pubkey3, _ := hex.DecodeString(pubkeyStr3)

		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey1)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey2)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey3)})

	case "LockWithPublicKeyHash":
		pubkeyHashStr := args[1]
		if len(pubkeyHashStr) != 64 {
			err = errors.WithDetailf(errors.New("mismatched arguments"), "the length of byte pubkeyHash[%d] is not equal 64\n", len(pubkeyHashStr))
		}

		pubkeyHash, _ := hex.DecodeString(pubkeyHashStr)
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkeyHash)})

	case "RevealPreimage":
		valueHashStr := args[0]
		if len(valueHashStr) != 64 {
			err = errors.WithDetailf(errors.New("mismatched arguments"), "the length of byte valueHash[%d] is not equal 64\n", len(valueHashStr))
		}

		valueHash, _ := hex.DecodeString(valueHashStr)
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&valueHash)})

	case "TradeOffer":
		assetStr := args[1]
		amountStr := args[2]
		sellerStr := args[3]
		pubkeyStr := args[4]
		if len(assetStr) != 64 || len(pubkeyStr) != 64 {
			err = errors.WithDetailf(errors.New("mismatched arguments"), "the length of byte assetID[%d] or pubkey[%s] is not equal 64", len(assetStr), len(pubkeyStr))
		}

		assetByte, _ := hex.DecodeString(assetStr)
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)
		amount, _ := strconv.ParseUint(amountStr, 10, 64)
		seller, _ := hex.DecodeString(sellerStr)
		pubkey, _ := hex.DecodeString(pubkeyStr)

		_assetRequested := assetID.Bytes()
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&_assetRequested)})
		_amountRequested := int64(amount)
		contractArgs = append(contractArgs, compiler.ContractArg{I: &_amountRequested})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&seller)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey)})

	case "Escrow":
		err = errors.New("Invalid contract!")
	case "LoanCollateral":
		err = errors.New("Invalid contract!")
	case "CallOption":
		err = errors.New("Invalid contract!")
	default:
		err = errors.New("Invalid contract!")
	}

	return
}
