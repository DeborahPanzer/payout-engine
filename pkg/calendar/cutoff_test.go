package calendar

import (
	"testing"
	"time"
)

func TestSelectUSDDisbursementRail_WindowsAndHolidayFailover(t *testing.T) {
	et := EasternLocation()

	// 1. Regular business day (Wednesday 2026-11-04 09:15 ET) -> Same-Day ACH Window 1
	morningET := time.Date(2026, time.November, 4, 9, 15, 0, 0, et).UTC()
	dec1 := SelectUSDDisbursementRail(morningET, 250_000, true)
	if dec1.SelectedRail != RailSameDayACHWindow1 || dec1.PromotedToInstantRail {
		t.Fatalf("expected Window 1 Same-Day ACH, got %+v", dec1)
	}

	// 2. Veterans Day Federal Reserve Holiday (Wednesday 2026-11-11 11:00 ET) with Same-Day SLA -> FedNow
	veteransDayET := time.Date(2026, time.November, 11, 11, 0, 0, 0, et).UTC()
	decHoliday := SelectUSDDisbursementRail(veteransDayET, 500_000, true)
	if decHoliday.SelectedRail != RailFedNowInstant || !decHoliday.PromotedToInstantRail || decHoliday.IsBankingBusinessDay {
		t.Fatalf("expected FedNow promotion on Veterans Day holiday, got %+v", decHoliday)
	}

	// 3. Friday after 16:45 ET cut-off without urgent SLA -> Roll forward to Monday Next-Day ACH
	fridayLateET := time.Date(2026, time.November, 6, 18, 30, 0, 0, et).UTC()
	decDeferred := SelectUSDDisbursementRail(fridayLateET, 500_000, false)
	if decDeferred.SelectedRail != RailNextDayACH || decDeferred.SettlementDateISO != "2026-11-09" {
		t.Fatalf("expected roll-forward to Monday 2026-11-09, got %+v", decDeferred)
	}

	// 4. Thanksgiving Day (4th Thursday of November: 2026-11-26) -> Recognized as Federal Reserve holiday
	thanksgiving := time.Date(2026, time.November, 26, 10, 0, 0, 0, et)
	if IsUSBankingDay(thanksgiving) {
		t.Fatal("expected Thanksgiving 2026-11-26 to be closed Federal Reserve holiday")
	}
}

func TestSelectEURDisbursementRail_Target2AndInstantPromotion(t *testing.T) {
	// Saturday -> SEPA Instant promotion for eligible amount
	satUTC := time.Date(2026, time.November, 14, 12, 0, 0, 0, time.UTC)
	eurDec := SelectEURDisbursementRail(satUTC, 150_000, false)
	if eurDec.SelectedRail != RailSEPAInstant || !eurDec.PromotedToInstantRail {
		t.Fatalf("expected SEPA Instant promotion on weekend, got %+v", eurDec)
	}
}
