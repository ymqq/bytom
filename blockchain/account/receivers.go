package account

import (
	"context"
	"time"
	"github.com/bytom/blockchain/signers"
	"github.com/bytom/blockchain/txbuilder"
	"github.com/bytom/crypto/ed25519"
	"github.com/bytom/crypto/ed25519/chainkd"
	"github.com/bytom/errors"
)

const defaultReceiverExpiry = 30 * 24 * time.Hour // 30 days

func (m *Manager) CreatePubkey(ctx context.Context, accID, accAlias string) (rootXPub chainkd.XPub, pubkey ed25519.PublicKey, path [][]byte, err error) {
	if accAlias != "" {
		var s *signers.Signer
		s, err = m.FindByAlias(ctx, accAlias)
		if err != nil {
			return
		}
		accID = s.ID
	}

	return m.createPubkey(ctx, accID)
}

// CreateReceiver creates a new account receiver for an account
// with the provided expiry. If a zero time is provided for the
// expiry, a default expiry of 30 days from the current time is
// used.
func (m *Manager) CreateReceiver(ctx context.Context, accID, accAlias string, expiresAt time.Time) (*txbuilder.Receiver, error) {
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(defaultReceiverExpiry)
	}

	if accAlias != "" {
		s, err := m.FindByAlias(ctx, accAlias)
		if err != nil {
			return nil, err
		}
		accID = s.ID
	}

	cp, err := m.CreateControlProgram(ctx, accID, false, expiresAt)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &txbuilder.Receiver{
		ControlProgram: cp,
		ExpiresAt:      expiresAt,
	}, nil
}
