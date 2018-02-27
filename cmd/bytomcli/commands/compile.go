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
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkey[%d] is not equal 64\n", len(pubkeyStr))
		}
		pubkey, _ := hex.DecodeString(pubkeyStr)
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkey)})

	case "LockWithMultiSig":
		pubkeyStr1 := args[1]
		pubkeyStr2 := args[2]
		pubkeyStr3 := args[3]
		if len(pubkeyStr1) != 64 || len(pubkeyStr2) != 64 || len(pubkeyStr3) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkey1[%d] or pubkey2[%d] or pubkey3[%d] is not equal 64\n",
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
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte pubkeyHash[%d] is not equal 64\n", len(pubkeyHashStr))
		}

		pubkeyHash, _ := hex.DecodeString(pubkeyHashStr)
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&pubkeyHash)})

	case "RevealPreimage":
		valueHashStr := args[1]
		if len(valueHashStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte valueHash[%d] is not equal 64\n", len(valueHashStr))
		}

		valueHash, _ := hex.DecodeString(valueHashStr)
		contractArgs = append(contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&valueHash)})

	case "TradeOffer":
		assetStr := args[1]
		amountStr := args[2]
		sellerStr := args[3]
		pubkeyStr := args[4]
		if len(assetStr) != 64 || len(pubkeyStr) != 64 {
			err = errors.WithDetailf(ErrInvalidLength, "the length of byte assetID[%d] or pubkey[%s] is not equal 64", len(assetStr), len(pubkeyStr))
		}

		assetByte, _ := hex.DecodeString(assetStr)
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)
		amount, _ := strconv.ParseUint(amountStr, 10, 64)
		seller, _ := hex.DecodeString(sellerStr)
		pubkey, _ := hex.DecodeString(pubkeyStr)

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
		}

		pubkey, _ := hex.DecodeString(pubkeyStr)
		sender, _ := hex.DecodeString(senderStr)
		recipient, _ := hex.DecodeString(recipientStr)

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
		}

		assetByte, _ := hex.DecodeString(assetStr)
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)

		amount, _ := strconv.ParseUint(amountStr, 10, 64)
		dueTimeStr = strings.Replace(dueTimeStr, "*", " ", -1)
		loc, _ := time.LoadLocation("Local")
		var dueTime time.Time
		dueTime, err = time.ParseInLocation(TimeLayout, dueTimeStr, loc)

		lender, _ := hex.DecodeString(lenderStr)
		borrower, _ := hex.DecodeString(borrowerStr)

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
		}

		amountPrice, _ := strconv.ParseUint(amountPriceStr, 10, 64)
		assetByte, _ := hex.DecodeString(assetStr)
		var b [32]byte
		copy(b[:], assetByte[:32])
		assetID := bc.NewAssetID(b)

		seller, _ := hex.DecodeString(sellerStr)
		buyerPubkey, _ := hex.DecodeString(buyerPubkeyStr)
		deadlineStr = strings.Replace(deadlineStr, "*", " ", -1)
		loc, _ := time.LoadLocation("Local")
		var deadline time.Time
		deadline, err = time.ParseInLocation(TimeLayout, deadlineStr, loc)

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
	}

	return
}
