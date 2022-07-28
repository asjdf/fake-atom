package person

import "github.com/hduhelp/api_open_sdk/transfer"

type infoFromSalmonBase struct {
	StaffId    string `json:"staffId"`
	StaffName  string `json:"staffName"`
	StaffState string `json:"staffState"`
	StaffType  string `json:"staffType"`
	UnitCode   string `json:"unitCode"`
}

func GetInfo(staffId string) (d *infoFromSalmonBase, err error) {
	d = new(infoFromSalmonBase)
	_, _, err = transfer.Get("salmon_base", "/person/info", nil, staffId).
		EndStruct(&d)
	return
}
