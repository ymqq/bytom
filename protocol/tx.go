package protocol

import (
	"github.com/bytom/errors"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/protocol/bc/legacy"
	"github.com/bytom/protocol/validation"
	"fmt"
)

// ErrBadTx is returned for transactions failing validation
var ErrBadTx = errors.New("invalid transaction")

// ValidateTx validates the given transaction. A cache holds
// per-transaction validation results and is consulted before
// performing full validation.
func (c *Chain) ValidateTx(tx *legacy.Tx) error {
	newTx := tx.Tx
	if err := c.checkIssuanceWindow(newTx); err != nil {
		return err
	}
	if ok := c.txPool.HaveTransaction(&newTx.ID); ok {
		return c.txPool.GetErrCache(&newTx.ID)
	}

	oldBlock, err := c.GetBlockByHash(c.state.hash)
	if err != nil {
		return err
	}
	block := legacy.MapBlock(oldBlock)
	fee, err := validation.ValidateTx(newTx, block)

	if err != nil {
		c.txPool.AddErrCache(&newTx.ID, err)
		return err
	}
	fmt.Println("after run validation.ValidateTx, fee:", fee)

	c.txPool.AddTransaction(tx, block.BlockHeader.Height, fee)
	return errors.Sub(ErrBadTx, err)
}

func (c *Chain) checkIssuanceWindow(tx *bc.Tx) error {
	if c.MaxIssuanceWindow == 0 {
		return nil
	}
	for _, entryID := range tx.InputIDs {
		if _, err := tx.Issuance(entryID); err == nil {
			if tx.MinTimeMs+bc.DurationMillis(c.MaxIssuanceWindow) < tx.MaxTimeMs {
				return errors.WithDetailf(ErrBadTx, "issuance input's time window is larger than the network maximum (%s)", c.MaxIssuanceWindow)
			}
		}
	}
	return nil
}

//get the used gas for program
func (c *Chain) GetUsedGas(tx *legacy.Tx) (uint64, error) {
	newTx := tx.Tx

	oldBlock, err := c.GetBlockByHash(c.state.hash)
	if err != nil {
		return 0, err
	}
	block := legacy.MapBlock(oldBlock)
	usedgas, err := validation.GetTxUsedGas(newTx, block)

	if err != nil {
		c.txPool.AddErrCache(&newTx.ID, err)
		return 0, err
	}
	fmt.Println("after program run in the virtual Machine, UsedGas:", usedgas)

	return usedgas, errors.Sub(ErrBadTx, err)
}