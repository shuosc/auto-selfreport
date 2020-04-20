package main

import (
	"auto-selfreport/config"
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
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
	to = strings.TrimSpace(to)
	if to == "" || to == "true" {
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

func getViewParam() map[string]string {
	var resp *http.Response
	err := retry(func() (err error) {
		resp, err = client.Get(dayReportURL)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body := resp.Body
	rand.Seed(time.Now().UnixNano())
	doc, _ := goquery.NewDocumentFromReader(body)
	html, _ := doc.Html()
	zxMatch := regexp.MustCompile(`f8_state={.*?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	gnMatch := regexp.MustCompile(`f14_state={.*?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	// szMatch := regexp.MustCompile(`f9_state={.+?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	shengMatch := regexp.MustCompile(`f16_state={.+?"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	shiMatch := regexp.MustCompile(`f17_state={.*?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	xianMatch := regexp.MustCompile(`f18_state={.*?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	tzMatch := regexp.MustCompile(`f19_state={.*?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	xxMatch := regexp.MustCompile(`f20_state={.*?"Text":"(.+?)"`).FindStringSubmatch(html)
	// jcMatch := regexp.MustCompile(`f15_state={.*?"SelectedValueArray":\["(.+?)"]`).FindStringSubmatch(html)
	ssMatch := regexp.MustCompile(`f43_state={.*?"SelectedValue":"(.+?)"`).FindStringSubmatch(html)
	date := time.Now().Format("2006-01-02")
	var F_State string
	var shanghai bool
	if (len(tzMatch) < 2 || len(ssMatch) < 2) {
		shanghai = false
		F_State = fmt.Sprintf(template_0, date, zxMatch[1], gnMatch[1], shengMatch[1], shiMatch[1], shiMatch[2], xianMatch[1], xianMatch[2], xxMatch[1], "否")
	} else {
		shanghai = true
		F_State = fmt.Sprintf(template_1, date, zxMatch[1], gnMatch[1], shengMatch[1], shiMatch[1], shiMatch[2], xianMatch[1], xianMatch[2], tzMatch[1], xxMatch[1], "否", ssMatch[1])
	}

	err = retry(func() (err error) {
		resp, err = client.Get(dayReportURL)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// body = resp.Body
	// doc, _ = goquery.NewDocumentFromReader(body)
	// html, _ = doc.Html()

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
		"p1$MingTDX":           "不到校",
		"p1$MingTJC":           "否",
		"p1$BanChe_1$Value":    "0",
		"p1$BanChe_1":          "不需要乘班车",
		"p1$BanChe_2$Value":    "0",
		"p1$BanChe_2":          "不需要乘班车",
		"p1$GuoNei":            gnMatch[1],
		"p1$ddlGuoJia$Value":   "-1",
		"p1$ddlGuoJia":         "选择国家",
		//"p1$DangQSZD":          szMatch[1],
		"p1$ddlSheng$Value": shengMatch[1],
		"p1$ddlShi$Value":   shiMatch[2],
		"p1$ddlXian$Value":  xianMatch[2],
		"p1$XiangXDZ":       xxMatch[1],
		"p1$QueZHZJC$Value": jcMatch[1],
		"p1$SuiSM":          "绿色",       // 随申码颜色
		"p1$LvMa14Days":     "是",    // 截止今天是否连续14天健康码为绿色
		"p1$QueZHZJC":       "否", //返沪
		"p1$DangRGL":	     "否", //是否隔离
		"p1$DaoXQLYGJ":      "",  //旅游国家
		"p1$DaoXQLYCS":      "",  //旅游城市
		"p1$Address2":       "中国",
		"p1_SuiSMSM_Collapsed": "false",
	}
	if shanghai {
		m["p1$TongZWDLH"] = tzMatch[1]
		m["p1$SuiSM"] = ssMatch[1]
	}
	return m

}

func dayReport() (msg string) {
	var resp *http.Response
	m := getViewParam()
	data := url.Values{}
	for k, v := range m {
		data.Set(k, v)
	}
	err := retry(func() (err error) {
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
	if (!strings.Contains(s, "提交成功")) {
		panic(s)
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
			os.Exit(1)
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
