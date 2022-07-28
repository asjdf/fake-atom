package student

import "github.com/hduhelp/api_open_sdk/transfer"

type Info struct {
	ClassID     string `json:"classId"`
	MajorID     string `json:"majorId"`
	MajorName   string `json:"majorName"`
	StaffID     string `json:"staffId"`
	StaffName   string `json:"staffName"`
	TeacherID   string `json:"teacherId"`
	TeacherName string `json:"teacherName"`
	UnitID      string `json:"unitId"`
	UnitName    string `json:"unitName"`
}

func GetInfo(staffId string) (d *Info, err error) {
	d = new(Info)
	_, _, err = transfer.Get("salmon_base", "/student/info", nil, staffId).
		EndStruct(&d)
	return
}
