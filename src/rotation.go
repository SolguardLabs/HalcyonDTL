package main

type RouteRotation struct {
	AccountID        AccountID `json:"account_id"`
	PreviousRoute    RouteID   `json:"previous_route"`
	NextRoute        RouteID   `json:"next_route"`
	Epoch            EpochID   `json:"epoch"`
	PreviousSnapshot int64     `json:"previous_snapshot"`
	NextSnapshot     int64     `json:"next_snapshot"`
	Reason           string    `json:"reason"`
}

func (engine *Engine) PreviewRotation(accountID AccountID, routeID RouteID, reason string) (RouteRotation, error) {
	account, err := engine.Account(accountID)
	if err != nil {
		return RouteRotation{}, err
	}
	route, err := engine.Route(routeID)
	if err != nil {
		return RouteRotation{}, err
	}
	return RouteRotation{
		AccountID:        account.ID,
		PreviousRoute:    account.ActiveRoute,
		NextRoute:        route.ID,
		Epoch:            engine.Clock,
		PreviousSnapshot: account.FundingSnapshot,
		NextSnapshot:     route.FundingAccumulator,
		Reason:           reason,
	}, nil
}

func (engine *Engine) ApplyRotation(accountID AccountID, routeID RouteID, reason string) (RouteRotation, error) {
	preview, err := engine.PreviewRotation(accountID, routeID, reason)
	if err != nil {
		return RouteRotation{}, err
	}
	if err := engine.RotateAccountRoute(accountID, routeID, reason); err != nil {
		return RouteRotation{}, err
	}
	return preview, nil
}
