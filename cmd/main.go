package main

import (
	"errors"
	"fmt"
	"gobanksystem/currency"
	"gobanksystem/data"
	"gobanksystem/domain"
	"gobanksystem/systems"
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

	savingsAccounts := make(map[string]*domain.SavingsAccount)
	currentAccounts := make(map[string]*domain.CurrentAccount)

	for _, seedCustomer := range seedData.Customers {
		customer := domain.Customer{
			Name:       seedCustomer.Name,
			CustomerID: seedCustomer.CustomerID,
		}

		for _, seedAccount := range seedCustomer.Accounts {
			curr, ok := currency.CurrencyExists(seedAccount.Currency)
			if !ok {
				fmt.Printf("Unknown currency %s for account %s — skipping\n",
					seedAccount.Currency, seedAccount.AccountID)
				continue
			}

			base := domain.BaseAccount{
				Customer:     customer,
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
				savingsAccounts[seedAccount.AccountID] = acc

			case "current":
				acc := &domain.CurrentAccount{
					Account: domain.Account{
						BaseAccount: base,
						Balance:     currency.NormalizeAmountByRound(seedAccount.Balance, curr),
					},
					OverdraftLimit: currency.NormalizeAmountByRound(seedAccount.OverdraftLimit, curr),
				}
				currentAccounts[seedAccount.AccountID] = acc
			}
		}
	}

	fmt.Println("═══════════════════════════════════")
	fmt.Println("        RUNNING OPERATIONS         ")
	fmt.Println("═══════════════════════════════════")

	for _, op := range seedData.Operations {
		wg.Add(1)
		go func(op data.SeedOperation) {
			defer wg.Done()
			fmt.Printf("\n▶ %s", op.Type)
			if op.Note != "" {
				fmt.Printf(" (%s)", op.Note)
			}
			fmt.Println()

			switch op.Type {

			case "deposit":
				time.Sleep(10 * time.Millisecond)
				acc, ok := currentAccounts[op.AccountID]
				if !ok {
					sacc, ok := savingsAccounts[op.AccountID]
					if !ok {
						fmt.Printf("  Account %s not found\n", op.AccountID)
						return
					}
					resp, err := sacc.Deposit(op.Amount)
					printResult(resp.Message, err)
					return
				}
				resp, err := acc.Deposit(op.Amount)
				printResult(resp.Message, err)

			case "withdraw":
				time.Sleep(10 * time.Millisecond)
				acc, ok := currentAccounts[op.AccountID]
				if !ok {
					sacc, ok := savingsAccounts[op.AccountID]
					if !ok {
						fmt.Printf("  Account %s not found\n", op.AccountID)
						return
					}
					resp, err := sacc.Withdraw(op.Amount)
					printResult(resp.Message, err)
					return
				}
				resp, err := acc.Withdraw(op.Amount)
				printResult(resp.Message, err)

			case "transfer":
				time.Sleep(10 * time.Millisecond)
				from, ok := currentAccounts[op.FromAccountID]
				if !ok {
					fmt.Printf("  From account %s not found\n", op.FromAccountID)
					return
				}
				to, ok := currentAccounts[op.ToAccountID]
				if !ok {
					fmt.Printf("  To account %s not found\n", op.ToAccountID)
					return
				}

				fromCode := from.CurrencyCode
				toCode := to.CurrencyCode
				rate, ok := seedData.ExchangeRates[fromCode][toCode]
				if !ok {
					fmt.Printf("  No exchange rate found for %s → %s\n", fromCode, toCode)
					return
				}

				resp, err := from.Transfer(to, op.Amount, rate)
				printResult(resp.Message, err)

			case "apply_interest":
				time.Sleep(10 * time.Millisecond)
				sacc, ok := savingsAccounts[op.AccountID]
				if !ok {
					fmt.Printf("  Savings account %s not found\n", op.AccountID)
					return
				}
				resp, err := sacc.ApplyDailyInterest()
				printResult(resp.Message, err)
			}
		}(op)
	}

	wg.Wait()

	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("            REPORTS                ")
	fmt.Println("═══════════════════════════════════")

	for _, acc := range savingsAccounts {
		systems.ReportSystem(acc)
	}
	for _, acc := range currentAccounts {
		systems.ReportSystem(acc)
	}

	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("           AUDIT LOGS              ")
	fmt.Println("═══════════════════════════════════")

	for _, acc := range savingsAccounts {
		fmt.Printf("\n— %s (%s) —\n", acc.Name, acc.AccountID)
		systems.AuditSystem(acc)
	}
	for _, acc := range currentAccounts {
		fmt.Printf("\n— %s (%s) —\n", acc.Name, acc.AccountID)
		systems.AuditSystem(acc)
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
		fmt.Printf("  ✗ %s\n", err)
		return
	} else {
		fmt.Printf("  ✓ %s\n", message)
	}
}
