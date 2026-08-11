package domain

import (
	"fmt"
	"gobanksystem/currency"
	"gobanksystem/util"
	"sync"
)

type AuditLogs []string

type BaseAccount struct {
	Customer
	AccountID    string
	Currency     currency.Currency
	CurrencyCode string
	AuditLogs
}

type Account struct {
	BaseAccount
	mu             sync.Mutex
	Balance        int64
	LockedReason   string
	FailedAttempts int32
	IsLocked       bool
}

func (b *BaseAccount) Audit() []string {
	return b.AuditLogs
}

type methodResponse = util.MethodResponse

func (a *Account) Lock(reason string) {
	//No lock here because if mutex from higher scope already holds, redudant could cause deadlock.
	a.IsLocked = true
	a.LockedReason = reason
	a.AuditLogs = append(a.AuditLogs, fmt.Sprintf("Account locked: %s", reason))
}
