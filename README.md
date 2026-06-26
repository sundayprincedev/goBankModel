# BankSystem

This isn't a CRUD app. There's no REST API, no frontend, no product brief behind it.

I built this as a learning sandbox. Every time I pick up a new concept in systems engineering or financial infrastructure, I implement it here — on a real-ish domain where the stakes of getting it wrong are obvious. A bank ledger is the perfect pressure test for that. Rounding errors matter. Bad control flow matters. Domain boundaries matter.

So this project grows with me. I add a stair each time I learn something worth implementing.

I also write about this stuff weekly on Medium — the thinking behind the decisions, the concepts I'm picking up, how I'm making sense of it all. If the code interests you, the writing might too: [medium.com/@sundayprincedev](https://medium.com/@sundayprincedev)

---

## Why Go

Go forces you to be intentional. There's no runtime magic absorbing your mistakes — what you write is what runs. For the kind of systems I want to understand deeply, that's the right constraint to work under. I wanted to feel every design decision.

---

## How the Code Is Organized

I structured this so each layer has one job and can't reach sideways into something it shouldn't touch:

```
Bank/
├── cmd/
│   └── main.go              # Entry point — bootstraps the system, runs the execution script
├── currency/
│   └── currency.go          # Helpers for scaling, rounding, and formatting money
├── data/
│   ├── data.json            # Seed file — initial account state and exchange rates
│   └── reader.go            # Hydrates the system from that seed
├── domain/
│   ├── base.go              # Shared fields both account types embed
│   ├── current.go           # Current account logic (overdrafts, network transfers)
│   ├── savings.go           # Savings account logic (compound interest, lockout)
│   └── errors.go            # Structured errors used for control flow, not just logging
├── interfaces/
│   ├── auditable.go         # Contract for anything that can emit an audit log
│   └── reportable.go        # Contract for anything that can produce a summary
├── systems/
│   ├── audit.go             # Processes and prints immutable audit trails
│   ├── report.go            # Aggregates final balance sheets
│   └── transfer.go          # Core orchestrator — handles cross-currency settlements
└── util/
    └── handleMoney.go       # Isolated math helper for shifting balances safely
```

---

## Concepts I've Implemented So Far

Each of these is something I learned, then built into the system as a way of actually understanding it.

### Money Is Not a Float

`float64` has rounding errors baked into it at the binary level. Acceptable for a lot of use cases — completely unacceptable for a ledger. This was one of the first things I implemented after reading about it.

Every balance in this system lives as an `int64`, representing the smallest currency unit (Kobo, cents, depending on the currency). When a raw float comes in from outside, it gets normalized immediately — either rounded or floored — before anything touches it mathematically. It only converts back to a float at the very edge of the system, when printing to a string.

### Errors as Control Flow

I learned that in resilient systems, errors shouldn't just be messages — they should carry enough information to make decisions with. So I implemented two patterns:

**Sentinel errors** — simple constants like `ErrInvalidAmount` for fast binary checks.

**Struct errors** — like `InsufficientFundsError`, which carries the account ID and the exact deficit as fields.

The payoff is in the transfer orchestrator. When a Current Account transfer fails due to a shortfall, I use `errors.As()` to read the exact deficit out of the error struct, check whether the overdraft buffer can cover it, and re-route the execution to complete the transfer anyway. The error becomes a branch in the logic, not a dead end.

### Domain Boundaries Between Account Types

**Savings:**
- Hard floor on withdrawals — can never exceed the settled balance
- Daily compound interest — fractional gain against the annual rate, rounded to a clean integer before committing to state
- Three consecutive failed withdrawals flips the account to locked
- No network access — deposits and withdrawals only

**Current:**
- Can spend past zero up to a defined overdraft cap
- Network-enabled — this is the account that moves money across the system to other clients
- Same lockout logic as savings

### Cross-Currency Settlement Without Hardcoding Pairs

Maintaining a direct conversion table for every possible currency pair doesn't scale and couples the engine to specific market assumptions. Instead I implemented a universal anchor strategy:

1. Divide the source amount by the sender's exchange rate → value is now expressed in the anchor unit (dollar-based)
2. Run all balance verification against that base value
3. Multiply by the recipient's exchange rate → inflate to their currency, round to a clean integer, credit the account

The engine itself is currency-agnostic. Rates come in from the JSON seed file and that's the only coupling point.

---

## Running It

There's no server to start. `cmd/main.go` runs a deterministic execution script that puts everything through its paces:

1. Parses the seed data, stands up accounts, loads exchange rates into memory
2. Runs a scripted sequence: standard transfers, cross-currency settlements, intentional consecutive failures to trigger the security lockout, a daily interest clock tick
3. Passes everything through the `Auditable` and `Reportable` interfaces, printing the full transaction trail and final balance sheets

Read the output top to bottom — it tells the whole story of what the system did and why.

---

## Follow the Learning

I write weekly on Medium about the concepts I'm picking up and how I'm thinking through implementing them. If you want to follow along as this project grows:

[medium.com/@sundayprincedev](https://medium.com/@sundayprincedev)

If you want to talk about anything you see here — the design decisions, the Go specifics, the concepts — I'm easy to reach.
