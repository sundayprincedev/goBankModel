package domain

import (
	"fmt"
	"gobanksystem/currency"
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
	Balance        int64
	LockedReason   string
	FailedAttempts int32
	IsLocked       bool
}

func (b *BaseAccount) Audit() []string {
	return b.AuditLogs
}

func (a *Account) Lock(reason string) {
	a.IsLocked = true
	a.LockedReason = reason
	a.AuditLogs = append(a.AuditLogs, fmt.Sprintf("Account locked: %s", reason))
}
