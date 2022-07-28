package joinGroupVerify

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/guonaihong/gout"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"strings"
	"testing"
)

func TestGetExamIdInfo(t *testing.T) {
	body := &bytes.Buffer{}
	err := gout.POST("http://zhaosheng0.hdu.edu.cn/Template/Default/search.asp?Action=OK").
		SetWWWForm(gout.H{
			"txtKSH":  "22331605150185",
			"txtname": "",
			"Submit":  "%CC%E1%BD%BB%B2%E9%D1%AF%C4%DA%C8%DD",
		}).BindBody(body).Do()
	if err != nil {
		return
	}

	utfBody := transform.NewReader(body, simplifiedchinese.GBK.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		return
	}
	table := doc.Find("body > table > tbody > tr")
	if table.Length() <= 1 {
		return
	}
	table.NextAll().EachWithBreak(func(i int, s *goquery.Selection) bool {
		s.Find("td").EachWithBreak(func(j int, s *goquery.Selection) bool {
			fmt.Println(strings.Trim(s.Text(), " \n"))
			return true
		})
		return true
	})
}

func TestVerifyByKsh(t *testing.T) {
	fmt.Println(verifyByKSH("朱晟豪", "22332504150334"))
	fmt.Println(verifyByKSH("蒋涵羽", "22338801150225"))
}
