package pacs002

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrEmptyStatusXML indicates an empty or whitespace-only ISO-20022 callback body.
	ErrEmptyStatusXML = errors.New("pacs002: empty ISO-20022 XML payload")
	// ErrMalformedStatusReport indicates invalid XML structure or missing transaction status nodes.
	ErrMalformedStatusReport = errors.New("pacs002: malformed FIToFIPmtStsRpt or missing TxInfAndSts")
)

// TreasuryDisposition specifies the deterministic compensating action required by payout-engine.
type TreasuryDisposition string

const (
	DispositionFinalized         TreasuryDisposition = "SETTLEMENT_FINALIZED"
	DispositionInFlightHold      TreasuryDisposition = "IN_FLIGHT_PENDING_HOLD"
	DispositionReverseAndFreeze  TreasuryDisposition = "REVERSE_HOLD_AND_FREEZE_BANK_ACCOUNT"
	DispositionReverseCompliance TreasuryDisposition = "REVERSE_HOLD_AND_ESCALATE_COMPLIANCE"
	DispositionReverseRetryable  TreasuryDisposition = "REVERSE_HOLD_ELIGIBLE_FOR_FAILOVER"
)

// EvaluatedStatus represents the normalized outcome of an ISO-20022 pacs.002 status callback.
type EvaluatedStatus struct {
	MessageID            string
	OriginalInstruction  string
	OriginalEndToEndID   string
	RawStatus            string
	ReasonCode           string
	AdditionalInfo       string
	Disposition          TreasuryDisposition
	ReleaseNostroBuffer  bool
	TriggerRailFailover  bool
	FreezeBeneficiaryAccount bool
}

type xmlDocument struct {
	XMLName xml.Name        `xml:"Document"`
	Report  xmlStatusReport `xml:"FIToFIPmtStsRpt"`
}

type xmlStatusReport struct {
	GroupHeader xmlGroupHeader `xml:"GrpHdr"`
	TxStatus    []xmlTxStatus  `xml:"TxInfAndSts"`
}

type xmlGroupHeader struct {
	MessageID string `xml:"MsgId"`
}

type xmlTxStatus struct {
	OriginalInstrID    string          `xml:"OrgnlInstrId"`
	OriginalEndToEndID string          `xml:"OrgnlEndToEndId"`
	TransactionStatus  string          `xml:"TxSts"`
	StatusReasons      []xmlReasonInfo `xml:"StsRsnInf"`
}

type xmlReasonInfo struct {
	Reason         xmlReasonCode `xml:"Rsn"`
	AdditionalInfo []string      `xml:"AddtlInf"`
}

type xmlReasonCode struct {
	Code        string `xml:"Cd"`
	Proprietary string `xml:"Prtry"`
}

// EvaluateStatusReport parses an ISO-20022 pacs.002.001.10 XML payload and classifies settlement or compensation actions.
func EvaluateStatusReport(rawXML []byte) (EvaluatedStatus, error) {
	if len(strings.TrimSpace(string(rawXML))) == 0 {
		return EvaluatedStatus{}, ErrEmptyStatusXML
	}

	var doc xmlDocument
	if err := xml.Unmarshal(rawXML, &doc); err != nil {
		return EvaluatedStatus{}, fmt.Errorf("%w: %v", ErrMalformedStatusReport, err)
	}

	msgID := strings.TrimSpace(doc.Report.GroupHeader.MessageID)
	if msgID == "" || len(doc.Report.TxStatus) == 0 {
		return EvaluatedStatus{}, ErrMalformedStatusReport
	}

	tx := doc.Report.TxStatus[0]
	e2e := strings.TrimSpace(tx.OriginalEndToEndID)
	txSts := strings.TrimSpace(strings.ToUpper(tx.TransactionStatus))
	if e2e == "" || txSts == "" {
		return EvaluatedStatus{}, ErrMalformedStatusReport
	}

	var reasonCode, addtlInfo string
	if len(tx.StatusReasons) > 0 {
		reasonCode = strings.TrimSpace(strings.ToUpper(tx.StatusReasons[0].Reason.Code))
		if reasonCode == "" {
			reasonCode = strings.TrimSpace(strings.ToUpper(tx.StatusReasons[0].Reason.Proprietary))
		}
		if len(tx.StatusReasons[0].AdditionalInfo) > 0 {
			addtlInfo = strings.TrimSpace(tx.StatusReasons[0].AdditionalInfo[0])
		}
	}

	out := EvaluatedStatus{
		MessageID:           msgID,
		OriginalInstruction: strings.TrimSpace(tx.OriginalInstrID),
		OriginalEndToEndID:  e2e,
		RawStatus:           txSts,
		ReasonCode:          reasonCode,
		AdditionalInfo:      addtlInfo,
	}

	switch txSts {
	case "ACCC", "ACSC":
		out.Disposition = DispositionFinalized
		out.ReleaseNostroBuffer = false
		return out, nil

	case "ACSP", "ACTC", "PDNG":
		out.Disposition = DispositionInFlightHold
		return out, nil

	case "RJCT":
		out.ReleaseNostroBuffer = true
		switch reasonCode {
		case "RR04", "RR01", "RR02", "RR03": // Regulatory / AML / Sanctions rejection
			out.Disposition = DispositionReverseCompliance
			out.FreezeBeneficiaryAccount = true
			out.TriggerRailFailover = false
		case "AC01", "AC04", "AC06", "AG01", "BE04": // Closed/invalid/blocked beneficiary account
			out.Disposition = DispositionReverseAndFreeze
			out.FreezeBeneficiaryAccount = true
			out.TriggerRailFailover = false
		default: // Transient switch timeout, liquidity window, or creditor agent offline (AB05, AM04, TM01)
			out.Disposition = DispositionReverseRetryable
			out.FreezeBeneficiaryAccount = false
			out.TriggerRailFailover = true
		}
		return out, nil

	default:
		return EvaluatedStatus{}, fmt.Errorf("%w: unsupported TxSts %q", ErrMalformedStatusReport, txSts)
	}
}
