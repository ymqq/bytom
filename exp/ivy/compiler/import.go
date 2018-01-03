package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"bufio"
	"io/ioutil"
)

func parsePath(p *parser) []*Contract {
	path := parseImport(p)
	if path == nil {
		panic("parseImport failed")
	}
	fmt.Println("path:", string(path))

	filename := absolutePath(string(path))
	if filename == "" {
		panic("check absolute path failed")
	}

	inputFile, inputError := os.Open(filename)
	if inputError != nil {
		errmsg := fmt.Sprintf("Open the file [%v] error, err:[%v]\n", filename, inputError)
		panic(errmsg)
	}
	defer inputFile.Close()

	inputReader := bufio.NewReader(inputFile)
	inp, err := ioutil.ReadAll(inputReader)
	if err != nil {
		errmsg := fmt.Sprintf("reading input error:[%v]\n", filename, err)
		panic(errmsg)
	}

	contracts, err := parse(inp)
	if err != nil {
		errmsg := fmt.Sprintf("parse input error:[%v]\n", err)
		panic(errmsg)
	}

	var result []*Contract
	for _, contract := range contracts {
		result = append(result, contract)
	}

	return result
}

func parseImport(p *parser) []byte {
	consumeKeyword(p, "import")
	strliteral, newOffset := scanStrLiteral(p.buf, p.pos)
	if newOffset < 0 {
		return nil
	}
	p.pos = newOffset

	//check the quote for path
	if strliteral[0] != '\'' || strliteral[len(strliteral) - 1] != '\'' {
		return nil
	}
	importPath := strliteral[1 : len(strliteral)-1]
	return importPath
}


func parseContractImport(p *parser) []*Contract {
	var result []*Contract
	for peekKeyword(p) == "import" {
		contracts := parsePath(p)
		for _, contract := range contracts {
			result = append(result, contract)
		}
	}
	return result
}

func absolutePath(path string) string {
	fpath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("err:", err)
		return ""
	}
	fpath = strings.Replace(fpath, "\\", "/", -1)

	if ok := checkPath(fpath); !ok {
		fmt.Println("check file path failed")
		return ""
	}

	return fpath
}

//check whether the path is valid
func checkPath(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

/*
//check whether the contract is valid
func scanContract(p *parser) error {
	fmt.Println("right")
	return nil
}

//get the current directory
func getCurrentDirectory(path string) string {
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

//get the parent directory
func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}
*/