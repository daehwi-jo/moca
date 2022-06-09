package sms

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"math/rand"
	"mocaApi/src/controller/cls"
	"net/http"
	"strconv"
	"strings"
	"time"

	smssql "mocaApi/query/sms"
	"mocaApi/src/controller"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

type SmsAuthResult struct {
	Access_token     string `json:"access_token"` //
	Expires_in       string `json:"expires_in"`   //
	Scope            string `json:"scope"`        //
	Create_on        string `json:"create_on"`    //
	Is_expires       string `json:"is_expires"`   //
	Token_type       string `json:"token_type"`   //
	PCD_PAYcode_TYPE string `json:"code"`         //
}

// sms 토큰 인증
func GetSmsAuthToken() (key string) {

	//fname := cls.Cls_conf(os.Args)
	fname := cls.Fname
	smsId, _ := cls.GetTokenValue("SMS.API.ID", fname)
	smsApiKey, _ := cls.GetTokenValue("SMS.API.KEY", fname)
	smsURL := "https://sms.gabia.com/oauth/token"

	encoded := base64.StdEncoding.EncodeToString([]byte(smsId + ":" + smsApiKey))

	payload := strings.NewReader("grant_type=client_credentials")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", smsURL, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", encoded)

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var authResult SmsAuthResult
	err = json.Unmarshal(body, &authResult)

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	key = authResult.Access_token

	return key

}

func SendSms(c echo.Context) error {

	// sms 토큰 get
	params := cls.GetParamJsonMap(c)
	callSendSms(params)

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)
}

func callSendSms(params map[string]string) {

	// sms 토큰 get
	authToken := GetSmsAuthToken()
	// 전송 시작
	//fname := cls.Cls_conf(os.Args)
	fname := cls.Fname
	smsId, _ := cls.GetTokenValue("SMS.API.ID", fname)
	smsApiCallBack, _ := cls.GetTokenValue("SMS.API.CALLBACK", fname)
	sendMsg := params["sendMsg"]
	sendPhone := params["sendPhone"]

	now := time.Now()
	nanos := now.UnixNano()
	millis := strconv.FormatInt(nanos/1000000, 10)

	refkey := "RESTAPI" + millis

	encoded := base64.StdEncoding.EncodeToString([]byte(smsId + ":" + authToken))

	payload := strings.NewReader("phone=" + sendPhone + "&callback=" + smsApiCallBack + "&message=" + sendMsg + "&refkey=[[" + refkey + "]]")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", "https://sms.gabia.com/api/send/sms", payload)
	if err != nil {
		//fmt.Println(err)
		lprintf(1, "[ERROR] error sms send : %s\n", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", encoded)

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		lprintf(1, "[ERROR] error sms send : %s\n", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		lprintf(1, "[ERROR] error sms send : %s\n", err)
	}
	lprintf(4, "[INFO] gabia sms resp(%s)\n", string(body))

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = string(body)

}

func smsConfirmNum(size int) string {
	var alpha = "1234567890"

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alpha[r1.Intn(len(alpha))]
	}
	return string(buf)
}

// 문자 인증 요청 // sms api 개발 필요
func SendSmsConfirm(c echo.Context) error {

	dprintf(4, c, "call SendSmsConfirm\n")

	// sms 전송 시작

	// sms api 개발 필요
	// sms 전송 끝

	params := cls.GetParamJsonMap(c)

	stype := params["stype"]

	//테스트면
	confirmNum := "123456"

	//리얼일경우
	if stype == "r" {
		confirmNum = smsConfirmNum(6)
	}

	//if len(os.Getenv("SERVER_TYPE")) > 0{
	//confirmNum = smsConfirmNum(6)
	//}

	params["confirmDiv"] = "1"
	params["confirmNum"] = confirmNum

	if stype == "r" {
		params["sendPhone"] = params["telNum"]
		params["sendMsg"] = "달아요 인증번호 [ " + confirmNum + " ]"
		callSendSms(params)
	}

	resultData, err := cls.GetSelectData(smssql.SelectConfirmCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {

		selectQuery, err := cls.GetQueryJson(smssql.InsertConfirmData, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("97", "query parameter fail"))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(selectQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}

	} else {

		selectQuery, err := cls.GetQueryJson(smssql.UpdateConfirmDataReset, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("97", "query parameter fail"))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(selectQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}

	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

// 문자 인증 확인
func ConfirmCheck(c echo.Context) error {

	dprintf(4, c, "call ConfirmCheck\n")
	params := cls.GetParamJsonMap(c)
	params["confirmDiv"] = "1"

	resultData, err := cls.GetSelectData(smssql.SelectConfirmCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "응답 실패"))
	}

	SND_DTM := resultData[0]["SND_DTM"]
	NOW_DATE := resultData[0]["NOW_DATE"]

	//시간 체크 5분
	if SND_DTM < NOW_DATE {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "시간초과"))
	}

	CONFIRM_NUM := resultData[0]["CONFIRM_NUM"]
	smsNum := params["smsNum"]

	//인증문자번호 비교
	if strings.Compare(CONFIRM_NUM, smsNum) == 0 {

		selectQuery, err := cls.GetQueryJson(smssql.UpdateConfirmData, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "query parameter fail"))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(selectQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	} else {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "인증 번호를 확인해주세요."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}
