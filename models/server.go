package models

type Server struct {
	Port        int     `json:"port"`
	IP          string  `json:"ip"`
	Protocol    string  `json:"protocol"`
	StatusCode  int     `json:"statusCode"`
	Response    string  `json:"response"`
	ContentType string  `json:"contentType"`
	Process     *string `json:"process"`
	PID         *int    `json:"pid"`
	UID         uint32  `json:"uid"`
	SockState   string  `json:"sockState"`
	Name        string  `json:"name"`
	Icon        *string `json:"icon"`
}
