package domain

func lockInOrder(a, b *Account) {
	if a.AccountID < b.AccountID {
		a.mu.Lock()
		b.mu.Lock()
	} else {
		b.mu.Lock()
		a.mu.Lock()
	}
}

func unlockInOrder(a, b *Account) {
	a.mu.Unlock()
	b.mu.Unlock()
}
