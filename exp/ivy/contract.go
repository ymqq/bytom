package ivy

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
		if(len(os.Args) != 2) {
			fmt.Println("add args: [pubkey]")
		}
		pubkey := os.Args[2]
		pubkeyvalue, _:= hex.DecodeString(pubkey)
		out, _ := instance.PayToRevealPreimage(pubkeyvalue)
		result = hex.EncodeToString(out)
	case "LockWithMultiSig":
		result = hex.EncodeToString(instance.LockWithMultiSig_body_bytes)
	case "LockWithPublicKeyHash":
		result = hex.EncodeToString(instance.LockWithPublicKeyHash_body_bytes)
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
