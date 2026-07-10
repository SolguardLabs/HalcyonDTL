package main

type Pool struct {
	FeesCollected      map[string]Amount `json:"fees_collected"`
	InsuranceBalance   map[string]Amount `json:"insurance_balance"`
	SocializedDebt     map[string]Amount `json:"socialized_debt"`
	UncollectedFunding map[string]Amount `json:"uncollected_funding"`
}

func NewPool() *Pool {
	return &Pool{
		FeesCollected:      make(map[string]Amount),
		InsuranceBalance:   make(map[string]Amount),
		SocializedDebt:     make(map[string]Amount),
		UncollectedFunding: make(map[string]Amount),
	}
}

func (pool *Pool) CreditFee(asset string, amount Amount) {
	pool.FeesCollected[asset] += amount
}

func (pool *Pool) CreditInsurance(asset string, amount Amount) {
	pool.InsuranceBalance[asset] += amount
}

func (pool *Pool) AbsorbDebt(asset string, amount Amount) {
	if amount == 0 {
		return
	}
	covered := pool.InsuranceBalance[asset].Min(amount)
	pool.InsuranceBalance[asset] -= covered
	uncovered := amount - covered
	pool.SocializedDebt[asset] += uncovered
}

func (pool *Pool) RecordUncollectedFunding(asset string, amount Amount) {
	pool.UncollectedFunding[asset] += amount
}

func (pool *Pool) Report() PoolReport {
	return PoolReport{
		FeesCollected:      sortedAmountBuckets(pool.FeesCollected),
		InsuranceBalance:   sortedAmountBuckets(pool.InsuranceBalance),
		SocializedDebt:     sortedAmountBuckets(pool.SocializedDebt),
		UncollectedFunding: sortedAmountBuckets(pool.UncollectedFunding),
	}
}
