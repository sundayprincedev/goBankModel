package domain

import (
	"errors"
	"fmt"
)

var ErrInvalidAmount = errors.New("invalid amount")
var ErrAccountLocked = errors.New("account locked")

type InsufficientFundsError struct {
	AccountID       string
	AmountRequested int64
	Deficit         int64
}

type AccountLockedError struct {
	AccountID string
	Reason    string
}

type TransferError struct {
	FromID string
	ToID   string
	Amount int64
	Cause  error
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient baseline balance: short of transaction target by %d units", e.Deficit)

}

func (e *AccountLockedError) Error() string {
	return fmt.Sprintf("account locked: %s", e.Reason)
}

func (e *TransferError) Error() string {
	return fmt.Sprintf("transfer from %s to %s of %d failed: %s",
		e.FromID, e.ToID, e.Amount, e.Cause)
}

func (e *TransferError) Unwrap() error {
	return e.Cause
}
