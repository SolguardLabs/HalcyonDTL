package main

import "fmt"

type ScenarioFunc func() (*Engine, error)

func ScenarioNames() []string {
	names := []string{"funding", "routes", "open-close", "liquidation", "rotation"}
	sortStrings(names)
	return names
}

func BuildScenario(name string) (*Engine, error) {
	switch name {
	case "funding":
		return ScenarioFunding()
	case "routes":
		return ScenarioRoutes()
	case "open-close":
		return ScenarioOpenClose()
	case "liquidation":
		return ScenarioLiquidation()
	case "rotation":
		return ScenarioRotation()
	default:
		return nil, fmt.Errorf("unknown scenario %s", name)
	}
}

func BaseEngine() (*Engine, error) {
	engine := NewEngine("halcyon-local-funding")
	asset := NewAsset(DefaultAsset, 6, 180_000, 45_000)
	asset.InitialMarginBps = 1_200
	asset.MaintenanceMarginBps = 650
	asset.LiquidationPenaltyBps = 450
	if err := engine.RegisterAsset(asset); err != nil {
		return nil, err
	}
	atlas := NewOperator("atlas", 8, 2_000)
	boreal := NewOperator("boreal", 10, 2_000)
	cirrus := NewOperator("cirrus", 12, 2_500)
	for _, operator := range []*Operator{atlas, boreal, cirrus} {
		if err := engine.AddOperator(operator); err != nil {
			return nil, err
		}
	}
	vaults := []*Vault{
		NewVault("atlas-eu-source", asset.Symbol, atlas.ID, MustAmount(1_600_000), MustAmount(120_000)),
		NewVault("atlas-eu-destination", asset.Symbol, atlas.ID, MustAmount(1_400_000), MustAmount(120_000)),
		NewVault("boreal-us-source", asset.Symbol, boreal.ID, MustAmount(1_250_000), MustAmount(90_000)),
		NewVault("boreal-us-destination", asset.Symbol, boreal.ID, MustAmount(1_100_000), MustAmount(90_000)),
		NewVault("cirrus-apac-source", asset.Symbol, cirrus.ID, MustAmount(1_050_000), MustAmount(80_000)),
		NewVault("cirrus-apac-destination", asset.Symbol, cirrus.ID, MustAmount(990_000), MustAmount(80_000)),
	}
	for _, vault := range vaults {
		if err := engine.AddVault(vault); err != nil {
			return nil, err
		}
	}
	routes := []*Route{
		NewRoute(DefaultRouteSpec("atlas-eu", vaults[0].ID, vaults[1].ID, atlas.ID, asset, MustAmount(1_000_000))),
		NewRoute(DefaultRouteSpec("boreal-us", vaults[2].ID, vaults[3].ID, boreal.ID, asset, MustAmount(900_000))),
		NewRoute(DefaultRouteSpec("cirrus-apac", vaults[4].ID, vaults[5].ID, cirrus.ID, asset, MustAmount(760_000))),
	}
	routes[1].TargetUtilizationBps = 6_200
	routes[1].FundingSlopePPM = 150_000
	routes[2].TargetUtilizationBps = 5_800
	routes[2].FundingSlopePPM = 130_000
	for _, route := range routes {
		if err := engine.AddRoute(route); err != nil {
			return nil, err
		}
	}
	accounts := []*Account{
		NewAccount("alice"),
		NewAccount("bob"),
		NewAccount("carol"),
		NewAccount("liquidator"),
	}
	deposits := []Amount{MustAmount(260_000), MustAmount(96_000), MustAmount(180_000), MustAmount(40_000)}
	for i, account := range accounts {
		account.Deposit(asset.Symbol, deposits[i])
		if err := engine.AddAccount(account); err != nil {
			return nil, err
		}
	}
	engine.Note("base engine initialized with three dynamic funding routes")
	return engine, nil
}

func ScenarioFunding() (*Engine, error) {
	engine, err := BaseEngine()
	if err != nil {
		return nil, err
	}
	alice := NewAccountID("alice")
	route := NewRouteID("atlas-eu")
	if _, err := engine.OpenPosition(alice, route, MustAmount(820_000), MustAmount(100_000)); err != nil {
		return nil, err
	}
	engine.AdvanceEpoch("atlas utilization above target")
	engine.AdvanceEpoch("second funding epoch")
	engine.AdvanceEpoch("third funding epoch")
	engine.Note("funding scenario leaves the position open so reports expose unrealized funding")
	return engine, nil
}

func ScenarioRoutes() (*Engine, error) {
	engine, err := BaseEngine()
	if err != nil {
		return nil, err
	}
	alice := NewAccountID("alice")
	carol := NewAccountID("carol")
	atlas := NewRouteID("atlas-eu")
	boreal := NewRouteID("boreal-us")
	cirrus := NewRouteID("cirrus-apac")
	if _, err := engine.OpenPosition(alice, atlas, MustAmount(520_000), MustAmount(75_000)); err != nil {
		return nil, err
	}
	if _, err := engine.OpenPosition(carol, boreal, MustAmount(410_000), MustAmount(60_000)); err != nil {
		return nil, err
	}
	if _, err := engine.OpenPosition(carol, cirrus, MustAmount(220_000), MustAmount(34_000)); err != nil {
		return nil, err
	}
	engine.AdvanceEpoch("mixed utilization before route review")
	if _, err := engine.ApplyRotation(alice, boreal, "operator preference rebalance"); err != nil {
		return nil, err
	}
	engine.Note("routes scenario validates route opening and benign account route rotation")
	return engine, nil
}

func ScenarioOpenClose() (*Engine, error) {
	engine, err := BaseEngine()
	if err != nil {
		return nil, err
	}
	alice := NewAccountID("alice")
	route := NewRouteID("atlas-eu")
	position, err := engine.OpenPosition(alice, route, MustAmount(760_000), MustAmount(120_000))
	if err != nil {
		return nil, err
	}
	engine.AdvanceEpoch("close path funding one")
	engine.AdvanceEpoch("close path funding two")
	if _, err := engine.ClosePosition(alice, position.ID, "normal_close"); err != nil {
		return nil, err
	}
	engine.Note("open-close scenario closes without route rotation")
	return engine, nil
}

func ScenarioLiquidation() (*Engine, error) {
	engine, err := BaseEngine()
	if err != nil {
		return nil, err
	}
	bob := NewAccountID("bob")
	route := NewRouteID("atlas-eu")
	position, err := engine.OpenPosition(bob, route, MustAmount(900_000), MustAmount(70_000))
	if err != nil {
		return nil, err
	}
	engine.AdvanceEpoch("liquidation funding shock")
	if _, err := engine.LiquidatePosition(position.ID, NewAccountID("liquidator")); err != nil {
		return nil, err
	}
	engine.Note("liquidation scenario uses active-route funding and should not require rotation")
	return engine, nil
}

func ScenarioRotation() (*Engine, error) {
	engine, err := BaseEngine()
	if err != nil {
		return nil, err
	}
	alice := NewAccountID("alice")
	atlas := NewRouteID("atlas-eu")
	boreal := NewRouteID("boreal-us")
	if _, err := engine.OpenPosition(alice, atlas, MustAmount(780_000), MustAmount(112_000)); err != nil {
		return nil, err
	}
	if _, err := engine.ApplyRotation(alice, boreal, "pre-epoch operator rotation"); err != nil {
		return nil, err
	}
	engine.AdvanceEpoch("route rotated before atlas funding accrues")
	engine.Note("rotation scenario exposes route/account accumulator drift without closing the old exposure")
	return engine, nil
}
