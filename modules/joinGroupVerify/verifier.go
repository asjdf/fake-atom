package joinGroupVerify

import (
	"errors"
	"regexp"

	"fake-atom/pkg/person"
)

func verifyByStuffID(name string, stuffID string) (pass bool, err error) {
	data, err := person.GetInfo(stuffID)
	if err != nil {
		if err.Error() == "json.Marsha.Error" {
			return false, nil
		}
		return false, errors.New("上游校验服务错误") // 这里应该还要处理一下网络错误，暂时不写
	}
	if data != nil {
		if filterDotInName(data.StaffName) == filterDotInName(name) {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, errors.New("遭遇其他阴间错误，请联系243852814")
}

func filterDotInName(name string) string {
	temp := regexp.MustCompile("[\u4e00-\u9fa5]*").FindAllString(name, -1)
	nameOutput := ""
	for _, v := range temp {
		nameOutput += v
	}
	return nameOutput
}
