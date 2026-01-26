package domain

type Status string

const (
	StatusPaymentPending string = "pending"
	StatusSucceeded      string = "succeeded"
	StatusFailed         string = "failed"
)
