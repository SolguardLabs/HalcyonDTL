package main

type AuditSeverity string

const (
	AuditInfo     AuditSeverity = "info"
	AuditLow      AuditSeverity = "low"
	AuditMedium   AuditSeverity = "medium"
	AuditHigh     AuditSeverity = "high"
	AuditCritical AuditSeverity = "critical"
)

type AuditFinding struct {
	ID         string        `json:"id"`
	Severity   AuditSeverity `json:"severity"`
	Component  string        `json:"component"`
	Subject    string        `json:"subject"`
	Message    string        `json:"message"`
	Epoch      EpochID       `json:"epoch"`
	Metric     int64         `json:"metric"`
	Threshold  int64         `json:"threshold"`
	Actionable bool          `json:"actionable"`
}

type AuditProfile struct {
	MaxUtilizationBps      int64  `json:"max_utilization_bps"`
	MaxSnapshotDriftPPM    int64  `json:"max_snapshot_drift_ppm"`
	MaxSocializedDebt      Amount `json:"max_socialized_debt"`
	MinVaultHealthBps      int64  `json:"min_vault_health_bps"`
	RequireActiveSnapshots bool   `json:"require_active_snapshots"`
	WarnOnClosedFundingGap bool   `json:"warn_on_closed_funding_gap"`
}

func DefaultAuditProfile() AuditProfile {
	return AuditProfile{
		MaxUtilizationBps:      9_600,
		MaxSnapshotDriftPPM:    35_000,
		MaxSocializedDebt:      MustAmount(5_000),
		MinVaultHealthBps:      9_000,
		RequireActiveSnapshots: true,
		WarnOnClosedFundingGap: true,
	}
}

func (engine *Engine) Audit(profile AuditProfile) []AuditFinding {
	findings := make([]AuditFinding, 0)
	findings = append(findings, engine.auditRoutes(profile)...)
	findings = append(findings, engine.auditVaults(profile)...)
	findings = append(findings, engine.auditAccounts(profile)...)
	findings = append(findings, engine.auditPositions(profile)...)
	findings = append(findings, engine.auditPool(profile)...)
	return findings
}

func (engine *Engine) auditRoutes(profile AuditProfile) []AuditFinding {
	out := make([]AuditFinding, 0)
	for _, routeID := range sortedRouteIDs(engine.Routes) {
		route := engine.Routes[routeID]
		util := route.UtilizationBps()
		if util > profile.MaxUtilizationBps {
			out = append(out, AuditFinding{
				ID:         "route-utilization-" + route.ID.String(),
				Severity:   AuditHigh,
				Component:  "route",
				Subject:    route.ID.String(),
				Message:    "route utilization exceeds profile limit",
				Epoch:      engine.Clock,
				Metric:     util,
				Threshold:  profile.MaxUtilizationBps,
				Actionable: true,
			})
		}
		if route.Status == RoutePaused && route.OpenInterest > 0 {
			out = append(out, AuditFinding{
				ID:         "paused-route-open-interest-" + route.ID.String(),
				Severity:   AuditMedium,
				Component:  "route",
				Subject:    route.ID.String(),
				Message:    "paused route still carries open interest",
				Epoch:      engine.Clock,
				Metric:     int64(route.OpenInterest),
				Threshold:  0,
				Actionable: true,
			})
		}
	}
	return out
}

func (engine *Engine) auditVaults(profile AuditProfile) []AuditFinding {
	out := make([]AuditFinding, 0)
	for _, vaultID := range sortedVaultIDs(engine.Vaults) {
		vault := engine.Vaults[vaultID]
		health := vault.HealthBps()
		if health < profile.MinVaultHealthBps {
			out = append(out, AuditFinding{
				ID:         "vault-health-" + vault.ID.String(),
				Severity:   AuditHigh,
				Component:  "vault",
				Subject:    vault.ID.String(),
				Message:    "vault health below profile minimum",
				Epoch:      engine.Clock,
				Metric:     health,
				Threshold:  profile.MinVaultHealthBps,
				Actionable: true,
			})
		}
		if vault.SocializedDebt > profile.MaxSocializedDebt {
			out = append(out, AuditFinding{
				ID:         "vault-socialized-debt-" + vault.ID.String(),
				Severity:   AuditCritical,
				Component:  "vault",
				Subject:    vault.ID.String(),
				Message:    "vault socialized debt exceeds profile maximum",
				Epoch:      engine.Clock,
				Metric:     int64(vault.SocializedDebt),
				Threshold:  int64(profile.MaxSocializedDebt),
				Actionable: true,
			})
		}
	}
	return out
}

func (engine *Engine) auditAccounts(profile AuditProfile) []AuditFinding {
	out := make([]AuditFinding, 0)
	for _, accountID := range sortedAccountIDs(engine.Accounts) {
		account := engine.Accounts[accountID]
		if profile.RequireActiveSnapshots && !account.ActiveRoute.Empty() {
			route := engine.Routes[account.ActiveRoute]
			if route == nil {
				out = append(out, AuditFinding{
					ID:         "account-active-route-missing-" + account.ID.String(),
					Severity:   AuditHigh,
					Component:  "account",
					Subject:    account.ID.String(),
					Message:    "account active route does not exist",
					Epoch:      engine.Clock,
					Actionable: true,
				})
				continue
			}
			drift := absInt64(route.FundingAccumulator - account.FundingSnapshot)
			if drift > profile.MaxSnapshotDriftPPM {
				out = append(out, AuditFinding{
					ID:         "account-snapshot-drift-" + account.ID.String(),
					Severity:   AuditMedium,
					Component:  "account",
					Subject:    account.ID.String(),
					Message:    "account snapshot drifts from active route accumulator",
					Epoch:      engine.Clock,
					Metric:     drift,
					Threshold:  profile.MaxSnapshotDriftPPM,
					Actionable: true,
				})
			}
		}
	}
	return out
}

func (engine *Engine) auditPositions(profile AuditProfile) []AuditFinding {
	out := make([]AuditFinding, 0)
	for _, positionID := range sortedPositionIDs(engine.Positions) {
		position := engine.Positions[positionID]
		route := engine.Routes[position.RouteID]
		if route == nil {
			out = append(out, AuditFinding{
				ID:         "position-route-missing-" + position.ID.String(),
				Severity:   AuditHigh,
				Component:  "position",
				Subject:    position.ID.String(),
				Message:    "position route does not exist",
				Epoch:      engine.Clock,
				Actionable: true,
			})
			continue
		}
		if position.IsOpen() {
			ratio := position.MarginRatioBps(route)
			asset, err := engine.Asset(position.Asset)
			if err == nil && ratio < asset.MaintenanceMarginBps {
				out = append(out, AuditFinding{
					ID:         "position-below-maintenance-" + position.ID.String(),
					Severity:   AuditHigh,
					Component:  "position",
					Subject:    position.ID.String(),
					Message:    "position is below maintenance margin",
					Epoch:      engine.Clock,
					Metric:     ratio,
					Threshold:  asset.MaintenanceMarginBps,
					Actionable: true,
				})
			}
		}
		if profile.WarnOnClosedFundingGap && !position.IsOpen() {
			expected, err := route.FundingFor(position.Notional, position.EntryAccumulator)
			if err == nil && expected != position.FundingPaid {
				out = append(out, AuditFinding{
					ID:         "position-funding-gap-" + position.ID.String(),
					Severity:   AuditCritical,
					Component:  "position",
					Subject:    position.ID.String(),
					Message:    "closed position funding differs from route accumulator expectation",
					Epoch:      engine.Clock,
					Metric:     expected - position.FundingPaid,
					Threshold:  0,
					Actionable: true,
				})
			}
		}
	}
	return out
}

func (engine *Engine) auditPool(profile AuditProfile) []AuditFinding {
	out := make([]AuditFinding, 0)
	for asset, amount := range engine.Pool.SocializedDebt {
		if amount > profile.MaxSocializedDebt {
			out = append(out, AuditFinding{
				ID:         "pool-socialized-debt-" + asset,
				Severity:   AuditCritical,
				Component:  "pool",
				Subject:    asset,
				Message:    "pool socialized debt exceeds profile maximum",
				Epoch:      engine.Clock,
				Metric:     int64(amount),
				Threshold:  int64(profile.MaxSocializedDebt),
				Actionable: true,
			})
		}
	}
	return out
}
