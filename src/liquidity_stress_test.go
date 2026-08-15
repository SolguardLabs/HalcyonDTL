package main

import "testing"

func standardStress() LiquidityStressInput {
	return LiquidityStressInput{
		RouteID: NewRouteID("atlas-eu"), Liquidity: 1_000_000, Utilized: 600_000,
		Reserved: 50_000, ShockNotional: 100_000, FundingShockPPM: 20_000,
		OracleHaircutBps: 300, LiquidationCostBps: 120, InsuranceBalance: 80_000,
		InsuranceRecoveryBps: 8_000, MaxUtilizationBps: 9_600,
	}
}

func TestLiquidityStressProjection(t *testing.T) {
	got, err := ProjectLiquidityStress(standardStress())
	if err != nil {
		t.Fatal(err)
	}
	if got.GrossExposure != 750_000 || got.ProjectedUtilizationBps != 7_500 {
		t.Fatalf("unexpected exposure: %+v", got)
	}
	if got.FundingLoss != 14_000 || got.OracleLoss != 22_500 || got.LiquidationCost != 8_400 {
		t.Fatalf("unexpected loss waterfall: %+v", got)
	}
	if got.GrossLoss != 44_900 || got.InsuranceUsed != 44_900 || got.ResidualDeficit != 0 {
		t.Fatalf("unexpected coverage: %+v", got)
	}
}

func TestLiquidityStressCapsInsuranceUsage(t *testing.T) {
	input := standardStress()
	input.InsuranceBalance = 10_000
	input.InsuranceRecoveryBps = 5_000
	got, err := ProjectLiquidityStress(input)
	if err != nil {
		t.Fatal(err)
	}
	if got.InsuranceUsed != 5_000 || got.ResidualDeficit != 39_900 || got.Severity != "critical" {
		t.Fatalf("unexpected deficit: %+v", got)
	}
}

func TestLiquidityStressFlagsCapacity(t *testing.T) {
	input := standardStress()
	input.ShockNotional = 400_000
	got, err := ProjectLiquidityStress(input)
	if err != nil {
		t.Fatal(err)
	}
	if !got.CapacityBreached || got.ProjectedUtilizationBps != 10_500 {
		t.Fatalf("capacity not detected: %+v", got)
	}
}

func TestLiquidityStressRejectsInvalidInputs(t *testing.T) {
	cases := []LiquidityStressInput{standardStress(), standardStress(), standardStress(), standardStress()}
	cases[0].RouteID = ""
	cases[1].Liquidity = 0
	cases[2].FundingShockPPM = PPMScale + 1
	cases[3].OracleHaircutBps = BpsScale + 1
	for index, input := range cases {
		if _, err := ProjectLiquidityStress(input); err == nil {
			t.Fatalf("case %d accepted", index)
		}
	}
}

func TestPortfolioStressAggregatesRoutes(t *testing.T) {
	first := standardStress()
	second := standardStress()
	second.RouteID = NewRouteID("boreal-us")
	second.InsuranceBalance = 0
	got, err := ProjectPortfolioStress([]LiquidityStressInput{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Routes) != 2 || got.GrossLoss != 89_800 || got.ResidualDeficit != 44_900 || got.CriticalRoutes != 1 {
		t.Fatalf("unexpected portfolio: %+v", got)
	}
}

func TestPortfolioStressRejectsDuplicateRoutes(t *testing.T) {
	input := standardStress()
	if _, err := ProjectPortfolioStress([]LiquidityStressInput{input, input}); err == nil {
		t.Fatal("duplicate route accepted")
	}
}

func TestPortfolioStressRejectsEmptySet(t *testing.T) {
	if _, err := ProjectPortfolioStress(nil); err == nil {
		t.Fatal("empty portfolio accepted")
	}
}

func TestStressSeverityGuarded(t *testing.T) {
	input := standardStress()
	input.Utilized = 820_000
	input.Reserved = 50_000
	input.ShockNotional = 20_000
	got, err := ProjectLiquidityStress(input)
	if err != nil {
		t.Fatal(err)
	}
	if got.CapitalBufferBps != 1_100 || got.Severity != "guarded" {
		t.Fatalf("unexpected severity: %+v", got)
	}
}

func TestStressSeverityNormal(t *testing.T) {
	got, err := ProjectLiquidityStress(standardStress())
	if err != nil {
		t.Fatal(err)
	}
	if got.Severity != "normal" || got.LiquidLiquidity != 250_000 {
		t.Fatalf("unexpected normal state: %+v", got)
	}
}
