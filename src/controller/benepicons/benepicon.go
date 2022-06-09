package benepicons

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	ordersql "mocaApi/query/orders"
	"mocaApi/src/controller"
	commons "mocaApi/src/controller"
	"mocaApi/src/controller/cls"
	"net/http"
	"net/url"
)


var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf


type OrderData struct {
	CampId       string `json:"CAMP_ID"`       //
	TrId      	 string `json:"TR_ID"`      //
	ProdId 		 string `json:"PROD_ID"` //
	ProdQty 	 string `json:"PROD_QTY"` //
	RcvrInFO  	 RcvrInFO `json:"RCVR_INFO"` //
	EncYn 		 string `json:"ENC_YN"` //
	CallBack 	 string `json:"CALLBACK"` //
	SendType 	 string `json:"SEND_TYPE"` //
	CpnoReqYn 	 string `json:"CPNO_REQ_YN"` //
}

type RcvrInFO struct {
	RcvrMnd     string `json:"RCVR_MDN"`       //
	UserId      string `json:"USER_ID"`      //
}

type CouponChkData struct {
	CampId       string `json:"CAMP_ID"`       //
	OrdNo      	 string `json:"ORD_NO"`      //
	EncYn 		 string `json:"ENC_YN"` //
	Cpno 	 	 string `json:"CPNO"` //
	TrId	 	 string `json:"TR_ID"` //
}


type CancelData struct {
	CampId       string `json:"CAMP_ID"`       //
	OrdNo      	 string `json:"ORD_NO"`      //
	Cpno 	 	 string `json:"CPNO"` //
	TrId	 	 string `json:"TR_ID"` //
}


// 선물 내역
func GetTest(c echo.Context) error {

	dprintf(4, c, "call GetTest\n")

	//params := cls.GetParamJsonMap(c)

	pURL:="https://nstglsend.gifticon.com:443/b2b/order.gc"

	var reqData OrderData
	reqData.CampId = "M1001658"
	reqData.TrId = "0000003920U0000000543"
	reqData.ProdId = "S0005003"
	reqData.ProdQty="1"
	reqData.RcvrInFO.RcvrMnd="01020047042"
	reqData.RcvrInFO.UserId="U019291010"
	reqData.EncYn="N"
	reqData.CallBack="01020047042"
	reqData.SendType="CAB06"
	reqData.CpnoReqYn="Y"

	pbytes, _ := json.Marshal(reqData)

	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest("POST", pURL, buff)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	str := ""
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		str = string(respBody)
	}

	println(str)


	ss ,_ :=url.QueryUnescape(str)


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = ss

	return c.JSON(http.StatusOK, m)

}



// 기프티콘 주문
func SetBeneficonOrder(c echo.Context) error {

	dprintf(4, c, "call SetBeneficonOrder\n")


	params := cls.GetParamJsonMap(c)

	//orderNo:="000000107101U0000003593"
	//prodId:="S0005003"
	//prodQty:="1"
	//userTel:="01020047042"
	//userId:="U0000003493"

	orderNo:=params["orderNo"]
	prodId:=params["prodId"]
	prodQty:=params["prodQty"]
	userTel:=params["userTel"]
	userId:=params["userId"]



	//9005 상품 미존재
	//9002 시스템 내부 오류
	callresultMap :=CallBenepiconOrder(orderNo, prodId,prodQty,userTel,userId )


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = callresultMap


	return c.JSON(http.StatusOK, m)

}


// 기프티콘 주문
func SetBeneficonCancel(c echo.Context) error {

	dprintf(4, c, "call SetBeneficonCancel\n")


	params := cls.GetParamJsonMap(c)

	//orderNo:="000000107101U0000003593"
	//prodId:="S0005003"
	//prodQty:="1"
	//userTel:="01020047042"
	//userId:="U0000003493"

	ordNo:=params["ordNo"]
	cpNo:=params["cpNo"]
	trId:=params["trId"]




	//9005 상품 미존재
	//9002 시스템 내부 오류
	callresultMap :=CallBenepiconCancel(ordNo,cpNo ,trId)


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = callresultMap


	return c.JSON(http.StatusOK, m)

}


// 기프티콘 취소
func SetGifticonCancel(c echo.Context) error {

	dprintf(4, c, "call SetGifticonCancel\n")
	params := cls.GetParamJsonMap(c)


	couPonInfo,err := cls.GetSelectData(ordersql.SelectCouponInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if couPonInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "취소 가능한 쿠폰이 없습니다."
		return c.JSON(http.StatusOK, m)
	}

	ordNo  :=couPonInfo[0]["ORD_NO"]
	cpNo  :=couPonInfo[0]["CPNO"]
	trId  :=params["orderNo"]

	callresultMap :=CallBenepiconCancel(ordNo,cpNo ,trId)

	resultCd :=callresultMap["RESULT_CD"]

	if resultCd=="0000"{

		UpdateCouponCancelQuery, err := cls.GetQueryJson(ordersql.UpdateCouponCancel, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateCouponCancelQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}else{
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = callresultMap["RESULT_MSG"]
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultMsg"] = callresultMap


	return c.JSON(http.StatusOK, m)

}

// 기프티콘 상태 업데이트
func SetGifticonStatus(c echo.Context) error {

	dprintf(4, c, "call SetGifticonStatus\n")

	params := cls.GetParamJsonMap(c)
	ordNo:=params["ordNo"]
	cpNo:=params["cpNo"]
	orderNo:=params["orderNo"]
	result :=CallBenepiconCheck(ordNo,cpNo,orderNo)


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = result

	return c.JSON(http.StatusOK, m)

}

// 기프티콘 사용체크
func SetGifticonCheck(c echo.Context) error {

	dprintf(4, c, "call SetGifticonCheck\n")

	params := cls.GetParamJsonMap(c)
	//ordNo:=params["ordNo"]
	//cpNo:=params["cpNo"]
	//orderNo:=params["orderNo"]

	couPonInfo,err := cls.GetSelectData(ordersql.SelectCouponCheckList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if couPonInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "미사용 쿠폰이 없습니다."
		return c.JSON(http.StatusOK, m)
	}

	for i, _ := range couPonInfo {

		ordNo:=couPonInfo[i]["ORD_NO"]
		cpNo:=couPonInfo[i]["CPNO"]
		orderNo:=couPonInfo[i]["ORDER_NO"]

		params["ordNo"] =ordNo
		params["cpNo"] =cpNo
		params["orderNo"] =orderNo


		chkResult :=CallBenepiconCheck(ordNo,cpNo,orderNo)

		cpnoStatusCd 		:=chkResult["CPNO_STATUS_CD"]
		params["exchplc"]		=chkResult["EXCHPLC"]
		params["exchcoNm"]		=chkResult["EXCHCO_NM"]
		params["cpnoExchDt"]	=chkResult["CPNO_EXCH_DT"]
		params["cpnoStatus"]	=chkResult["CPNO_STATUS"]
		params["cpnoStatusCd"]	=cpnoStatusCd
		params["balance"]		=chkResult["BALANCE"]



		switch cpnoStatusCd {

		case "IAD10": //IAD10 발행완료
			params["cpStatus"]="0"
			break
		case "IAD03":  //IAD03 쿠폰 사용불가-주문취소
			params["cpStatus"]="1"
			break
		case "IAD05": //쿠폰사용불가_유효기간 경과
			params["cpStatus"]="9"
			break
		case "IAD06": //쿠폰사용불가_환불
			params["cpStatus"]="1"
			break
		case "IAD09": //교환완료
			params["cpStatus"]="2"
			break
		case "IAD08": //다운로드완료
			params["cpStatus"]="0"
			break
		}

		UpdateCouponQuery, err := cls.GetQueryJson(ordersql.UpdateCoupon, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateCouponQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}


	}



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"


	return c.JSON(http.StatusOK, m)

}



func CallBenepiconCheck(ordNo string,cpNo string ,trId string) (map[string]string)  {


	burl:= commons.BENEPICON_URL
	campId := commons.BENEPICON_CAMP_ID
	statusUrl:= commons.BENEPICON_STATUS_URL
	result := make(map[string]string)
	pURL:=burl+statusUrl

	var reqData CouponChkData
	reqData.CampId 	= campId
	reqData.OrdNo	= ordNo
	reqData.EncYn	= "N"
	reqData.Cpno	= cpNo
	reqData.TrId 	= trId

	cls.Lprintf(4, "[INFO] pURL(%s)\n", pURL)
	cls.Lprintf(4, "[INFO] campId(%s)\n", campId)
	cls.Lprintf(4, "[INFO] ordNo(%s)\n", ordNo)
	cls.Lprintf(4, "[INFO] cpNo(%s)\n", cpNo)
	cls.Lprintf(4, "[INFO] trId(%s)\n", trId)

	pbytes, _ := json.Marshal(reqData)

	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest("POST", pURL, buff)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")


	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()


	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	var chkResult ChkResultData
	err = json.Unmarshal(respBody, &chkResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(2)"
		return result
	}


	ResultMsg ,_ :=url.QueryUnescape(chkResult.ResultMsg)

	if chkResult.ResultCd=="0000"{
		result["RESULT_CD"]= chkResult.ResultCd
		result["RESULT_MSG"],_= url.QueryUnescape(chkResult.ResultMsg)
		result["PROD_NM"],_= url.QueryUnescape(chkResult.ProdNm)										// 상품명
		result["PROD_PRC"]= chkResult.ProdPrc 															//상품금액
		result["TR_ID"]= chkResult.TrID 																// 고객사 연동 번호 (Unique ID)
		result["ORD_NO"]= chkResult.OrdInfo[0].OrdNo 													// 기프티콘 주문번호
		result["RCVR_MDN"]= chkResult.OrdInfo[0].RcvrMdn 												// 수신자 전화번호
		result["CPNO"]= chkResult.OrdInfo[0].CpnoInfo[0].Cpno 											// 기프티콘 쿠폰 번호
		result["CPNO_SEQ"]= chkResult.OrdInfo[0].CpnoInfo[0].CpnoSeq 									// 기프티콘 쿠폰 순번
		result["EXCH_FR_DY"]= chkResult.OrdInfo[0].CpnoInfo[0].ExchFrDy 								// 쿠폰 교환 시작일자 (YYYYMMDD)
		result["EXCH_TO_DY"]= chkResult.OrdInfo[0].CpnoInfo[0].ExchToDy 								// 쿠폰 교환 종료일자 (YYYYMMDD)
		result["EXCHPLC"],_= url.QueryUnescape(chkResult.OrdInfo[0].CpnoInfo[0].Exchplc) 				// 교환처
		result["EXCHCO_NM"],_= url.QueryUnescape(chkResult.OrdInfo[0].CpnoInfo[0].ExchcoNm) 			// 교환제휴사명 (쿠폰을 교환한 매장정보)
		result["CPNO_EXCH_DT"]= chkResult.OrdInfo[0].CpnoInfo[0].CpnoExchDt 							// 쿠폰 교환 일시 (YYYYMMDDHH24MISS)
		result["CPNO_STATUS"],_= url.QueryUnescape(chkResult.OrdInfo[0].CpnoInfo[0].CpnoStatus) 		// 쿠폰상태
		result["CPNO_STATUS_CD"],_= url.QueryUnescape(chkResult.OrdInfo[0].CpnoInfo[0].CpnoStatusCd) 	// 쿠폰상태코드
		result["UNIT_PROD_NM"],_= url.QueryUnescape(chkResult.OrdInfo[0].CpnoInfo[0].UnitProdNm) 		// 쿠폰단위상품명
		result["BALANCE"]= chkResult.OrdInfo[0].CpnoInfo[0].Balance  									// 쿠폰잔액 또는 횟수 (차감식 상품인 경우)

	}else{
		result["RESULT_CD"] = chkResult.ResultCd
		result["RESULT_MSG"] =ResultMsg
		return result
	}

	return result

}


func CallBenepiconOrder(orderNo string, prodId string,prodQty string,userTel string,userId string  ) (map[string]string)  {

	lprintf(4, "[INFO] Call CallBenepiconOrder \n")

	burl:= commons.BENEPICON_URL
	campId := commons.BENEPICON_CAMP_ID
	orderUrl:= commons.BENEPICON_ORDER_URL

	result := make(map[string]string)


	pURL:=burl+orderUrl

	var reqData OrderData
	reqData.CampId =campId
	reqData.TrId = orderNo
	reqData.ProdId = prodId
	reqData.ProdQty= prodQty
	reqData.RcvrInFO.RcvrMnd= userTel
	reqData.RcvrInFO.UserId= userId
	reqData.EncYn="N"
	reqData.CallBack="0264500601"
	reqData.SendType="CAB06"
	reqData.CpnoReqYn="Y"

	pbytes, _ := json.Marshal(reqData)

	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest("POST", pURL, buff)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()


	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	var orderResult OrderResultData
	err = json.Unmarshal(respBody, &orderResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(2)"
		return result
	}

	ResultMsg ,_ :=url.QueryUnescape(orderResult.ResultMsg)

	if orderResult.ResultCd=="0000"{
		result["RESULT_CD"] = orderResult.ResultCd
		result["RESULT_MSG"] = ResultMsg
		result["CAMP_ID"] = orderResult.CampID
		result["TR_ID"] = orderResult.TrID
		result["ORD_NO"] = orderResult.OrdInfo[0].OrdNo
		result["RCVR_MDN"] = orderResult.OrdInfo[0].RcvrMdn
		result["CPNO"] = orderResult.OrdInfo[0].CpnoInfo[0].Cpno
		result["CPNO_SEQ"] = orderResult.OrdInfo[0].CpnoInfo[0].CpnoSeq
		result["EXCH_FR_DY"] = orderResult.OrdInfo[0].CpnoInfo[0].ExchFrDy
		result["EXCH_TO_DY"] = orderResult.OrdInfo[0].CpnoInfo[0].ExchToDy
	}else{
		result["RESULT_CD"] = orderResult.ResultCd
		result["RESULT_MSG"] =ResultMsg
		return result
	}

	return result

}

type OrderResultData struct {
	CampID  string `json:"CAMP_ID"`
	OrdInfo []struct {
		CpnoInfo []struct {
			Cpno     string `json:"CPNO"`
			CpnoSeq  string `json:"CPNO_SEQ"`
			ExchFrDy string `json:"EXCH_FR_DY"`
			ExchToDy string `json:"EXCH_TO_DY"`
		} `json:"CPNO_INFO"`
		OrdNo   string `json:"ORD_NO"`
		RcvrMdn string `json:"RCVR_MDN"`
	} `json:"ORD_INFO"`
	ResultCd  string `json:"RESULT_CD"`
	ResultMsg string `json:"RESULT_MSG"`
	TrID      string `json:"TR_ID"`
	OrdResult bool   `json:"ordResult"`
}

type CancelResultData struct {
	ResultCd    string      `json:"RESULT_CD"`
	ResultMsg   string      `json:"RESULT_MSG"`
	OrdNo    	string      `json:"ORD_NO"`
	Cpno   		string      `json:"CPNO"`
	TrID    	string      `json:"TR_ID"`

}
type ChkResultData struct {
	OrdInfo []struct {
		CpnoInfo []struct {
			Balance      string 	 `json:"BALANCE"`
			Cpno         string      `json:"CPNO"`
			CpnoExchDt   string		 `json:"CPNO_EXCH_DT"`
			CpnoSeq      string      `json:"CPNO_SEQ"`
			CpnoStatus   string      `json:"CPNO_STATUS"`
			CpnoStatusCd string      `json:"CPNO_STATUS_CD"`
			ExchcoNm     string 	 `json:"EXCHCO_NM"`
			Exchplc      string      `json:"EXCHPLC"`
			ExchFrDy     string      `json:"EXCH_FR_DY"`
			ExchToDy     string      `json:"EXCH_TO_DY"`
			UnitProdNm   string      `json:"UNIT_PROD_NM"`
		} `json:"CPNO_INFO"`
		OrdNo      string      `json:"ORD_NO"`
		RcvrMdn    string      `json:"RCVR_MDN"`
		ShppAuthNo string	   `json:"SHPP_AUTH_NO"`
		ShppURL    string	   `json:"SHPP_URL"`
	} `json:"ORD_INFO"`
	ProdNm    string      `json:"PROD_NM"`
	ProdPrc   string      `json:"PROD_PRC"`
	ResultCd  string      `json:"RESULT_CD"`
	ResultMsg string      `json:"RESULT_MSG"`
	TrID      string	  `json:"TR_ID"`
}




func CallBenepiconCancel(ordNo string,cpNo string ,trId string) (map[string]string)  {

	lprintf(4, "[INFO] Call BenepiconCancel \n")
	burl:= commons.BENEPICON_URL
	campId := commons.BENEPICON_CAMP_ID
	cancelUrl:= commons.BENEPICON_CANCEL_URL
	result := make(map[string]string)
	pURL:=burl+cancelUrl

	var reqData CancelData
	reqData.CampId 	= campId
	reqData.OrdNo	= ordNo
	reqData.Cpno	= cpNo
	reqData.TrId 	= trId


	pbytes, _ := json.Marshal(reqData)

	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest("POST", pURL, buff)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()


	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	var chkResult CancelResultData
	err = json.Unmarshal(respBody, &chkResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(2)"
		return result
	}
	ResultMsg ,_ :=url.QueryUnescape(chkResult.ResultMsg)

	if chkResult.ResultCd=="0000"{
		result["RESULT_CD"]= chkResult.ResultCd
		result["RESULT_MSG"],_= url.QueryUnescape(chkResult.ResultMsg)
		result["ORD_NO"] = chkResult.OrdNo
		result["CPNO"] = chkResult.Cpno
		result["TR_ID"] = chkResult.TrID
	}else{
		result["RESULT_CD"] = chkResult.ResultCd
		result["RESULT_MSG"] =ResultMsg
		return result
	}

	return result

}
