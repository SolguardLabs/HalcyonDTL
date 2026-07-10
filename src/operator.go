package main

type Operator struct {
	ID              OperatorID        `json:"id"`
	Name            string            `json:"name"`
	FeeBps          int64             `json:"fee_bps"`
	FundingShareBps int64             `json:"funding_share_bps"`
	RiskWeightBps   int64             `json:"risk_weight_bps"`
	Balances        map[string]Amount `json:"balances"`
	RouteCount      int               `json:"route_count"`
	Status          string            `json:"status"`
}

func NewOperator(label string, feeBps int64, fundingShareBps int64) *Operator {
	return &Operator{
		ID:              NewOperatorID(label),
		Name:            label,
		FeeBps:          feeBps,
		FundingShareBps: fundingShareBps,
		RiskWeightBps:   BpsScale,
		Balances:        make(map[string]Amount),
		Status:          "active",
	}
}

func (op *Operator) Validate() error {
	if op.ID.Empty() {
		return fail("operator.validate", "missing operator id")
	}
	if op.FeeBps < 0 || op.FeeBps > 1_000 {
		return fail("operator.validate", "fee outside bounds for %s", op.ID)
	}
	if op.FundingShareBps < 0 || op.FundingShareBps > BpsScale {
		return fail("operator.validate", "funding share outside bounds for %s", op.ID)
	}
	if op.RiskWeightBps <= 0 {
		return fail("operator.validate", "risk weight must be positive for %s", op.ID)
	}
	return nil
}

func (op *Operator) Credit(asset string, amount Amount) {
	op.Balances[asset] += amount
}

func (op *Operator) Debit(asset string, amount Amount) error {
	current := op.Balances[asset]
	if current < amount {
		return fail("operator.debit", "insufficient operator balance for %s", op.ID)
	}
	op.Balances[asset] = current - amount
	return nil
}

func (op *Operator) FeeFor(amount Amount) (Amount, error) {
	return amount.MulBps(op.FeeBps)
}

func (op *Operator) FundingShare(amount Amount) (Amount, error) {
	return amount.MulBps(op.FundingShareBps)
}

func (op *Operator) Report() OperatorReport {
	return OperatorReport{
		ID:              op.ID.String(),
		Name:            op.Name,
		FeeBps:          op.FeeBps,
		FundingShareBps: op.FundingShareBps,
		RiskWeightBps:   op.RiskWeightBps,
		Balances:        sortedAmountBuckets(op.Balances),
		RouteCount:      op.RouteCount,
		Status:          op.Status,
	}
}
