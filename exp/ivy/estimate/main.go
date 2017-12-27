package main

import (
	"fmt"
	"os"
	"bufio"
	"github.com/bytom/exp/ivy/compiler"
	"encoding/hex"
)

const (
	PARAM_GAS = "--gas"
	PARAM_SHIFT = "--shift"
	PARAM_BIN = "--bin"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Println("command args: [command] [contract_file] arguments")
		fmt.Println("the command arguments:")
		fmt.Printf("\t %s \t\t estimate gas of contract\n", PARAM_GAS)
		fmt.Printf("\t %s \t the shift of clause for contract\n", PARAM_SHIFT)
		fmt.Printf("\t %s \t\t the program for contract\n", PARAM_BIN)
		fmt.Println("\n")
		os.Exit(0)
	}

	filename := os.Args[1]
	inputFile, inputError := os.Open(filename)
	if inputError != nil {
		fmt.Printf("An error occurred on opening the inputfile\n" +
			"Does the file exist?\n" +
			"Have you got acces to it?\n")
		os.Exit(0)
	}
	defer inputFile.Close()

	inputReader := bufio.NewReader(inputFile)
	contracts, err := compiler.Compile(inputReader)
	if err != nil {
		fmt.Println("Compile contract failed, err:", err)
		os.Exit(0)
	}

	//the compile contract can adapt to that multiple contracts are compiled at the same time,
	//but this place can only use a single contract
	contract := contracts[0]
	prog := contract.Body

	fmt.Printf("======= %v =======\n", contract.Name)
	arguments := os.Args[2]
	if arguments == PARAM_GAS {
		if err := estimate(contract, prog); err != nil {
			fmt.Println("Error:", err)
			os.Exit(0)
		}
	} else if arguments == PARAM_SHIFT {
		if err := shift(contract, prog); err != nil {
			fmt.Println("Error:", err)
			os.Exit(0)
		}
	} else if arguments == PARAM_BIN {
		fmt.Println("Contract program:")
		fmt.Printf("%v\n\n", hex.EncodeToString(contract.Body))
	} else {
		fmt.Println("the command arguments is not used\n")
	}
}