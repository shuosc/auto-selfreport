package main

const (
	loginURL     = "https://newsso.shu.edu.cn/login"
	homeURL      = "http://selfreport.shu.edu.cn/Default.aspx"
	dayReportURL = "http://selfreport.shu.edu.cn/DayReport.aspx"
)

//TODO:
const (
	username = ""
	password = ""
	email    = "" //收信地址
)

var info = map[string]string{
	"p1$DangQSZD":       "其他",             //TODO: 当天所在地，必须在选项内，上海、湖北、其他
	"p1$ddlSheng$Value": "四川",             //TODO: 当天所在地，必须在选项内
	"p1$ddlShi$Value":   "广元市",            //TODO: 当天所在地，必须在选项内
	"p1$ddlXian$Value":  "利州区",            //TODO: 当天所在地，必须在选项内
	"p1$XiangXDZ":       "四川省广元市利州区宝轮镇",   //TODO: 国内详细地址
	"p1$QueZHZJC$Value": "否",              //TODO: 是否曾与确诊患者密切接触，必须在选项内，是、否
	"p1$QueZHZJC":       "否",              //TODO: 是否在返校或返沪途中，必须在选项内，是、否
	"p1$DaoXQLYGJ":      "",               //TODO: 到校前旅游过的国家
	"p1$DaoXQLYCS":      "",               //TODO: 到校前旅游过的城市
	"p1$Address2":       "中国四川省成都市锦江区督院街", //TODO: 通过ip地址获取的具体位置，我们可以填一个差不多的具体位置
}
