package main

import (
	"fmt"
	"github.com/bytom/errors"
	"encoding/hex"
	"github.com/bytom/exp/ivy/compiler"
	"github.com/bytom/protocol/vm"
)

func shift(contract *compiler.Contract, prog []byte) error {
	fmt.Println("Clause shift:")

	instructions, err := vm.ParseProgram(prog)
	if err != nil {
		fmt.Println("ParseProgram err:", err)
		return err
	}

	var shiftData [][]byte
	var jumpifData [][]byte
	var jumpData []byte
	jumpifCount := 0
	jumpCount := 0

	for _, inst := range instructions {
		switch inst.Op.String() {
		case "JUMPIF":
			jumpifData = append(jumpifData, inst.Data)
			jumpifCount++
		case "JUMP":
			//if If there is more than one JUMP instruction, all the data of instruction is the same
			if jumpCount > 0 {
				continue
			}
			jumpData = inst.Data
			jumpCount++
		}
	}

	//if If there is more than one JUMPIF instruction, the order of clause is adverse
	if jumpifCount >= 2 {
		length := len(jumpifData)
		for j:= 0; j < length/2; j++ {
			jumpifData[j], jumpifData[length-j-1] = jumpifData[length-j-1], jumpifData[j]
		}
	}

	//the first clause is 00000000
	firstData, _ := hex.DecodeString("00000000")
	shiftData = append(shiftData, firstData)

	//the second or more clause
	if len(jumpifData) > 0 {
		for _, data := range jumpifData {
			shiftData = append(shiftData, data)
		}
	}

	if len(shiftData) != len(contract.Clauses) {
		errmsg := fmt.Sprintf("the number of clause_data for program [%d] is not equal to the number of contract clause [%d]\n\n",
			len(shiftData), len(contract.Clauses))
		err := errors.New(errmsg)
		return err
	}

	for i, _ := range contract.Clauses {
		fmt.Printf("    %s:  %v\n", contract.Clauses[i].Name, hex.EncodeToString(shiftData[i]))
	}

	//the ending of clause
	if jumpData != nil {
		fmt.Printf("    ending:  %v\n\n", hex.EncodeToString(jumpData))
	}

	if len(contract.Clauses) == 1 {
		fmt.Println("\nNOTE: \n    The contract contain only one clause, Users don't need to input clause selector!!!\n")
	}

	return nil
}