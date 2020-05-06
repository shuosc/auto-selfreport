package main

const (
	loginURL     = "https://newsso.shu.edu.cn/login"
	homeURL      = "https://selfreport.shu.edu.cn/Default.aspx"
	dayReportURL = "https://selfreport.shu.edu.cn/DayReport.aspx"
	historyURL   = "https://selfreport.shu.edu.cn/ViewDayReport.aspx?day=%v"
)
type Area uint
const (
	AreaDefault Area = iota
	AreaShanghai
	AreaGuowai
)
