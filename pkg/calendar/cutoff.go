package calendar

import "time"

// RailSelection identifies the settlement rail selected by the banking calendar router.
type RailSelection string

const (
	RailSameDayACHWindow1 RailSelection = "US_ACH_SAME_DAY_W1_1030_ET"
	RailSameDayACHWindow2 RailSelection = "US_ACH_SAME_DAY_W2_1445_ET"
	RailSameDayACHWindow3 RailSelection = "US_ACH_SAME_DAY_W3_1645_ET"
	RailNextDayACH        RailSelection = "US_ACH_NEXT_BANKING_DAY"
	RailFedNowInstant     RailSelection = "US_FEDNOW_24X7_INSTANT"
	RailSEPAStandard      RailSelection = "EU_SEPA_TARGET2_SCT"
	RailSEPAInstant       RailSelection = "EU_SEPA_SCT_INST_24X7"
)

// RoutingDecision encapsulates the banking cut-off evaluation and target settlement date.
type RoutingDecision struct {
	SelectedRail          RailSelection
	IsBankingBusinessDay  bool
	PromotedToInstantRail bool
	SettlementDateISO     string
	ReasonCode            string
}

// EasternLocation returns America/New_York or a deterministic UTC-5 EST fallback in minimal containers.
func EasternLocation() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.FixedZone("EST", -5*3600)
	}
	return loc
}

// IsFederalReserveHoliday evaluates the 11 US Federal Reserve holidays including Sunday->Monday observance.
func IsFederalReserveHoliday(t time.Time) bool {
	y, m, d := t.Date()
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}

	matchesFixedOrSundayObserved := func(month time.Month, day int) bool {
		if m == month && d == day {
			return true
		}
		holiday := time.Date(y, month, day, 0, 0, 0, 0, t.Location())
		if holiday.Weekday() == time.Sunday && m == month && d == day+1 {
			return true
		}
		return false
	}

	// Fixed-date Federal Reserve holidays (Sunday holidays observed on following Monday).
	if matchesFixedOrSundayObserved(time.January, 1) || // New Year's Day
		matchesFixedOrSundayObserved(time.June, 19) || // Juneteenth
		matchesFixedOrSundayObserved(time.July, 4) || // Independence Day
		matchesFixedOrSundayObserved(time.November, 11) || // Veterans Day
		matchesFixedOrSundayObserved(time.December, 25) { // Christmas Day
		return true
	}

	nthWeekday := (d-1)/7 + 1
	// Martin Luther King Jr. Day (3rd Monday in January)
	if m == time.January && wd == time.Monday && nthWeekday == 3 {
		return true
	}
	// Washington's Birthday / Presidents' Day (3rd Monday in February)
	if m == time.February && wd == time.Monday && nthWeekday == 3 {
		return true
	}
	// Memorial Day (Last Monday in May)
	if m == time.May && wd == time.Monday && d+7 > 31 {
		return true
	}
	// Labor Day (1st Monday in September)
	if m == time.September && wd == time.Monday && nthWeekday == 1 {
		return true
	}
	// Columbus Day (2nd Monday in October)
	if m == time.October && wd == time.Monday && nthWeekday == 2 {
		return true
	}
	// Thanksgiving Day (4th Thursday in November)
	if m == time.November && wd == time.Thursday && nthWeekday == 4 {
		return true
	}

	return false
}

// IsUSBankingDay returns true if the date in Eastern Time is a weekday and not a Federal Reserve holiday.
func IsUSBankingDay(t time.Time) bool {
	et := t.In(EasternLocation())
	if et.Weekday() == time.Saturday || et.Weekday() == time.Sunday {
		return false
	}
	return !IsFederalReserveHoliday(et)
}

// NextUSBankingDay advances day-by-day until reaching an open Federal Reserve business day.
func NextUSBankingDay(t time.Time) time.Time {
	candidate := t.In(EasternLocation()).AddDate(0, 0, 1)
	for !IsUSBankingDay(candidate) {
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate
}

// SelectUSDDisbursementRail evaluates Federal Reserve Same-Day ACH windows (10:30, 14:45, 16:45 ET)
// and promotes urgent payouts ($<= 1,000,000 FedNow limit) to 24x7 FedNow after cut-off or on holidays.
func SelectUSDDisbursementRail(nowUTC time.Time, amountMinor int64, requireSameDaySLA bool) RoutingDecision {
	const fedNowMaxLimitMinor int64 = 100_000_000 // $1,000,000.00 in cents
	const sameDayACHMaxMinor int64 = 100_000_000  // $1,000,000.00 Nacha Same-Day cap

	et := nowUTC.In(EasternLocation())
	todayISO := et.Format("2006-01-02")
	bankingDay := IsUSBankingDay(et)
	minutesOfDay := et.Hour()*60 + et.Minute()

	if bankingDay && amountMinor <= sameDayACHMaxMinor {
		switch {
		case minutesOfDay < 10*60+30: // Before 10:30 ET
			return RoutingDecision{
				SelectedRail:          RailSameDayACHWindow1,
				IsBankingBusinessDay:  true,
				PromotedToInstantRail: false,
				SettlementDateISO:     todayISO,
				ReasonCode:            "WITHIN_FED_ACH_SDA_WINDOW_1",
			}
		case minutesOfDay < 14*60+45: // Before 14:45 ET
			return RoutingDecision{
				SelectedRail:          RailSameDayACHWindow2,
				IsBankingBusinessDay:  true,
				PromotedToInstantRail: false,
				SettlementDateISO:     todayISO,
				ReasonCode:            "WITHIN_FED_ACH_SDA_WINDOW_2",
			}
		case minutesOfDay < 16*60+45: // Before 16:45 ET
			return RoutingDecision{
				SelectedRail:          RailSameDayACHWindow3,
				IsBankingBusinessDay:  true,
				PromotedToInstantRail: false,
				SettlementDateISO:     todayISO,
				ReasonCode:            "WITHIN_FED_ACH_SDA_WINDOW_3",
			}
		}
	}

	if requireSameDaySLA && amountMinor <= fedNowMaxLimitMinor {
		return RoutingDecision{
			SelectedRail:          RailFedNowInstant,
			IsBankingBusinessDay:  bankingDay,
			PromotedToInstantRail: true,
			SettlementDateISO:     todayISO,
			ReasonCode:            "PROMOTED_TO_FEDNOW_AFTER_ACH_CUTOFF_OR_HOLIDAY",
		}
	}

	nextOpen := NextUSBankingDay(et)
	return RoutingDecision{
		SelectedRail:          RailNextDayACH,
		IsBankingBusinessDay:  bankingDay,
		PromotedToInstantRail: false,
		SettlementDateISO:     nextOpen.Format("2006-01-02"),
		ReasonCode:            "DEFERRED_TO_NEXT_FED_BANKING_DAY",
	}
}

// SelectEURDisbursementRail routes SEPA Credit Transfers before 16:00 CET/UTC+1 or upgrades to SEPA Instant.
func SelectEURDisbursementRail(nowUTC time.Time, amountMinor int64, requireInstant bool) RoutingDecision {
	const sepaInstantMaxMinor int64 = 10_000_000 // EUR 100,000.00 SCT Inst scheme cap
	cet := nowUTC.In(time.FixedZone("CET", 3600))
	todayISO := cet.Format("2006-01-02")

	isWeekend := cet.Weekday() == time.Saturday || cet.Weekday() == time.Sunday
	isTarget2Holiday := (cet.Month() == time.January && cet.Day() == 1) ||
		(cet.Month() == time.May && cet.Day() == 1) ||
		(cet.Month() == time.December && (cet.Day() == 25 || cet.Day() == 26))

	openDay := !isWeekend && !isTarget2Holiday
	beforeCutoff := (cet.Hour()*60 + cet.Minute()) < 16*60

	if openDay && beforeCutoff && !requireInstant {
		return RoutingDecision{
			SelectedRail:          RailSEPAStandard,
			IsBankingBusinessDay:  true,
			PromotedToInstantRail: false,
			SettlementDateISO:     todayISO,
			ReasonCode:            "WITHIN_TARGET2_SCT_CUTOFF",
		}
	}

	if amountMinor <= sepaInstantMaxMinor {
		return RoutingDecision{
			SelectedRail:          RailSEPAInstant,
			IsBankingBusinessDay:  openDay,
			PromotedToInstantRail: true,
			SettlementDateISO:     todayISO,
			ReasonCode:            "ROUTED_TO_SEPA_SCT_INST_24X7",
		}
	}

	next := cet.AddDate(0, 0, 1)
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.AddDate(0, 0, 1)
	}
	return RoutingDecision{
		SelectedRail:          RailSEPAStandard,
		IsBankingBusinessDay:  openDay,
		PromotedToInstantRail: false,
		SettlementDateISO:     next.Format("2006-01-02"),
		ReasonCode:            "QUEUED_NEXT_TARGET2_SETTLEMENT_DAY",
	}
}
