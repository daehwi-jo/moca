package controller

import (
	"fmt"
	"mocaApi/src/controller/cls"
	"strconv"
	"strings"
)

var RedisAddr []string


func SetErrResult(code, msg string) ResponseHeader {
	var resCode ResponseHeader
	resCode.ResultCode = code
	resCode.ResultMsg = msg

	cls.Fprint(code, msg)

	return resCode
}

func SetResult(code string, data []byte) Response {
	var resCode Response
	resCode.ResultCode = code
	resCode.ResultMsg = "성공"
	resCode.ResultData = data
	return resCode
}

/*
func getPage(rowNum, sPage, lPage, cPage int) Page {
	var pageInfo Page
	pageInfo.TotalBlock = 1
	pageInfo.StartPage = sPage
	pageInfo.LastPage = lPage
	pageInfo.CurrentBlock = 1
	pageInfo.TotalPage = 1
	pageInfo.NextPage = 0
	pageInfo.PrevPage = 0
	pageInfo.TotalCount = rowNum
	pageInfo.CurrentPage = cPage

	var pageInfoList PageList
	pageInfoList.PageNoText = cPage
	pageInfoList.PageNo = cPage
	pageInfoList.ClassName = "on"

	pageInfo.PagingList = append(pageInfo.PagingList, pageInfoList)

	return pageInfo
}

*/

// common
type ResponseHeader struct {
	ResultCode string `json:"resultCode"` // result code
	ResultMsg  string `json:"resultMsg"`  // result msg

}

// common
type Response struct {
	ResultCode string      `json:"resultCode"` // result
	ResultMsg  string      `json:"resultMsg"`  // result code
	ResultData interface{} `json:"resultData"` // result data
}

type PageList struct {
	PageNoText int    `json:"pageNoText"`
	PageNo     int    `json:"pageNo"`
	ClassName  string `json:"className"`
}

type UseHistory struct {
	UserId    string `json:"UserId"`    // user id
	UserNm    string `json:"UserNm"`    // user name
	RestNm    string `json:"RestNm"`    // 가맹점 이름
	ItemNm    string `json:"ItemNm"`    // 구매 아이템 이름
	OrderDate string `json:"OrderDate"` // 주문날자시간
	UseAmt    string `json:"UseAmt"`    // 주문 금액
}


var BENEPICON_URL string
var BENEPICON_CAMP_ID string
var BENEPICON_STATUS_URL string
var BENEPICON_CANCEL_URL string
var BENEPICON_ORDER_URL string

// 베네피콘 설정
func Benepicon_conf(fname string) int {
	lprintf(4, "[INFO] Benepicon conf start (%s)\n", fname)

	BENEPICON_URL, _ = cls.GetTokenValue("BENEPICON.URL", fname)
	BENEPICON_CAMP_ID, _ = cls.GetTokenValue("BENEPICON.CAMP_ID", fname)
	BENEPICON_STATUS_URL, _ = cls.GetTokenValue("BENEPICON.STATUS_URL", fname)
	BENEPICON_CANCEL_URL, _ = cls.GetTokenValue("BENEPICON.CANCEL_URL", fname)
	BENEPICON_ORDER_URL, _ = cls.GetTokenValue("BENEPICON.ORDER_URL", fname)

	return 0
}


var WINCUBE_URL string
var WINCUBE_MEDIA_URL string
var WINCUBE_MDCODE string
var WINCUBE_AUTKEY string


// 윈큐브 설정
func Wincube_conf(fname string) int {
	lprintf(4, "[INFO] Wincube conf start (%s)\n", fname)

	v,r := cls.GetTokenValue("REDIS_INFO", fname)
	if r != cls.CONF_ERR {
		rCnt, err := strconv.Atoi(v)
		if err == nil && rCnt > 0{
			for i:=0; i<rCnt; i++{
				v,r = cls.GetTokenValue(fmt.Sprintf("REDIS_INF0%d",i), fname)
				if r != cls.CONF_ERR{
					rConfig := strings.Split(v, ",")
					if len(rConfig) == 2{
						RedisAddr = append(RedisAddr, fmt.Sprintf("%s:%s", rConfig[1], rConfig[0]))
					}
				}
			}
		}
	}

	WINCUBE_URL, _ = cls.GetTokenValue("WINCUBE.URL", fname)
	WINCUBE_MEDIA_URL, _ = cls.GetTokenValue("WINCUBE.MEDIA_URL", fname)
	WINCUBE_MDCODE, _ = cls.GetTokenValue("WINCUBE.MDCODE", fname)
	WINCUBE_AUTKEY, _ = cls.GetTokenValue("WINCUBE.AUTKEY", fname)

	return 0
}


var TPAY_MID_SIMPLE_PAY string
var TPAY_SIMPLE_PAY_MERCHANT_KEY string
var TPAY_SIMPLE_PAY_GEN_URL string
var TPAY_SIMPLE_PAY_DEL_URL string
var TPAY_SIMPLE_PAY_PAYMENT_URL string
var TPAY_SIMPLE_PAY_CANCEL_URL string
var TPAY_SIMPLE_PAY_CANCEL_PWD string


// Tpay 간편결제 설정
func Tpay_conf(fname string) int {
	lprintf(4, "[INFO] Tpay_conf start (%s)\n", fname)

	TPAY_MID_SIMPLE_PAY, _ = cls.GetTokenValue("TPAY.MID_SIMPLE_PAY", fname)
	TPAY_SIMPLE_PAY_MERCHANT_KEY, _ = cls.GetTokenValue("TPAY.SIMPLE_PAY_MERCHANT_KEY", fname)
	TPAY_SIMPLE_PAY_GEN_URL, _ = cls.GetTokenValue("TPAY.SIMPLE_PAY_GEN_URL", fname)
	TPAY_SIMPLE_PAY_DEL_URL, _ = cls.GetTokenValue("TPAY.SIMPLE_PAY_DEL_URL", fname)
	TPAY_SIMPLE_PAY_PAYMENT_URL, _ = cls.GetTokenValue("TPAY.SIMPLE_PAY_PAYMENT_URL", fname)
	TPAY_SIMPLE_PAY_CANCEL_URL, _ = cls.GetTokenValue("TPAY.SIMPLE_PAY_CANCEL_URL", fname)
	TPAY_SIMPLE_PAY_CANCEL_PWD, _ = cls.GetTokenValue("TPAY.SIMPLE_PAY_CANCEL_PWD", fname)

	return 0
}

