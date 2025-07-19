package model

type ErrorInfo struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type ErrorResponse struct {
	Error ErrorInfo `json:"error,omitempty"`
}

type Response struct {
	Response interface{} `json:"response,omitempty"`
}

type DataResponse struct {
	Data interface{} `json:"data,omitempty"`
}
