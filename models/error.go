package models

type APIError struct {
	ErrorCode        string  `json:"errorcode"`
	Message          string  `json:"message"`
	FAPIErrorMessage *string `json:"fapiMessage,omitempty"`
}
