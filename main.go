package main

import (
	"auto-selfreport/config"
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var client *http.Client

func init() {
	client = http.DefaultClient
	client.Jar, _ = cookiejar.New(nil)
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func retry(f func() (err error), maxTimes int) (err error) {
	t := 0
	for {
		t++
		err = f()
		if err != nil {
			if t >= maxTimes {
				return
			}
		} else {
			return nil
		}
	}
}

func login(username, password string) {
	var resp *http.Response
	err := retry(func() (err error) {
		req, _ := http.NewRequest("GET", loginURL, nil)
		req.Header.Set("Referer", homeURL)
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Host", req.URL.Host)
		resp, err = client.Do(req)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//client.Jar.SetCookies(resp.Request.URL, resp.Cookies())
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("login_submit", "登录")
	err = retry(func() (err error) {
		req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
		req.Header.Set("Referer", loginURL)
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Host", req.URL.Host)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err = client.Do(req)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 500:
		//由于没有cookie，500可能登录成功了
		err = retry(func() (err error) {
			resp, err = client.Get(homeURL)
			return err
		}, 5)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return
		}
	case 200:
		panic("用户名或密码错误")
	default:
		panic(fmt.Sprint("POST ", loginURL, " returns ", resp.Status))
	}
}

func sendMail(to, content string) (err error) {
	if len(strings.trimSpace(to)) == 0{
		return
	}
	inf := map[string]string{
		"user": "dayreport@mzz.pub",
		"pass": "GoodGoodStudyDayDayReport!",
		"host": "smtp.mxhichina.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(inf["port"])
	m := gomail.NewMessage()
	m.SetHeader("From", inf["user"])
	m.SetHeader("To", to)
	m.SetHeader("Subject", "每日一报")
	m.SetBody("text/plain", content)
	d := gomail.NewDialer(inf["host"], port, inf["user"], inf["pass"])
	err = d.DialAndSend(m)
	return err
}

func getViewParam(body io.Reader) map[string]string {
	rand.Seed(time.Now().UnixNano())
	doc, _ := goquery.NewDocumentFromReader(body)
	html, _ := doc.Html()
	zxMatch := regexp.MustCompile(`f7_state={.+?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	gnMatch := regexp.MustCompile(`f8_state={.+?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	//szMatch := regexp.MustCompile(`f9_state={.+?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	shengMatch := regexp.MustCompile(`f9_state={.+?"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	shiMatch := regexp.MustCompile(`f10_state={.+?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	xianMatch := regexp.MustCompile(`f11_state={.+?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	xxMatch := regexp.MustCompile(`f12_state={.+?"Text":"(.+?)"`).FindStringSubmatch(html)
	jcMatch := regexp.MustCompile(`f13_state={.+?"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	date := time.Now().Format("2006-01-02")
	F_State := fmt.Sprintf(template, date, zxMatch[1], gnMatch[1], shengMatch[1], shiMatch[1], shiMatch[2], xianMatch[1], xianMatch[2], xxMatch[1], jcMatch[1])
	m := map[string]string{
		"F_State":              base64.StdEncoding.EncodeToString([]byte(F_State)),
		"__VIEWSTATE":          doc.Find("#__VIEWSTATE").AttrOr("value", ""),
		"__EVENTTARGET":        "p1$ctl00$btnSubmit",
		"__EVENTARGUMENT":      "",
		"__VIEWSTATEGENERATOR": doc.Find("#__VIEWSTATEGENERATOR").AttrOr("value", ""),
		"p1$ChengNuo":          "p1_ChengNuo",
		"p1$BaoSRQ":            date,
		"p1$DangQSTZK":         "良好",
		"p1$TiWen":             fmt.Sprintf("%.1f", float64(362+rand.Int()%5)/10),
		"F_TARGET":             "p1_ctl00_btnSubmit",
		"p1_Collapsed":         "false",
		"p1$CengFWH_RiQi":      "",
		"p1$CengFWH_BeiZhu":    "",
		"p1$JieChu_RiQi":       "",
		"p1$JieChu_BeiZhu":     "",
		"p1$TuJWH_RiQi":        "",
		"p1$TuJWH_BeiZhu":      "",
		"p1$JiaRen_BeiZhu":     "",
		"p1$ZaiXiao":           zxMatch[1],
		"p1$GuoNei":            gnMatch[1],
		//"p1$DangQSZD":          szMatch[1],
		"p1$ddlSheng$Value": shengMatch[1],
		"p1$ddlShi$Value":   shiMatch[2],
		"p1$ddlXian$Value":  xianMatch[2],
		"p1$XiangXDZ":       xxMatch[1],
		"p1$QueZHZJC$Value": jcMatch[1],
		"p1$QueZHZJC":       "否", //返沪
		"p1$DaoXQLYGJ":      "",  //旅游国家
		"p1$DaoXQLYCS":      "",  //旅游城市
		"p1$Address2":       "中国",
	}

	return m

}

func dayReport() (msg string) {
	var resp *http.Response
	err := retry(func() (err error) {
		resp, err = client.Get(dayReportURL)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	m := getViewParam(resp.Body)
	data := url.Values{}
	for k, v := range m {
		data.Set(k, v)
	}
	err = retry(func() (err error) {
		req, _ := http.NewRequest("POST", dayReportURL, strings.NewReader(data.Encode()))
		req.Header.Set("Referer", dayReportURL)
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Host", req.URL.Host)
		req.Header.Set("Origin", req.URL.Scheme+"://"+req.URL.Host)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("X-FineUI-Ajax", "true")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		resp, err = client.Do(req)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	s := string(b)
	left := strings.Index(s, "alert(")
	if left >= 0 {
		s = s[left:]
		right := strings.Index(s, ");")
		if right >= 0 {
			s = s[:right]
		}
	}
	return s
}

func handleRecover(msg interface{}) {
	var content string
	switch msg := msg.(type) {
	case error:
		content = msg.Error()
	case string:
		content = msg
	case fmt.Stringer:
		content = msg.String()
	default:
		content = "未知错误"
	}
	cfg := config.Get()
	err := retry(func() (err error) {
		err = sendMail(cfg.Email, content)
		return
	}, 3)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(debug.Stack()))
	log.Println(content)
}

func main() {
	cfg := config.Get()
	defer func() {
		if msg := recover(); msg != nil {
			handleRecover(msg)
		}
	}()
	login(cfg.Username, cfg.Password)
	msg := dayReport()
	log.Println(msg)
	err := retry(func() (err error) {
		err = sendMail(cfg.Email, msg)
		return
	}, 3)
	if err != nil {
		panic(err)
	}
}
