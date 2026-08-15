package main

import "fmt"

// LiquidityStressInput describes a deterministic route-level solvency shock.
type LiquidityStressInput struct {
	RouteID              RouteID `json:"route_id"`
	Liquidity            Amount  `json:"liquidity"`
	Utilized             Amount  `json:"utilized"`
	Reserved             Amount  `json:"reserved"`
	ShockNotional        Amount  `json:"shock_notional"`
	FundingShockPPM      int64   `json:"funding_shock_ppm"`
	OracleHaircutBps     int64   `json:"oracle_haircut_bps"`
	LiquidationCostBps   int64   `json:"liquidation_cost_bps"`
	InsuranceBalance     Amount  `json:"insurance_balance"`
	InsuranceRecoveryBps int64   `json:"insurance_recovery_bps"`
	MaxUtilizationBps    int64   `json:"max_utilization_bps"`
}

// LiquidityStressProjection exposes every intermediate accounting term.
type LiquidityStressProjection struct {
	RouteID                 RouteID `json:"route_id"`
	GrossExposure           Amount  `json:"gross_exposure"`
	ProjectedUtilized       Amount  `json:"projected_utilized"`
	ProjectedUtilizationBps int64   `json:"projected_utilization_bps"`
	FundingLoss             Amount  `json:"funding_loss"`
	OracleLoss              Amount  `json:"oracle_loss"`
	LiquidationCost         Amount  `json:"liquidation_cost"`
	GrossLoss               Amount  `json:"gross_loss"`
	RecoverableInsurance    Amount  `json:"recoverable_insurance"`
	InsuranceUsed           Amount  `json:"insurance_used"`
	ResidualDeficit         Amount  `json:"residual_deficit"`
	LiquidLiquidity         Amount  `json:"liquid_liquidity"`
	CapitalBufferBps        int64   `json:"capital_buffer_bps"`
	CapacityBreached        bool    `json:"capacity_breached"`
	Severity                string  `json:"severity"`
}

func (input LiquidityStressInput) Validate() error {
	if input.RouteID.Empty() {
		return fmt.Errorf("route id is required")
	}
	if input.Liquidity == 0 {
		return fmt.Errorf("route liquidity is required")
	}
	if input.Utilized+input.Reserved > input.Liquidity {
		return fmt.Errorf("route commitments exceed liquidity")
	}
	if input.FundingShockPPM < 0 || input.FundingShockPPM > PPMScale {
		return fmt.Errorf("funding shock must be between 0 and %d ppm", PPMScale)
	}
	if err := validateBps("oracle haircut", input.OracleHaircutBps); err != nil {
		return err
	}
	if err := validateBps("liquidation cost", input.LiquidationCostBps); err != nil {
		return err
	}
	if err := validateBps("insurance recovery", input.InsuranceRecoveryBps); err != nil {
		return err
	}
	if input.MaxUtilizationBps <= 0 || input.MaxUtilizationBps > BpsScale {
		return fmt.Errorf("invalid maximum utilization")
	}
	return nil
}

func validateBps(label string, value int64) error {
	if value < 0 || value > BpsScale {
		return fmt.Errorf("%s must be between 0 and %d bps", label, BpsScale)
	}
	return nil
}

// ProjectLiquidityStress applies losses in accounting order: funding, oracle,
// liquidation, and finally recoverable insurance.
func ProjectLiquidityStress(input LiquidityStressInput) (LiquidityStressProjection, error) {
	if err := input.Validate(); err != nil {
		return LiquidityStressProjection{}, err
	}
	grossExposure := input.Utilized + input.Reserved + input.ShockNotional
	projectedUtilized := input.Utilized + input.ShockNotional
	fundingLoss, err := projectedUtilized.MulPPM(input.FundingShockPPM)
	if err != nil {
		return LiquidityStressProjection{}, err
	}
	oracleLoss, err := grossExposure.MulBps(input.OracleHaircutBps)
	if err != nil {
		return LiquidityStressProjection{}, err
	}
	liquidationCost, err := projectedUtilized.MulBps(input.LiquidationCostBps)
	if err != nil {
		return LiquidityStressProjection{}, err
	}
	grossLoss, err := fundingLoss.Add(oracleLoss)
	if err != nil {
		return LiquidityStressProjection{}, err
	}
	grossLoss, err = grossLoss.Add(liquidationCost)
	if err != nil {
		return LiquidityStressProjection{}, err
	}
	recoverableInsurance, err := input.InsuranceBalance.MulBps(input.InsuranceRecoveryBps)
	if err != nil {
		return LiquidityStressProjection{}, err
	}
	insuranceUsed := recoverableInsurance.Min(grossLoss)
	residualDeficit, _ := grossLoss.Sub(insuranceUsed)
	liquidLiquidity := Amount(0)
	if grossExposure < input.Liquidity {
		liquidLiquidity = input.Liquidity - grossExposure
	}
	utilizationBps := int64(grossExposure) * BpsScale / int64(input.Liquidity)
	bufferBps := int64(0)
	if input.Liquidity > 0 {
		bufferBps = int64(liquidLiquidity) * BpsScale / int64(input.Liquidity)
	}
	projection := LiquidityStressProjection{
		RouteID: input.RouteID, GrossExposure: grossExposure,
		ProjectedUtilized: projectedUtilized, ProjectedUtilizationBps: utilizationBps,
		FundingLoss: fundingLoss, OracleLoss: oracleLoss, LiquidationCost: liquidationCost,
		GrossLoss: grossLoss, RecoverableInsurance: recoverableInsurance,
		InsuranceUsed: insuranceUsed, ResidualDeficit: residualDeficit,
		LiquidLiquidity: liquidLiquidity, CapitalBufferBps: bufferBps,
		CapacityBreached: utilizationBps > input.MaxUtilizationBps,
	}
	projection.Severity = stressSeverity(projection)
	return projection, nil
}

func stressSeverity(projection LiquidityStressProjection) string {
	if projection.ResidualDeficit > 0 || projection.ProjectedUtilizationBps > BpsScale {
		return "critical"
	}
	if projection.CapacityBreached || projection.CapitalBufferBps < 500 {
		return "high"
	}
	if projection.CapitalBufferBps < 1_500 {
		return "guarded"
	}
	return "normal"
}

type PortfolioStressProjection struct {
	Routes          []LiquidityStressProjection `json:"routes"`
	GrossLoss       Amount                      `json:"gross_loss"`
	InsuranceUsed   Amount                      `json:"insurance_used"`
	ResidualDeficit Amount                      `json:"residual_deficit"`
	BreachedRoutes  int                         `json:"breached_routes"`
	CriticalRoutes  int                         `json:"critical_routes"`
	WorstBufferBps  int64                       `json:"worst_buffer_bps"`
}

// ProjectPortfolioStress aggregates independently parameterized route shocks.
func ProjectPortfolioStress(inputs []LiquidityStressInput) (PortfolioStressProjection, error) {
	out := PortfolioStressProjection{Routes: make([]LiquidityStressProjection, 0, len(inputs)), WorstBufferBps: BpsScale}
	if len(inputs) == 0 {
		return out, fmt.Errorf("at least one route is required")
	}
	seen := make(map[RouteID]bool, len(inputs))
	for _, input := range inputs {
		if seen[input.RouteID] {
			return PortfolioStressProjection{}, fmt.Errorf("duplicate route %s", input.RouteID)
		}
		seen[input.RouteID] = true
		projection, err := ProjectLiquidityStress(input)
		if err != nil {
			return PortfolioStressProjection{}, err
		}
		out.Routes = append(out.Routes, projection)
		out.GrossLoss += projection.GrossLoss
		out.InsuranceUsed += projection.InsuranceUsed
		out.ResidualDeficit += projection.ResidualDeficit
		if projection.CapacityBreached {
			out.BreachedRoutes++
		}
		if projection.Severity == "critical" {
			out.CriticalRoutes++
		}
		if projection.CapitalBufferBps < out.WorstBufferBps {
			out.WorstBufferBps = projection.CapitalBufferBps
		}
	}
	return out, nil
}
