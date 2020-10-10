package main

import (
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/samuncle-jqk/httpProxyPool/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/samuncle-jqk/httpProxyPool/config"
)

func main() {
	var (
		cfg *config.Config
		err error
	)

	// 加载配置文件
	if cfg, err = config.Init(""); err != nil {
		panic(err)
	}

	// 初始化日志配置
	cfg.InitLog()

	abTest()
}

func abTest() {
	url := viper.GetString("api.url")
	method := viper.GetString("api.method")
	params := viper.GetStringMapString("api.params")
	token := viper.GetStringMapString("api.token")

	concurrency := viper.GetInt("concurrency")

	begin := time.Now()
	logrus.WithFields(logrus.Fields{"begin": begin}).Info("Begin abTest")

	wg := &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go httpRequest(wg, url, method, params, token)
	}
	wg.Wait()

	end := time.Now()
	logrus.WithFields(logrus.Fields{
		"end":         end,
		"duration":    end.Sub(begin).String(),
		"concurrency": concurrency,
	}).Info("End abTest")
}

func httpRequest(wg *sync.WaitGroup, url string, method string, params map[string]string, token map[string]string) {
	defer wg.Done()

	var (
		rsp *resty.Response
		err error
	)
	r := utils.NewRestyRequestChrome(token)
	if strings.ToUpper(method) == "GET" {
		rsp, err = r.SetQueryParams(params).Get(url)
	} else {
		rsp, err = r.SetFormData(params).Post(url)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"url":    url,
			"method": method,
			"params": params,
			"token":  token,
			"t":      time.Now(),
		}).Error("http request fail")
		return
	}

	logrus.WithFields(logrus.Fields{
		"url":      url,
		"method":   method,
		"params":   params,
		"token":    token,
		"status":   rsp.StatusCode(),
		"response": string(rsp.Body()),
		"t":        time.Now(),
	}).Info("http request ok")
}
