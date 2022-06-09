package cls

import (
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"strconv"
	"strings"
	"time"
)

var FluentdChan chan Fluentd

type Fluentd struct {
	Query      string
	ResultCode string
	ResultMsg  string
}

func Fprint(code, msg string) {
	var f Fluentd
	f.ResultCode = code
	f.ResultMsg = msg

	if FluentdChan != nil {
		FluentdChan <- f
	}

	return
}

func LogCollect(appName string, fname string) {

	var fPort, fHost string

	f, cErr := GetTokenValues("FLUENTD_INFO", fname, ",")
	if cErr != CONF_OK || len(f) != 2 {
		fHost = "localhost"
		fPort = "24224"
	} else {
		fHost = f[1]
		fPort = f[0]
	}

	fmt.Println("[INFO] FLUENTD HOST : ", fHost)
	fmt.Println("[INFO] FLUENTD PORT : ", fPort)

	port, err := strconv.Atoi(fPort)
	if port == 0 || err != nil {
		port = 24224
	}

	FluentdChan = make(chan Fluentd)

	for {
		select {
		// channel에 msg 전송 요청이 들어올 경우
		case f := <-FluentdChan:
			if fHost != "localhost" {

				var data = map[string]string{
					"name":       appName,
					"query":      f.Query,
					"resultCode": f.ResultCode,
					"resultMsg":  f.ResultMsg,
					"kst":        time.Now().Format("2006-01-02 15:04:05"),
				}

				if len(f.ResultMsg) > 0 {
					/*
						msg := strings.ReplaceAll(f.ResultMsg, "\n", "")
						msg = strings.ReplaceAll(msg, "\t", "")
						msg = strings.ReplaceAll(msg, "( ", "(")
						msg = strings.ReplaceAll(msg, "(  ", "(")
						msg = strings.ReplaceAll(msg, "((", "(")
						fmt.Printf(fmt.Sprintf("WARN:%s\n", strings.TrimSpace(msg)))
					*/

					fmt.Println(strings.TrimSpace(f.ResultMsg))
				}

				go sendFluentd(appName, fHost, port, data)
			}
		default:
			time.Sleep(2 * time.Millisecond)
			//goto EXIST
		}
	}
	//EXIST:
	//fmt.Println("routine finish")
}

func sendFluentd(appName, fHost string, fPort int, data map[string]string) int {

	logger, err := fluent.New(fluent.Config{FluentPort: fPort, FluentHost: fHost})
	if err != nil {
		return -1
	}
	defer logger.Close()

	tag := fmt.Sprintf("darayo.%s", appName)

	err = logger.Post(tag, data)
	// err := logger.PostWithTime(tag, time.Now(), data)
	if err != nil {
		return -1
	}

	return 1
}
