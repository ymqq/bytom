package blockchain

import (
	"bytes"
	"fmt"
	"strings"

	chainjson "github.com/bytom/encoding/json"
	"github.com/bytom/exp/ivy/compiler"
)

type (
	compileReq struct {
		Contract string                 `json:"contract"`
		Args     []compiler.ContractArg `json:"args"`
	}

	compileResp struct {
		Name    string             `json:"name"`
		Source  string             `json:"source"`
		Program chainjson.HexBytes `json:"program"`
		Params  []compiler.Param   `json:"params"`
		Value   string             `json:"value"`
		Clauses []ClauseInfo       `json:"clause_info"`
		Opcodes string             `json:"opcodes"`
		Error   string             `json:"error"`
	}

	ClauseInfo struct {
		Name      string               `json:"name"`
		Args      []compiler.Param     `json:"args"`
		Values    []compiler.ValueInfo `json:"value_info"`
		Mintimes  []string             `json:"mintimes"`
		Maxtimes  []string             `json:"maxtimes"`
		HashCalls []compiler.HashCall  `json:"hash_calls"`
	}
)

func compileIvy(req compileReq) (compileResp, error) {
	var resp compileResp
	compiled, err := compiler.Compile(strings.NewReader(req.Contract))
	if err != nil {
		resp.Error = err.Error()
	}

	// The contract in dashboard only support single contract instance
	contract := compiled[0]

	resp.Name = contract.Name
	resp.Source = req.Contract
	resp.Value = contract.Value

	for _, param := range contract.Params {
		resp.Params = append(resp.Params, *param)
	}

	resp.Program, err = compiler.Instantiate(contract.Body, contract.Params, false, req.Args)
	if err != nil {
		resp.Error = err.Error()
	}

	for _, contract := range compiled {
		for _, clause := range contract.Clauses {
			info := ClauseInfo{
				Name:      clause.Name,
				Args:      []compiler.Param{},
				Mintimes:  clause.MinTimes,
				Maxtimes:  clause.MaxTimes,
				HashCalls: clause.HashCalls,
			}
			if info.Mintimes == nil {
				info.Mintimes = []string{}
			}
			if info.Maxtimes == nil {
				info.Maxtimes = []string{}
			}

			for _, p := range clause.Params {
				info.Args = append(info.Args, compiler.Param{Name: p.Name, Type: p.Type})
			}

			for _, value := range clause.Values {
				info.Values = append(info.Values, value)
			}

			resp.Clauses = append(resp.Clauses, info)
		}
	}

	buf := new(bytes.Buffer)
	for _, step := range contract.Steps {
		fmt.Fprintf(buf, "%s ", step.Opcodes)
	}
	resp.Opcodes = buf.String()

	return resp, nil
}
