package main

type Engine struct {
	NetworkID string
	Clock     EpochID
	Assets    AssetBook
	Operators map[OperatorID]*Operator
	Vaults    map[VaultID]*Vault
	Routes    map[RouteID]*Route
	Accounts  map[AccountID]*Account
	Positions map[PositionID]*Position
	Funding   *FundingEngine
	Pool      *Pool
	Journal   *Journal
	Notes     []string
	nonce     int64
}

func NewEngine(networkID string) *Engine {
	return &Engine{
		NetworkID: networkID,
		Assets:    NewAssetBook(),
		Operators: make(map[OperatorID]*Operator),
		Vaults:    make(map[VaultID]*Vault),
		Routes:    make(map[RouteID]*Route),
		Accounts:  make(map[AccountID]*Account),
		Positions: make(map[PositionID]*Position),
		Funding:   NewFundingEngine(),
		Pool:      NewPool(),
		Journal:   NewJournal(),
		Notes:     make([]string, 0),
	}
}

func (engine *Engine) RegisterAsset(asset Asset) error {
	if err := engine.Assets.Register(asset); err != nil {
		return err
	}
	engine.Journal.Append(engine.Clock, "asset.registered", eventFields("asset", asset.Symbol))
	return nil
}

func (engine *Engine) AddOperator(operator *Operator) error {
	if operator == nil {
		return fail("engine.operator", "nil operator")
	}
	if err := operator.Validate(); err != nil {
		return err
	}
	if _, exists := engine.Operators[operator.ID]; exists {
		return fail("engine.operator", "duplicate operator %s", operator.ID)
	}
	engine.Operators[operator.ID] = operator
	engine.Journal.Append(engine.Clock, "operator.registered", eventFields("operator", operator.ID.String()))
	return nil
}

func (engine *Engine) AddVault(vault *Vault) error {
	if vault == nil {
		return fail("engine.vault", "nil vault")
	}
	if vault.ID.Empty() {
		return fail("engine.vault", "empty vault id")
	}
	if !engine.Assets.Has(vault.Asset) {
		return fail("engine.vault", "unknown asset %s", vault.Asset)
	}
	if _, ok := engine.Operators[vault.Operator]; !ok {
		return fail("engine.vault", "unknown operator %s", vault.Operator)
	}
	if _, exists := engine.Vaults[vault.ID]; exists {
		return fail("engine.vault", "duplicate vault %s", vault.ID)
	}
	engine.Vaults[vault.ID] = vault
	engine.Journal.Append(engine.Clock, "vault.opened", eventFields("vault", vault.ID.String(), "asset", vault.Asset))
	return nil
}

func (engine *Engine) AddAccount(account *Account) error {
	if account == nil {
		return fail("engine.account", "nil account")
	}
	if account.ID.Empty() {
		return fail("engine.account", "empty account id")
	}
	if _, exists := engine.Accounts[account.ID]; exists {
		return fail("engine.account", "duplicate account %s", account.ID)
	}
	engine.Accounts[account.ID] = account
	engine.Journal.Append(engine.Clock, "account.opened", eventFields("account", account.ID.String()))
	return nil
}

func (engine *Engine) AddRoute(route *Route) error {
	if route == nil {
		return fail("engine.route", "nil route")
	}
	if err := route.Validate(); err != nil {
		return err
	}
	if _, exists := engine.Routes[route.ID]; exists {
		return fail("engine.route", "duplicate route %s", route.ID)
	}
	if !engine.Assets.Has(route.Asset) {
		return fail("engine.route", "unknown route asset %s", route.Asset)
	}
	if _, ok := engine.Vaults[route.SourceVault]; !ok {
		return fail("engine.route", "unknown source vault %s", route.SourceVault)
	}
	if _, ok := engine.Vaults[route.DestinationVault]; !ok {
		return fail("engine.route", "unknown destination vault %s", route.DestinationVault)
	}
	operator, ok := engine.Operators[route.Operator]
	if !ok {
		return fail("engine.route", "unknown operator %s", route.Operator)
	}
	engine.Routes[route.ID] = route
	operator.RouteCount++
	engine.Journal.Append(engine.Clock, "route.opened", eventFields(
		"route", route.ID.String(),
		"asset", route.Asset,
		"liquidity", route.Liquidity,
		"target_bps", route.TargetUtilizationBps,
	))
	return nil
}

func (engine *Engine) Asset(symbol string) (Asset, error) {
	return engine.Assets.Get(symbol)
}

func (engine *Engine) Account(id AccountID) (*Account, error) {
	account, ok := engine.Accounts[id]
	if !ok {
		return nil, fail("engine.account", "unknown account %s", id)
	}
	return account, nil
}

func (engine *Engine) Route(id RouteID) (*Route, error) {
	route, ok := engine.Routes[id]
	if !ok {
		return nil, fail("engine.route", "unknown route %s", id)
	}
	return route, nil
}

func (engine *Engine) Vault(id VaultID) (*Vault, error) {
	vault, ok := engine.Vaults[id]
	if !ok {
		return nil, fail("engine.vault", "unknown vault %s", id)
	}
	return vault, nil
}

func (engine *Engine) Operator(id OperatorID) (*Operator, error) {
	operator, ok := engine.Operators[id]
	if !ok {
		return nil, fail("engine.operator", "unknown operator %s", id)
	}
	return operator, nil
}

func (engine *Engine) Position(id PositionID) (*Position, error) {
	position, ok := engine.Positions[id]
	if !ok {
		return nil, fail("engine.position", "unknown position %s", id)
	}
	return position, nil
}

func (engine *Engine) nextPositionID(account AccountID, route RouteID) PositionID {
	engine.nonce++
	return NewPositionID(account, route, engine.nonce)
}

func (engine *Engine) OpenPosition(accountID AccountID, routeID RouteID, notional Amount, marginOverride Amount) (*Position, error) {
	account, err := engine.Account(accountID)
	if err != nil {
		return nil, err
	}
	route, err := engine.Route(routeID)
	if err != nil {
		return nil, err
	}
	if route.Status != RouteOpen {
		return nil, fail("engine.open", "route %s is not open", route.ID)
	}
	asset, err := engine.Asset(route.Asset)
	if err != nil {
		return nil, err
	}
	margin := marginOverride
	if margin == 0 {
		margin, err = asset.InitialMargin(notional)
		if err != nil {
			return nil, err
		}
	}
	if margin == 0 {
		return nil, fail("engine.open", "margin cannot be zero")
	}
	if err := route.Reserve(notional); err != nil {
		return nil, err
	}
	if err := account.ReserveMargin(route.Asset, margin); err != nil {
		route.Release(notional)
		return nil, err
	}
	source, err := engine.Vault(route.SourceVault)
	if err != nil {
		return nil, err
	}
	if err := source.LockMargin(margin); err != nil {
		account.ReleaseMargin(route.Asset, margin)
		route.Release(notional)
		return nil, err
	}
	route.ActivateReservation(notional)
	account.SetActiveRoute(route, engine.Clock)
	id := engine.nextPositionID(account.ID, route.ID)
	position := NewPosition(id, account.ID, route, notional, margin, engine.Clock)
	engine.Positions[id] = position
	account.AddPosition(id)
	engine.Journal.Append(engine.Clock, "position.opened", eventFields(
		"account", account.ID.String(),
		"route", route.ID.String(),
		"position", id.String(),
		"notional", notional,
		"margin", margin,
		"entry_accumulator", position.EntryAccumulator,
	))
	return position, nil
}

func (engine *Engine) AdvanceEpoch(label string) []FundingRecord {
	engine.Clock++
	records := make([]FundingRecord, 0, len(engine.Routes))
	for _, routeID := range sortedRouteIDs(engine.Routes) {
		route := engine.Routes[routeID]
		if route.Status == RouteClosed {
			continue
		}
		record := route.ApplyFunding(engine.Clock)
		engine.Funding.Append(record)
		records = append(records, record)
		engine.Journal.Append(engine.Clock, "funding.applied", eventFields(
			"route", route.ID.String(),
			"rate_ppm", record.RatePPM,
			"accumulator", record.AccumulatorAfter,
			"utilization_bps", record.UtilizationBps,
			"label", label,
		))
	}
	for _, accountID := range sortedAccountIDs(engine.Accounts) {
		account := engine.Accounts[accountID]
		route, ok := engine.Routes[account.ActiveRoute]
		if !ok {
			continue
		}
		account.SyncIfActive(route, engine.Clock)
	}
	return records
}

func (engine *Engine) RotateAccountRoute(accountID AccountID, routeID RouteID, reason string) error {
	account, err := engine.Account(accountID)
	if err != nil {
		return err
	}
	route, err := engine.Route(routeID)
	if err != nil {
		return err
	}
	if route.Status != RouteOpen {
		return fail("engine.rotate", "cannot rotate account %s to route %s with status %s", account.ID, route.ID, route.Status)
	}
	previous := account.ActiveRoute
	account.SetActiveRoute(route, engine.Clock)
	engine.Journal.Append(engine.Clock, "account.route_rotated", eventFields(
		"account", account.ID.String(),
		"previous_route", previous.String(),
		"active_route", route.ID.String(),
		"snapshot", account.FundingSnapshot,
		"reason", reason,
	))
	return nil
}

func (engine *Engine) ClosePosition(accountID AccountID, positionID PositionID, reason string) (FundingSettlement, error) {
	account, err := engine.Account(accountID)
	if err != nil {
		return FundingSettlement{}, err
	}
	position, err := engine.Position(positionID)
	if err != nil {
		return FundingSettlement{}, err
	}
	if position.AccountID != account.ID {
		return FundingSettlement{}, fail("engine.close", "position %s does not belong to %s", position.ID, account.ID)
	}
	if !position.IsOpen() {
		return FundingSettlement{}, fail("engine.close", "position %s is not open", position.ID)
	}
	route, err := engine.Route(position.RouteID)
	if err != nil {
		return FundingSettlement{}, err
	}
	operator, err := engine.Operator(route.Operator)
	if err != nil {
		return FundingSettlement{}, err
	}
	asset, err := engine.Asset(position.Asset)
	if err != nil {
		return FundingSettlement{}, err
	}
	position.MarkClosing()
	settlement := SettlementUsesAccountSnapshot(account, position, route)
	expected := SettlementUsesRouteAccumulator(account, position, route)
	closeFee, err := operator.FeeFor(position.Notional)
	if err != nil {
		return FundingSettlement{}, err
	}
	socialized, err := engine.applyPositionSettlement(account, position, route, operator, asset, settlement, closeFee, 0)
	if err != nil {
		return FundingSettlement{}, err
	}
	settlement.SocializedDebt = socialized
	if expected.Funding < settlement.Funding {
		engine.Pool.RecordUncollectedFunding(position.Asset, Amount(settlement.Funding-expected.Funding))
	}
	position.MarkClosed(engine.Clock, settlement, closeFee, nonEmpty(reason, "user_close"))
	account.RemovePosition(position.ID)
	route.Release(position.Notional)
	account.FundingSnapshot = route.FundingAccumulator
	account.SnapshotRoute = route.ID
	account.SnapshotEpoch = engine.Clock
	account.LastSnapshotByRoute[route.ID] = route.FundingAccumulator
	engine.Journal.Append(engine.Clock, "position.closed", eventFields(
		"account", account.ID.String(),
		"route", route.ID.String(),
		"position", position.ID.String(),
		"funding", settlement.Funding,
		"expected_funding", expected.Funding,
		"close_fee", closeFee,
		"socialized_debt", socialized,
		"reason", position.CloseReason,
	))
	return settlement, nil
}

func (engine *Engine) applyPositionSettlement(
	account *Account,
	position *Position,
	route *Route,
	operator *Operator,
	asset Asset,
	settlement FundingSettlement,
	closeFee Amount,
	liquidationFee Amount,
) (Amount, error) {
	vault, err := engine.Vault(route.SourceVault)
	if err != nil {
		return 0, err
	}
	reserved := account.ReservedMargin[position.Asset].Min(position.Margin)
	account.ReservedMargin[position.Asset] -= reserved
	vault.ReleaseMargin(reserved)
	balance := int64(reserved) + settlement.Funding - int64(closeFee) - int64(liquidationFee)
	account.RealizeFunding(position.Asset, settlement.Funding)
	account.PayFee(position.Asset, closeFee+liquidationFee)
	if closeFee > 0 {
		engine.Pool.CreditFee(position.Asset, closeFee)
		insuranceShare, err := closeFee.MulBps(asset.InsuranceShareBps)
		if err != nil {
			return 0, err
		}
		engine.Pool.CreditInsurance(position.Asset, insuranceShare)
		vault.CreditInsurance(insuranceShare)
		operatorShare := closeFee - insuranceShare
		operator.Credit(position.Asset, operatorShare)
	}
	if liquidationFee > 0 {
		engine.Pool.CreditFee(position.Asset, liquidationFee)
		operator.Credit(position.Asset, liquidationFee)
	}
	if balance >= 0 {
		account.Credit(position.Asset, Amount(balance))
		return 0, nil
	}
	debt := Amount(-balance)
	uncovered := vault.AbsorbDebt(debt)
	engine.Pool.AbsorbDebt(position.Asset, debt)
	return uncovered, nil
}

func (engine *Engine) Note(message string) {
	engine.Notes = append(engine.Notes, message)
}
