package commands

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bytom/encoding/json"
	"github.com/bytom/errors"
	"github.com/bytom/exp/ivy/compiler"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/util"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	// ErrInvalidNumber means that the number of arguments is illegal
	ErrInvalidNumber = errors.New("invalid number of arguments")
	// ErrInvalidLength means that the length is illegal
	ErrInvalidLength = errors.New("invalid length")
)

// the TimeLayout by time Template
const (
	TimeLayout string = "2006-01-02 15:05:05"
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
			jww.FEEDBACK.Printf("\n\n")
			os.Exit(util.ErrLocalExe)
		}

		contractArgs, err := BuildContractArgs(args)
		if err != nil {
			fmt.Print(err)
			jww.FEEDBACK.Printf("\n\n")
			os.Exit(util.ErrLocalExe)
		}

		var compileReq = struct {
			Contract string                 `json:"contract"`
			Args     []compiler.ContractArg `json:"args"`
		}{Contract: string(inputFile), Args: contractArgs}

		jww.FEEDBACK.Printf("\n\n")
		data, exitCode := util.ClientCall("/compile", &compileReq)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

// CheckCompileArgs check the number of contract's arguments
func CheckCompileArgs(contractName string, args []string) (err error) {
	usage := "Usage:\n  bytomcli compile <contractPathFile>"
	switch contractName {
	case "LockWithPublicKey":
		if len(args) != 2 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <pubkey> [flags]\n", usage)
		}
	case "LockWithMultiSig":
		if len(args) != 4 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <pubkey1> <pubkey2> <pubkey3> [flags]\n", usage)
		}
	case "LockWithPublicKeyHash":
		if len(args) != 2 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <pubkeyHash> [flags]\n", usage)
		}
	case "RevealPreimage":
		if len(args) != 2 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <valueHash> [flags]\n", usage)
		}
	case "TradeOffer":
		if len(args) != 5 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <assetID> <amount> <seller> <pubkey> [flags]\n", usage)
		}
	case "Escrow":
		if len(args) != 4 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <pubkey> <sender> <recipient> [flags]\n", usage)
		}
	case "LoanCollateral":
		if len(args) != 6 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <assetID> <amount> <dueTime> <lender> <borrower> [flags]\n", usage)
		}
	case "CallOption":
		if len(args) != 6 {
			err = errors.WithDetailf(ErrInvalidNumber, "%s <amountPrice> <assetID> <seller> <buyerPubkey> <deadline> [flags]\n", usage)
		}
	}

	return
}

// BuildContractArgs build the ContractArg for contact
func BuildContractArgs(args []string) ([]compiler.ContractArg, error) {
	var contractArgs []compiler.ContractArg

	inputFile, err := os.Open(args[0])
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	inputReader := bufio.NewReader(inputFile)
	contracts, err := compiler.Compile(inputReader)
	if err != nil {
		return nil, err
	}

	if len(contracts) != 1 {
		err = errors.New("Invalid contract format, because of support only one contract!")
		return nil, err
	}

	contract := contracts[0]
	if err := CheckCompileArgs(contract.Name, args); err != nil {
		return nil, err
	}

	switch contract.Name {
	case "LockWithPublicKey":
		pubkeyStr := args[1]
		if len(pubkeyStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkey[%d] is not equal 64\n", len(pubkeyStr))
			return nil, err
		}

		pubkey, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return nil, err
		}
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey)})

	case "LockWithMultiSig":
		pubkeyStr1 := args[1]
		pubkeyStr2 := args[2]
		pubkeyStr3 := args[3]
		if len(pubkeyStr1) != 64 || len(pubkeyStr2) != 64 || len(pubkeyStr3) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkey1[%d] or pubkey2[%d] or pubkey3[%d] is not equal 64\n",
				len(pubkeyStr1), len(pubkeyStr2), len(pubkeyStr3))
			return nil, err
		}

		pubkey1, err := hex.DecodeString(pubkeyStr1)
		if err != nil {
			return nil, err
		}

		pubkey2, err := hex.DecodeString(pubkeyStr2)
		if err != nil {
			return nil, err
		}

		pubkey3, err := hex.DecodeString(pubkeyStr3)
		if err != nil {
			return nil, err
		}

		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey1)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey2)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey3)})

	case "LockWithPublicKeyHash":
		pubkeyHashStr := args[1]
		if len(pubkeyHashStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkeyHash[%d] is not equal 64\n", len(pubkeyHashStr))
			return nil, err
		}

		pubkeyHash, err := hex.DecodeString(pubkeyHashStr)
		if err != nil {
			return nil, err
		}

		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkeyHash)})

	case "RevealPreimage":
		valueHashStr := args[1]
		if len(valueHashStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte valueHash[%d] is not equal 64\n", len(valueHashStr))
			return nil, err
		}

		valueHash, err := hex.DecodeString(valueHashStr)
		if err != nil {
			return nil, err
		}

		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&valueHash)})

	case "TradeOffer":
		assetStr := args[1]
		amountStr := args[2]
		sellerStr := args[3]
		pubkeyStr := args[4]
		if len(assetStr) != 64 || len(pubkeyStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte assetID[%d] or pubkey[%s] is not equal 64", len(assetStr), len(pubkeyStr))
			return nil, err
		}

		assetByte, err := hex.DecodeString(assetStr)
		if err != nil {
			return nil, err
		}
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)

		amount, err := strconv.ParseUint(amountStr, 10, 64)
		if err != nil {
			return nil, err
		}

		seller, err := hex.DecodeString(sellerStr)
		if err != nil {
			return nil, err
		}

		pubkey, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return nil, err
		}

		assetRequested := assetID.Bytes()
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&assetRequested)})
		amountRequested := int64(amount)
		contractArgs = append(contractArgs, compiler.ContractArg{I: &amountRequested})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&seller)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey)})

	case "Escrow":
		pubkeyStr := args[1]
		senderStr := args[2]
		recipientStr := args[3]
		if len(pubkeyStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkey[%d] is not equal 64\n", len(pubkeyStr))
			return nil, err
		}

		pubkey, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return nil, err
		}

		sender, err := hex.DecodeString(senderStr)
		if err != nil {
			return nil, err
		}

		recipient, err := hex.DecodeString(recipientStr)
		if err != nil {
			return nil, err
		}

		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&sender)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&recipient)})

	case "LoanCollateral":
		assetStr := args[1]
		amountStr := args[2]
		dueTimeStr := args[3]
		lenderStr := args[4]
		borrowerStr := args[5]
		if len(assetStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte assetID[%d] is not equal 64\n", len(assetStr))
			return nil, err
		}

		assetByte, err := hex.DecodeString(assetStr)
		if err != nil {
			return nil, err
		}
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)

		amount, err := strconv.ParseUint(amountStr, 10, 64)
		if err != nil {
			return nil, err
		}

		dueTimeStr = strings.Replace(dueTimeStr, "*", " ", -1)
		loc, _ := time.LoadLocation("Local")
		dueTime, err := time.ParseInLocation(TimeLayout, dueTimeStr, loc)
		if err != nil {
			return nil, err
		}

		lender, err := hex.DecodeString(lenderStr)
		if err != nil {
			return nil, err
		}

		borrower, err := hex.DecodeString(borrowerStr)
		if err != nil {
			return nil, err
		}

		assetLoaned := assetID.Bytes()
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&assetLoaned)})
		amountLoaned := int64(amount)
		contractArgs = append(contractArgs, compiler.ContractArg{I: &amountLoaned})
		repaymentDue := dueTime.UnixNano() / int64(time.Millisecond)
		contractArgs = append(contractArgs, compiler.ContractArg{I: &repaymentDue})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&lender)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&borrower)})

	case "CallOption":
		amountPriceStr := args[1]
		assetStr := args[2]
		sellerStr := args[3]
		buyerPubkeyStr := args[4]
		deadlineStr := args[5]
		if len(assetStr) != 64 || len(buyerPubkeyStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte assetID[%d] or buyerPubkey[%d] is not equal 64\n", len(assetStr), len(buyerPubkeyStr))
			return nil, err
		}

		amountPrice, err := strconv.ParseUint(amountPriceStr, 10, 64)
		if err != nil {
			return nil, err
		}

		assetByte, err := hex.DecodeString(assetStr)
		if err != nil {
			return nil, err
		}
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)

		seller, err := hex.DecodeString(sellerStr)
		if err != nil {
			return nil, err
		}

		buyerPubkey, err := hex.DecodeString(buyerPubkeyStr)
		if err != nil {
			return nil, err
		}

		deadlineStr = strings.Replace(deadlineStr, "*", " ", -1)
		loc, _ := time.LoadLocation("Local")
		deadline, err := time.ParseInLocation(TimeLayout, deadlineStr, loc)
		if err != nil {
			return nil, err
		}

		strikePrice := int64(amountPrice)
		contractArgs = append(contractArgs, compiler.ContractArg{I: &strikePrice})
		strikeCurrency := assetID.Bytes()
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&strikeCurrency)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&seller)})
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&buyerPubkey)})
		_deadline := deadline.UnixNano() / int64(time.Millisecond)
		contractArgs = append(contractArgs, compiler.ContractArg{I: &_deadline})

	default:
		err = errors.New("Invalid contract name!")
		return nil, err
	}

	return contractArgs, nil
}
