package api

import (
	"context"
	"encoding/hex"

	"github.com/bytom/blockchain/txbuilder"
	"github.com/bytom/crypto/ed25519/chainkd"
)

func (a *API) createAccountReceiver(ctx context.Context, ins struct {
	AccountID    string `json:"account_id"`
	AccountAlias string `json:"account_alias"`
}) Response {
	accountID := ins.AccountID
	if ins.AccountAlias != "" {
		account, err := a.wallet.AccountMgr.FindByAlias(ctx, ins.AccountAlias)
		if err != nil {
			return NewErrorResponse(err)
		}

		accountID = account.ID
	}

	program, err := a.wallet.AccountMgr.CreateAddress(ctx, accountID, false)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(&txbuilder.Receiver{
		ControlProgram: program.ControlProgram,
		Address:        program.Address,
	})
}

// AccountPubkey is structure of account pubkey
type AccountPubkey struct {
	Root   chainkd.XPub `json:"root_xpub"`
	Pubkey string       `json:"pubkey"`
	Path   []string     `json:"pubkey_derivation_path"`
}

// CreatePubkeyInfo creates a new public key for an account
func (a *API) createAccountPubkey(ctx context.Context, ins struct {
	AccountID    string `json:"account_id"`
	AccountAlias string `json:"account_alias"`
}) Response {
	accountID := ins.AccountID
	if ins.AccountAlias != "" {
		account, err := a.wallet.AccountMgr.FindByAlias(ctx, ins.AccountAlias)
		if err != nil {
			return NewErrorResponse(err)
		}

		accountID = account.ID
	}

	rootXPub, pubkey, path, err := a.wallet.AccountMgr.CreatePubkey(ctx, accountID)
	if err != nil {
		return NewErrorResponse(err)
	}

	var pathStr []string
	for _, p := range path {
		pathStr = append(pathStr, hex.EncodeToString(p))
	}

	return NewSuccessResponse(&AccountPubkey{
		Root:   rootXPub,
		Pubkey: hex.EncodeToString(pubkey),
		Path:   pathStr,
	})
}
