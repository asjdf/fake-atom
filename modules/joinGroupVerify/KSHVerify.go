package joinGroupVerify

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/guonaihong/gout"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"strings"
)

type kshInfo struct {
	Ksh  string //考生号
	Name string //姓名
	Spec string //专业
	From string //生源地
}

var specShort = map[string]string{
	"计算机科学与技术(中外合作办学)": "计算机",
	"自动化(中外合作办学)":      "自动化",
}

//专门用于通过考生号验证无学号的新生
func verifyByKSH(name string, ksh string) (bool, *kshInfo, error) {
	body := &bytes.Buffer{}
	err := gout.POST("http://zhaosheng0.hdu.edu.cn/Template/Default/search.asp?Action=OK").
		SetWWWForm(gout.H{
			"txtKSH":  ksh,
			"txtname": "",
			"Submit":  "%CC%E1%BD%BB%B2%E9%D1%AF%C4%DA%C8%DD",
		}).BindBody(body).Do()
	if err != nil {
		return false, nil, errors.New("network error")
	}

	utfBody := transform.NewReader(body, simplifiedchinese.GBK.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		return false, nil, errors.New("network error")
	}
	table := doc.Find("body > table > tbody > tr")
	if table.Length() <= 1 {
		return false, nil, errors.New("network error")
	}

	info := kshInfo{}
	pass := false
	table.NextAll().EachWithBreak(func(i int, s *goquery.Selection) bool {
		s.Find("td").Each(func(j int, s *goquery.Selection) {
			switch j {
			case 0:
				info.Ksh = strings.Trim(s.Text(), " \n")
			case 1:
				info.Name = strings.Trim(s.Text(), " \n")
			case 2:
				info.Spec = strings.Trim(s.Text(), " \n")
			case 3:
				info.From = strings.Trim(s.Text(), " \n")
			}
		})
		if info.Ksh == ksh && info.Name == name {
			pass = true
			return false
		}
		info = kshInfo{}
		return true
	})
	return pass, &info, nil
}
