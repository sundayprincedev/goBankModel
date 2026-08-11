package domain

import (
	"fmt"
	"gobanksystem/currency"
)

var normalizeAmountByFloor = currency.NormalizeAmountByFloor

type CurrentAccount struct {
	Account
	OverdraftLimit int64
}

func (c *CurrentAccount) GenerateSummary() string {
	balance := currency.DenormalizeAmount(c.Balance, c.Currency)
	return fmt.Sprintf("Current Account — %s (ID: %s)\nBalance: %.2f %s",
		c.Name,
		c.AccountID,
		balance,
		c.Currency.Name,
	)
}

func (c *CurrentAccount) Deposit(amount float64) (methodResponse, error) {
	if amount <= 0 {
		return methodResponse{}, ErrInvalidAmount
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsLocked {
		return methodResponse{}, &AccountLockedError{
			AccountID: c.AccountID,
			Reason:    c.LockedReason,
		}
	}

	c.Balance += normalizeAmountByFloor(amount, c.Currency)
	c.AuditLogs = append(c.AuditLogs, fmt.Sprintf("Deposited: %.2f", amount))

	return methodResponse{Message: "Deposit Successful", Status: true}, nil
}

func (c *CurrentAccount) Withdraw(amount float64) (methodResponse, error) {
	if amount <= 0 {
		return methodResponse{}, ErrInvalidAmount
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsLocked {
		return methodResponse{}, &AccountLockedError{
			AccountID: c.AccountID,
			Reason:    c.LockedReason,
		}
	}

	amountInt := currency.NormalizeAmountByRound(amount, c.Currency)
	if c.Balance < amountInt {
		c.FailedAttempts++
		c.AuditLogs = append(c.AuditLogs, fmt.Sprintf("Failed withdrawal attempt: %.2f", amount))
		if c.FailedAttempts > 3 {
			c.Lock("Too many attempts to withdraw on insufficient funds")
		}
		return methodResponse{}, &InsufficientFundsError{
			AccountID:       c.AccountID,
			AmountRequested: amountInt,
			Deficit:         amountInt - c.Balance,
		}
	}

	c.Balance -= amountInt
	c.FailedAttempts = 0
	c.AuditLogs = append(c.AuditLogs, fmt.Sprintf("Withdrew: %.2f", amount))

	return methodResponse{Message: "Withdrawal Successful", Status: true}, nil
}

func (c *CurrentAccount) Transfer(
	to *CurrentAccount,
	amount float64,
	exchangeRate float64,
) (methodResponse, error) {

	if amount <= 0 {
		return methodResponse{}, ErrInvalidAmount
	}

	lockInOrder(&c.Account, &to.Account)
	defer unlockInOrder(&c.Account, &to.Account)

	if c.IsLocked {
		return methodResponse{}, &TransferError{
			FromID: c.AccountID,
			ToID:   to.AccountID,
			Amount: currency.NormalizeAmountByRound(amount, c.Currency),
			Cause: &AccountLockedError{
				AccountID: c.AccountID,
				Reason:    c.LockedReason,
			},
		}
	}
	if to.IsLocked {
		return methodResponse{}, &TransferError{
			FromID: c.AccountID,
			ToID:   to.AccountID,
			Amount: currency.NormalizeAmountByRound(amount, c.Currency),
			Cause: &AccountLockedError{
				AccountID: to.AccountID,
				Reason:    to.LockedReason,
			},
		}
	}

	debitAmount := currency.NormalizeAmountByRound(amount, c.Currency)
	if c.Balance < debitAmount {
		c.FailedAttempts++
		c.AuditLogs = append(c.AuditLogs, fmt.Sprintf("Failed transfer attempt: %.2f", amount))
		if c.FailedAttempts > 3 {
			c.Lock("Too many failed transfer attempts")
		}

		insufficientErr := &InsufficientFundsError{
			AccountID:       c.AccountID,
			AmountRequested: debitAmount,
			Deficit:         debitAmount - c.Balance,
		}
		return methodResponse{}, &TransferError{
			FromID: c.AccountID,
			ToID:   to.AccountID,
			Amount: debitAmount,
			Cause:  insufficientErr,
		}
	}

	convertedAmount := amount * exchangeRate
	creditAmount := currency.NormalizeAmountByRound(convertedAmount, to.Currency)

	c.Balance -= debitAmount
	to.Balance += creditAmount
	c.FailedAttempts = 0

	c.AuditLogs = append(c.AuditLogs, fmt.Sprintf(
		"Transferred: %.2f %s to account %s (rate: %.4f)",
		amount, c.Currency.Name, to.AccountID, exchangeRate,
	))
	to.AuditLogs = append(to.AuditLogs, fmt.Sprintf(
		"Received: %.2f %s from account %s",
		convertedAmount, to.Currency.Name, c.AccountID,
	))

	return methodResponse{Message: "Transfer Successful", Status: true}, nil
}
