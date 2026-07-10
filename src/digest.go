package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func (engine *Engine) StateDigest() string {
	payload := map[string]any{
		"network":   engine.NetworkID,
		"clock":     engine.Clock,
		"assets":    engine.Assets.List(),
		"operators": engine.OperatorReports(),
		"vaults":    engine.VaultReports(),
		"accounts":  engine.AccountReports(),
		"routes":    engine.RouteReports(),
		"positions": engine.PositionReports(),
		"pool":      engine.Pool.Report(),
		"events":    engine.Journal.Events(),
	}
	sum := sha256.Sum256(stableJSON(payload))
	return hex.EncodeToString(sum[:16])
}
