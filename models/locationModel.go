package models

type LocationInfo struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	LastIP string
}
