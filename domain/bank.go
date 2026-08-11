package domain

import (
	"fmt"
	"gobanksystem/currency"
	"gobanksystem/util"
	"time"
)

type Staff struct {
	StaffID              string
	Name                 string
	Position             string
	SalaryValue          int64
	BankAccountID        string
	LastSalaryPaidAt     time.Time
	LastSalaryPaidAmount int64
}

type Bank struct {
	SavingsAccounts map[string]*SavingsAccount
	CurrentAccounts map[string]*CurrentAccount
	Customers       map[string]*Customer
	Staff           map[string]*Staff
	Ledger          *TransactionLedger
}

// Helper
func (b *Bank) recordAndReturn(accountID, opType string, amount float64, accBalanceBefore, accBalanceAfter int64, curr currency.Currency, resp methodResponse, err error) (methodResponse, error) {
	b.Ledger.RecordTransaction(TransactionRecord{
		TransactionID: util.GenerateTransactionID(),
		AccountID:     accountID,
		OperationType: opType,
		Amount:        currency.NormalizeAmountByRound(amount, curr),
		BalanceBefore: accBalanceBefore,
		BalanceAfter:  accBalanceAfter,
		Timestamp:     time.Now(),
		Success:       err == nil,
		Note: func() string {
			if err != nil {
				return err.Error()
			}
			return resp.Message
		}(),
	})
	return resp, err
}

func (b *Bank) Deposit(accountID string, amount float64) (methodResponse, error) {
	if acc, ok := b.SavingsAccounts[accountID]; ok {
		before := acc.Balance
		resp, err := acc.Deposit(amount)
		return b.recordAndReturn(accountID, "deposit", amount, before, acc.Balance, acc.Currency, resp, err)
	}
	if acc, ok := b.CurrentAccounts[accountID]; ok {
		before := acc.Balance
		resp, err := acc.Deposit(amount)
		return b.recordAndReturn(accountID, "deposit", amount, before, acc.Balance, acc.Currency, resp, err)
	}
	return methodResponse{}, fmt.Errorf("account %s not found", accountID)
}

func (b *Bank) Withdraw(accountID string, amount float64) (methodResponse, error) {
	if acc, ok := b.SavingsAccounts[accountID]; ok {
		before := acc.Balance
		resp, err := acc.Withdraw(amount)
		return b.recordAndReturn(accountID, "withdraw", amount, before, acc.Balance, acc.Currency, resp, err)
	}
	if acc, ok := b.CurrentAccounts[accountID]; ok {
		before := acc.Balance
		resp, err := acc.Withdraw(amount)
		return b.recordAndReturn(accountID, "withdraw", amount, before, acc.Balance, acc.Currency, resp, err)
	}
	return methodResponse{}, fmt.Errorf("account %s not found", accountID)
}

func (b *Bank) Transfer(fromAccountID, toAccountID string, amount float64, exchangeRate float64) (methodResponse, error) {
	from, okFrom := b.CurrentAccounts[fromAccountID]
	to, okTo := b.CurrentAccounts[toAccountID]

	if !okFrom || !okTo {
		return methodResponse{}, fmt.Errorf("both accounts must be current accounts for transfer")
	}

	beforeFrom := from.Balance
	beforeTo := to.Balance

	resp, err := from.Transfer(to, amount, exchangeRate)

	// Record debit side
	b.Ledger.RecordTransaction(TransactionRecord{
		TransactionID: util.GenerateTransactionID(),
		AccountID:     fromAccountID,
		OperationType: "transfer_debit",
		Amount:        currency.NormalizeAmountByRound(amount, from.Currency),
		BalanceBefore: beforeFrom,
		BalanceAfter:  from.Balance,
		Timestamp:     time.Now(),
		Success:       err == nil,
		Note: func() string {
			if err != nil {
				return err.Error()
			}
			return fmt.Sprintf("Transfer to %s", toAccountID)
		}(),
	})

	// Record credit side only if transfer succeeded
	if err == nil {
		b.Ledger.RecordTransaction(TransactionRecord{
			TransactionID: util.GenerateTransactionID(),
			AccountID:     toAccountID,
			OperationType: "transfer_credit",
			Amount:        currency.NormalizeAmountByRound(amount*exchangeRate, to.Currency),
			BalanceBefore: beforeTo,
			BalanceAfter:  to.Balance,
			Timestamp:     time.Now(),
			Success:       true,
			Note:          fmt.Sprintf("Transfer from %s", fromAccountID),
		})
	}

	return resp, err
}

func (b *Bank) ApplyDailyInterest(accountID string) (methodResponse, error) {
	acc, ok := b.SavingsAccounts[accountID]
	if !ok {
		return methodResponse{}, fmt.Errorf("savings account %s not found", accountID)
	}

	before := acc.Balance
	resp, err := acc.ApplyDailyInterest()

	b.Ledger.RecordTransaction(TransactionRecord{
		TransactionID: util.GenerateTransactionID(),
		AccountID:     accountID,
		OperationType: "interest",
		Amount:        acc.Balance - before,
		BalanceBefore: before,
		BalanceAfter:  acc.Balance,
		Timestamp:     time.Now(),
		Success:       err == nil,
		Note: func() string {
			if err != nil {
				return err.Error()
			}
			return resp.Message
		}(),
	})

	return resp, err
}
