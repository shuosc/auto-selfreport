package main

const (
	loginURL     = "https://newsso.shu.edu.cn/login"
	homeURL      = "http://selfreport.shu.edu.cn/Default.aspx"
	dayReportURL = "http://selfreport.shu.edu.cn/DayReport.aspx"
	historyURL   = "http://selfreport.shu.edu.cn/ViewDayReport.aspx?day=%v"
)
type Area uint
const (
	AreaDefault Area = iota
	AreaShanghai
	AreaGuowai
)
