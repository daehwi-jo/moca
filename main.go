package main

import (
	echopprof "github.com/hiko1129/echo-pprof"
	"net/http"
	"os"

	// MocaApp "daragoApi/mocaapi/appapi"
	ApiSrc "mocaApi/src"

	"mocaApi/src/controller/cls"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.elastic.co/apm/module/apmechov4"
	commons "mocaApi/src/controller"
)

var lprintf func(int, string, ...interface{}) = cls.Lprintf

func main() {

	fname := cls.Cls_conf(os.Args)
	lprintf(4, "fname : %s\n", fname)

	domains := cls.WebConf(fname)

	if cls.Db_conf(fname) < 0 {
		lprintf(4, "DataBase not setting \n")
	}

	if commons.Benepicon_conf(fname) < 0 {
		lprintf(4, "Benepicon not setting \n")
	}
	if commons.Wincube_conf(fname) < 0 {
		lprintf(4, "Wincube not setting \n")
	}

	if commons.Tpay_conf(fname) < 0 {
		lprintf(4, "Tpay_conf not setting \n")
	}

	go cls.LogCollect("mocaapi", fname)

	// single domain
	if len(domains) == 1 {
		e := echo.New()
		e.Use(apmechov4.Middleware())
		e.Static("/public", "src/public")
		e.Static("/SharedStorage", "/app/SharedStorage")

		// http://localhost:7788/debug/pprof/
		echopprof.Wrap(e)

		/*
			e.Renderer = echotemplate.New(echotemplate.TemplateConfig{
				Root:         "src/templates",
				Extension:    ".htm",
				Master:       "master",
				Partials:     []string{"/master"},
				DisableCache: true,
				Delims:       echotemplate.Delims{Left: "[[", Right: "]]"},
			})
		*/

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			//	AllowOrigins: []string{"https://172.30.1.22", "https://172.30.1.22:8080"},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		}))

		// Web page Service
		e = ApiSrc.SvcSetting(e, fname)
		// e = MocaApp.SvcSetting(e, fname)
		// start Web Server
		domains[0].EchoData = e
		cls.StartDomain(domains)
		return
	}

	for i := 0; i < len(domains); i++ {
		cls.Lprintf(4, "domains : %s\n", domains[i].Domain)
	}
	cls.StartDomain(domains)
}
