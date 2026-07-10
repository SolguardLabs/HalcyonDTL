package main

type AccountStatus string

const (
	AccountActive     AccountStatus = "active"
	AccountRestricted AccountStatus = "restricted"
	AccountLiquidated AccountStatus = "liquidated"
)

type Account struct {
	ID                  AccountID           `json:"id"`
	Owner               string              `json:"owner"`
	Collateral          map[string]Amount   `json:"collateral"`
	ReservedMargin      map[string]Amount   `json:"reserved_margin"`
	RealizedFunding     map[string]int64    `json:"realized_funding"`
	FeesPaid            map[string]Amount   `json:"fees_paid"`
	Positions           map[PositionID]bool `json:"positions"`
	ActiveRoute         RouteID             `json:"active_route"`
	FundingSnapshot     int64               `json:"funding_snapshot"`
	SnapshotEpoch       EpochID             `json:"snapshot_epoch"`
	SnapshotRoute       RouteID             `json:"snapshot_route"`
	LastSnapshotByRoute map[RouteID]int64   `json:"last_snapshot_by_route"`
	Status              AccountStatus       `json:"status"`
	Warnings            []string            `json:"warnings"`
}

func NewAccount(label string) *Account {
	return &Account{
		ID:                  NewAccountID(label),
		Owner:               label,
		Collateral:          make(map[string]Amount),
		ReservedMargin:      make(map[string]Amount),
		RealizedFunding:     make(map[string]int64),
		FeesPaid:            make(map[string]Amount),
		Positions:           make(map[PositionID]bool),
		LastSnapshotByRoute: make(map[RouteID]int64),
		Status:              AccountActive,
		Warnings:            make([]string, 0),
	}
}

func (account *Account) Deposit(asset string, amount Amount) {
	account.Collateral[asset] += amount
}

func (account *Account) Withdraw(asset string, amount Amount) error {
	if account.Collateral[asset] < amount {
		return fail("account.withdraw", "insufficient collateral in %s", account.ID)
	}
	account.Collateral[asset] -= amount
	return nil
}

func (account *Account) ReserveMargin(asset string, amount Amount) error {
	if account.Collateral[asset] < amount {
		return fail("account.reserve", "insufficient collateral in %s", account.ID)
	}
	account.Collateral[asset] -= amount
	account.ReservedMargin[asset] += amount
	return nil
}

func (account *Account) ReleaseMargin(asset string, amount Amount) {
	released := amount.Min(account.ReservedMargin[asset])
	account.ReservedMargin[asset] -= released
	account.Collateral[asset] += released
}

func (account *Account) ConsumeMargin(asset string, amount Amount) Amount {
	available := account.ReservedMargin[asset].Min(amount)
	account.ReservedMargin[asset] -= available
	return amount - available
}

func (account *Account) Credit(asset string, amount Amount) {
	account.Collateral[asset] += amount
}

func (account *Account) Debit(asset string, amount Amount) Amount {
	fromCollateral := account.Collateral[asset].Min(amount)
	account.Collateral[asset] -= fromCollateral
	remaining := amount - fromCollateral
	if remaining == 0 {
		return 0
	}
	fromMargin := account.ReservedMargin[asset].Min(remaining)
	account.ReservedMargin[asset] -= fromMargin
	return remaining - fromMargin
}

func (account *Account) AddPosition(id PositionID) {
	account.Positions[id] = true
}

func (account *Account) RemovePosition(id PositionID) {
	delete(account.Positions, id)
}

func (account *Account) OpenPositionCount() int {
	return len(account.Positions)
}

func (account *Account) SetActiveRoute(route *Route, epoch EpochID) {
	account.ActiveRoute = route.ID
	account.FundingSnapshot = route.FundingAccumulator
	account.SnapshotRoute = route.ID
	account.SnapshotEpoch = epoch
	account.LastSnapshotByRoute[route.ID] = route.FundingAccumulator
}

func (account *Account) SyncIfActive(route *Route, epoch EpochID) {
	if account.ActiveRoute != route.ID {
		return
	}
	account.FundingSnapshot = route.FundingAccumulator
	account.SnapshotRoute = route.ID
	account.SnapshotEpoch = epoch
	account.LastSnapshotByRoute[route.ID] = route.FundingAccumulator
}

func (account *Account) RealizeFunding(asset string, delta int64) {
	account.RealizedFunding[asset] += delta
}

func (account *Account) PayFee(asset string, amount Amount) {
	account.FeesPaid[asset] += amount
}

func (account *Account) MarginBalance(asset string) Amount {
	return account.Collateral[asset] + account.ReservedMargin[asset]
}

func (account *Account) Equity(asset string) int64 {
	return int64(account.Collateral[asset] + account.ReservedMargin[asset])
}

func (account *Account) AddWarning(message string) {
	account.Warnings = append(account.Warnings, message)
}

func (account *Account) Report() AccountReport {
	positions := make([]string, 0, len(account.Positions))
	for id := range account.Positions {
		positions = append(positions, id.String())
	}
	sortStrings(positions)
	realized := make([]SignedBucket, 0, len(account.RealizedFunding))
	for _, asset := range sortedKeys(account.RealizedFunding) {
		realized = append(realized, signedBucket(asset, account.RealizedFunding[asset]))
	}
	return AccountReport{
		ID:              account.ID.String(),
		Owner:           account.Owner,
		Collateral:      sortedAmountBuckets(account.Collateral),
		ReservedMargin:  sortedAmountBuckets(account.ReservedMargin),
		RealizedFunding: realized,
		FeesPaid:        sortedAmountBuckets(account.FeesPaid),
		Positions:       positions,
		ActiveRoute:     account.ActiveRoute.String(),
		FundingSnapshot: account.FundingSnapshot,
		SnapshotEpoch:   account.SnapshotEpoch,
		SnapshotRoute:   account.SnapshotRoute.String(),
		Status:          string(account.Status),
		Warnings:        append([]string(nil), account.Warnings...),
	}
}
