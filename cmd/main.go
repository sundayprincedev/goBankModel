package main

import (
	"errors"
	"fmt"
	"gobanksystem/currency"
	"gobanksystem/data"
	"gobanksystem/domain"
	"gobanksystem/systems"
	"gobanksystem/util"
	"sync"
	"time"
)

var printMu sync.Mutex
var wg sync.WaitGroup

func main() {
	seedData, err := data.Load("data/data.json")
	if err != nil {
		fmt.Println("Failed to load seed data:", err)
		return
	}

	bank := domain.NewBank()

	// Load customers and accounts
	for _, seedCustomer := range seedData.Customers {
		customer := &domain.Customer{
			Name:       seedCustomer.Name,
			CustomerID: seedCustomer.CustomerID,
		}
		bank.Customers[seedCustomer.CustomerID] = customer

		for _, seedAccount := range seedCustomer.Accounts {
			curr, ok := currency.CurrencyExists(seedAccount.Currency)
			if !ok {
				fmt.Printf("Unknown currency %s for account %s — skipping\n",
					seedAccount.Currency, seedAccount.AccountID)
				continue
			}

			base := domain.BaseAccount{
				Customer:     *customer,
				AccountID:    seedAccount.AccountID,
				Currency:     curr,
				CurrencyCode: seedAccount.Currency,
				AuditLogs:    make(domain.AuditLogs, 0, 50),
			}

			switch seedAccount.Type {
			case "savings":
				acc := &domain.SavingsAccount{
					Account: domain.Account{
						BaseAccount: base,
						Balance:     currency.NormalizeAmountByRound(seedAccount.Balance, curr),
					},
					InterestRate: seedAccount.InterestRate,
				}
				bank.SavingsAccounts[seedAccount.AccountID] = acc

			case "current":
				acc := &domain.CurrentAccount{
					Account: domain.Account{
						BaseAccount: base,
						Balance:     currency.NormalizeAmountByRound(seedAccount.Balance, curr),
					},
					OverdraftLimit: currency.NormalizeAmountByRound(seedAccount.OverdraftLimit, curr),
				}
				bank.CurrentAccounts[seedAccount.AccountID] = acc
			}
		}
	}

	// Load staff
	for _, seedStaff := range seedData.Staff {
		staff := &domain.Staff{
			StaffID:       seedStaff.StaffID,
			Name:          seedStaff.Name,
			Position:      seedStaff.Position,
			SalaryValue:   seedStaff.SalaryValue,
			BankAccountID: seedStaff.BankAccountID,
		}
		bank.Staff[seedStaff.StaffID] = staff
	}

	fmt.Println("═══════════════════════════════════")
	fmt.Println("        RUNNING OPERATIONS         ")
	fmt.Println("═══════════════════════════════════")

	for _, op := range seedData.Operations {
		wg.Add(1)
		go func(op data.SeedOperation) {
			defer wg.Done()

			fmt.Printf("\n▶ %s", op.Type)
			if op.StaffID != "" {
				fmt.Printf(" [%s]", op.StaffID)
			}
			if op.Note != "" {
				fmt.Printf(" (%s)", op.Note)
			}
			fmt.Println()

			var resp util.MethodResponse
			var opErr error

			switch op.Type {
			case "deposit":
				time.Sleep(10 * time.Millisecond)
				resp, opErr = bank.Deposit(op.AccountID, op.Amount)

			case "withdraw":
				time.Sleep(10 * time.Millisecond)
				resp, opErr = bank.Withdraw(op.AccountID, op.Amount)

			case "transfer":
				time.Sleep(10 * time.Millisecond)
				fromCode := ""
				toCode := ""
				if from, ok := bank.CurrentAccounts[op.FromAccountID]; ok {
					fromCode = from.CurrencyCode
				}
				if to, ok := bank.CurrentAccounts[op.ToAccountID]; ok {
					toCode = to.CurrencyCode
				}
				if fromCode == "" || toCode == "" {
					printResult("Transfer failed", fmt.Errorf("account not found"))
					return
				}
				rate, ok := seedData.ExchangeRates[fromCode][toCode]
				if !ok {
					printResult("Transfer failed", fmt.Errorf("no exchange rate for %s → %s", fromCode, toCode))
					return
				}
				resp, opErr = bank.Transfer(op.FromAccountID, op.ToAccountID, op.Amount, rate)

			case "apply_interest":
				time.Sleep(10 * time.Millisecond)
				resp, opErr = bank.ApplyDailyInterest(op.AccountID)

			case "pay_salary":
				time.Sleep(10 * time.Millisecond)
				resp, opErr = bank.PayStaffSalary(op.StaffID)

			default:
				printResult("", fmt.Errorf("unknown operation: %s", op.Type))
				return
			}

			printResult(resp.Message, opErr)
		}(op)
	}

	wg.Wait()

	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("            REPORTS                ")
	fmt.Println("═══════════════════════════════════")

	for _, acc := range bank.SavingsAccounts {
		systems.ReportSystem(acc)
	}
	for _, acc := range bank.CurrentAccounts {
		systems.ReportSystem(acc)
	}

	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("           AUDIT LOGS              ")
	fmt.Println("═══════════════════════════════════")

	for _, acc := range bank.SavingsAccounts {
		fmt.Printf("\n— %s (%s) —\n", acc.Name, acc.AccountID)
		systems.AuditSystem(acc)
	}
	for _, acc := range bank.CurrentAccounts {
		fmt.Printf("\n— %s (%s) —\n", acc.Name, acc.AccountID)
		systems.AuditSystem(acc)
	}

	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("        TRANSACTION LEDGER         ")
	fmt.Println("═══════════════════════════════════")

	allAccountIDs := make([]string, 0, len(bank.SavingsAccounts)+len(bank.CurrentAccounts))
	for id := range bank.SavingsAccounts {
		allAccountIDs = append(allAccountIDs, id)
	}
	for id := range bank.CurrentAccounts {
		allAccountIDs = append(allAccountIDs, id)
	}

	for _, accountID := range allAccountIDs {
		records, err := bank.Ledger.AccountTransactions(accountID)
		if err != nil {
			fmt.Printf("\n— Account %s —\n  (no transactions recorded)\n", accountID)
			continue
		}
		fmt.Printf("\n— Account %s (%d transactions) —\n", accountID, len(records))
		for _, r := range records {
			status := "✓"
			if !r.Success {
				status = "✗"
			}
			fmt.Printf("  %s [%s] %-16s | amount: %8d | %8d → %8d | %s\n",
				status,
				r.Timestamp.Format(time.RFC3339),
				r.OperationType,
				r.Amount,
				r.BalanceBefore,
				r.BalanceAfter,
				r.Note,
			)
		}
	}

	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("         SALARY AUDIT              ")
	fmt.Println("═══════════════════════════════════")

	for _, staff := range bank.Staff {
		status := "NEVER PAID"
		if !staff.LastSalaryPaidAt.IsZero() {
			status = fmt.Sprintf("last paid %s (%d units)",
				staff.LastSalaryPaidAt.Format(time.RFC3339),
				staff.LastSalaryPaidAmount)
		}

		warning := ""
		if staff.LastSalaryPaidAmount != 0 && staff.LastSalaryPaidAmount != staff.SalaryValue {
			warning = fmt.Sprintf(" ⚠ MISMATCH: expected %d, got %d",
				staff.SalaryValue, staff.LastSalaryPaidAmount)
		}

		records, _ := bank.Ledger.AccountTransactions(staff.BankAccountID)
		salaryCredits := 0
		totalPaid := int64(0)
		for _, r := range records {
			if r.OperationType == "salary_credit" && r.Success {
				salaryCredits++
				totalPaid += r.Amount
			}
		}

		fmt.Printf("\n  %s (%s) — %s\n", staff.Name, staff.Position, status)
		fmt.Printf("    Account: %s | Salary: %d | Credits: %d | Total Paid: %d%s\n",
			staff.BankAccountID, staff.SalaryValue, salaryCredits, totalPaid, warning)
	}
}

func printResult(message string, err error) {
	printMu.Lock()
	defer printMu.Unlock()
	if err != nil {
		var fundErr *domain.InsufficientFundsError
		if errors.As(err, &fundErr) {
			fmt.Printf("  ✗ Insufficient funds — short by %d units on account %s\n",
				fundErr.Deficit, fundErr.AccountID)
			return
		}
		var lockedErr *domain.AccountLockedError
		if errors.As(err, &lockedErr) {
			fmt.Printf("  ✗ Account locked — %s (account %s)\n",
				lockedErr.Reason, lockedErr.AccountID)
			return
		}
		fmt.Printf("  ✗ %s\n", err)
		return
	}
	fmt.Printf("  ✓ %s\n", message)
}
