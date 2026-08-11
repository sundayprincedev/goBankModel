package util

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"time"
)

var txnCounter atomic.Uint64

func GenerateTransactionID() string {
	b := make([]byte, 4)
	rand.Read(b)
	seq := txnCounter.Add(1)
	return fmt.Sprintf("TXN-%s-%06d-%x",
		time.Now().UTC().Format("20060102-150405"),
		seq,
		b,
	)
}
