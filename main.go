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
	"path"
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

func getRedirectURL(resp *http.Response) string {
	p := resp.Header.Get("location")
	if strings.HasPrefix(p, "/") {
		p = path.Clean(p)
	}
	if strings.HasPrefix(p, ".") {
		p = path.Clean(path.Join(path.Dir(resp.Request.URL.Path), p))
	}
	return resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + p
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
	client.Jar.SetCookies(resp.Request.URL, resp.Cookies())
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
	case 301, 302:
	case 500:
		//500可能也登录成功了
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
	//redirect to /oauth/authorize
	u := getRedirectURL(resp)
	err = retry(func() (err error) {
		resp, err = client.Get(u)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 301, 302:
	default:
		panic(fmt.Sprint("GET ", u, " returns ", resp.Status))
	}
	//redirect to selfreport.shu.edu.cn/LoginSSO.aspx and set cookies
	u = getRedirectURL(resp)
	err = retry(func() (err error) {
		resp, err = client.Get(u)
		return err
	}, 5)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 301, 302:
	default:
		panic(fmt.Sprint("GET ", u, " returns ", resp.Status))
	}
	client.Jar.SetCookies(resp.Request.URL, resp.Cookies())
}

func sendMail(to, content string) (err error) {
	//定义邮箱服务器连接信息，如果是阿里邮箱 pass填密码，qq邮箱填授权码
	mailConn := map[string]string{
		"user": "dayreport@mzz.pub",
		"pass": "GoodGoodStudyDayDayReport!",
		"host": "smtp.mxhichina.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()
	m.SetHeader("From", "<"+mailConn["user"]+">") //这种方式可以添加别名，即“XD Game”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	m.SetHeader("To", to)                         //发送给多个用户
	m.SetHeader("Subject", "每日一报")                //设置邮件主题
	m.SetBody("text/plain", content)              //设置邮件正文
	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	err = d.DialAndSend(m)
	return err
}

func getViewParam(body io.Reader) map[string]string {
	rand.Seed(time.Now().UnixNano())
	doc, _ := goquery.NewDocumentFromReader(body)
	html, _ := doc.Html()
	zx := regexp.MustCompile(`f7_state={.+?"SelectedValue":"(.+?)"`)
	gn := regexp.MustCompile(`f8_state={.+?"SelectedValue":"(.+?)"`)
	sz := regexp.MustCompile(`f9_state={.+?"SelectedValue":"(.+?)"`)
	sheng := regexp.MustCompile(`f10_state={.+?"SelectedValueArray":\["(.+?)"]`)
	shi := regexp.MustCompile(`f11_state={.+?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`)
	xian := regexp.MustCompile(`f12_state={.+?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`)
	xx := regexp.MustCompile(`f13_state={.+?"Text":"(.+?)"`)
	jc := regexp.MustCompile(`f14_state={.+?"SelectedValueArray":\["(.+?)"]`)
	zxMatch := zx.FindStringSubmatch(html)
	gnMatch := gn.FindStringSubmatch(html)
	szMatch := sz.FindStringSubmatch(html)
	shengMatch := sheng.FindStringSubmatch(html)
	shiMatch := shi.FindStringSubmatch(html)
	xianMatch := xian.FindStringSubmatch(html)
	xxMatch := xx.FindStringSubmatch(html)
	jcMatch := jc.FindStringSubmatch(html)
	F_State := fmt.Sprintf(template, zxMatch[1], gnMatch[1], szMatch[1], shengMatch[1], shiMatch[1], shiMatch[2], xianMatch[1], xianMatch[2], xxMatch[1], jcMatch[1])
	m := map[string]string{
		"F_State":              base64.StdEncoding.EncodeToString([]byte(F_State)),
		"__VIEWSTATE":          doc.Find("#__VIEWSTATE").AttrOr("value", ""),
		"__EVENTTARGET":        "p1$ctl00$btnSubmit",
		"__EVENTARGUMENT":      "",
		"__VIEWSTATEGENERATOR": doc.Find("#__VIEWSTATEGENERATOR").AttrOr("value", ""),
		"p1$ChengNuo":          "p1_ChengNuo",
		"p1$BaoSRQ":            time.Now().Format("2006-01-02"),
		"p1$DangQSTZK":         "良好",
		"p1$TiWen":             fmt.Sprintf("%.1f", float64(362+rand.Int()%7)/10),
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
		"p1$DangQSZD":          szMatch[1],
		"p1$ddlSheng$Value":    shengMatch[1],
		"p1$ddlShi$Value":      shiMatch[2],
		"p1$ddlXian$Value":     xianMatch[2],
		"p1$XiangXDZ":          xxMatch[1],
		"p1$QueZHZJC$Value":    jcMatch[1],
		"p1$QueZHZJC":          "否", //返沪
		"p1$DaoXQLYGJ":         "",  //旅游国家
		"p1$DaoXQLYCS":         "",  //旅游城市
		"p1$Address2":          "中国",
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

func main() {
	cfg := config.Get()
	defer func() {
		if e := recover(); e != nil {
			var content string
			switch e := e.(type) {
			case error:
				content = e.Error()
			case string:
				content = e
			case fmt.Stringer:
				content = e.String()
			default:
				content = "未知错误"
			}
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
