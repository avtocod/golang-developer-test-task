package structs

type (
	// Info is struct for information parsing from json
	Info struct {
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

	// InfoList is alias for []Info
	//
	//easyjson:json
	InfoList []Info

	// URLObject is struct for getting URL from user
	URLObject struct {
		URL string `json:"url"`
	}

	// SearchObject is struct for query data
	SearchObject struct {
		GlobalID       *int    `json:"global_id,omitempty"`
		SystemObjectID *string `json:"system_object_id,omitempty"`
		ID             *int    `json:"id,omitempty"`
		Mode           *string `json:"mode,omitempty"`
		IDEn           *int    `json:"id_en,omitempty"`
		ModeEn         *string `json:"mode_en,omitempty"`
		Offset         int64   `json:"offset"`
	}

	// PaginationObject contains info about data by query which is contained in DB
	PaginationObject struct {
		Size        int64    `json:"size"`
		Offset      int64    `json:"offset"`
		HasNext     bool     `json:"hasNext"`
		HasPrevious bool     `json:"hasPrevious"`
		Data        InfoList `json:"data"`
	}
)
