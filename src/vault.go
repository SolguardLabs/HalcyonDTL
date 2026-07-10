package main

type Vault struct {
	ID             VaultID    `json:"id"`
	Asset          string     `json:"asset"`
	Operator       OperatorID `json:"operator"`
	Reserve        Amount     `json:"reserve"`
	LockedMargin   Amount     `json:"locked_margin"`
	Insurance      Amount     `json:"insurance"`
	PendingRebate  Amount     `json:"pending_rebate"`
	SocializedDebt Amount     `json:"socialized_debt"`
	MinReserve     Amount     `json:"min_reserve"`
	Status         string     `json:"status"`
	LastAccounting EpochID    `json:"last_accounting"`
}

func NewVault(label string, asset string, operator OperatorID, reserve Amount, minReserve Amount) *Vault {
	return &Vault{
		ID:             NewVaultID(label),
		Asset:          asset,
		Operator:       operator,
		Reserve:        reserve,
		MinReserve:     minReserve,
		Status:         "open",
		LastAccounting: 0,
	}
}

func (vault *Vault) Available() Amount {
	if vault.Reserve <= vault.MinReserve+vault.LockedMargin {
		return 0
	}
	return vault.Reserve - vault.MinReserve - vault.LockedMargin
}

func (vault *Vault) LockMargin(amount Amount) error {
	if vault.Available() < amount {
		return fail("vault.lock", "insufficient available reserve in %s", vault.ID)
	}
	vault.LockedMargin += amount
	return nil
}

func (vault *Vault) ReleaseMargin(amount Amount) {
	if amount > vault.LockedMargin {
		vault.LockedMargin = 0
		return
	}
	vault.LockedMargin -= amount
}

func (vault *Vault) CreditReserve(amount Amount) {
	vault.Reserve += amount
}

func (vault *Vault) DebitReserve(amount Amount) error {
	if vault.Reserve < amount {
		return fail("vault.debit", "insufficient reserve in %s", vault.ID)
	}
	vault.Reserve -= amount
	return nil
}

func (vault *Vault) CreditInsurance(amount Amount) {
	vault.Insurance += amount
	vault.Reserve += amount
}

func (vault *Vault) AbsorbDebt(amount Amount) Amount {
	if amount == 0 {
		return 0
	}
	covered := vault.Insurance.Min(amount)
	vault.Insurance -= covered
	if vault.Reserve >= covered {
		vault.Reserve -= covered
	}
	uncovered := amount - covered
	vault.SocializedDebt += uncovered
	return uncovered
}

func (vault *Vault) RecordRebate(amount Amount) {
	vault.PendingRebate += amount
}

func (vault *Vault) HealthBps() int64 {
	obligation := vault.LockedMargin + vault.SocializedDebt + vault.PendingRebate
	if obligation == 0 {
		return BpsScale
	}
	return int64(vault.Reserve) * BpsScale / int64(obligation)
}

func (vault *Vault) Report() VaultReport {
	return VaultReport{
		ID:             vault.ID.String(),
		Asset:          vault.Asset,
		Operator:       vault.Operator.String(),
		Reserve:        vault.Reserve,
		LockedMargin:   vault.LockedMargin,
		Insurance:      vault.Insurance,
		PendingRebate:  vault.PendingRebate,
		SocializedDebt: vault.SocializedDebt,
		MinReserve:     vault.MinReserve,
		Available:      vault.Available(),
		HealthBps:      vault.HealthBps(),
		Status:         vault.Status,
		LastAccounting: vault.LastAccounting,
	}
}
