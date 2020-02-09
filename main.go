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
	sz := regexp.MustCompile(`f9_state={.+?"SelectedValue":"(.+?)"`)
	sheng := regexp.MustCompile(`f10_state={.+?"SelectedValueArray":\["(.+?)"]`)
	shi := regexp.MustCompile(`f11_state={.+?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`)
	xian := regexp.MustCompile(`f12_state={.+?"F_Items":(.+?),"SelectedValueArray":\["(.+?)"]`)
	xx := regexp.MustCompile(`f13_state={.+?"Text":"(.+?)"`)
	jc := regexp.MustCompile(`f14_state={.+?"SelectedValueArray":\["(.+?)"]`)
	szMatch := sz.FindStringSubmatch(html)
	shengMatch := sheng.FindStringSubmatch(html)
	shiMatch := shi.FindStringSubmatch(html)
	xianMatch := xian.FindStringSubmatch(html)
	xxMatch := xx.FindStringSubmatch(html)
	jcMatch := jc.FindStringSubmatch(html)
	F_State := fmt.Sprintf(template, szMatch[1], shengMatch[1], shiMatch[1], shiMatch[2], xianMatch[1], xianMatch[2], xxMatch[1], jcMatch[1])
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
		"p1$ZaiXiao":           "不在校",
		"p1$GuoNei":            "国内",
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
	//b, e := base64.StdEncoding.DecodeString(`eyJwMV9CYW9TUlEiOnsiVGV4dCI6IjIwMjAtMDItMDkifSwicDFfRGFuZ1FTVFpLIjp7IkZfSXRlbXMiOltbIuiJr+WlvSIsIuiJr+WlvSIsMV0sWyLkuI3pgIIiLCLkuI3pgIIiLDFdXSwiU2VsZWN0ZWRWYWx1ZSI6IuiJr+WlvSJ9LCJwMV9aaGVuZ1podWFuZyI6eyJIaWRkZW4iOnRydWUsIkZfSXRlbXMiOltbIuaEn+WGkiIsIuaEn+WGkiIsMV0sWyLlkrPll70iLCLlkrPll70iLDFdLFsi5Y+R54OtIiwi5Y+R54OtIiwxXV0sIlNlbGVjdGVkVmFsdWVBcnJheSI6W119LCJwMV9aYWlYaWFvIjp7IkZfSXRlbXMiOltbIuS4jeWcqOagoSIsIuS4jeWcqOagoSIsMV0sWyLlrp3lsbEiLCLlrp3lsbHmoKHljLoiLDFdLFsi5bu26ZW/Iiwi5bu26ZW/5qCh5Yy6IiwxXSxbIuWYieWumiIsIuWYieWumuagoeWMuiIsMV0sWyLmlrDpl7jot68iLCLmlrDpl7jot6/moKHljLoiLDFdXSwiU2VsZWN0ZWRWYWx1ZSI6IuS4jeWcqOagoSJ9LCJwMV9HdW9OZWkiOnsiRl9JdGVtcyI6W1si5Zu95YaFIiwi5Zu95YaFIiwxXSxbIuWbveWkliIsIuWbveWkliIsMV1dLCJTZWxlY3RlZFZhbHVlIjoi5Zu95YaFIn0sInAxX0RhbmdRU1pEIjp7IlJlcXVpcmVkIjp0cnVlLCJTZWxlY3RlZFZhbHVlIjoi5YW25LuWIiwiRl9JdGVtcyI6W1si5LiK5rW3Iiwi5LiK5rW3IiwxXSxbIua5luWMlyIsIua5luWMlyIsMV0sWyLlhbbku5YiLCLlhbbku5YiLDFdXX0sInAxX2RkbFNoZW5nIjp7IkZfSXRlbXMiOltbIi0xIiwi6YCJ5oup55yB5Lu9IiwxLCIiLCIiXSxbIuWMl+S6rCIsIuWMl+S6rCIsMSwiIiwiIl0sWyLlpKnmtKUiLCLlpKnmtKUiLDEsIiIsIiJdLFsi5LiK5rW3Iiwi5LiK5rW3IiwxLCIiLCIiXSxbIumHjeW6hiIsIumHjeW6hiIsMSwiIiwiIl0sWyLmsrPljJciLCLmsrPljJciLDEsIiIsIiJdLFsi5bGx6KW/Iiwi5bGx6KW/IiwxLCIiLCIiXSxbIui+veWugSIsIui+veWugSIsMSwiIiwiIl0sWyLlkInmnpciLCLlkInmnpciLDEsIiIsIiJdLFsi6buR6b6Z5rGfIiwi6buR6b6Z5rGfIiwxLCIiLCIiXSxbIuaxn+iLjyIsIuaxn+iLjyIsMSwiIiwiIl0sWyLmtZnmsZ8iLCLmtZnmsZ8iLDEsIiIsIiJdLFsi5a6J5b69Iiwi5a6J5b69IiwxLCIiLCIiXSxbIuemj+W7uiIsIuemj+W7uiIsMSwiIiwiIl0sWyLmsZ/opb8iLCLmsZ/opb8iLDEsIiIsIiJdLFsi5bGx5LicIiwi5bGx5LicIiwxLCIiLCIiXSxbIuays+WNlyIsIuays+WNlyIsMSwiIiwiIl0sWyLmuZbljJciLCLmuZbljJciLDEsIiIsIiJdLFsi5rmW5Y2XIiwi5rmW5Y2XIiwxLCIiLCIiXSxbIuW5v+S4nCIsIuW5v+S4nCIsMSwiIiwiIl0sWyLmtbfljZciLCLmtbfljZciLDEsIiIsIiJdLFsi5Zub5bedIiwi5Zub5bedIiwxLCIiLCIiXSxbIui0teW3niIsIui0teW3niIsMSwiIiwiIl0sWyLkupHljZciLCLkupHljZciLDEsIiIsIiJdLFsi6ZmV6KW/Iiwi6ZmV6KW/IiwxLCIiLCIiXSxbIueUmOiCgyIsIueUmOiCgyIsMSwiIiwiIl0sWyLpnZLmtbciLCLpnZLmtbciLDEsIiIsIiJdLFsi5YaF6JKZ5Y+kIiwi5YaF6JKZ5Y+kIiwxLCIiLCIiXSxbIuW5v+ilvyIsIuW5v+ilvyIsMSwiIiwiIl0sWyLopb/ol48iLCLopb/ol48iLDEsIiIsIiJdLFsi5a6B5aSPIiwi5a6B5aSPIiwxLCIiLCIiXSxbIuaWsOeWhiIsIuaWsOeWhiIsMSwiIiwiIl0sWyLpppnmuK8iLCLpppnmuK8iLDEsIiIsIiJdLFsi5r6z6ZeoIiwi5r6z6ZeoIiwxLCIiLCIiXSxbIuWPsOa5viIsIuWPsOa5viIsMSwiIiwiIl1dLCJTZWxlY3RlZFZhbHVlQXJyYXkiOlsi5Zub5bedIl19LCJwMV9kZGxTaGkiOnsiRW5hYmxlZCI6dHJ1ZSwiRl9JdGVtcyI6W1siLTEiLCLpgInmi6nluIIiLDEsIiIsIiJdLFsi5oiQ6YO95biCIiwi5oiQ6YO95biCIiwxLCIiLCIiXSxbIuiHqui0oeW4giIsIuiHqui0oeW4giIsMSwiIiwiIl0sWyLmlIDmnp3oirHluIIiLCLmlIDmnp3oirHluIIiLDEsIiIsIiJdLFsi5rO45bee5biCIiwi5rO45bee5biCIiwxLCIiLCIiXSxbIuW+t+mYs+W4giIsIuW+t+mYs+W4giIsMSwiIiwiIl0sWyLnu7XpmLPluIIiLCLnu7XpmLPluIIiLDEsIiIsIiJdLFsi5bm/5YWD5biCIiwi5bm/5YWD5biCIiwxLCIiLCIiXSxbIumBguWugeW4giIsIumBguWugeW4giIsMSwiIiwiIl0sWyLlhoXmsZ/luIIiLCLlhoXmsZ/luIIiLDEsIiIsIiJdLFsi5LmQ5bGx5biCIiwi5LmQ5bGx5biCIiwxLCIiLCIiXSxbIuWNl+WFheW4giIsIuWNl+WFheW4giIsMSwiIiwiIl0sWyLnnInlsbHluIIiLCLnnInlsbHluIIiLDEsIiIsIiJdLFsi5a6c5a6+5biCIiwi5a6c5a6+5biCIiwxLCIiLCIiXSxbIuW5v+WuieW4giIsIuW5v+WuieW4giIsMSwiIiwiIl0sWyLovr7lt57luIIiLCLovr7lt57luIIiLDEsIiIsIiJdLFsi6ZuF5a6J5biCIiwi6ZuF5a6J5biCIiwxLCIiLCIiXSxbIuW3tOS4reW4giIsIuW3tOS4reW4giIsMSwiIiwiIl0sWyLotYTpmLPluIIiLCLotYTpmLPluIIiLDEsIiIsIiJdLFsi6Zi/5Z2d6JeP5peP576M5peP6Ieq5rK75beeIiwi6Zi/5Z2d6JeP5peP576M5peP6Ieq5rK75beeIiwxLCIiLCIiXSxbIueUmOWtnOiXj+aXj+iHquayu+W3niIsIueUmOWtnOiXj+aXj+iHquayu+W3niIsMSwiIiwiIl0sWyLlh4nlsbHlvZ3ml4/oh6rmsrvlt54iLCLlh4nlsbHlvZ3ml4/oh6rmsrvlt54iLDEsIiIsIiJdXSwiU2VsZWN0ZWRWYWx1ZUFycmF5IjpbIuW5v+WFg+W4giJdfSwicDFfZGRsWGlhbiI6eyJFbmFibGVkIjp0cnVlLCJGX0l0ZW1zIjpbWyItMSIsIumAieaLqeWOv+WMuiIsMSwiIiwiIl0sWyLliKnlt57ljLoiLCLliKnlt57ljLoiLDEsIiIsIiJdLFsi5YWD5Z2d5Yy6Iiwi5YWD5Z2d5Yy6IiwxLCIiLCIiXSxbIuacneWkqeWMuiIsIuacneWkqeWMuiIsMSwiIiwiIl0sWyLpnZLlt53ljr8iLCLpnZLlt53ljr8iLDEsIiIsIiJdLFsi5pe66IuN5Y6/Iiwi5pe66IuN5Y6/IiwxLCIiLCIiXSxbIuWJkemYgeWOvyIsIuWJkemYgeWOvyIsMSwiIiwiIl0sWyLoi43muqrljr8iLCLoi43muqrljr8iLDEsIiIsIiJdXSwiU2VsZWN0ZWRWYWx1ZUFycmF5IjpbIuWIqeW3nuWMuiJdfSwicDFfWGlhbmdYRFoiOnsiTGFiZWwiOiLlm73lhoXor6bnu4blnLDlnYAiLCJUZXh0Ijoi5Zub5bed55yB5bm/5YWD5biC5Yip5bee5Yy65a6d6L2u6ZWHIn0sInAxX1F1ZVpIWkpDIjp7IkZfSXRlbXMiOltbIuaYryIsIuaYryIsMSwiIiwiIl0sWyLlkKYiLCLlkKYiLDEsIiIsIiJdXSwiU2VsZWN0ZWRWYWx1ZUFycmF5IjpbIuWQpiJdfSwicDFfQ2VuZ0ZXSCI6eyJMYWJlbCI6IjIwMjDlubQx5pyIMTDml6XlkI7mmK/lkKblnKjmuZbljJfpgJfnlZnov4cifSwicDFfQ2VuZ0ZXSF9SaVFpIjp7IkhpZGRlbiI6dHJ1ZX0sInAxX0NlbmdGV0hfQmVpWmh1Ijp7IkhpZGRlbiI6dHJ1ZX0sInAxX0ppZUNodSI6eyJMYWJlbCI6IjAx5pyIMjbml6Xoh7MwMuaciDA55pel5piv5ZCm5LiO5p2l6Ieq5rmW5YyX5Y+R54Ot5Lq65ZGY5a+G5YiH5o6l6KemIn0sInAxX0ppZUNodV9SaVFpIjp7IkhpZGRlbiI6dHJ1ZX0sInAxX0ppZUNodV9CZWlaaHUiOnsiSGlkZGVuIjp0cnVlfSwicDFfVHVKV0giOnsiTGFiZWwiOiIwMeaciDI25pel6IezMDLmnIgwOeaXpeaYr+WQpuS5mOWdkOWFrOWFseS6pOmAmumAlOW+hOa5luWMlyJ9LCJwMV9UdUpXSF9SaVFpIjp7IkhpZGRlbiI6dHJ1ZX0sInAxX1R1SldIX0JlaVpodSI6eyJIaWRkZW4iOnRydWV9LCJwMV9KaWFSZW4iOnsiTGFiZWwiOiIwMeaciDI25pel6IezMDLmnIgwOeaXpeWutuS6uuaYr+WQpuacieWPkeeDreetieeXh+eKtiJ9LCJwMV9KaWFSZW5fQmVpWmh1Ijp7IkhpZGRlbiI6dHJ1ZX0sInAxIjp7IklGcmFtZUF0dHJpYnV0ZXMiOnt9fX0=`)
	//log.Println(string(b), e)
	//ioutil.WriteFile("test.json", b, os.FileMode(0644))
	//return
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
