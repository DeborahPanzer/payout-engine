package rtp

import (
	"errors"
	"fmt"
)

type Pacs008Transfer struct {
	MsgID          string
	EndToEndID     string
	AmountCents    int64
	Currency       string
	DebtorABA      string
	CreditorABA    string
	CreditorName   string
	SettlementDate string
}

func BuildFedNowPacs008XML(tx Pacs008Transfer) (string, error) {
	if tx.AmountCents <= 0 {
		return "", errors.New("fednow: transfer amount must be positive")
	}
	major := tx.AmountCents / 100
	minor := tx.AmountCents % 100
	amtStr := fmt.Sprintf("%d.%02d", major, minor)

	xml := fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>`+
			`<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.008.001.08">`+
			`<FIToFICstmrCdtTrf>`+
			`<GrpHdr><MsgId>%s</MsgId><NbOfTxs>1</NbOfTxs><SttlmInf><SttlmMtd>CLRG</SttlmMtd></SttlmInf></GrpHdr>`+
			`<CdtTrfTxInf>`+
			`<PmtId><EndToEndId>%s</EndToEndId></PmtId>`+
			`<IntrBkSttlmAmt Ccy="%s">%s</IntrBkSttlmAmt>`+
			`<IntrBkSttlmDt>%s</IntrBkSttlmDt>`+
			`<DbtrAgt><FinInstnId><ClrSysMmbId><MmbId>%s</MmbId></ClrSysMmbId></FinInstnId></DbtrAgt>`+
			`<CdtrAgt><FinInstnId><ClrSysMmbId><MmbId>%s</MmbId></ClrSysMmbId></FinInstnId></CdtrAgt>`+
			`<Cdtr><Nm>%s</Nm></Cdtr>`+
			`</CdtTrfTxInf>`+
			`</FIToFICstmrCdtTrf></Document>`,
		tx.MsgID, tx.EndToEndID, tx.Currency, amtStr, tx.SettlementDate, tx.DebtorABA, tx.CreditorABA, tx.CreditorName,
	)
	return xml, nil
}
