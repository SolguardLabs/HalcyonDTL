package main

import "testing"

func auditFindingByID(findings []AuditFinding, id string) (AuditFinding, bool) {
	for _, finding := range findings {
		if finding.ID == id {
			return finding, true
		}
	}
	return AuditFinding{}, false
}

func TestAuditSeparatesCapacityAndVaultAlerts(t *testing.T) {
	engine, err := BaseEngine()
	if err != nil {
		t.Fatal(err)
	}
	route := engine.Routes[NewRouteID("atlas-eu")]
	route.Utilized = MustAmount(970_000)
	route.OpenInterest = 1
	route.Status = RoutePaused
	vault := engine.Vaults[NewVaultID("atlas-eu-source")]
	vault.SocializedDebt = MustAmount(6_000)

	findings := engine.Audit(DefaultAuditProfile())
	cases := []struct {
		id       string
		severity AuditSeverity
	}{
		{"route-utilization-" + route.ID.String(), AuditHigh},
		{"paused-route-open-interest-" + route.ID.String(), AuditMedium},
		{"vault-socialized-debt-" + vault.ID.String(), AuditCritical},
	}
	for _, test := range cases {
		finding, ok := auditFindingByID(findings, test.id)
		if !ok {
			t.Fatalf("missing audit finding %s", test.id)
		}
		if finding.Severity != test.severity || !finding.Actionable {
			t.Fatalf("unexpected finding %s: %+v", test.id, finding)
		}
	}
}

func TestAuditFindsBrokenRouteReferences(t *testing.T) {
	engine, err := BaseEngine()
	if err != nil {
		t.Fatal(err)
	}
	position, err := engine.OpenPosition(
		NewAccountID("alice"), NewRouteID("atlas-eu"),
		MustAmount(120_000), MustAmount(20_000),
	)
	if err != nil {
		t.Fatal(err)
	}
	delete(engine.Routes, position.RouteID)

	findings := engine.Audit(DefaultAuditProfile())
	for _, id := range []string{
		"account-active-route-missing-" + position.AccountID.String(),
		"position-route-missing-" + position.ID.String(),
	} {
		finding, ok := auditFindingByID(findings, id)
		if !ok || finding.Severity != AuditHigh {
			t.Fatalf("missing route reference alert %s: %+v", id, finding)
		}
	}
}

func TestAuditRespectsConfiguredUtilizationThreshold(t *testing.T) {
	engine, err := BaseEngine()
	if err != nil {
		t.Fatal(err)
	}
	route := engine.Routes[NewRouteID("atlas-eu")]
	route.Utilized = MustAmount(940_000)
	id := "route-utilization-" + route.ID.String()
	if _, ok := auditFindingByID(engine.Audit(DefaultAuditProfile()), id); ok {
		t.Fatal("default profile rejected an allowed utilization")
	}
	strict := DefaultAuditProfile()
	strict.MaxUtilizationBps = 9_000
	finding, ok := auditFindingByID(engine.Audit(strict), id)
	if !ok || finding.Metric != 9_400 || finding.Threshold != 9_000 {
		t.Fatalf("strict utilization finding is incorrect: %+v", finding)
	}
}
