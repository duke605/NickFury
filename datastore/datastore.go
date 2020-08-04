package datastore

import (
	"context"

	"github.com/boltdb/bolt"
)

type contextKey int

var transactionKey contextKey

// Datastore stores data
type Datastore struct {
	*bolt.DB
}

// InTransaction pulls a existing transaction from the context or creates
// a new one if it does not exist and passes the transaction along with the
// context containing the transaction to the provided function.
//
// If writable is false, the created transaction will not be able to write to the
// datastore and attemoting to do so will return an error
func (ds *Datastore) InTransaction(ctx context.Context, writable bool, fn func(context.Context, *bolt.Tx) error) (err error) {
	var tx *bolt.Tx
	txI := ctx.Value(transactionKey)
	managed := false

	// Creating a new transaction and managing rollback when one
	// does not exist
	if txI == nil {
		txI, err = ds.Begin(writable)
		if err != nil {
			return err
		}

		defer func() {
			if tx.DB() != nil || err != nil {
				if writable {
					tx.Rollback()
				}
			} else if perr := recover(); perr != nil {
				if writable {
					tx.Rollback()
				}
				panic(perr)
			}
		}()

		managed = true
		ctx = context.WithValue(ctx, transactionKey, txI)
	}
	tx = txI.(*bolt.Tx)

	// Checking if the writable state of the tx is compatible
	if !tx.Writable() && writable {
		return ErrIncompatibleTransaction
	}

	err = fn(ctx, tx)
	if err == nil && managed {
		if writable {
			return tx.Commit()
		}

		return tx.Rollback()
	}

	return err
}
