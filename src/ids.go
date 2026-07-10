package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type AccountID string
type OperatorID string
type VaultID string
type RouteID string
type PositionID string
type EpochID int64
type EventID int64

func NewAccountID(label string) AccountID {
	return AccountID("acct-" + slug(label))
}

func NewOperatorID(label string) OperatorID {
	return OperatorID("op-" + slug(label))
}

func NewVaultID(label string) VaultID {
	return VaultID("vault-" + slug(label))
}

func NewRouteID(label string) RouteID {
	return RouteID("route-" + slug(label))
}

func NewPositionID(account AccountID, route RouteID, nonce int64) PositionID {
	return PositionID(fmt.Sprintf("pos-%s-%s-%04d", trimPrefix(string(account), "acct-"), trimPrefix(string(route), "route-"), nonce))
}

func (id AccountID) String() string {
	return string(id)
}

func (id OperatorID) String() string {
	return string(id)
}

func (id VaultID) String() string {
	return string(id)
}

func (id RouteID) String() string {
	return string(id)
}

func (id PositionID) String() string {
	return string(id)
}

func (id AccountID) Empty() bool {
	return id == ""
}

func (id OperatorID) Empty() bool {
	return id == ""
}

func (id VaultID) Empty() bool {
	return id == ""
}

func (id RouteID) Empty() bool {
	return id == ""
}

func (id PositionID) Empty() bool {
	return id == ""
}

func slug(input string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(input)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case !lastDash:
			builder.WriteRune('-')
			lastDash = true
		}
	}
	value := strings.Trim(builder.String(), "-")
	if value == "" {
		return "id"
	}
	return value
}

func trimPrefix(value string, prefix string) string {
	return strings.TrimPrefix(value, prefix)
}

func sortedKeys[V any](input map[string]V) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedAccountIDs(input map[AccountID]*Account) []AccountID {
	keys := make([]AccountID, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedOperatorIDs(input map[OperatorID]*Operator) []OperatorID {
	keys := make([]OperatorID, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedVaultIDs(input map[VaultID]*Vault) []VaultID {
	keys := make([]VaultID, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedRouteIDs(input map[RouteID]*Route) []RouteID {
	keys := make([]RouteID, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedPositionIDs(input map[PositionID]*Position) []PositionID {
	keys := make([]PositionID, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
