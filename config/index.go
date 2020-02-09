package config

import (
	"github.com/stevenroose/gonfig"
	"log"
	"sync"
)

type Params struct {
	Username string `id:"username" short:"u" desc:"学号"`
	Password string `id:"password" short:"p" desc:"密码"`
	Email    string `id:"email" short:"e" desc:"接收邮件的地址"`
}

var params Params

func initFunc() {
	err := gonfig.Load(&params, gonfig.Conf{
		FileDisable: true,
		//FlagIgnoreUnknown: true,
	})
	if params.Username == "" || params.Password == "" {
		log.Fatal("Either USERNAME or PASSWORD can not be empty! See help by command line flag --help.")
	}
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log.Fatal(err)
		}
	}
}

var once sync.Once

func Get() *Params {
	once.Do(initFunc)
	return &params
}
