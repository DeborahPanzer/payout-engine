package pacs002

import (
	"errors"
	"testing"
)

func TestEvaluateStatusReport_SettlementAndRejectionDispositions(t *testing.T) {
	acccXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.002.001.10">
  <FIToFIPmtStsRpt>
    <GrpHdr><MsgId>MSG-FEDNOW-20261112-001</MsgId></GrpHdr>
    <TxInfAndSts>
      <OrgnlInstrId>INSTR-9001</OrgnlInstrId>
      <OrgnlEndToEndId>E2E-PAYOUT-9001</OrgnlEndToEndId>
      <TxSts>ACCC</TxSts>
    </TxInfAndSts>
  </FIToFIPmtStsRpt>
</Document>`)

	settled, err := EvaluateStatusReport(acccXML)
	if err != nil {
		t.Fatalf("expected ACCC XML to parse cleanly, got: %v", err)
	}
	if settled.Disposition != DispositionFinalized || settled.TriggerRailFailover {
		t.Fatalf("expected finalized settlement disposition, got %+v", settled)
	}

	rr04XML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.002.001.10">
  <FIToFIPmtStsRpt>
    <GrpHdr><MsgId>MSG-FEDNOW-20261112-002</MsgId></GrpHdr>
    <TxInfAndSts>
      <OrgnlInstrId>INSTR-9002</OrgnlInstrId>
      <OrgnlEndToEndId>E2E-PAYOUT-9002</OrgnlEndToEndId>
      <TxSts>RJCT</TxSts>
      <StsRsnInf>
        <Rsn><Cd>RR04</Cd></Rsn>
        <AddtlInf>REGULATORY_SANCTIONS_BLOCK</AddtlInf>
      </StsRsnInf>
    </TxInfAndSts>
  </FIToFIPmtStsRpt>
</Document>`)

	regBlock, err := EvaluateStatusReport(rr04XML)
	if err != nil {
		t.Fatalf("expected RR04 XML to parse cleanly, got: %v", err)
	}
	if regBlock.Disposition != DispositionReverseCompliance || !regBlock.FreezeBeneficiaryAccount || regBlock.TriggerRailFailover {
		t.Fatalf("expected compliance escalation with zero failover on RR04, got %+v", regBlock)
	}

	ab05XML := []byte(`<Document><FIToFIPmtStsRpt><GrpHdr><MsgId>M3</MsgId></GrpHdr><TxInfAndSts><OrgnlEndToEndId>E2E-3</OrgnlEndToEndId><TxSts>RJCT</TxSts><StsRsnInf><Rsn><Cd>AB05</Cd></Rsn></StsRsnInf></TxInfAndSts></FIToFIPmtStsRpt></Document>`)
	timeoutRej, err := EvaluateStatusReport(ab05XML)
	if err != nil {
		t.Fatalf("expected AB05 XML to parse cleanly, got: %v", err)
	}
	if timeoutRej.Disposition != DispositionReverseRetryable || !timeoutRej.TriggerRailFailover {
		t.Fatalf("expected AB05 switch timeout to allow ACH rail failover, got %+v", timeoutRej)
	}

	_, badErr := EvaluateStatusReport([]byte(`<Document><Broken>`))
	if !errors.Is(badErr, ErrMalformedStatusReport) {
		t.Fatalf("expected ErrMalformedStatusReport on broken XML, got: %v", badErr)
	}
}
