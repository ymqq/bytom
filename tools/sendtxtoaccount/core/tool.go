package core

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/bytom/util"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

type config struct {
	SendAcct    string `toml:"send_acct_id"`
	Sendasset   string `toml:"send_asset_id"`
	BtmGas      uint64 `toml:"btm_gas"`
	OutputNum   int    `toml:"output_num"`
	OnlyBuildTx bool   `toml:"only_build_tx"`
}

func init() {
	sendTxCmd.PersistentFlags().StringVar(&configFile, "config", "./config.toml", "config file")
	sendTxCmd.PersistentFlags().StringVar(&accountFile, "accountinfo", "./accountinfo.csv", "acoount info(format: csv)")
}

var (
	sendAcct    string
	sendasset   string
	configFile  string
	accountFile string
	cfg         config
	acctInfo    []accountInfo
	totalBtm    uint64
	outputNum   = 10
)

var sendTxCmd = &cobra.Command{
	Use:   "sendtxttoaccount",
	Short: "send tx to account",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		bs, err := ioutil.ReadFile(configFile)
		if err = toml.Unmarshal(bs, &cfg); err != nil {
			fmt.Println(err)
			return
		}
		sendAcct = cfg.SendAcct
		sendasset = cfg.Sendasset
		readAccoutInfo()
		fmt.Println("*****************send tx start*****************")
		// send btm to account
		acctNum := len(acctInfo)
		if cfg.OutputNum > 0 {
			outputNum = cfg.OutputNum
		}

		for i := 0; i < acctNum; i += outputNum {
			if (i + outputNum) > acctNum {
				outputNum = acctNum - i
			}
			arr := acctInfo[i : i+outputNum]
			Sendtx(sendAcct, sendasset, arr)
		}
		fmt.Println("Total number of users:", acctNum)
		fmt.Println("Total btm[neu]:", totalBtm)
		fmt.Println("*****************send tx end*****************")
	},
}

func readAccoutInfo() {
	file, _ := os.Open(accountFile)
	defer file.Close()
	reader := csv.NewReader(file)
	// generate data
	i := 0
	totalBtm = 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("记录集错误:", err)
			os.Exit(1)
		}
		if len(record[0]) == 42 || len(record[0]) == 62 {
			var acct accountInfo
			acct.address = record[0]
			//amount, _ := strconv.Atoi(record[1])
			amount, _ := strconv.ParseUint(record[1], 10, 64)
			if amount < 0 {
				fmt.Println("address:[", record[0], "] amount < 0")
				os.Exit(1)
			}
			acct.amount = amount
			totalBtm += amount
			//acctInfo[i] = acct
			acctInfo = append(acctInfo, acct)
			i++
		} else {
			fmt.Println("account:", record[0], " is error")
			os.Exit(1)
		}
	}
}

// Execute send tx
func Execute() {
	if _, err := sendTxCmd.ExecuteC(); err != nil {
		os.Exit(util.ErrLocalExe)
	}
}
