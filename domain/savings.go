package domain

import (
	"fmt"
	"gobanksystem/currency"
)

type SavingsAccount struct {
	Account
	InterestEarned int64
	InterestRate   float64
}

func (s *SavingsAccount) GenerateSummary() string {
	balance := currency.DenormalizeAmount(s.Balance, s.Currency)
	return fmt.Sprintf("Savings Account — %s (ID: %s)\nBalance: %.2f %s | Interest Earned: %.2f %s",
		s.Name,
		s.AccountID,
		balance,
		s.Currency.Name,
		currency.DenormalizeAmount(s.InterestEarned, s.Currency),
		s.Currency.Name,
	)
}

func (s *SavingsAccount) Deposit(amount float64) (methodResponse, error) {
	if amount <= 0 {
		return methodResponse{}, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsLocked {
		return methodResponse{}, &AccountLockedError{
			AccountID: s.AccountID,
			Reason:    s.LockedReason,
		}
	}

	s.Balance += normalizeAmountByFloor(amount, s.Currency)
	s.AuditLogs = append(s.AuditLogs, fmt.Sprintf("Deposited: %.2f", amount))

	return methodResponse{Message: "Deposit Successful", Status: true}, nil
}

func (s *SavingsAccount) Withdraw(amount float64) (methodResponse, error) {
	if amount <= 0 {
		return methodResponse{}, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsLocked {
		return methodResponse{}, &AccountLockedError{
			AccountID: s.AccountID,
			Reason:    s.LockedReason,
		}
	}

	amountInt := currency.NormalizeAmountByRound(amount, s.Currency)
	if s.Balance < amountInt {
		s.FailedAttempts++
		s.AuditLogs = append(s.AuditLogs, fmt.Sprintf("Failed withdrawal attempt: %.2f", amount))

		if s.FailedAttempts > 3 {
			s.Lock("Too many attempts to withdraw on insufficient funds")
			return methodResponse{}, &AccountLockedError{
				AccountID: s.AccountID,
				Reason:    "Too many attempts to withdraw on insufficient funds",
			}
		}
		return methodResponse{}, &InsufficientFundsError{
			AccountID:       s.AccountID,
			AmountRequested: amountInt,
			Deficit:         amountInt - s.Balance,
		}
	}

	s.Balance -= amountInt
	s.FailedAttempts = 0
	s.AuditLogs = append(s.AuditLogs, fmt.Sprintf("Withdrew: %.2f", amount))

	return methodResponse{Message: "Withdrawal Successful", Status: true}, nil
}

func (s *SavingsAccount) ApplyDailyInterest() (methodResponse, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsLocked {
		return methodResponse{}, &AccountLockedError{
			AccountID: s.AccountID,
			Reason:    "cannot apply interest: account is locked",
		}
	}

	if s.Balance <= 0 {
		return methodResponse{Message: "No balance to accrue interest", Status: false}, nil
	}

	dailyRate := s.InterestRate / 365.0
	earnedFloat := float64(s.Balance) * dailyRate

	earnedInt := currency.NormalizeAmountByRound(earnedFloat, s.Currency)
	if earnedInt <= 0 {
		return methodResponse{Message: "Interest accrued is below minimum curreny unit", Status: false}, nil
	}

	s.Balance += earnedInt
	s.InterestEarned += earnedInt
	s.AuditLogs = append(s.AuditLogs, fmt.Sprintf("Daily interest applied: %d units accrued", earnedInt))

	return methodResponse{Message: "Daily Interest Applied Successfully", Status: true}, nil
}
