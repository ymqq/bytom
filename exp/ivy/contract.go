package main

import (
	"fmt"
	"os"
	"encoding/hex"
	"github.com/bytom/exp/ivy/instance"
	"strings"
	"strconv"
	"time"
	"github.com/bytom/protocol/bc"
	"io"
)

// the TimeLayout by time Template
const (
	TimeLayout string = "2006-01-02 15:05:05"
)

func main() {
	if(len(os.Args) < 2) {
		help(os.Stdout)
		os.Exit(0)
	}

	var tmp [32]byte
	var result string

	template_contract_name := strings.TrimSpace(os.Args[1])

	switch template_contract_name {
	case "LockWithPublicKey":
		if(len(os.Args) != 3) {
			fmt.Println("args: [pubkey]\n\n")
			os.Exit(0)
		}

		pubkey := os.Args[2]
		if CheckLength(pubkey) == false {
			fmt.Println("the length of pubkey is not equal 32\n")
			os.Exit(0)
		}

		pubkeyvalue, _ := hex.DecodeString(pubkey)

		out, _ := instance.PayToLockWithPublicKey(pubkeyvalue)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToLockWithPublicKey(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "LockWithMultiSig":
		if(len(os.Args) != 5) {
			fmt.Println("args: [pubkey1] [pubkey2] [pubkey3]\n\n")
			os.Exit(0)
		}
		pubkey1 := os.Args[2]
		pubkey2 := os.Args[3]
		pubkey3 := os.Args[4]
		if CheckLength(pubkey1) == false || CheckLength(pubkey2) == false || CheckLength(pubkey3) == false {
			fmt.Println("the length of pubkey is not equal 32\n")
			os.Exit(0)
		}

		pub1, _:= hex.DecodeString(pubkey1)
		pub2, _:= hex.DecodeString(pubkey2)
		pub3, _:= hex.DecodeString(pubkey3)

		out, _ := instance.PayToLockWithMultiSig(pub1, pub2, pub3)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToLockWithMultiSig(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "LockWithPublicKeyHash":
		if(len(os.Args) != 3) {
			fmt.Println("args: [pubKeyHash]\n\n")
			os.Exit(0)
		}
		pubkeyhash := os.Args[2]
		if CheckLength(pubkeyhash) == false {
			fmt.Println("the length of pubKeyHash is not equal 32\n")
			os.Exit(0)
		}

		hashvalue, _:= hex.DecodeString(pubkeyhash)

		out, _ := instance.PayToLockWithPublicKeyHash(hashvalue)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToLockWithPublicKeyHash(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "TradeOffer":
		if(len(os.Args) != 6) {
			fmt.Println("args: [assetid] [amount] [seller] [pubkey]\n\n")
			os.Exit(0)
		}
		assetRequested := os.Args[2]
		amountRequested := os.Args[3]
		seller := os.Args[4]
		pubkey := os.Args[5]
		if CheckLength(assetRequested) == false || CheckLength(pubkey) == false {
			fmt.Println("the length of assetid or pubkey is not equal 32\n")
			os.Exit(0)
		}

		asset, _:= hex.DecodeString(assetRequested)
		copy(tmp[:], asset[:32])
		assetid := bc.NewAssetID(tmp)
		//fmt.Println("assetid:", assetid)

		amount, _ := strconv.ParseUint(amountRequested, 10, 64)
		sell, _:= hex.DecodeString(seller)
		pub, _:= hex.DecodeString(pubkey)

		out, _ := instance.PayToTradeOffer(assetid, amount, sell, pub)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToTradeOffer(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "Escrow":
		if(len(os.Args) != 5) {
			fmt.Println("args: [pubkey] [sender] [recipient]\n\n")
			os.Exit(0)
		}
		pubkey := os.Args[2]
		sender := os.Args[3]
		recipient := os.Args[4]
		if CheckLength(pubkey) == false {
			fmt.Println("the length of pubkey is not equal 32\n")
			os.Exit(0)
		}

		pub, _:= hex.DecodeString(pubkey)
		send, _:= hex.DecodeString(sender)
		recip, _:= hex.DecodeString(recipient)

		out, _ := instance.PayToEscrow(pub, send, recip)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToEscrow(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "CallOption":
		if(len(os.Args) != 7) {
			fmt.Println("args: [price] [assetid] [seller] [buyerKey] [deadline]\n\n")
			os.Exit(0)
		}
		strikePrice := os.Args[2]
		strikeCurrency := os.Args[3]
		seller := os.Args[4]
		buyerKey := os.Args[5]
		deadline := os.Args[6]
		if CheckLength(strikeCurrency) == false || CheckLength(buyerKey) == false {
			fmt.Println("the length of assetid or pubkey is not equal 32\n")
			os.Exit(0)
		}

		asset, _:= hex.DecodeString(strikeCurrency)
		copy(tmp[:], asset[:32])
		assetid := bc.NewAssetID(tmp)

		price, _:= strconv.ParseUint(strikePrice, 10, 64)
		sell, _:= hex.DecodeString(seller)
		pub, _:= hex.DecodeString(buyerKey)

		deadline = strings.Replace(deadline, "*", " ", -1)
		loc, _ := time.LoadLocation("Local")
		expiretime, _ := time.ParseInLocation(TimeLayout, deadline, loc)
		//fmt.Println("expiretime:", expiretime)

		out, _ := instance.PayToCallOption(price, assetid, sell, pub, expiretime)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToCallOption(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "LoanCollateral":
		if(len(os.Args) != 7) {
			fmt.Println("args: [assetid] [amount] [duetime] [lender] [borrower]\n\n")
			os.Exit(0)
		}
		assetLoaned := os.Args[2]
		amountLoaned := os.Args[3]
		repaymentDue := os.Args[4]
		lender := os.Args[5]
		borrower := os.Args[6]
		if CheckLength(assetLoaned) == false {
			fmt.Println("the length of assetid is not equal 32\n")
			os.Exit(0)
		}

		asset, _:= hex.DecodeString(assetLoaned)
		copy(tmp[:], asset[:32])
		assetid := bc.NewAssetID(tmp)

		amount, _:= strconv.ParseUint(amountLoaned, 10, 64)
		lend, _:= hex.DecodeString(lender)
		borrow, _:= hex.DecodeString(borrower)

		repaymentDue = strings.Replace(repaymentDue, "*", " ", -1)
		loc, _ := time.LoadLocation("Local")
		duetime, _ := time.ParseInLocation(TimeLayout, repaymentDue, loc)
		//fmt.Println("duetime:", duetime)

		out, _ := instance.PayToLoanCollateral(assetid, amount, duetime, lend, borrow)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToLoanCollateral(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "RevealPreimage":
		if(len(os.Args) != 3) {
			fmt.Println("args: [hash]\n\n")
			os.Exit(0)
		}
		hash := os.Args[2]
		if CheckLength(hash) == false {
			fmt.Println("the length of hash is not equal 32\n")
			os.Exit(0)
		}

		hashvalue, _:= hex.DecodeString(hash)

		out, _ := instance.PayToRevealPreimage(hashvalue)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToRevealPreimage(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	case "PriceChanger":
		if(len(os.Args) != 6) {
			fmt.Println("args: [amount] [assetid] [pubkey] [seller]\n\n")
			os.Exit(0)
		}
		askAmount := os.Args[2]
		askAsset := os.Args[3]
		sellerKey := os.Args[4]
		sellerProg := os.Args[5]
		if CheckLength(askAsset) == false || CheckLength(sellerKey) == false {
			fmt.Println("the length of assetid or pubkey is not equal 32\n")
			os.Exit(0)
		}

		asset, _:= hex.DecodeString(askAsset)
		copy(tmp[:], asset[:32])
		assetid := bc.NewAssetID(tmp)

		amount, _:= strconv.ParseUint(askAmount, 10, 64)
		pubkey, _:= hex.DecodeString(sellerKey)
		seller, _:= hex.DecodeString(sellerProg)

		out, _ := instance.PayToPriceChanger(amount, assetid, pubkey, seller)
		result = hex.EncodeToString(out)

		//check the program
		_, err := instance.ParsePayToPriceChanger(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	default:
		fmt.Printf("Error: the contract [%s] is not in ivy template contract\n\n\n", template_contract_name)
		os.Exit(0)
	}

	fmt.Printf("The Result ControlProgram:\n%s\n\n", result)
}

func help(w io.Writer) {
	fmt.Fprintln(w, "usage: ivy [command] [arguments]")
	fmt.Fprint(w, "\nThe commands are:\n\n")
	fmt.Fprintln(w, "\t LockWithPublicKey")
	fmt.Fprintln(w, "\t LockWithMultiSig")
	fmt.Fprintln(w, "\t LockWithPublicKeyHash")
	fmt.Fprintln(w, "\t TradeOffer")
	fmt.Fprintln(w, "\t Escrow")
	fmt.Fprintln(w, "\t CallOption")
	fmt.Fprintln(w, "\t LoanCollateral")
	fmt.Fprintln(w, "\t RevealPreimage")
	fmt.Fprintln(w, "\t PriceChanger")
	fmt.Fprintln(w)
}

func CheckLength(str string) bool {
	length := len(str)
	if length == 64 { //the length of 32-bytes string is 64, because of a byte compose with two charactor
		return true
	} else {
		return false
	}
}