package returncodes

type ReturnDisposition struct {
	Code                  string
	Retryable             bool
	InvalidateBankAccount bool
	Description           string
}

func ClassifyACHReturnCode(code string) ReturnDisposition {
	switch code {
	case "R01":
		return ReturnDisposition{Code: "R01", Retryable: true, InvalidateBankAccount: false, Description: "Insufficient Funds"}
	case "R02":
		return ReturnDisposition{Code: "R02", Retryable: false, InvalidateBankAccount: true, Description: "Account Closed"}
	case "R03":
		return ReturnDisposition{Code: "R03", Retryable: false, InvalidateBankAccount: true, Description: "No Account / Unable to Locate"}
	case "R04":
		return ReturnDisposition{Code: "R04", Retryable: false, InvalidateBankAccount: true, Description: "Invalid Account Number Structure"}
	case "R29":
		return ReturnDisposition{Code: "R29", Retryable: false, InvalidateBankAccount: true, Description: "Corporate Customer Advises Not Authorized"}
	default:
		return ReturnDisposition{Code: code, Retryable: false, InvalidateBankAccount: true, Description: "Unmapped Terminal ACH Return"}
	}
}
