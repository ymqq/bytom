package main

import (
	"fmt"
	"os"
	"bufio"

	"github.com/bytom/exp/ivy/compiler"
	"github.com/bytom/protocol/vm"
)

func main() {
	if len(os.Args) <= 2 || os.Args[1] != "-f" {
		fmt.Println("command args: [command] -f [contract_file]")
		os.Exit(0)
	}

	filename := os.Args[2]
	inputFile, inputError := os.Open(filename)
	if inputError != nil {
		fmt.Printf("An error occurred on opening the inputfile\n" +
			"Does the file exist?\n" +
			"Have you got acces to it?\n")
		return // exit the function on error
	}
	defer inputFile.Close()

	inputReader := bufio.NewReader(inputFile)
	contracts, err := compiler.Compile(inputReader)
	if err != nil {
		fmt.Println("Compile contract failed, err:", err)
		return
	}

	//the compile contract can adapt to that multiple contracts are compiled at the same time,
	//but this place can only use a single contract
	contract := contracts[0]
	prog := contract.Body
	fmt.Printf("======= %v =======\n", contract.Name)
	fmt.Println("Gas estimation:")

	//claculate the contract paraments consumed gas
	var contractParamGas int64
	contractParamGas = 0
	for _, cparam := range contract.Params {
		if cgas := vm.GetContractParamGas(string(cparam.Type)); cgas != -1 {
			contractParamGas = contractParamGas + cgas
		} else {
			fmt.Printf("the type of parament [%v] is error\n", cparam.Type)
			return
		}
	}
	fmt.Println("contractParamGas:", contractParamGas)

	//claculate the clause paraments consumed gas
	var clauseParamGasList []int64
	for i, _ := range contract.Clauses {
		clauseParamGas := int64(0)
		for _, fparam := range contract.Clauses[i].Params {
			if fgas := vm.GetClauseParamGas(string(fparam.Type)); fgas != -1 {
				clauseParamGas = clauseParamGas + fgas
			} else {
				fmt.Printf("the type of parament [%v] is error\n", fparam.Type)
				return
			}
		}
		clauseParamGasList = append(clauseParamGasList, clauseParamGas)
	}

	//print the clause paraments consumed gas
	fmt.Println("clauseParamGas:")
	for i, _ := range clauseParamGasList {
		clause := fmt.Sprintf("%s", contract.Clauses[i].Name)
		fmt.Printf("    %v:  %v\n", clause, clauseParamGasList[i])
	}

	//estimate gas
	result := calculate(prog)
	if len(result) != len(clauseParamGasList) {
		fmt.Printf("the length of result[%d] is not equal to the number of clause[%d]\n", len(result), len(clauseParamGasList))
		return
	}

	//print the estimation result
	fmt.Println("\nEstimation result:")
	for i, _ := range result {
		//print the clause paraments type
		var paramlist string
		for j , p := range contract.Clauses[i].Params {
			if j != len(contract.Clauses[i].Params) - 1 {
				paramlist = paramlist + string(p.Type) + ", "
			} else {
				paramlist = paramlist + string(p.Type)
			}

		}

		clause := fmt.Sprintf("%s(%s)", contract.Clauses[i].Name, paramlist)
		fmt.Printf("\t%v:  %v\n", clause, result[i] + contractParamGas + clauseParamGasList[i])
	}

	fmt.Println("\nNOTICE: \n    Estimated results for reference only, Please check the execution program consumed gas!!!\n")
}

func calculate( prog []byte) []int64 {
	instructions, err := vm.ParseProgram(prog)
	if err != nil {
		fmt.Println("ParseProgram err:", err)
		os.Exit(-1)
	}

	//init the gas of instruction
	vm.InitGas()

	var clauseResult []int64
	var childClauseResult []int64
	var result int64
	var gas int64
	var count int
	var intermediate int64
	result = 0
	gas = 0
	count = 0
	intermediate = 0

	//calculate the instruction consumed gas
	for i, inst := range instructions {
		switch inst.Op.String() {
		case "PUSHDATA1":
			if len(inst.Data) != 0 {
				gas = int64(10 + len(inst.Data))
			} else {
				gas = vm.GetGas(inst.Op)
			}
		case "PUSHDATA2":
			if len(inst.Data) != 0 {
				gas = int64(11 + len(inst.Data))
			} else {
				gas = vm.GetGas(inst.Op)
			}
		case "PUSHDATA4":
			if len(inst.Data) != 0 {
				gas = int64(13 + len(inst.Data))
			} else {
				gas = vm.GetGas(inst.Op)
			}
		case "CHECKPREDICATE":
			childprog := instructions[i-2].Data
			fmt.Println("\nstart childVM instructions")
			tmpclauseResult := calculate(childprog)
			for _, tmp := range tmpclauseResult{
				childClauseResult = append(childClauseResult, tmp)
			}
			fmt.Println("end childVM instructions")
			fmt.Printf("The result of childVM estimate gas: %v\n\n", childClauseResult)
			gas = vm.GetGas(inst.Op)
		case "JUMPIF":
			gas = vm.GetGas(inst.Op)
			//fmt.Printf("%v:  %d\n", inst.Op.String(), gas)
			if instructions[i+1].Op.String() != "JUMPIF" {
				intermediate = result + gas
				result = 0
				gas = 0
				//fmt.Printf("intermediate result: %d\n", intermediate)
			}
		case "JUMP":
			count = count + 1
			gas = vm.GetGas(inst.Op)
			//fmt.Printf("%v:  %d\n", inst.Op.String(), gas)
			result = intermediate + result + gas
			//fmt.Printf("the %d clause estimate gas: %d\n", count, result)
			clauseResult = append(clauseResult, result)
			result = 0
			gas = 0
		default:
			gas = vm.GetGas(inst.Op)
		}

		//if inst.Op.String() != "JUMP" && inst.Op.String() != "JUMPIF" {
		//	fmt.Printf("%v:  %d\n", inst.Op.String(), gas)
		//}
		result = result + gas
	}

	if len(childClauseResult) > 0 {
		for i, _ := range childClauseResult {
			childClauseResult[i] = childClauseResult[i] + result
		}
		clauseResult = childClauseResult
	} else {
		//fmt.Println("The ending clause(or only one clause) estimate gas:", result)
		clauseResult = append(clauseResult, result)
	}

	return clauseResult
}