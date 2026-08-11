package data

import (
	"encoding/json"
	"fmt"
	"os"
)

type SeedAccount struct {
	AccountID      string  `json:"account_id"`
	Type           string  `json:"type"`
	Currency       string  `json:"currency"`
	Balance        float64 `json:"balance"`
	InterestRate   float64 `json:"interest_rate"`
	OverdraftLimit float64 `json:"overdraft_limit"`
}

type SeedCustomer struct {
	CustomerID string        `json:"customer_id"`
	Name       string        `json:"name"`
	Accounts   []SeedAccount `json:"accounts"`
}

type SeedStaff struct {
	StaffID       string `json:"staff_id"`
	Name          string `json:"name"`
	Position      string `json:"position"`
	SalaryValue   int64  `json:"salary_value"`
	BankAccountID string `json:"bank_account_id"`
}

type SeedOperation struct {
	Type          string  `json:"type"`
	AccountID     string  `json:"account_id"`
	FromAccountID string  `json:"from_account_id"`
	ToAccountID   string  `json:"to_account_id"`
	Amount        float64 `json:"amount"`
	StaffID       string  `json:"staff_id"`
	Note          string  `json:"note"`
}

type SeedData struct {
	ExchangeRates map[string]map[string]float64 `json:"exchange_rates"`
	Customers     []SeedCustomer                `json:"customers"`
	Staff         []SeedStaff                   `json:"staff"`
	Operations    []SeedOperation               `json:"operations"`
}

func Load(path string) (SeedData, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return SeedData{}, fmt.Errorf("failed to read seed file: %w", err)
	}

	var data SeedData
	if err := json.Unmarshal(file, &data); err != nil {
		return SeedData{}, fmt.Errorf("failed to parse seed data: %w", err)
	}

	return data, nil
}
