package main

type LedgerEntry struct {
	Index      int64          `json:"index"`
	Epoch      EpochID        `json:"epoch"`
	AccountID  string         `json:"account_id"`
	RouteID    string         `json:"route_id"`
	PositionID string         `json:"position_id"`
	Asset      string         `json:"asset"`
	Delta      int64          `json:"delta"`
	Kind       string         `json:"kind"`
	Reference  string         `json:"reference"`
	Metadata   map[string]any `json:"metadata"`
}

type DerivedLedger struct {
	Entries      []LedgerEntry    `json:"entries"`
	AccountDelta map[string]int64 `json:"account_delta"`
	RouteDelta   map[string]int64 `json:"route_delta"`
	AssetDelta   map[string]int64 `json:"asset_delta"`
}

func (engine *Engine) BuildDerivedLedger() DerivedLedger {
	ledger := DerivedLedger{
		Entries:      make([]LedgerEntry, 0),
		AccountDelta: make(map[string]int64),
		RouteDelta:   make(map[string]int64),
		AssetDelta:   make(map[string]int64),
	}
	for _, event := range engine.Journal.Events() {
		ledger.consumeEvent(event)
	}
	return ledger
}

func (ledger *DerivedLedger) consumeEvent(event Event) {
	switch event.Kind {
	case "position.opened":
		ledger.addFromEvent(event, "margin_locked", -amountField(event, "margin"))
		ledger.addRouteFromEvent(event, "notional_opened", amountField(event, "notional"))
	case "position.closed":
		ledger.addFromEvent(event, "funding_settled", signedField(event, "funding"))
		ledger.addFromEvent(event, "fee_paid", -amountField(event, "close_fee"))
		ledger.addRouteFromEvent(event, "notional_closed", -amountField(event, "notional"))
	case "position.liquidated":
		ledger.addFromEvent(event, "funding_settled", signedField(event, "funding"))
		ledger.addFromEvent(event, "fee_paid", -amountField(event, "close_fee"))
		ledger.addFromEvent(event, "liquidation_fee_paid", -amountField(event, "liquidation_fee"))
	case "funding.applied":
		ledger.addRouteFromEvent(event, "funding_accumulator", signedField(event, "rate_ppm"))
	}
}

func (ledger *DerivedLedger) addFromEvent(event Event, kind string, delta int64) {
	accountID := stringField(event, "account")
	if accountID == "" {
		return
	}
	entry := LedgerEntry{
		Index:      int64(len(ledger.Entries) + 1),
		Epoch:      event.Epoch,
		AccountID:  accountID,
		RouteID:    stringField(event, "route"),
		PositionID: stringField(event, "position"),
		Asset:      stringField(event, "asset"),
		Delta:      delta,
		Kind:       kind,
		Reference:  event.Kind,
		Metadata:   event.Fields,
	}
	ledger.Entries = append(ledger.Entries, entry)
	ledger.AccountDelta[accountID] += delta
	if entry.Asset != "" {
		ledger.AssetDelta[entry.Asset] += delta
	}
}

func (ledger *DerivedLedger) addRouteFromEvent(event Event, kind string, delta int64) {
	routeID := stringField(event, "route")
	if routeID == "" {
		return
	}
	entry := LedgerEntry{
		Index:      int64(len(ledger.Entries) + 1),
		Epoch:      event.Epoch,
		AccountID:  stringField(event, "account"),
		RouteID:    routeID,
		PositionID: stringField(event, "position"),
		Asset:      stringField(event, "asset"),
		Delta:      delta,
		Kind:       kind,
		Reference:  event.Kind,
		Metadata:   event.Fields,
	}
	ledger.Entries = append(ledger.Entries, entry)
	ledger.RouteDelta[routeID] += delta
	if entry.Asset != "" {
		ledger.AssetDelta[entry.Asset] += delta
	}
}

func stringField(event Event, key string) string {
	value, ok := event.Fields[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case AccountID:
		return typed.String()
	case RouteID:
		return typed.String()
	case PositionID:
		return typed.String()
	case VaultID:
		return typed.String()
	case OperatorID:
		return typed.String()
	default:
		return ""
	}
}

func amountField(event Event, key string) int64 {
	value, ok := event.Fields[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case Amount:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}

func signedField(event Event, key string) int64 {
	return amountField(event, key)
}
