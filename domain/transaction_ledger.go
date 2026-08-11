package domain

import (
	"sync"
	"time"
)

type TransactionRecord struct {
	TransactionID string
	AccountID     string
	OperationType string
	Amount        int64
	BalanceBefore int64
	BalanceAfter  int64
	Timestamp     time.Time
	Note          string
	Success       bool
}
type TransactionLedger struct {
	mu           sync.RWMutex
	transactions map[string][]TransactionRecord
}

func NewTransactionLedger() *TransactionLedger {
	return &TransactionLedger{
		transactions: make(map[string][]TransactionRecord),
	}
}
func (t *TransactionLedger) RecordTransaction(data TransactionRecord) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.transactions[data.AccountID] = append(t.transactions[data.AccountID], data)
}
func (t *TransactionLedger) AccountTransactions(accountID string) ([]TransactionRecord, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tran, ok := t.transactions[accountID]
	if !ok {
		return nil, &AccountTransactionsError{
			AccountID: accountID,
			Reason:    "no entry found in ledger",
		}
	}
	return tran, nil
}

func (t *TransactionLedger) AccountTransactionByStatus(accountID string, status bool) ([]TransactionRecord, error) {
	transactions, err := t.AccountTransactions(accountID)
	if err != nil {
		return nil, &LedgerError{
			Level:     "filter by status function",
			Operation: "get account transactions by status",
			Cause:     err,
		}
	}
	statusTransactions := make([]TransactionRecord, 0, len(transactions))
	for _, val := range transactions {
		if val.Success == status {
			statusTransactions = append(statusTransactions, val)
		}
	}
	return statusTransactions, nil
}
