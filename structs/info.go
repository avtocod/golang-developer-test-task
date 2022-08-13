package structs

type Info struct {
	GlobalID         int    `json:"global_id"`
	SystemObjectID   string `json:"system_object_id"`
	ID               int    `json:"ID"`
	Name             string `json:"Name"`
	AdmArea          string `json:"AdmArea"`
	District         string `json:"District"`
	Address          string `json:"Address"`
	LongitudeWGS84   string `json:"Longitude_WGS84"`
	LatitudeWGS84    string `json:"Latitude_WGS84"`
	CarCapacity      int    `json:"CarCapacity"`
	Mode             string `json:"Mode"`
	IDEn             int    `json:"ID_en"`
	NameEn           string `json:"Name_en"`
	AdmAreaEn        string `json:"AdmArea_en"`
	DistrictEn       string `json:"District_en"`
	AddressEn        string `json:"Address_en"`
	LongitudeWGS84En string `json:"Longitude_WGS84_en"`
	LatitudeWGS84En  string `json:"Latitude_WGS84_en"`
	CarCapacityEn    int    `json:"CarCapacity_en"`
	ModeEn           string `json:"Mode_en"`
}

//easyjson:json
type InfoList []Info
