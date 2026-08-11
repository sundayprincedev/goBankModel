package domain

import (
	"errors"
	"fmt"
)

var ErrInvalidAmount = errors.New("invalid amount")
var ErrAccountLocked = errors.New("account locked")
var ErrAccountNotFound = errors.New("account not found")

type InsufficientFundsError struct {
	AccountID       string
	AmountRequested int64
	Deficit         int64
}

type AccountLockedError struct {
	AccountID string
	Reason    string
}

type AccountTransactionsError struct {
	AccountID string
	Reason    string
}

type TransferError struct {
	FromID string
	ToID   string
	Amount int64
	Cause  error
}

type LedgerError struct {
	Level     string
	Operation string
	Cause     error
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient baseline balance: short of transaction target by %d units", e.Deficit)

}

func (e *AccountLockedError) Error() string {
	return fmt.Sprintf("account locked: %s", e.Reason)
}

func (e *AccountTransactionsError) Error() string {
	return fmt.Sprintf("unable to get account transactions: %s", e.Reason)
}

func (e *TransferError) Error() string {
	return fmt.Sprintf("transfer from %s to %s of %d failed: %s",
		e.FromID, e.ToID, e.Amount, e.Cause)
}

func (e *TransferError) Unwrap() error {
	return e.Cause
}

func (e *LedgerError) Error() string {
	return fmt.Sprintf("ledger op failed at %s for %s operation",
		e.Level, e.Operation)
}

func (e *LedgerError) Unwrap() error {
	return e.Cause
}
