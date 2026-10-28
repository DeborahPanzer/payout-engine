package ofac

import "strings"

type ScreeningDecision struct {
	Allowed        bool
	MatchedEntity  string
	SimilarityBps  int
	ComplianceHold bool
}

func ScreenBeneficiary(beneficiaryName string, sdnBlocklist []string) ScreeningDecision {
	normTarget := normalizeEntityName(beneficiaryName)
	if normTarget == "" {
		return ScreeningDecision{Allowed: false, ComplianceHold: true}
	}

	for _, entry := range sdnBlocklist {
		normEntry := normalizeEntityName(entry)
		if normTarget == normEntry || strings.Contains(normTarget, normEntry) {
			return ScreeningDecision{
				Allowed:        false,
				MatchedEntity:  entry,
				SimilarityBps:  10000,
				ComplianceHold: true,
			}
		}
	}

	return ScreeningDecision{
		Allowed:        true,
		SimilarityBps:  0,
		ComplianceHold: false,
	}
}

func normalizeEntityName(s string) string {
	upper := strings.ToUpper(strings.TrimSpace(s))
	replacer := strings.NewReplacer(".", "", ",", "", "-", " ", "LLC", "", "INC", "")
	return strings.Join(strings.Fields(replacer.Replace(upper)), " ")
}
