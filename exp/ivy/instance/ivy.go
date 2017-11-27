package instance

import (
	"fmt"
	"os"
	"encoding/hex"
	"strings"
)

func main() {
	if(len(os.Args) != 2) {
		fmt.Println("command args: [template_contract_name]")
		os.Exit(1)
	}

	template_contract_name := strings.TrimSpace(os.Args[1])

	var result string
	switch template_contract_name {
		case "LockWithPublicKey":
			result = hex.EncodeToString(LockWithPublicKey_body_bytes)
		case "LockWithMultiSig":
			result = hex.EncodeToString(LockWithMultiSig_body_bytes)
		case "LockWithPublicKeyHash":
			result = hex.EncodeToString(LockWithPublicKeyHash_body_bytes)
		case "TradeOffer":
			result = hex.EncodeToString(TradeOffer_body_bytes)
		case "Escrow":
			result = hex.EncodeToString(Escrow_body_bytes)
		case "CallOption":
			result = hex.EncodeToString(CallOption_body_bytes)
		case "LoanCollateral":
			result = hex.EncodeToString(LoanCollateral_body_bytes)
		case "RevealPreimage":
			result = hex.EncodeToString(RevealPreimage_body_bytes)
		default:
			fmt.Printf("Error: the contract[%s] is not in ivy template contract\n", template_contract_name)
			os.Exit(0)
	}

	fmt.Printf("the result: %s\n", result)
}
