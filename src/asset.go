package main

import "fmt"

type Asset struct {
	Symbol                string `json:"symbol"`
	Decimals              int    `json:"decimals"`
	FundingSlopePPM       int64  `json:"funding_slope_ppm"`
	FundingClampPPM       int64  `json:"funding_clamp_ppm"`
	InitialMarginBps      int64  `json:"initial_margin_bps"`
	MaintenanceMarginBps  int64  `json:"maintenance_margin_bps"`
	LiquidationPenaltyBps int64  `json:"liquidation_penalty_bps"`
	InsuranceShareBps     int64  `json:"insurance_share_bps"`
}

func NewAsset(symbol string, decimals int, slopePPM int64, clampPPM int64) Asset {
	return Asset{
		Symbol:                symbol,
		Decimals:              decimals,
		FundingSlopePPM:       slopePPM,
		FundingClampPPM:       clampPPM,
		InitialMarginBps:      1_000,
		MaintenanceMarginBps:  650,
		LiquidationPenaltyBps: 450,
		InsuranceShareBps:     2_500,
	}
}

func (asset Asset) Validate() error {
	if asset.Symbol == "" {
		return fmt.Errorf("missing asset symbol")
	}
	if asset.Decimals < 0 || asset.Decimals > 18 {
		return fmt.Errorf("invalid decimals for %s", asset.Symbol)
	}
	if asset.FundingSlopePPM <= 0 {
		return fmt.Errorf("invalid funding slope for %s", asset.Symbol)
	}
	if asset.FundingClampPPM <= 0 {
		return fmt.Errorf("invalid funding clamp for %s", asset.Symbol)
	}
	if asset.InitialMarginBps <= 0 || asset.InitialMarginBps > BpsScale {
		return fmt.Errorf("invalid initial margin for %s", asset.Symbol)
	}
	if asset.MaintenanceMarginBps <= 0 || asset.MaintenanceMarginBps >= asset.InitialMarginBps {
		return fmt.Errorf("invalid maintenance margin for %s", asset.Symbol)
	}
	return nil
}

func (asset Asset) InitialMargin(notional Amount) (Amount, error) {
	return notional.MulBps(asset.InitialMarginBps)
}

func (asset Asset) MaintenanceMargin(notional Amount) (Amount, error) {
	return notional.MulBps(asset.MaintenanceMarginBps)
}

func (asset Asset) LiquidationPenalty(notional Amount) (Amount, error) {
	return notional.MulBps(asset.LiquidationPenaltyBps)
}

type AssetBook struct {
	assets map[string]Asset
	order  []string
}

func NewAssetBook() AssetBook {
	return AssetBook{assets: make(map[string]Asset), order: make([]string, 0)}
}

func (book *AssetBook) Register(asset Asset) error {
	if err := asset.Validate(); err != nil {
		return err
	}
	if _, exists := book.assets[asset.Symbol]; !exists {
		book.order = append(book.order, asset.Symbol)
	}
	book.assets[asset.Symbol] = asset
	return nil
}

func (book *AssetBook) MustRegister(asset Asset) {
	if err := book.Register(asset); err != nil {
		panic(err)
	}
}

func (book AssetBook) Get(symbol string) (Asset, error) {
	asset, ok := book.assets[symbol]
	if !ok {
		return Asset{}, fmt.Errorf("unknown asset %s", symbol)
	}
	return asset, nil
}

func (book AssetBook) Has(symbol string) bool {
	_, ok := book.assets[symbol]
	return ok
}

func (book AssetBook) List() []Asset {
	out := make([]Asset, 0, len(book.order))
	for _, symbol := range book.order {
		out = append(out, book.assets[symbol])
	}
	return out
}

func (book AssetBook) Symbols() []string {
	out := make([]string, len(book.order))
	copy(out, book.order)
	return out
}
