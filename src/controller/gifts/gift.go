package gifts

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"hash/crc32"
	giftsql "mocaApi/query/gifts"
	ordersql "mocaApi/query/orders"
	restsql "mocaApi/query/rests"
	"mocaApi/src/controller"
	apiPush "mocaApi/src/controller/api/push"
	"mocaApi/src/controller/cls"
	"net/http"
	"strconv"
	"time"
)


var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf


// 선물 내역
func GetGiftHistory(c echo.Context) error {

	dprintf(4, c, "call GetGiftHistory\n")

	params := cls.GetParamJsonMap(c)

	resultList, err := cls.GetSelectType(giftsql.SelectGiftHist, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] =  []string{}
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}




// 선물 보내기 준비
func GetGiftReady(c echo.Context) error {

	dprintf(4, c, "call GetGiftReady\n")

	params := cls.GetParamJsonMap(c)

	resultData, err := cls.GetSelectType(giftsql.SelectGiftReady, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98","연결 정보가 없습니다."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultData[0]

	return c.JSON(http.StatusOK, m)

}


// 선물 상세
func GetGiftMsg(c echo.Context) error {

	dprintf(4, c, "call GetGiftMsg\n")

	params := cls.GetParamJsonMap(c)

	resultData, err := cls.GetSelectType(giftsql.SelectGiftDesc, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98","선물정보가 없습니다."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultData[0]

	return c.JSON(http.StatusOK, m)

}



// 선물 하기
func SetGiftSend(c echo.Context) error {

	dprintf(4, c, "call SetGiftSend\n")

	params := cls.GetParamJsonMap(c)

	giftSeq, err := cls.GetSelectData(giftsql.SelectGetGiftSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if giftSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선물번호 생성 실패(2)"))
	}

	giftNum := GenerateCoupon(giftSeq[0]["GIFT_SEQ"])


	linkInfo, err := cls.GetSelectData(giftsql.SelectGiftLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "연결정보 DB fail"))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결정보 응답 실패"))
	}

	reqStat :=linkInfo[0]["REQ_STAT"]
	grpPayTy :=linkInfo[0]["GRP_PAY_TY"]
	grpUserId:=linkInfo[0]["GRP_USER_ID"]

	if reqStat !="1"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부연결이 필요합니다."))
	}
	if grpPayTy !="0"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선물하기는 선불장부만 가능합니다."))
	}

	if grpUserId != params["sndUserId"]{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선물하기는 장부장만 가능합니다."))
	}

	prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
	giftAmt, _ := strconv.Atoi(params["giftAmt"])

	if giftAmt > prepaidAmt {
		return c.JSON(http.StatusOK, controller.SetErrResult("01", "충전된 금액이 모자라 결제 화면으로 이동합니다."))
	}

	params["agrmId"] = linkInfo[0]["AGRM_ID"]


	// 선물 처리 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	// 오류 처리
	defer func() {
		if err != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -선물하기(SetGiftSend)  \n")
			tx.Rollback()
		}
	}()

	//충전금 차감

	params["prepaidAmt"] = strconv.Itoa(prepaidAmt - giftAmt)
	LinkUpdateQuery, err := cls.GetQueryJson(restsql.UpdateLinkInfo, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(LinkUpdateQuery)
	dprintf(4, c, "call set Query UpdateLinkInfo (%s)\n", LinkUpdateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", LinkUpdateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	//주문번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["sndUserId"]

	//주문 생성

	params["moid"] = giftNum
	params["userId"] = params["sndUserId"]
	params["orderAmt"] = params["giftAmt"]
	params["creditAmt"] = params["giftAmt"]
	params["discount"] = "0"
	params["orderTy"] = "4"
	params["payTy"] = "0"
	params["qrOrderTy"] = "0"
	params["pointUse"] = "0"

	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}

	_, err = tx.Exec(OrderCreateQuery)
	dprintf(4, c, "call set Query InsertOrder (%s)\n", OrderCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 상세 메뉴 등록

	params["orderSeq"] = "0"
	params["itemNo"] = "0000000000"
	params["itemPrice"] = params["giftAmt"]
	params["itemCount"] = "1"
	params["itemUserId"] = params["sndUserId"]

	OrderDetailCreateQuery, err := cls.GetQueryJson(ordersql.InsertOrderDetail, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}

	_, err = tx.Exec(OrderDetailCreateQuery)
	dprintf(4, c, "call set Query InsertOrderDetail (%s)\n", OrderDetailCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderDetailCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	params["rcvUserId"] = ""
	params["giftOwnerDiv"] = "3"
	params["giftMethodDiv"] = "1"

	// 선물 테이블에 등록
	giftCreateQuery, err := cls.SetUpdateParam(giftsql.InsertGift, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(giftCreateQuery)
	dprintf(4, c, "call set Query InsertGift (%s)\n", giftCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", giftCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["giftNum"] = giftNum

	return c.JSON(http.StatusOK, m)

}


func GenerateCoupon(keyNum string) string {
	bizKey := keyNum
	if bizKey == "" {
		return ""
	}

	//bizKey = bizKey[8:]

	t := time.Now().Format("2006-0102 150405")
	crc32q := crc32.MakeTable(0xD5828281)
	f3 := fmt.Sprintf("%.8X", crc32.Checksum([]byte(t[10:12]+t[14:16]+bizKey), crc32q))
	final := fmt.Sprintf("%s-%s-%s-%s", t[5:9], f3[0:4], f3[4:8], t[12:16])

	return final

}



func SetGiftRecv(c echo.Context) error {

	dprintf(4, c, "call SetGiftRecv\n")

	params := cls.GetParamJsonMap(c)

	grpData, err := cls.GetSelectData(giftsql.SelectGiftGrpCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if grpData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}

	grpPayTy :=grpData[0]["GRP_PAY_TY"]
	grpUserId:=grpData[0]["GRP_USER_ID"]

	if grpPayTy !="0"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선물받기는 선불장부만 가능합니다."))
	}
	if grpUserId != params["rcvUserId"]{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선물받기는 장부장만 가능합니다."))
	}


	resultData, err := cls.GetSelectDataRequire(giftsql.SelectGiftRecvInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 선물 번호 입니다."))
	}

	giftStsCd := resultData[0]["GIFT_STS_CD"] //fmt.Sprintf("%v", resultData[0]["GIFT_STS_CD"])
	rcvStstCd := resultData[0]["RCV_STS_CD"]
	giftId := resultData[0]["GIFT_ID"]

	sndUserId := resultData[0]["SND_USER_ID"]
	rcvUserNm := resultData[0]["rcvUserNm"]




	if giftStsCd == "2" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소된 선물입니다."))
	}

	if giftStsCd == "3" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "회수된 선물입니다."))
	}

	if rcvStstCd == "1" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "이미 사용된 선물입니다."))
	}

	if rcvStstCd == "3" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "이미 거절된 선물입니다."))
	}

	// 선물 받기 업데이트

	// 선물 처리 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	// 오류 처리
	defer func() {
		if err != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -선물받기(SetGiftRecv)  \n")
			tx.Rollback()
		}
	}()


	// 선물 테이블에 등록
	params["giftId"] = giftId
	params["rcvStsCd"] = "1"
	giftRecvUpdateQuery, err := cls.SetUpdateParam(giftsql.UpdateGiftInfo, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}

	_, err = tx.Exec(giftRecvUpdateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", giftRecvUpdateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	params["restId"] = resultData[0]["REST_ID"]
	linkInfo, err := cls.GetSelectData(restsql.SelectLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	// 협약 정보 없을시 insert
	if linkInfo == nil {
		params["agrmId"] = params["grpId"] + resultData[0]["REST_ID"]
		params["reqStat"] = "1"
		params["reqTy"] = "1"
		params["reqComment"] = "선물받기요청"
		params["payTy"] = "0"
		params["prepaidAmt"] = resultData[0]["GIFT_AMT"]

		linkInsertQuery, err := cls.SetUpdateParam(restsql.InsertLink, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(linkInsertQuery)
		if err != nil {
			dprintf(1, c, "Query(%s) -> error (%s) \n", linkInsertQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	} else {

		prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
		giftAmt, _ := strconv.Atoi(resultData[0]["GIFT_AMT"])

		params["reqStat"] = "1"
		params["prepaidAmt"] = strconv.Itoa(prepaidAmt + giftAmt)
		params["agrmId"] =linkInfo[0]["AGRM_ID"]

		linkUpdateQuery, err := cls.GetQueryJson(restsql.UpdateLinkInfo, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(linkUpdateQuery)
		if err != nil {
			dprintf(1, c, "Query(%s) -> error (%s) \n", linkUpdateQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}


	// transaction commit
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}



	// 푸쉬 전송 시작
	pushMsg:= rcvUserNm+"님이 선물을 받으셨습니다."
	apiPush.SendPush_Msg_V1("선물",pushMsg,"M","0",sndUserId,"","gift")
	// 푸쉬 전송 완료

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}


func SetGiftCancel(c echo.Context) error {


	dprintf(4, c, "call SetGiftCancel\n")

	params := cls.GetParamJsonMap(c)

	giftData, err := cls.GetSelectData(giftsql.SelectGiftInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if giftData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선물 정보가 없습니다."))
	}

	giftStsCd := giftData[0]["GIFT_STS_CD"]
	rcvStstCd := giftData[0]["RCV_STS_CD"]

	if giftStsCd == "2" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소된 선물입니다."))
	}

	if giftStsCd == "3" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "회수된 선물입니다."))
	}

	if rcvStstCd == "1" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "이미 사용된 선물입니다."))
	}

	params["restId"] = giftData[0]["REST_ID"]
	params["grpId"] = giftData[0]["SND_GRP_ID"]


	LinkData, err := cls.GetSelectData(giftsql.SelectGiftLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if LinkData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결 정보가 없습니다."))
	}



	// 선물 처리 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	// 오류 처리
	defer func() {
		if err != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -선물하기 취소 (SetGiftCancel)  \n")
			tx.Rollback()
		}
	}()

	params["giftStsCd"]="2"
	params["giftId"]=giftData[0]["GIFT_ID"]
	UpdateGiftCancelQuery, err := cls.SetUpdateParam(giftsql.UpdateGiftCancel, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(UpdateGiftCancelQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", UpdateGiftCancelQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	// 취소 금액 복원
	prepaidAmt, _ := strconv.Atoi(LinkData[0]["PREPAID_AMT"])
	giftAmt, _ := strconv.Atoi(giftData[0]["GIFT_AMT"])
	params["prepaidAmt"] = strconv.Itoa(prepaidAmt + giftAmt)
	params["agrmId"] = LinkData[0]["AGRM_ID"]


	UpdateLinkQuery, err := cls.SetUpdateParam(restsql.UpdateLinkInfo, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(UpdateLinkQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", UpdateLinkQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}



	params["orderNo"] = giftData[0]["ORDER_NO"]
	params["orderStat"] ="21"
	UpdateOrder, err := cls.SetUpdateParam(ordersql.UpdateOrder, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(UpdateOrder)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", UpdateOrder, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}



	// transaction commit
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}
