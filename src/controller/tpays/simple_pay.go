package tpays

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"mocaApi/query/payments"
	"mocaApi/src/controller"
	"mocaApi/src/controller/cls"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TpayGenBillingData struct {
	Result_cd  string `json:"result_cd"`  //
	Result_msg string `json:"result_msg"` //
	Card_token string `json:"card_token"` //
	Card_num   string `json:"card_num"`   //
	Card_code  string `json:"card_code"`  //
	Card_name  string `json:"card_name"`  //
	Tid        string `json:"tid"`        //
	Paid_amt   string `json:"paid_amt"`   //
	Moid       string `json:"moid"`       //
	CancelDate string `json:"CancelDate"` //
	CancelTime string `json:"CancelTime"` //
}

// 간편결제 카드 리스트
func GetBillingCardList(c echo.Context) error {

	dprintf(4, c, "call GetBillingCardList\n")

	params := cls.GetParamJsonMap(c)

	m := make(map[string]interface{})

	myBookList, err := cls.GetSelectType(payments.SelectTpayBillingCardList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if myBookList == nil {
		m["resultList"] = []string{}
	} else {
		m["resultList"] = myBookList
	}

	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

func TpayGenBillikey(c echo.Context) error {

	dprintf(4, c, "call TpayGenBillikey\n")

	params := cls.GetParamJsonMap(c)
	mid := controller.TPAY_MID_SIMPLE_PAY
	merchantKey := controller.TPAY_SIMPLE_PAY_MERCHANT_KEY
	genUrl := controller.TPAY_SIMPLE_PAY_GEN_URL

	new_card_num := params["cardNum"]
	buyer_auth_num := params["buyerAuthNum"]
	card_exp := params["cardExp"]
	card_pwd := params["cardPwd"]
	userId := params["userId"]
	cardType := params["cardType"]

	//card_code:=params["card_code"]

	cardSeq := 1

	resultList, err := cls.GetSelectData(payments.SelectTpayBillingKey, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList != nil {

		for i, _ := range resultList {

			cardNum := resultList[i]["CARD_NUM"]
			cardNum_len := len(resultList[i]["CARD_NUM"])

			if cardNum_len > 6 {
				cardNum_start := cardNum[:4]
				cardNum_end := cardNum[(cardNum_len - 4):cardNum_len]

				newCardNum_len := len(new_card_num)
				newCardNum_start := new_card_num[:4]
				newCardNum_end := new_card_num[(newCardNum_len - 4):newCardNum_len]

				if cardNum_start == newCardNum_start && cardNum_end == newCardNum_end {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "이미 등록된 카드입니다."))
				}
			}

			seq, _ := strconv.Atoi(resultList[i]["SEQ"])
			cardSeq = seq + 1
		}

	}

	uValue := url.Values{
		"api_key":        {merchantKey},
		"mid":            {mid},
		"card_num":       {new_card_num},
		"buyer_auth_num": {buyer_auth_num},
		"card_exp":       {card_exp},
		"card_pwd":       {card_pwd},
	}

	resp, err := http.PostForm(genUrl, uValue)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 발급 요청 오류"
		return c.JSON(http.StatusOK, m)
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "전송 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	println(string(respBody))

	var result TpayGenBillingData
	err = json.Unmarshal(respBody, &result)
	if err != nil {

		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "전송 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})

	if result.Result_cd == "0000" {

		insertParam := make(map[string]string)
		insertParam["userId"] = userId
		insertParam["seq"] = strconv.Itoa(cardSeq)
		insertParam["cardToken"] = result.Card_token
		insertParam["cardName"] = result.Card_name
		insertParam["cardNum"] = result.Card_num
		insertParam["cardCode"] = result.Card_code
		insertParam["cardType"] = cardType

		InsertTpayBillingKeyQuery, err := cls.SetUpdateParam(payments.InsertTpayBillingKey, insertParam)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(InsertTpayBillingKeyQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		m["resultCode"] = "00"
		m["resultMsg"] = result.Result_msg
	} else {
		m["resultCode"] = "99"
		m["resultMsg"] = result.Result_msg
	}

	return c.JSON(http.StatusOK, m)
}

// 빌링키 삭제
func TpayDelbillkey(c echo.Context) error {

	dprintf(4, c, "call TpayDelbillkey\n")

	params := cls.GetParamJsonMap(c)

	mid := controller.TPAY_MID_SIMPLE_PAY
	merchantKey := controller.TPAY_SIMPLE_PAY_MERCHANT_KEY
	delUrl := controller.TPAY_SIMPLE_PAY_DEL_URL

	cardInfo, err := cls.GetSelectData(payments.SelectTpayBillingCardInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	card_token := cardInfo[0]["CARD_TOKEN"]

	uValue := url.Values{
		"api_key":    {merchantKey},
		"mid":        {mid},
		"card_token": {card_token},
	}

	resp, err := http.PostForm(delUrl, uValue)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 삭제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 삭제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	println(string(respBody))

	var result TpayGenBillingData
	err = json.Unmarshal(respBody, &result)
	if err != nil {

		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 삭제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})

	if result.Result_cd == "0000" {

		UpdateTpayBillingKeyQuery, err := cls.SetUpdateParam(payments.UpdateTpayBillingKey, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateTpayBillingKeyQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		m["resultCode"] = "00"
		m["resultMsg"] = result.Result_msg
	} else {
		m["resultCode"] = "99"
		m["resultMsg"] = result.Result_msg
	}

	return c.JSON(http.StatusOK, m)
}

// 간편결제 비밀 설정 여부 확인
func TpayBillingPwdYn(c echo.Context) error {

	dprintf(4, c, "call TpayBillingPwdYn\n")

	params := cls.GetParamJsonMap(c)

	cardInfo, err := cls.GetSelectDataRequire(payments.SelectTpayBillingCardPwq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	BILLING_PWD := cardInfo[0]["BILLING_PWD"]

	pwdYn := "Y"

	if BILLING_PWD == "NONE" || BILLING_PWD == "" {
		pwdYn = "N"
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "정상"
	m["pwdYn"] = pwdYn
	return c.JSON(http.StatusOK, m)
}

// 빌링키 비밀번호 확인
func TpayBillingPwdChk(c echo.Context) error {

	dprintf(4, c, "call TpayBillingPwdChk\n")

	params := cls.GetParamJsonMap(c)

	cardInfo, err := cls.GetSelectDataRequire(payments.SelectTpayBillingCardPwq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	BILLING_PWD := cardInfo[0]["BILLING_PWD"]

	if BILLING_PWD == "NONE" || BILLING_PWD == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "결제 비밀번호가 없습니다."))
	}

	if BILLING_PWD != params["billPwd"] {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "비밀번호를 확인해주세요."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "정상 처리 "
	return c.JSON(http.StatusOK, m)
}

// 빌링키 비밀번호 등록 및 수정
func TpayBillingPwd(c echo.Context) error {

	dprintf(4, c, "call TpayBillingPwd\n")

	params := cls.GetParamJsonMap(c)

	UpdateTpayBillingPwdQuery, err := cls.SetUpdateParam(payments.UpdateTpayBillingPwd, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdateTpayBillingPwdQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "정상 처리 "

	return c.JSON(http.StatusOK, m)
}

// 빌링키 결제
func TpayBillingPay(c echo.Context) error {

	dprintf(4, c, "call TpayBillingPay\n")

	params := cls.GetParamJsonMap(c)

	mid := controller.TPAY_MID_SIMPLE_PAY
	merchantKey := controller.TPAY_SIMPLE_PAY_MERCHANT_KEY
	payUrl := controller.TPAY_SIMPLE_PAY_PAYMENT_URL

	card_token := params["card_token"]
	goods_nm := params["goods_nm"]
	amt := params["amt"]
	moid := params["moid"]
	mall_user_id := params["mall_user_id"]
	buyer_name := params["buyer_name"]
	buyer_tel := params["buyer_tel"]
	buyer_email := params["buyer_email"]
	batch_div := "1"

	uValue := url.Values{
		"api_key":      {merchantKey},
		"mid":          {mid},
		"card_token":   {card_token},
		"goods_nm":     {goods_nm},
		"amt":          {amt},
		"moid":         {moid},
		"mall_user_id": {mall_user_id},
		"buyer_name":   {buyer_name},
		"buyer_tel":    {buyer_tel},
		"buyer_email":  {buyer_email},
		"batch_div":    {batch_div},
	}

	resp, err := http.PostForm(payUrl, uValue)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	println(string(respBody))

	var result TpayGenBillingData
	err = json.Unmarshal(respBody, &result)
	if err != nil {

		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})
	if result.Result_cd != "0000" {
		m["resultCode"] = "99"
		m["resultMsg"] = result.Result_msg
	} else {
		m["resultCode"] = "00"
		m["resultMsg"] = result.Result_msg
		m["tid"] = result.Tid
		m["moid"] = result.Moid
		m["amt"] = result.Paid_amt

		sendResultConfirm(result.Tid, "3001")
	}

	return c.JSON(http.StatusOK, m)
}

func BillingPayCancel(moid string, cancelAmt string, cancelMsg string, tid string) map[string]string {

	cancel_pw := controller.TPAY_SIMPLE_PAY_CANCEL_PWD
	mid := controller.TPAY_MID_SIMPLE_PAY
	merchantKey := controller.TPAY_SIMPLE_PAY_MERCHANT_KEY
	cancelUrl := controller.TPAY_SIMPLE_PAY_CANCEL_URL

	partial_cancel := "0"

	uValue := url.Values{
		"api_key":        {merchantKey},
		"mid":            {mid},
		"moid":           {moid},
		"cancel_pw":      {cancel_pw},
		"cancel_amt":     {cancelAmt},
		"cancel_msg":     {cancelMsg},
		"partial_cancel": {partial_cancel},
		"tid":            {tid},
	}

	resp, err := http.PostForm(cancelUrl, uValue)
	if err != nil {
		m := make(map[string]string)
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return m
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m := make(map[string]string)
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return m
	}

	println(string(respBody))

	var result TpayRecvData
	err = json.Unmarshal(respBody, &result)
	if err != nil {

		m := make(map[string]string)
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return m
	}

	m := make(map[string]string)
	if result.Result_cd != "2001" {
		m["resultCode"] = "99"
		m["resultMsg"] = result.Result_msg
	} else {
		m["resultCode"] = "00"
		m["resultMsg"] = result.Result_msg
		m["Canceltid"] = result.Tid
		m["CancelDate"] = result.CancelDate
		m["CancelTime"] = result.CancelTime

		sendCancelResultConfirmNew(merchantKey, mid, result.Tid, "3001")
	}
	return m
}

// 빌링키 결제 취소
func TpayBillingPayCancel(c echo.Context) error {

	dprintf(4, c, "call TpayBillingPayCancel\n")

	params := cls.GetParamJsonMap(c)

	cancel_pw := controller.TPAY_SIMPLE_PAY_CANCEL_PWD
	mid := controller.TPAY_MID_SIMPLE_PAY
	merchantKey := controller.TPAY_SIMPLE_PAY_MERCHANT_KEY
	cancelUrl := controller.TPAY_SIMPLE_PAY_CANCEL_URL

	cancel_amt := params["cancel_amt"]
	moid := params["moid"]
	cancel_msg := params["cancel_msg"]
	partial_cancel := "0"
	tid := params["tid"]

	uValue := url.Values{
		"api_key":        {merchantKey},
		"mid":            {mid},
		"moid":           {moid},
		"cancel_pw":      {cancel_pw},
		"cancel_amt":     {cancel_amt},
		"cancel_msg":     {cancel_msg},
		"partial_cancel": {partial_cancel},
		"tid":            {tid},
	}

	resp, err := http.PostForm(cancelUrl, uValue)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	println(string(respBody))

	var result TpayRecvData
	err = json.Unmarshal(respBody, &result)
	if err != nil {

		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "Tpay 빌링키 결제 요청 오류"
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})
	if result.Result_cd != "2001" {
		m["resultCode"] = "99"
		m["resultMsg"] = result.Result_msg
	} else {
		m["resultCode"] = "00"
		m["resultMsg"] = result.Result_msg
		m["tid"] = result.Tid
		m["CancelDate"] = result.CancelDate
		m["CancelTime"] = result.CancelTime

		sendCancelResultConfirmNew(merchantKey, mid, result.Tid, "3001")
	}

	return c.JSON(http.StatusOK, m)
}

func sendResultConfirm(tid string, result string) {

	payload := strings.NewReader("tid=" + tid + "&result=" + result)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", "https://webtx.tpay.co.kr/resultConfirm", payload)
	if err != nil {
		lprintf(1, "[ERROR] Tpay resultConfirm Send1 : %s\n", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		lprintf(1, "[ERROR] Tpay resultConfirm Send2 : %s\n", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		lprintf(1, "[ERROR] Tpay resultConfirm Send3 : %s\n", err)
	}
	lprintf(4, "[INFO] Tpay resultConfirm resp(%s)\n", string(body))
	println(string(body))

}

func sendCancelResultConfirmNew(apikey string, mid string, tid string, result string) {

	payload := strings.NewReader("api_key=" + apikey + "&mid=" + mid + "&tid=" + tid + "&result_code=" + result + "&result_cd=" + result + "&result=" + result)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", "https://webtx.tpay.co.kr/api/v1/result_confirm", payload)
	if err != nil {
		lprintf(1, "[ERROR] Tpay Cancel resultConfirm Send1 : %s\n", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		lprintf(1, "[ERROR] Tpay Cancel resultConfirm Send2 : %s\n", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		lprintf(1, "[ERROR] Tpay Cancel resultConfirm Send3 : %s\n", err)
	}
	lprintf(4, "[INFO] Tpay Cancel resultConfirm resp(%s)\n", string(body))
	println(string(body))

}
