package bot

import (
	"github.com/Mrs4s/MiraiGo/client"
	_ "unsafe"
)

//go:linkname GetCookiesWithDomain github.com/Mrs4s/MiraiGo/client.(*QQClient).getCookiesWithDomain
func GetCookiesWithDomain(_ *client.QQClient, domain string) string
