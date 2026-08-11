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

func (b *Bank) Deposit(accountID string, amount float64) (methodResponse, error) {
	if acc, ok := b.SavingsAccounts[accountID]; ok {
		before := acc.Balance
		resp, err := acc.Deposit(amount)
		b.Ledger.RecordTransaction(TransactionRecord{
			TransactionID: util.GenerateTransactionID(),
			AccountID:     accountID,
			OperationType: "deposit",
			Amount:        currency.NormalizeAmountByRound(amount, acc.Currency),
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
	if acc, ok := b.CurrentAccounts[accountID]; ok {
		before := acc.Balance
		resp, err := acc.Deposit(amount)
		b.Ledger.RecordTransaction(TransactionRecord{
			TransactionID: util.GenerateTransactionID(),
			AccountID:     accountID,
			OperationType: "deposit",
			Amount:        currency.NormalizeAmountByRound(amount, acc.Currency),
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
	return methodResponse{}, fmt.Errorf("account %s not found", accountID)
}
