package noc

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrUnsupportedChangeCode indicates a non-standard or unrecognized Nacha COR C-code.
	ErrUnsupportedChangeCode = errors.New("noc: unsupported Nacha Notification of Change code")
	// ErrEmptyCorrectedData indicates Addenda Type 98 Corrected Data field is blank.
	ErrEmptyCorrectedData = errors.New("noc: corrected data field cannot be empty")
	// ErrInvalidABARoutingChecksum indicates the replacement 9-digit routing number failed Mod-10 check.
	ErrInvalidABARoutingChecksum = errors.New("noc: replacement ABA routing number failed Modulo-10 checksum")
	// ErrInvalidAccountFormat indicates an invalid replacement DFI account number or transaction code.
	ErrInvalidAccountFormat = errors.New("noc: invalid replacement DFI account number or transaction code")
)

// ChangeInstruction represents a validated beneficiary bank update parsed from a Nacha COR entry.
type ChangeInstruction struct {
	ChangeCode             string
	OriginalTraceNumber    string
	NewRoutingNumber       string
	NewAccountNumber       string
	NewTransactionCode     string
	RequiresRoutingUpdate  bool
	RequiresAccountUpdate  bool
	RequiresTranCodeUpdate bool
	ComplianceDeadlineUTC  time.Time
}

// ValidateABARoutingMod10 verifies a 9-digit US ABA routing number using Nacha [3, 7, 1] weights.
func ValidateABARoutingMod10(routing string) bool {
	if len(routing) != 9 || routing == "000000000" {
		return false
	}
	weights := [9]int{3, 7, 1, 3, 7, 1, 3, 7, 1}
	sum := 0
	for i := 0; i < 9; i++ {
		ch := routing[i]
		if ch < '0' || ch > '9' {
			return false
		}
		sum += int(ch-'0') * weights[i]
	}
	return sum%10 == 0
}

func isValidAccountToken(acct string) bool {
	if len(acct) < 4 || len(acct) > 17 {
		return false
	}
	for i := 0; i < len(acct); i++ {
		c := acct[i]
		isDigit := c >= '0' && c <= '9'
		isUpper := c >= 'A' && c <= 'Z'
		if !isDigit && !isUpper && c != '-' {
			return false
		}
	}
	return true
}

func isValidCreditTranCode(tc string) bool {
	// 22 = Demand/Checking Credit, 23 = Prenote Checking, 32 = Savings Credit, 33 = Prenote Savings
	return tc == "22" || tc == "23" || tc == "32" || tc == "33"
}

func addSixBankingDays(from time.Time) time.Time {
	cur := from.UTC()
	added := 0
	for added < 6 {
		cur = cur.AddDate(0, 0, 1)
		if cur.Weekday() != time.Saturday && cur.Weekday() != time.Sunday {
			added++
		}
	}
	return cur
}

// ParseAddenda98 parses Nacha Notification of Change (COR) codes C01 through C07 and computes the 6-banking-day SLA.
func ParseAddenda98(changeCode, correctedData, originalTrace string, receivedAt time.Time) (ChangeInstruction, error) {
	code := strings.TrimSpace(strings.ToUpper(changeCode))
	trimmed := strings.TrimSpace(correctedData)
	if trimmed == "" {
		return ChangeInstruction{}, ErrEmptyCorrectedData
	}

	inst := ChangeInstruction{
		ChangeCode:            code,
		OriginalTraceNumber:   strings.TrimSpace(originalTrace),
		ComplianceDeadlineUTC: addSixBankingDays(receivedAt),
	}

	fields := strings.Fields(trimmed)

	switch code {
	case "C01": // Incorrect DFI Account Number
		acct := strings.ToUpper(fields[0])
		if !isValidAccountToken(acct) {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		inst.NewAccountNumber = acct
		inst.RequiresAccountUpdate = true
		return inst, nil

	case "C02": // Incorrect Routing/Transit Number
		routing := fields[0]
		if !ValidateABARoutingMod10(routing) {
			return ChangeInstruction{}, fmt.Errorf("%w: %s", ErrInvalidABARoutingChecksum, routing)
		}
		inst.NewRoutingNumber = routing
		inst.RequiresRoutingUpdate = true
		return inst, nil

	case "C03": // Incorrect Routing/Transit Number and Incorrect DFI Account Number
		var routing, acct string
		if len(fields) >= 2 {
			routing = fields[0]
			acct = strings.ToUpper(fields[1])
		} else if len(trimmed) > 9 {
			routing = trimmed[:9]
			acct = strings.ToUpper(strings.TrimSpace(trimmed[9:]))
		} else {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		if !ValidateABARoutingMod10(routing) {
			return ChangeInstruction{}, fmt.Errorf("%w: %s", ErrInvalidABARoutingChecksum, routing)
		}
		if !isValidAccountToken(acct) {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		inst.NewRoutingNumber = routing
		inst.NewAccountNumber = acct
		inst.RequiresRoutingUpdate = true
		inst.RequiresAccountUpdate = true
		return inst, nil

	case "C05": // Incorrect Transaction Code
		tc := fields[0]
		if !isValidCreditTranCode(tc) {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		inst.NewTransactionCode = tc
		inst.RequiresTranCodeUpdate = true
		return inst, nil

	case "C06": // Incorrect DFI Account Number and Incorrect Transaction Code
		if len(fields) < 2 {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		acct := strings.ToUpper(fields[0])
		tc := fields[1]
		if !isValidAccountToken(acct) || !isValidCreditTranCode(tc) {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		inst.NewAccountNumber = acct
		inst.NewTransactionCode = tc
		inst.RequiresAccountUpdate = true
		inst.RequiresTranCodeUpdate = true
		return inst, nil

	case "C07": // Incorrect Routing/Transit Number, Account Number, and Transaction Code
		if len(fields) < 3 {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		routing := fields[0]
		acct := strings.ToUpper(fields[1])
		tc := fields[2]
		if !ValidateABARoutingMod10(routing) {
			return ChangeInstruction{}, fmt.Errorf("%w: %s", ErrInvalidABARoutingChecksum, routing)
		}
		if !isValidAccountToken(acct) || !isValidCreditTranCode(tc) {
			return ChangeInstruction{}, ErrInvalidAccountFormat
		}
		inst.NewRoutingNumber = routing
		inst.NewAccountNumber = acct
		inst.NewTransactionCode = tc
		inst.RequiresRoutingUpdate = true
		inst.RequiresAccountUpdate = true
		inst.RequiresTranCodeUpdate = true
		return inst, nil

	default:
		return ChangeInstruction{}, fmt.Errorf("%w: %s", ErrUnsupportedChangeCode, code)
	}
}
