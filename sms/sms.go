package sms

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
	"visitw_backend/misc"
)

const (
	charset = "0123456789"
)

var (
	acc        = misc.Config.Sms.Account
	pwd        = misc.Config.Sms.Password
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type (
	VerifyInfo struct {
		Code     string
		ExprTime time.Time
	}
)

func init() {

	fmt.Println("sms loading...")
}

func RandCode(_len int) string {
	b := make([]byte, _len)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SendVerifyCode(phone, code string) {
	cc := `歡迎使用Visitw服務，您的驗證碼為：` + code

	sms := "http://api.twsms.com/json/sms_send.php" +
		"?username=" + acc +
		"&password=" + pwd +
		"&mobile=" + phone +
		"&message=" + url.QueryEscape(cc)
	res, err := http.Get(sms)
	if err != nil {
		log.Fatal(err)
	}

	//http.PostForm("", url.Values{})
	//defer res.Body.Close()

	sitemap, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("%s", sitemap)
	s := string(sitemap)

	fmt.Println(s)
}
