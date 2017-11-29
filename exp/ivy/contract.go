package main

import (
	"fmt"
	"os"
	"encoding/hex"
	"github.com/bytom/exp/ivy/instance"
	"strings"
)

func main() {
	if(len(os.Args) < 2) {
		fmt.Println("command args: [template_contract_name]")
		os.Exit(1)
	}

	template_contract_name := strings.TrimSpace(os.Args[1])

	var result string
	switch template_contract_name {
	case "LockWithPublicKey":
		if(len(os.Args) != 3) {
			fmt.Println("add args: [pubkey]")
		}
		pubkey := os.Args[2]
		pubkeyvalue, _:= hex.DecodeString(pubkey)
		out, _ := instance.PayToLockWithPublicKey(pubkeyvalue)
		result = hex.EncodeToString(out)
	case "LockWithMultiSig":
		if(len(os.Args) != 5) {
			fmt.Println("add args: [pubkey1] [pubkey2] [pubkey3]")
		}
		pubkey1 := os.Args[2]
		pubkey2 := os.Args[3]
		pubkey3 := os.Args[4]
		pub1, _:= hex.DecodeString(pubkey1)
		pub2, _:= hex.DecodeString(pubkey2)
		pub3, _:= hex.DecodeString(pubkey3)

		out, _ := instance.PayToLockWithMultiSig(pub1, pub2, pub3)
		result = hex.EncodeToString(out)
	case "LockWithPublicKeyHash":
		if(len(os.Args) != 3) {
			fmt.Println("add args: [pubKeyHash]")
			os.Exit(1)
		}
		pubkeyhash := os.Args[2]
		hashvalue, _:= hex.DecodeString(pubkeyhash)

		out, _ := instance.PayToLockWithPublicKeyHash(hashvalue)
		result = hex.EncodeToString(out)
	case "TradeOffer":
		result = hex.EncodeToString(instance.TradeOffer_body_bytes)
	case "Escrow":
		result = hex.EncodeToString(instance.Escrow_body_bytes)
	case "CallOption":
		result = hex.EncodeToString(instance.CallOption_body_bytes)
	case "LoanCollateral":
		result = hex.EncodeToString(instance.LoanCollateral_body_bytes)
	case "RevealPreimage":
		if(len(os.Args) != 3) {
			fmt.Println("add args: [hash]")
			os.Exit(1)
		}
		hash := os.Args[2]
		hashvalue, _:= hex.DecodeString(hash)

		out, _ := instance.PayToRevealPreimage(hashvalue)
		result = hex.EncodeToString(out)
	default:
		fmt.Printf("Error: the contract[%s] is not in ivy template contract\n", template_contract_name)
		os.Exit(0)
	}

	fmt.Printf("the result: %s\n", result)
}
