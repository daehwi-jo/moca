package orders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	booksql "mocaApi/query/books"
	commons "mocaApi/query/commons"
	ordersql "mocaApi/query/orders"
	paymentsql "mocaApi/query/payments"
	restsql "mocaApi/query/rests"
	usersql "mocaApi/query/users"
	"mocaApi/src/controller"
	apiPush "mocaApi/src/controller/api/push"
	"mocaApi/src/controller/tpays"
	"mocaApi/src/controller/wincubes"
	"os"

	benepicons "mocaApi/src/controller/benepicons"
	"mocaApi/src/controller/cls"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

type empty struct {
}

// 가맹점 카테고리
func GetStoreCategory(c echo.Context) error {

	dprintf(4, c, "call GetStoreCategory\n")

	params := cls.GetParamJsonMap(c)

	m := make(map[string]interface{})

	resultList, err := cls.GetSelectType(ordersql.SelectStoreCategory, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {
		m["resultCode"] = "00"
		m["resultMsg"] = "카테고리 정보가 없습니다."
		m["resultData"] = []string{}
		return c.JSON(http.StatusOK, m)
	}

	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}

// 가맹점 메뉴
func GetStoreMenu(c echo.Context) error {

	dprintf(4, c, "call GetStoreMenu\n")

	params := cls.GetParamJsonMap(c)

	codeId := params["codeId"]
	query := ordersql.SelectStoreMenuList

	if codeId != "" {
		query = ordersql.SelectStoreMenuListSearch
	}

	m := make(map[string]interface{})

	resultList, err := cls.GetSelectType(query, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {
		m["resultCode"] = "00"
		m["resultMsg"] = "메뉴 정보가 없습니다."
		m["resultData"] = []string{}
		return c.JSON(http.StatusOK, m)
	}

	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}

// 가맹점 매뉴 즐겨찾기
func SetStoreMenuFavorite(c echo.Context) error {

	dprintf(4, c, "call SetStoreMenuFavorite\n")

	params := cls.GetParamJsonMap(c)
	menuChk, err := cls.GetSelectDataRequire(restsql.SelectFreqMenuChk, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	itemCnt, _ := strconv.Atoi(menuChk[0]["itemCnt"])

	favoriteYn := params["favoriteYn"]
	var sqlquery string
	if favoriteYn == "Y" {
		sqlquery = restsql.InsertFreqStoreMenu
	} else {
		sqlquery = restsql.DeleteFreqStoreMenu
	}
	//println(favoriteYn)
	if itemCnt > 0 && favoriteYn == "Y" {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		return c.JSON(http.StatusOK, m)
	}

	// 파라메터 맵으로 쿼리 변환
	selectQuery, err := cls.GetQueryJson(sqlquery, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(selectQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

// 주문 메뉴
type OrderMenu struct {
	ItemNo    string `json:"itemNo"`    //
	ItemPrice int    `json:"itemPrice"` //
	ItemCount int    `json:"itemCount"` //
	UserId    string `json:"userId"`    //
}

// N/1 데이터
type PaySplit struct {
	UserId   string `json:"userId"`   //
	SplitAmt int    `json:"splitAmt"` //
}

type Order struct {
	UserId    string `json:"userId"`   //
	GrpId     string `json:"grpId"`    //
	PayTy     int    `json:"payTy"`    //
	OrderTy   int    `json:"orderTy"`  //
	OrderAmt  int    `json:"orderAmt"` //
	OrderMenu []OrderMenu
	PaySplit  []PaySplit
	OrderMemo []OrderMemo
}

type OrderMemo struct {
	UserId string `json:"userId"` //
	Memo   string `json:"memo"`   //
}

// 주문하기
func SetOrderPay(c echo.Context) error {

	payTy := c.FormValue("payTy")

	if payTy == "0" {
		//선불결제
		return SetOrderPay_PREPAY(c)
	} else if payTy == "1" {
		//후불 결제
		return SetOrderPay_DEFERPAY(c)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "99"
	m["resultMsg"] = "잘못된 호출입니다."
	return c.JSON(http.StatusOK, m)
}

// 주문하기(선불)
func SetOrderPay_PREPAY(c echo.Context) error {

	dprintf(4, c, "call 주문하기 선불 \n")

	bodyBytes, _ := ioutil.ReadAll(c.Request().Body)

	// 상세 주문데이터 get
	var order Order
	err2 := json.Unmarshal(bodyBytes, &order)
	if err2 != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err2.Error()))
	}
	orderMenus := order.OrderMenu
	orderSplitAmt := order.PaySplit
	orderMemo := order.OrderMemo

	c.Request().Body.Close() //  must close
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// 상세 주문데이터 get 끝
	params := cls.GetParamJsonMap(c)

	//1.장부 유효성 체크
	bookInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다."))
	}

	limitYn := bookInfo[0]["LIMIT_YN"]
	orderUserNm := bookInfo[0]["USER_NM"]
	if limitYn == "Y" {
		newLayout := "15:04"
		now := time.Now().Format(newLayout)

		chkGrpUseTime, _ := time.Parse(newLayout, now)
		grpUseTimeStart, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_START"]+":00")
		grpUseTimeEnd, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_END"]+":00")
		chkTime := inTimeSpan(grpUseTimeStart, grpUseTimeEnd, chkGrpUseTime)

		if bookInfo[0]["LIMIT_USE_TIME_START"] == "0" && bookInfo[0]["LIMIT_USE_TIME_END"] == "0" {
			chkTime = true
		}

		if chkTime == false {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 사용 가능시간은 "+bookInfo[0]["LIMIT_USE_TIME_START"]+"시 부터 "+bookInfo[0]["LIMIT_USE_TIME_END"]+"시 까지입니다."))
		}

	}

	params["checkTime"] = bookInfo[0]["CHECK_TIME"]
	//2.선후불 체크
	grpPayTy := bookInfo[0]["GRP_PAY_TY"]
	if grpPayTy == "1" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "후불 결제만 가능한 장부 입니다."))
	}

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	limitAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_AMT"])
	limitDayAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_AMT"])
	limitDayCnt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_CNT"])
	supportYn := bookInfo[0]["SUPPORT_YN"]
	supportExeedYn := bookInfo[0]["SUPPORT_EXCEED_YN"]                  // 지원금 초과사용 여부
	mySupportBalance, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 내 지원금
	orderAmt, _ := strconv.Atoi(params["orderAmt"])                     // 결제 금액

	//4.장부 제한 사항 체크
	// 결제 타입

	myPrepay := 0
	myPoint := 0
	pointUse := 0

	//5.장부 한도 체크
	linkInfo, err := cls.GetSelectData(restsql.SelectStoreLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결된 가맹점 정보가 없습니다."))
	}
	prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
	prepaidPoint, _ := strconv.Atoi(linkInfo[0]["PREPAID_POINT"])
	agrmId := linkInfo[0]["AGRM_ID"]
	pointRate := linkInfo[0]["POINT_RATE"]

	if orderAmt > prepaidAmt {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불금액이 모자랍니다. 선불금액 충전 후 주문해주세요. "))
	} else {
		myPrepay = prepaidAmt - orderAmt

		if pointRate != "0" {

			pointUse = orderPoint(orderAmt, prepaidPoint, pointRate)
			myPoint = prepaidPoint - pointUse
		}

	}

	//아이템별 가격 체크
	itemList, err := cls.GetSelectData(restsql.SelectStoreItemList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	itemMap := make(map[string]interface{})
	for i, item := range itemList {
		menuitem := item["ITEM_NO"] + item["ITEM_PRICE"]
		itemMap[menuitem] = i
	}
	for i, _ := range orderMenus {

		itemPrice := orderMenus[i].ItemPrice
		itemNo := orderMenus[i].ItemNo

		//9999999999 는 금액아이템 가격 비교 안함
		if itemNo != "9999999999" {
			orderItem := itemNo + strconv.Itoa(itemPrice)
			isMenu := itemMap[orderItem]
			if isMenu == nil {
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
			}
		}

	}

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err

	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -주문 선불(SetPayment)  \n")
			tx.Rollback()
		}
	}()

	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	params["pointUse"] = strconv.Itoa(pointUse)

	qrOrderTy := params["qrOrderTy"]

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//개인 메모 등록
	for i, _ := range orderMemo {

		memoUserId := orderMemo[i].UserId
		userMemo := orderMemo[i].Memo

		params["memoUserId"] = memoUserId
		params["userMemo"] = userMemo

		if userMemo != "" {
			OrderMemoQuery, err := cls.SetUpdateParam(ordersql.InsertOrderMemo, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(OrderMemoQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		}
	}

	//주문 상세 메뉴 등록

	//8.order detail insert
	amtChk1 := 0
	for i, _ := range orderMenus {

		itemNo := orderMenus[i].ItemNo
		itemPrice := strconv.Itoa(orderMenus[i].ItemPrice)
		itemCount := strconv.Itoa(orderMenus[i].ItemCount)
		itemUserId := orderMenus[i].UserId

		amtChk1 = amtChk1 + (orderMenus[i].ItemPrice * orderMenus[i].ItemCount)

		params["orderSeq"] = strconv.Itoa(i + 1)
		params["itemNo"] = itemNo
		params["itemPrice"] = itemPrice
		params["itemCount"] = itemCount
		params["itemUserId"] = itemUserId

		OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(OrderDetailCreateQuery)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}
	if strconv.Itoa(amtChk1) != params["orderAmt"] {
		tx.Rollback()
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "총 주문 금액과 메뉴합산 금액이 다릅니다."))
	}

	// 엔빵인경우 SPLITPAY INSERT
	if qrOrderTy == "2" {
		amtChk2 := 0

		for i, _ := range orderSplitAmt {

			splitAmt := strconv.Itoa(orderSplitAmt[i].SplitAmt)
			splitUserId := orderSplitAmt[i].UserId

			params["splitAmt"] = splitAmt
			params["splitUserId"] = splitUserId

			amtChk2 = amtChk2 + orderSplitAmt[i].SplitAmt

			OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderSplitAmt, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(OrderDetailCreateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

		}

		if strconv.Itoa(amtChk2) != params["orderAmt"] {
			tx.Rollback()
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "총 주문 금액과 메뉴합산 금액이 다릅니다(2)."))
		}

	}

	//지원금 사용인경우
	//10.개인 지원금 차감
	if supportYn == "Y" {
		if qrOrderTy == "0" {

			//금액 제한 설정 체크
			if limitYn == "Y" {
				limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
				if limitChek == "N" {
					tx.Rollback()
					return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
				}
			}

			if supportExeedYn == "N" {
				if orderAmt > mySupportBalance {
					tx.Rollback()
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금 잔액이 부족합니다."))
				}
			}
			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else if qrOrderTy == "1" {

			params["orderAmt"] = ""
			params["userId"] = ""
			for i, _ := range orderMenus {

				itemUserId := orderMenus[i].UserId
				userOrderAmt := orderMenus[i].ItemPrice * orderMenus[i].ItemCount
				params["orderAmt"] = strconv.Itoa(userOrderAmt)
				params["userId"] = itemUserId

				//금액 제한 설정 체크
				if limitYn == "Y" {
					limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, userOrderAmt, itemUserId, params["grpId"])
					if limitChek == "N" {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
					}
				}

				// 장부정보 불러오기
				pBookInfo, err := cls.GetSelectData(booksql.SelectUserBookInfo, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if pBookInfo == nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다-"+itemUserId))
				}

				userSupportAmt, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 사용자별 지원금

				if supportExeedYn == "N" {
					if userOrderAmt > userSupportAmt {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금이 부족한 사용자가 있습니다."))
					}
				}
				UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
				}
				_, err = tx.Exec(UserSupportBalanceUpdateQuery)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
			}

		} else if qrOrderTy == "2" {

			params["orderAmt"] = ""
			params["userId"] = ""
			for i, _ := range orderSplitAmt {

				itemUserId := orderSplitAmt[i].UserId
				userOrderAmt := orderSplitAmt[i].SplitAmt
				params["orderAmt"] = strconv.Itoa(userOrderAmt)
				params["userId"] = itemUserId

				//금액 제한 설정 체크

				if limitYn == "Y" {
					limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, userOrderAmt, itemUserId, params["grpId"])
					if limitChek == "N" {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
					}
				}
				// 장부정보 불러오기
				pBookInfo, err := cls.GetSelectData(booksql.SelectUserBookInfo, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if pBookInfo == nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다-"+itemUserId))
				}
				userSupportAmt, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 사용자별 지원금

				if supportExeedYn == "N" {
					if userOrderAmt > userSupportAmt {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금이 부족한 사용자가 있습니다."))
					}
				}
				UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
				}
				_, err = tx.Exec(UserSupportBalanceUpdateQuery)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
			}

		} else {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 유형입니다."))
		}
	} else {
		if limitYn == "Y" {
			limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
			if limitChek == "N" {
				tx.Rollback()
				return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
			}
		}
	}

	//11. 충전금 차감
	params["agrmId"] = agrmId
	params["prepaidAmt"] = strconv.Itoa(myPrepay)
	params["prepaidPoint"] = strconv.Itoa(myPoint)

	LinkAmtUpdateQuery, err := cls.SetUpdateParam(restsql.UpdateLinkAmt, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(LinkAmtUpdateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	if len(os.Getenv("SERVER_TYPE")) > 0 {
		// 푸쉬 전송 시작
		pushMsg := orderUserNm + "님이 주문(계산)하였습니다."
		go apiPush.SendPush_Msg_V1("주문", pushMsg, "M", "1", params["restId"], "", "bookorder")
		// 푸쉬 전송 완료
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

// 주문하기(후불)
func SetOrderPay_DEFERPAY(c echo.Context) error {

	dprintf(4, c, "call 주문하기 후불\n")

	bodyBytes, _ := ioutil.ReadAll(c.Request().Body)

	// 상세 주문데이터 get
	var order Order
	err2 := json.Unmarshal(bodyBytes, &order)
	if err2 != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err2.Error()))
	}
	orderMenus := order.OrderMenu
	orderSplitAmt := order.PaySplit
	orderMemo := order.OrderMemo

	c.Request().Body.Close() //  must close
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// 상세 주문데이터 get 끝
	params := cls.GetParamJsonMap(c)

	//1.장부 유효성 체크
	bookInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다."))
	}

	if params["restId"] == "S0000000668" {
		params["orderTy"] = "3"
	}

	limitYn := bookInfo[0]["LIMIT_YN"]
	orderUserNm := bookInfo[0]["USER_NM"]

	if limitYn == "Y" {
		newLayout := "15:04"
		now := time.Now().Format(newLayout)

		chkGrpUseTime, _ := time.Parse(newLayout, now)
		grpUseTimeStart, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_START"]+":00")
		grpUseTimeEnd, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_END"]+":00")
		chkTime := inTimeSpan(grpUseTimeStart, grpUseTimeEnd, chkGrpUseTime)

		if bookInfo[0]["LIMIT_USE_TIME_START"] == "0" && bookInfo[0]["LIMIT_USE_TIME_END"] == "0" {
			chkTime = true
		}

		if chkTime == false {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 사용 가능시간은 "+bookInfo[0]["LIMIT_USE_TIME_START"]+"시 부터 "+bookInfo[0]["LIMIT_USE_TIME_END"]+"시 까지입니다."))
		}

	}

	params["checkTime"] = bookInfo[0]["CHECK_TIME"]
	//2.선후불 체크
	grpPayTy := bookInfo[0]["GRP_PAY_TY"]
	if grpPayTy == "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불 결제만 가능한 장부 입니다."))
	}

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	limitAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_AMT"])
	limitDayAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_AMT"])
	limitDayCnt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_CNT"])
	supportYn := bookInfo[0]["SUPPORT_YN"]
	supportExeedYn := bookInfo[0]["SUPPORT_EXCEED_YN"]                  // 지원금 초과사용 여부
	mySupportBalance, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 내 지원금
	orderAmt, _ := strconv.Atoi(params["orderAmt"])                     // 결제 금액

	//4.장부 제한 사항 체크
	// 결제 타입

	//아이템별 가격 체크
	itemList, err := cls.GetSelectData(restsql.SelectStoreItemList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	itemMap := make(map[string]interface{})
	for i, item := range itemList {
		menuitem := item["ITEM_NO"] + item["ITEM_PRICE"]
		itemMap[menuitem] = i
	}
	for i, _ := range orderMenus {

		itemPrice := orderMenus[i].ItemPrice
		itemNo := orderMenus[i].ItemNo

		//9999999999 는 금액아이템 가격 비교 안함
		if itemNo != "9999999999" {
			orderItem := itemNo + strconv.Itoa(itemPrice)
			isMenu := itemMap[orderItem]
			if isMenu == nil {
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
			}
		}
	}

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err

	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -주문 후불(SetOrderPay_DEFERPAY)  \n")
			tx.Rollback()
		}
	}()

	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	qrOrderTy := params["qrOrderTy"]
	params["pointUse"] = "0"

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//개인 메모 등록
	for i, _ := range orderMemo {

		memoUserId := orderMemo[i].UserId
		userMemo := orderMemo[i].Memo

		params["memoUserId"] = memoUserId
		params["userMemo"] = userMemo

		if userMemo != "" {
			OrderMemoQuery, err := cls.SetUpdateParam(ordersql.InsertOrderMemo, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(OrderMemoQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		}
	}

	//주문 상세 메뉴 등록

	//8.order detail insert
	amtChk1 := 0
	for i, _ := range orderMenus {

		itemNo := orderMenus[i].ItemNo
		itemPrice := strconv.Itoa(orderMenus[i].ItemPrice)
		itemCount := strconv.Itoa(orderMenus[i].ItemCount)
		itemUserId := orderMenus[i].UserId

		amtChk1 = amtChk1 + (orderMenus[i].ItemPrice * orderMenus[i].ItemCount)

		params["orderSeq"] = strconv.Itoa(i + 1)
		params["itemNo"] = itemNo
		params["itemPrice"] = itemPrice
		params["itemCount"] = itemCount
		params["itemUserId"] = itemUserId

		OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(OrderDetailCreateQuery)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}
	if strconv.Itoa(amtChk1) != params["orderAmt"] {
		tx.Rollback()
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "총 주문 금액과 메뉴합산 금액이 다릅니다."))
	}

	// 엔빵인경우 SPLITPAY INSERT
	if qrOrderTy == "2" {
		amtChk2 := 0

		for i, _ := range orderSplitAmt {

			splitAmt := strconv.Itoa(orderSplitAmt[i].SplitAmt)
			splitUserId := orderSplitAmt[i].UserId

			params["splitAmt"] = splitAmt
			params["splitUserId"] = splitUserId

			amtChk2 = amtChk2 + orderSplitAmt[i].SplitAmt

			OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderSplitAmt, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(OrderDetailCreateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

		}

		if strconv.Itoa(amtChk2) != params["orderAmt"] {
			tx.Rollback()
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "총 주문 금액과 메뉴합산 금액이 다릅니다(2)."))
		}

	}

	//지원금 사용인경우
	//11.개인 지원금 차감
	if supportYn == "Y" {
		if qrOrderTy == "0" {
			if supportExeedYn == "N" {
				if orderAmt > mySupportBalance {
					tx.Rollback()
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금 잔액이 부족합니다."))
				}
			}

			//금액 제한 설정 체크
			if limitYn == "Y" {
				limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
				if limitChek == "N" {
					tx.Rollback()
					return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
				}
			}

			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else if qrOrderTy == "1" {

			params["orderAmt"] = ""
			params["userId"] = ""
			for i, _ := range orderMenus {

				itemUserId := orderMenus[i].UserId
				userOrderAmt := orderMenus[i].ItemPrice * orderMenus[i].ItemCount
				params["orderAmt"] = strconv.Itoa(userOrderAmt)
				params["userId"] = itemUserId

				//금액 제한 설정 체크
				if limitYn == "Y" {
					limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, userOrderAmt, itemUserId, params["grpId"])
					if limitChek == "N" {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
					}
				}

				// 장부정보 불러오기
				pBookInfo, err := cls.GetSelectData(booksql.SelectUserBookInfo, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if pBookInfo == nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다-"+itemUserId))
				}

				userSupportAmt, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 사용자별 지원금

				if supportExeedYn == "N" {
					if userOrderAmt > userSupportAmt {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금이 부족한 사용자가 있습니다."))
					}
				}
				UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
				}
				_, err = tx.Exec(UserSupportBalanceUpdateQuery)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
			}

		} else if qrOrderTy == "2" {

			params["orderAmt"] = ""
			params["userId"] = ""
			for i, _ := range orderSplitAmt {

				itemUserId := orderSplitAmt[i].UserId
				userOrderAmt := orderSplitAmt[i].SplitAmt
				params["orderAmt"] = strconv.Itoa(userOrderAmt)
				params["userId"] = itemUserId

				//금액 제한 설정 체크
				if limitYn == "Y" {
					limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, userOrderAmt, itemUserId, params["grpId"])
					if limitChek == "N" {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
					}
				}

				// 장부정보 불러오기
				pBookInfo, err := cls.GetSelectData(booksql.SelectUserBookInfo, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if pBookInfo == nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다-"+itemUserId))
				}
				userSupportAmt, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 사용자별 지원금

				if supportExeedYn == "N" {
					if userOrderAmt > userSupportAmt {
						tx.Rollback()
						return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금이 부족한 사용자가 있습니다."))
					}
				}
				UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
				}
				_, err = tx.Exec(UserSupportBalanceUpdateQuery)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
			}

		} else {
			tx.Rollback()
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 유형입니다."))
		}
	} else {

		//금액 제한 설정 체크
		if limitYn == "Y" {
			limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
			if limitChek == "N" {
				tx.Rollback()
				return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
			}
		}

	}

	if params["orderTy"] == "3" {

		OrderPickupQuery, err := cls.SetUpdateParam(ordersql.InsertOrderPickupData, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(OrderPickupQuery)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	if len(os.Getenv("SERVER_TYPE")) > 0 {
		// 푸쉬 전송 시작
		pushMsg := orderUserNm + "님이 주문(계산)하였습니다."
		go apiPush.SendPush_Msg_V1("주문", pushMsg, "M", "1", params["restId"], "", "bookorder")
		// 푸쉬 전송 완료
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

//장부별 사용금액 체크
func grpUseAmtCheck(limitAmt int, limitDayAmt int, limitDayCnt int, orderAmt int, userId string, grpId string) (string, string) {

	//금액 제한 설정 체크
	if limitAmt > 0 && orderAmt > limitAmt {
		return "N", "1회 사용금액을 초과하였습니다."
	}

	if limitDayAmt > 0 {
		// 장부정보 불러오기
		params := make(map[string]string)
		params["grpId"] = grpId
		params["userId"] = userId
		myTodayOrderAmtChk, err := cls.SelectData(ordersql.SelectTodayOrderAmt, params)
		if err != nil {
			return "N", err.Error()
		}
		todayOrderAmt, _ := strconv.Atoi(myTodayOrderAmtChk[0]["TODAY_ORDER_AMT"])
		chkAmt := todayOrderAmt + orderAmt
		if chkAmt > limitDayAmt {
			return "N", "일일 사용금액을 초과하였습니다."
		}
	}

	if limitDayCnt > 0 {
		// 장부정보 불러오기
		params := make(map[string]string)
		params["grpId"] = grpId
		params["userId"] = userId
		myTodayOrderAmtChk, err := cls.SelectData(ordersql.SelectTodayOrderAmt, params)
		if err != nil {
			return "N", err.Error()
		}
		todayCount, _ := strconv.Atoi(myTodayOrderAmtChk[0]["TODAY_COUNT"])

		if todayCount >= limitDayCnt {
			return "N", "일일 사용횟수를 초과하였습니다."
		}
	}

	return "Y", "정상"
}

//  오늘 주문 내역
func GetOrderToday(c echo.Context) error {

	dprintf(4, c, "call GetOrderToday\n")

	params := cls.GetParamJsonMap(c)

	resultList, err := cls.GetSelectType(ordersql.SelectTodayOrder, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultList"] = []string{}
		return c.JSON(http.StatusOK, m)

	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}

//  장부별 주문내역
func GetBooksOrder(c echo.Context) error {

	dprintf(4, c, "call GetBooksOrder\n")

	params := cls.GetParamJsonMap(c)

	myGrpList, err := cls.GetSelectData(usersql.SelectMyGrpList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	for i := range myGrpList {

		params["grpId"] = myGrpList[i]["grpId"]
		params["grpAuth"] = myGrpList[i]["grpAuth"]
		queryStr := ""
		if myGrpList[i]["GRP_AUTH"] == "0" {
			queryStr = ordersql.SelectGrpLastOrder0
		} else {
			queryStr = ordersql.SelectGrpLastOrder1
		}
		lastOrder, err := cls.GetSelectData(queryStr, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		if lastOrder == nil {

		} else {
			myGrpList[i]["orderNo"] = lastOrder[0]["ORDER_NO"]
			myGrpList[i]["orderDate"] = lastOrder[0]["ORDER_DATE"]
			myGrpList[i]["restNm"] = lastOrder[0]["REST_NM"]
			myGrpList[i]["orderUserCnt"] = lastOrder[0]["ORDER_USER_CNT"]
			myGrpList[i]["totalAmt"] = lastOrder[0]["TOTAL_AMT"]
			myGrpList[i]["userNm"] = lastOrder[0]["USER_NM"]
			myGrpList[i]["orderStat"] = lastOrder[0]["ORDER_STAT"]
			myGrpList[i]["grpTypeCd"] = lastOrder[0]["GRP_TYPE_CD"]

		}
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = myGrpList

	return c.JSON(http.StatusOK, m)

}

// 장부별 주문 상세 리스트
func GetBookOrdersList(c echo.Context) error {

	dprintf(4, c, "call GetBookOrdersList\n")

	params := cls.GetParamJsonMap(c)

	//페이징처리
	pageSize, _ := strconv.Atoi(params["pageSize"])
	pageNo, _ := strconv.Atoi(params["pageNo"])

	offset := strconv.Itoa((pageNo - 1) * pageSize)
	if pageNo == 1 {
		offset = "0"
	}
	params["offSet"] = offset

	bookAuth, err := cls.GetSelectData(booksql.SelectBookMyAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	grpNm := bookAuth[0]["GRP_NM"]
	grpAuth := bookAuth[0]["GRP_AUTH"]

	// 장부장일때는 주문 전체
	// 장부원일때는 개인 주문
	countQuery := ordersql.SelectBookOrderCount
	totalQuery := ordersql.SelectBookOrderTotal
	listQuery := ordersql.SelectBookOrderList + commons.PagingQuery

	if grpAuth == "1" {
		countQuery = ordersql.SelectBookOrderCount_AUTH1
		totalQuery = ordersql.SelectBookOrderTotal_AUTH1
		listQuery = ordersql.SelectBookOrderList_AUTH1 + commons.PagingQuery
	}

	orderListCount, err := cls.GetSelectData(countQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	totalCount, _ := strconv.Atoi(orderListCount[0]["totalCount"])
	totalPage := strconv.Itoa((totalCount / pageSize) + 1)

	orderTotal, err := cls.GetSelectData(totalQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	orderList, err := cls.GetSelectType(listQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	if orderList == nil {
		LinkData := make(map[string]interface{})
		LinkData["totalPage"] = totalPage
		LinkData["currentPage"] = pageNo
		LinkData["totalCount"] = 0
		LinkData["totalAmt"] = 0
		LinkData["orderList"] = []string{}
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = LinkData
		return c.JSON(http.StatusOK, m)
	}

	totalAmt, _ := strconv.Atoi(orderTotal[0]["totalAmt"])
	LinkData := make(map[string]interface{})
	LinkData["totalAmt"] = totalAmt
	LinkData["totalPage"] = totalPage
	LinkData["totalCount"] = totalCount
	LinkData["grpNm"] = grpNm
	LinkData["grpAuth"] = grpAuth
	LinkData["currentPage"] = pageNo
	LinkData["orderList"] = orderList

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = LinkData

	return c.JSON(http.StatusOK, m)

}

func GetOrderInfo(c echo.Context) error {

	dprintf(4, c, "call GetOrderInfo\n")

	params := cls.GetParamJsonMap(c)

	orderInfo, err := cls.GetSelectData(ordersql.SelectOrderInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}
	if orderInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 정보 입니다."))
	}

	qrOrderTy := orderInfo[0]["QR_ORDER_TYPE"]

	totalMenu, err := cls.GetSelectType(ordersql.SelectOrderDetail, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}

	//	userList := make([]map[string]string)

	if qrOrderTy == "2" {

		userSplitList, err := cls.GetSelectType(ordersql.SelectOrderUserSplitAmt, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
		}

		order := make(map[string]interface{})
		order["orderNo"] = orderInfo[0]["ORDER_NO"]
		order["restNm"] = orderInfo[0]["REST_NM"]
		order["grpNm"] = orderInfo[0]["GRP_NM"]
		totalAmt, _ := strconv.Atoi(orderInfo[0]["TOTAL_AMT"])
		order["totalAmt"] = totalAmt
		order["orderStat"] = orderInfo[0]["ORDER_STAT"]
		order["orderDate"] = orderInfo[0]["ORDER_DATE"]
		order["totalMenu"] = totalMenu
		order["usersList"] = userSplitList

		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = order

		return c.JSON(http.StatusOK, m)

	} else {

		userDetail, err := cls.GetSelectData(ordersql.SelectOrderUserDetail, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
		}

		userList := make([]map[string]interface{}, len(userDetail))
		for i := range userDetail {

			params["userId"] = userDetail[i]["USER_ID"]
			userMenu, err := cls.GetSelectType(ordersql.SelectOrderUserMenu, params, c)
			if err != nil {
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

			pOrderAmt, _ := strconv.Atoi(userDetail[i]["ORDER_AMT"])
			order2 := make(map[string]interface{})
			order2["userNm"] = userDetail[i]["USER_NM"]
			order2["orderAmt"] = pOrderAmt
			order2["menus"] = userMenu
			order2["memo"] = userDetail[i]["MEMO"]
			userList[i] = order2

		}

		order := make(map[string]interface{})
		order["orderNo"] = orderInfo[0]["ORDER_NO"]
		order["restNm"] = orderInfo[0]["REST_NM"]
		order["grpNm"] = orderInfo[0]["GRP_NM"]
		totalAmt, _ := strconv.Atoi(orderInfo[0]["TOTAL_AMT"])
		order["totalAmt"] = totalAmt
		order["orderStat"] = orderInfo[0]["ORDER_STAT"]
		order["orderDate"] = orderInfo[0]["ORDER_DATE"]
		order["totalMenu"] = totalMenu
		order["usersList"] = userList

		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = order

		return c.JSON(http.StatusOK, m)

	}

}

// 결제 내역 - 장부별
func GetBooksPayment(c echo.Context) error {

	dprintf(4, c, "call GetBooksPayment\n")

	params := cls.GetParamJsonMap(c)

	myGrpArray, err := cls.GetSelectData(usersql.SelectMyGrpAuth0Array, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if myGrpArray == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부가 없습니다."))
	}
	params["myGrp"] = strings.ReplaceAll(myGrpArray[0]["myGrp"], ",", "','")

	booksPayment, err := cls.GetSelectData(ordersql.SelectBooksPayment, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	for i := range booksPayment {

		params["grpId"] = booksPayment[i]["grpId"]
		params["regDate"] = booksPayment[i]["rDate"]
		paymentData, err := cls.GetSelectData(ordersql.SelectLastPaymentData, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		if paymentData == nil {

		} else {
			booksPayment[i]["creditAmt"] = paymentData[0]["creditAmt"]
			booksPayment[i]["payCnt"] = paymentData[0]["payCnt"]
		}
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = booksPayment

	return c.JSON(http.StatusOK, m)

}

// 장부별 결제 상세 리스트
func GetBooksPaymentList(c echo.Context) error {

	dprintf(4, c, "call GetBooksPaymentList\n")

	params := cls.GetParamJsonMap(c)

	//페이징처리
	pageSize, _ := strconv.Atoi(params["pageSize"])
	pageNo, _ := strconv.Atoi(params["pageNo"])

	offset := strconv.Itoa((pageNo - 1) * pageSize)
	if pageNo == 1 {
		offset = "0"
	}
	params["offSet"] = offset

	bookAuth, err := cls.GetSelectDataRequire(booksql.SelectBookMyAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	grpNm := bookAuth[0]["GRP_NM"]
	//grpAuth :=bookAuth[0]["GRP_AUTH"]

	paymentListCount, err := cls.GetSelectData(ordersql.SelectBookPaymentListCount, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	totalCount, _ := strconv.Atoi(paymentListCount[0]["totalCount"])
	totalPage := strconv.Itoa((totalCount / pageSize) + 1)
	totalAmt, _ := strconv.Atoi(paymentListCount[0]["totalAmt"])

	paymentList, err := cls.GetSelectType(ordersql.SelectBookPaymentList+commons.PagingQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	if paymentList == nil {
		payment := make(map[string]interface{})
		payment["totalPage"] = totalPage
		payment["currentPage"] = pageNo
		payment["totalCount"] = 0
		payment["totalAmt"] = 0
		payment["paymentList"] = []string{}
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = payment
		return c.JSON(http.StatusOK, m)
	}

	payment := make(map[string]interface{})
	payment["totalPage"] = totalPage
	payment["totalCount"] = totalCount
	payment["totalAmt"] = totalAmt
	payment["grpNm"] = grpNm
	payment["currentPage"] = pageNo
	payment["paymentList"] = paymentList

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = payment

	return c.JSON(http.StatusOK, m)

}

func GetPaymentInfo(c echo.Context) error {

	dprintf(4, c, "call GetPaymentInfo\n")

	params := cls.GetParamJsonMap(c)

	paymentInfo, err := cls.GetSelectType(ordersql.SelectPaymentInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}
	if paymentInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 결제 정보 입니다."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = paymentInfo[0]

	return c.JSON(http.StatusOK, m)

}

func GetPaymentCancelInfo(c echo.Context) error {

	dprintf(4, c, "call GetPaymentCancelInfo\n")

	params := cls.GetParamJsonMap(c)

	paymentCancelData, err := cls.GetSelectType(ordersql.SelectPaymentCancelChk, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}
	if paymentCancelData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소 내역이 없습니다."))
	}

	paymentInfo, err := cls.GetSelectType(ordersql.SelectPaymentCancelInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}
	if paymentInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소 내역이 없습니다."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = paymentInfo[0]

	return c.JSON(http.StatusOK, m)

}

func inTimeSpan(start, end, check time.Time) bool {
	if check.After(end) {
		return false
	}
	if end.After(start) {
		return check.After(start)
	}
	return check.Before(start)
}

func orderPoint(orderAmt int, prepaidPoint int, pointRate string) int {
	point := 0

	value, err := strconv.ParseFloat(pointRate, 32)
	if err != nil {
		// do something sensible
	}
	floatPointRate := float32(value)

	//println(floatPointRate)

	floatOrderAmt := float32(orderAmt)

	if prepaidPoint == 0 {
		return point
	}

	floatOrderPoint := floatOrderAmt * floatPointRate / 100

	mPoint := fmt.Sprintf("%.0f", floatOrderPoint)
	smPoint, _ := strconv.Atoi(mPoint)

	//println(smPoint)

	if smPoint > prepaidPoint {
		point = prepaidPoint
	} else {
		point = smPoint
	}
	//println(point)

	return point
}

func GetBooksPaymentList_V2(c echo.Context) error {

	dprintf(4, c, "call GetBooksPaymentList_V2\n")

	params := cls.GetParamJsonMap(c)

	//페이징처리
	pageSize, _ := strconv.Atoi(params["pageSize"])
	pageNo, _ := strconv.Atoi(params["pageNo"])

	offset := strconv.Itoa((pageNo - 1) * pageSize)
	if pageNo == 1 {
		offset = "0"
	}
	params["offSet"] = offset

	lastDate := ""

	first := params["first"]

	bookAuth, err := cls.GetSelectDataRequire(booksql.SelectBookMyAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	grpNm := bookAuth[0]["GRP_NM"]
	//grpAuth :=bookAuth[0]["GRP_AUTH"]

	paymentListCount, err := cls.GetSelectData(ordersql.SelectBookPaymentListCount, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	totalCount, _ := strconv.Atoi(paymentListCount[0]["totalCount"])
	totalPage := strconv.Itoa((totalCount / pageSize) + 1)
	totalAmt, _ := strconv.Atoi(paymentListCount[0]["totalAmt"])

	if first == "Y" && totalCount == 0 {
		orderInfo, err := cls.GetSelectData(ordersql.SelectLastPayDate, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}

		lastDate = orderInfo[0]["LAST_PAY_DATE"]
		params["startDate"] = strings.Replace(lastDate, "-", "", -1)

		paymentListCount, err := cls.GetSelectData(ordersql.SelectBookPaymentListCount, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}
		totalCount, _ = strconv.Atoi(paymentListCount[0]["totalCount"])
		totalPage = strconv.Itoa((totalCount / pageSize) + 1)
		totalAmt, _ = strconv.Atoi(paymentListCount[0]["totalAmt"])

	}

	paymentList, err := cls.GetSelectType(ordersql.SelectBookPaymentList+commons.PagingQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	if paymentList == nil {
		payment := make(map[string]interface{})
		payment["totalPage"] = totalPage
		payment["currentPage"] = pageNo
		payment["totalCount"] = 0
		payment["totalAmt"] = 0
		payment["paymentList"] = []string{}
		payment["lastDate"] = ""
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = payment
		return c.JSON(http.StatusOK, m)
	}

	payment := make(map[string]interface{})
	payment["totalPage"] = totalPage
	payment["totalCount"] = totalCount
	payment["totalAmt"] = totalAmt
	payment["grpNm"] = grpNm
	payment["currentPage"] = pageNo
	payment["paymentList"] = paymentList
	payment["lastDate"] = lastDate

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = payment

	return c.JSON(http.StatusOK, m)

}

func GetBookOrdersList_V2(c echo.Context) error {

	dprintf(4, c, "call GetBookOrdersList_V2\n")

	params := cls.GetParamJsonMap(c)

	//페이징처리
	pageSize, _ := strconv.Atoi(params["pageSize"])
	pageNo, _ := strconv.Atoi(params["pageNo"])

	offset := strconv.Itoa((pageNo - 1) * pageSize)
	if pageNo == 1 {
		offset = "0"
	}
	params["offSet"] = offset

	first := params["first"]

	lastDate := ""

	bookAuth, err := cls.GetSelectData(booksql.SelectBookMyAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if bookAuth == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}

	grpNm := bookAuth[0]["GRP_NM"]
	grpAuth := bookAuth[0]["GRP_AUTH"]

	// 장부장일때는 주문 전체
	// 장부원일때는 개인 주문
	countQuery := ordersql.SelectBookOrderCount
	totalQuery := ordersql.SelectBookOrderTotal
	listQuery := ordersql.SelectBookOrderList + commons.PagingQuery

	if grpAuth == "1" {
		countQuery = ordersql.SelectBookOrderCount_AUTH1
		totalQuery = ordersql.SelectBookOrderTotal_AUTH1
		listQuery = ordersql.SelectBookOrderList_AUTH1 + commons.PagingQuery
	}

	orderListCount, err := cls.GetSelectData(countQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	totalCount, _ := strconv.Atoi(orderListCount[0]["totalCount"])
	totalPage := strconv.Itoa((totalCount / pageSize) + 1)

	if first == "Y" && totalCount == 0 {
		orderInfo, err := cls.GetSelectData(ordersql.SelectLastOrderDate, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}

		lastDate = orderInfo[0]["LAST_ORDER_DATE"]
		params["startDate"] = strings.Replace(lastDate, "-", "", -1)

		orderListCount, err := cls.GetSelectData(countQuery, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}
		totalCount, _ = strconv.Atoi(orderListCount[0]["totalCount"])
		totalPage = strconv.Itoa((totalCount / pageSize) + 1)

	}

	orderTotal, err := cls.GetSelectData(totalQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	orderList, err := cls.GetSelectType(listQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	if orderList == nil {
		LinkData := make(map[string]interface{})
		LinkData["totalPage"] = totalPage
		LinkData["currentPage"] = pageNo
		LinkData["totalCount"] = 0
		LinkData["totalAmt"] = 0
		LinkData["orderList"] = []string{}
		LinkData["lastOrderDate"] = ""
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = LinkData
		return c.JSON(http.StatusOK, m)
	}

	totalAmt, _ := strconv.Atoi(orderTotal[0]["totalAmt"])
	LinkData := make(map[string]interface{})
	LinkData["totalAmt"] = totalAmt
	LinkData["totalPage"] = totalPage
	LinkData["totalCount"] = totalCount
	LinkData["grpNm"] = grpNm
	LinkData["grpAuth"] = grpAuth
	LinkData["currentPage"] = pageNo
	LinkData["orderList"] = orderList
	LinkData["lastDate"] = lastDate

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = LinkData

	return c.JSON(http.StatusOK, m)

}

// 주문하기 - 기프티콘
func SetOrderGifticon(c echo.Context) error {

	payTy := c.FormValue("payTy")

	if payTy == "0" {
		//선불결제
		return SetOrderGifticon_PREPAY(c)
	} else if payTy == "1" {
		//후불 결제
		return SetOrderGifticon_DEFERPAY(c)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "99"
	m["resultMsg"] = "잘못된 호출입니다."
	return c.JSON(http.StatusOK, m)
}

// 주문하기(선불) - 기프티콘
func SetOrderGifticon_PREPAY(c echo.Context) error {

	dprintf(4, c, "call 기프티콘 주문하기 선불 \n")

	params := cls.GetParamJsonMap(c)

	//1.장부 유효성 체크
	bookInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다."))
	}

	userTel := bookInfo[0]["USER_TEL"]
	limitYn := bookInfo[0]["LIMIT_YN"]
	if limitYn == "Y" {
		newLayout := "15:04"
		now := time.Now().Format(newLayout)

		chkGrpUseTime, _ := time.Parse(newLayout, now)
		grpUseTimeStart, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_START"]+":00")
		grpUseTimeEnd, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_END"]+":00")
		chkTime := inTimeSpan(grpUseTimeStart, grpUseTimeEnd, chkGrpUseTime)

		if bookInfo[0]["LIMIT_USE_TIME_START"] == "0" && bookInfo[0]["LIMIT_USE_TIME_END"] == "0" {
			chkTime = true
		}

		if chkTime == false {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 사용 가능시간은 "+bookInfo[0]["LIMIT_USE_TIME_START"]+"시 부터 "+bookInfo[0]["LIMIT_USE_TIME_END"]+"시 까지입니다."))
		}

	}

	params["checkTime"] = bookInfo[0]["CHECK_TIME"]
	//2.선후불 체크
	grpPayTy := bookInfo[0]["GRP_PAY_TY"]
	if grpPayTy == "1" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "후불 결제만 가능한 장부 입니다."))
	}

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	limitAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_AMT"])
	limitDayAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_AMT"])
	limitDayCnt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_CNT"])
	supportYn := bookInfo[0]["SUPPORT_YN"]
	supportExeedYn := bookInfo[0]["SUPPORT_EXCEED_YN"]                  // 지원금 초과사용 여부
	mySupportBalance, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 내 지원금
	orderAmt, _ := strconv.Atoi(params["orderAmt"])                     // 결제 금액

	//4.장부 제한 사항 체크
	// 결제 타입

	myPrepay := 0
	myPoint := 0
	pointUse := 0

	//5.장부 한도 체크
	linkInfo, err := cls.GetSelectData(restsql.SelectStoreLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결된 가맹점 정보가 없습니다."))
	}
	prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
	prepaidPoint, _ := strconv.Atoi(linkInfo[0]["PREPAID_POINT"])
	agrmId := linkInfo[0]["AGRM_ID"]
	pointRate := linkInfo[0]["POINT_RATE"]

	if orderAmt > prepaidAmt {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불금액이 모자랍니다. 선불금액 충전 후 주문해주세요. "))
	} else {
		myPrepay = prepaidAmt - orderAmt
		if pointRate != "0" {
			pointUse = orderPoint(orderAmt, prepaidPoint, pointRate)
			myPoint = prepaidPoint - pointUse
		}

	}

	//아이템별 가격 체크
	itemInfo, err := cls.GetSelectData(restsql.SelectStoreItem, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if itemInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품 정보가 없습니다."))
	}

	itemPrice := itemInfo[0]["ITEM_PRICE"]
	itemName := itemInfo[0]["ITEM_NM"]
	prodId := itemInfo[0]["PROD_ID"]

	if params["itemPrice"] != itemPrice {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
	}
	if prodId == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품코드가 없습니다."))
	}

	params["prodId"] = prodId

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	if media == "Benepicon" {
		// 베네피콘 api 시작

		callresultMap := benepicons.CallBenepiconOrder(params["orderNo"], prodId, "1", userTel, params["userId"])
		if callresultMap["RESULT_CD"] != "0000" {
			dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
		}

		params["ordNo"] = callresultMap["ORD_NO"]
		params["cpNo"] = callresultMap["CPNO"]
		params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 베네피콘 api 끝

	} else if media == "Wincube" {

		tokenId, resultcd := wincubes.GetWincubeAuth()
		if resultcd == "99" {
			lprintf(1, "[INFO] token recv fail \n")
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
		}

		itemCheckResult := wincubes.CallWincubeItemCheck(tokenId, prodId)
		if itemCheckResult["RESULT_CD"] != "0" {
			dprintf(1, c, "기프티콘 주문 상품 체크 실패 -- %s \n", itemCheckResult["RESULT_MSG"])
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문(상품체크)에 실패하였습니다."))
		}

		callresultMap := wincubes.CallWincubeOrder(tokenId, prodId, params["orderNo"])

		if callresultMap["RESULT_CD"] != "1000" {
			dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
		}

		params["ordNo"] = callresultMap["ORD_NO"]
		params["cpNo"] = callresultMap["CPNO"]
		params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 쿠폰 일단 저장
		temp_param := make(map[string]string)

		temp_param["orderNo"] = params["orderNo"]
		temp_param["resultCd"] = callresultMap["RESULT_CD"]
		temp_param["reason"] = callresultMap["RESULT_MSG"]
		temp_param["trID"] = callresultMap["TR_ID"]
		temp_param["ctrID"] = callresultMap["ORD_NO"]
		temp_param["pinNo"] = callresultMap["CPNO"]
		temp_param["createDateTime"] = callresultMap["EXCH_FR_DY"]
		temp_param["expirationDate"] = callresultMap["EXCH_TO_DY"]

		wincubeCouponHistory, err := cls.GetQueryJson(ordersql.InsertWincubeCoupon, temp_param)
		if err != nil {
			dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
		}
		// 쿼리 실행
		_, err = cls.QueryDB(wincubeCouponHistory)
		if err != nil {
			dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
		}

	} else {
		dprintf(1, c, "기프티콘 주문 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

	}

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback

			//기프티콘 취소
			if media == "Benepicon" {
				benepicons.CallBenepiconCancel(params["ordNo"], params["cpNo"], params["orderNo"])
			} else if media == "Wincube" {

				tokenId, resultcd := wincubes.GetWincubeAuth()
				if resultcd == "99" {
					lprintf(1, "[INFO] token recv fail \n")
				}
				wincubes.CallWincubeCancel(tokenId, params["orderNo"])
			}

			dprintf(4, c, "do rollback - 기프티콘 선불 주문 (SetOrderGifticon_PREPAY)  \n")
			tx.Rollback()
		}
	}()

	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	params["pointUse"] = strconv.Itoa(pointUse)

	qrOrderTy := params["qrOrderTy"]

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 상세 메뉴 등록
	//8.order detail insert

	params["orderSeq"] = "1"
	params["itemUserId"] = params["userId"]

	OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderDetailCreateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//========== 임시 코드==========//
	couponQuery := ""

	chkItemNo := params["restId"]
	if chkItemNo == "C0000000001" {
		couponQuery = ordersql.InsertOrderCoupon_expireUnlimit
	} else {
		couponQuery = ordersql.InsertOrderCoupon
	}
	//========== 임시 코드==========//

	InsertOrderCouponQuery, err := cls.SetUpdateParam(couponQuery, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(InsertOrderCouponQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//지원금 사용인경우
	//10.개인 지원금 차감
	if supportYn == "Y" {
		if qrOrderTy == "0" {

			//금액 제한 설정 체크
			if limitYn == "Y" {
				limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
				if limitChek == "N" {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
				}
			}

			if supportExeedYn == "N" {
				if orderAmt > mySupportBalance {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금 잔액이 부족합니다."))
				}
			}
			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 유형입니다."))
		}
	} else {
		if limitYn == "Y" {
			limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
			if limitChek == "N" {
				return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
			}
		}
	}

	//11. 충전금 차감
	params["agrmId"] = agrmId
	params["prepaidAmt"] = strconv.Itoa(myPrepay)
	params["prepaidPoint"] = strconv.Itoa(myPoint)

	LinkAmtUpdateQuery, err := cls.SetUpdateParam(restsql.UpdateLinkAmt, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(LinkAmtUpdateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	expireDate := ""
	itemImg := ""
	memo := ""
	// 주문 번호 생성
	storeEtc, err := cls.GetSelectData(restsql.SelectStoreEtc, params, c)
	if err != nil {
		//return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if storeEtc != nil {
		expireDate = storeEtc[0]["expireDate"]
		itemImg = storeEtc[0]["itemImg"]
		memo = storeEtc[0]["memo"]
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]
	data["cpNo"] = params["cpNo"]
	data["ordNo"] = params["ordNo"]
	data["itemPrice"] = itemPrice
	data["itemName"] = itemName
	data["expireDate"] = expireDate
	data["itemImg"] = itemImg
	data["itemDesc"] = memo

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

// 기프티콘 주문하기(후불)
func SetOrderGifticon_DEFERPAY(c echo.Context) error {

	dprintf(4, c, "call 기프티콘 주문하기 후불 \n")

	// 상세 주문데이터 get 끝
	params := cls.GetParamJsonMap(c)

	//1.장부 유효성 체크
	bookInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다."))
	}

	userTel := bookInfo[0]["USER_TEL"]
	limitYn := bookInfo[0]["LIMIT_YN"]
	if limitYn == "Y" {
		newLayout := "15:04"
		now := time.Now().Format(newLayout)

		chkGrpUseTime, _ := time.Parse(newLayout, now)
		grpUseTimeStart, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_START"]+":00")
		grpUseTimeEnd, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_END"]+":00")
		chkTime := inTimeSpan(grpUseTimeStart, grpUseTimeEnd, chkGrpUseTime)

		if bookInfo[0]["LIMIT_USE_TIME_START"] == "0" && bookInfo[0]["LIMIT_USE_TIME_END"] == "0" {
			chkTime = true
		}

		if chkTime == false {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 사용 가능시간은 "+bookInfo[0]["LIMIT_USE_TIME_START"]+"시 부터 "+bookInfo[0]["LIMIT_USE_TIME_END"]+"시 까지입니다."))
		}

	}

	params["checkTime"] = bookInfo[0]["CHECK_TIME"]
	//2.선후불 체크
	grpPayTy := bookInfo[0]["GRP_PAY_TY"]
	if grpPayTy == "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불 결제만 가능한 장부 입니다."))
	}

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	limitAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_AMT"])
	limitDayAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_AMT"])
	limitDayCnt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_CNT"])
	supportYn := bookInfo[0]["SUPPORT_YN"]
	supportExeedYn := bookInfo[0]["SUPPORT_EXCEED_YN"]                  // 지원금 초과사용 여부
	mySupportBalance, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 내 지원금
	orderAmt, _ := strconv.Atoi(params["orderAmt"])                     // 결제 금액

	//4.장부 제한 사항 체크
	// 결제 타입

	//myPrepay := 0
	//myPoint := 0
	pointUse := 0

	//5.장부 한도 체크
	linkInfo, err := cls.GetSelectData(restsql.SelectStoreLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결된 가맹점 정보가 없습니다."))
	}

	//아이템별 가격 체크
	itemInfo, err := cls.GetSelectData(restsql.SelectStoreItem, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if itemInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품 정보가 없습니다."))
	}

	itemPrice := itemInfo[0]["ITEM_PRICE"]
	itemName := itemInfo[0]["ITEM_NM"]
	prodId := itemInfo[0]["PROD_ID"]

	if params["itemPrice"] != itemPrice {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
	}
	if prodId == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품코드가 없습니다."))
	}

	params["prodId"] = prodId

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	if media == "Benepicon" {
		// 베네피콘 api 시작

		callresultMap := benepicons.CallBenepiconOrder(params["orderNo"], prodId, "1", userTel, params["userId"])
		if callresultMap["RESULT_CD"] != "0000" {
			dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
		}

		params["ordNo"] = callresultMap["ORD_NO"]
		params["cpNo"] = callresultMap["CPNO"]
		params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 베네피콘 api 끝

	} else if media == "Wincube" {

		tokenId, resultcd := wincubes.GetWincubeAuth()
		if resultcd == "99" {
			lprintf(1, "[INFO] token recv fail \n")
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
		}

		itemCheckResult := wincubes.CallWincubeItemCheck(tokenId, prodId)
		if itemCheckResult["RESULT_CD"] != "0" {
			dprintf(1, c, "기프티콘 주문 상품 체크 실패 -- %s \n", itemCheckResult["RESULT_MSG"])
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문(상품체크)에 실패하였습니다."))
		}

		callresultMap := wincubes.CallWincubeOrder(tokenId, prodId, params["orderNo"])

		if callresultMap["RESULT_CD"] != "1000" {
			dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
		}

		params["ordNo"] = callresultMap["ORD_NO"]
		params["cpNo"] = callresultMap["CPNO"]
		params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 쿠폰 일단 저장
		temp_param := make(map[string]string)

		temp_param["orderNo"] = params["orderNo"]
		temp_param["resultCd"] = callresultMap["RESULT_CD"]
		temp_param["reason"] = callresultMap["RESULT_MSG"]
		temp_param["trID"] = callresultMap["TR_ID"]
		temp_param["ctrID"] = callresultMap["ORD_NO"]
		temp_param["pinNo"] = callresultMap["CPNO"]
		temp_param["createDateTime"] = callresultMap["EXCH_FR_DY"]
		temp_param["expirationDate"] = callresultMap["EXCH_TO_DY"]

		wincubeCouponHistory, err := cls.GetQueryJson(ordersql.InsertWincubeCoupon, temp_param)
		if err != nil {
			dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
		}
		// 쿼리 실행
		_, err = cls.QueryDB(wincubeCouponHistory)
		if err != nil {
			dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
		}

	} else {
		dprintf(1, c, "기프티콘 주문 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

	}

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback

			//기프티콘 취소
			if media == "Benepicon" {
				benepicons.CallBenepiconCancel(params["ordNo"], params["cpNo"], params["orderNo"])
			} else if media == "Wincube" {

				tokenId, resultcd := wincubes.GetWincubeAuth()
				if resultcd == "99" {
					lprintf(1, "[INFO] token recv fail \n")
				}
				wincubes.CallWincubeCancel(tokenId, params["orderNo"])
			}

			dprintf(4, c, "do rollback - 기프티콘 후불 주문 (SetOrderGifticon_DEFERPAY)  \n")
			tx.Rollback()
		}
	}()

	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	params["pointUse"] = strconv.Itoa(pointUse)

	qrOrderTy := params["qrOrderTy"]

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 상세 메뉴 등록
	//8.order detail insert

	params["orderSeq"] = "1"
	params["itemUserId"] = params["userId"]

	OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderDetailCreateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//========== 임시 코드==========//
	couponQuery := ""

	chkItemNo := params["restId"]
	if chkItemNo == "C0000000001" {
		couponQuery = ordersql.InsertOrderCoupon_expireUnlimit
	} else {
		couponQuery = ordersql.InsertOrderCoupon
	}
	//========== 임시 코드==========//

	InsertOrderCouponQuery, err := cls.SetUpdateParam(couponQuery, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(InsertOrderCouponQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//지원금 사용인경우
	//10.개인 지원금 차감
	if supportYn == "Y" {
		if qrOrderTy == "0" {

			//금액 제한 설정 체크
			if limitYn == "Y" {
				limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
				if limitChek == "N" {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
				}
			}

			if supportExeedYn == "N" {
				if orderAmt > mySupportBalance {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금 잔액이 부족합니다."))
				}
			}
			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 유형입니다."))
		}
	} else {
		if limitYn == "Y" {
			limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, orderAmt, params["userId"], params["grpId"])
			if limitChek == "N" {
				return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
			}
		}
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	expireDate := ""
	itemImg := ""
	memo := ""
	// 주문 번호 생성
	storeEtc, err := cls.GetSelectData(restsql.SelectStoreEtc, params, c)
	if err != nil {
		//return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if storeEtc != nil {
		expireDate = storeEtc[0]["expireDate"]
		itemImg = storeEtc[0]["itemImg"]
		memo = storeEtc[0]["memo"]
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]
	data["cpNo"] = params["cpNo"]
	data["ordNo"] = params["ordNo"]
	data["itemPrice"] = itemPrice
	data["itemName"] = itemName
	data["expireDate"] = expireDate
	data["itemImg"] = itemImg
	data["itemDesc"] = memo

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

// 기프티콘 주문취소
func SetGiftOrderCancel(c echo.Context) error {

	dprintf(4, c, "call SetGiftOrderCancel\n")

	params := cls.GetParamJsonMap(c)

	orderInfo, err := cls.GetSelectData(ordersql.SelectOrder, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문 내용이 없습니다."))
	}

	payTy := orderInfo[0]["PAY_TY"]
	orderStat := orderInfo[0]["ORDER_STAT"]
	totalAmt, _ := strconv.Atoi(orderInfo[0]["TOTAL_AMT"])
	bookId := orderInfo[0]["BOOK_ID"]
	storeId := orderInfo[0]["STORE_ID"]
	userId := orderInfo[0]["USER_ID"]
	pointUse, _ := strconv.Atoi(orderInfo[0]["POINT_USE"])
	//orderDate := orderInfo[0]["ORDER_DATE"]

	params["bookId"] = bookId
	params["storeId"] = storeId
	params["userId"] = userId
	params["grpId"] = bookId
	params["restId"] = storeId

	if orderStat != "20" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소가 불가능한 주문입니다."))
	}

	couPonInfo, err := cls.GetSelectData(ordersql.SelectCouponCancelInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if couPonInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "취소 가능한 쿠폰이 없습니다."
		return c.JSON(http.StatusOK, m)
	}

	ordNo := couPonInfo[0]["ORD_NO"]
	cpNo := couPonInfo[0]["CPNO"]
	trId := params["orderNo"]

	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	if media == "Benepicon" {
		callresultMap := benepicons.CallBenepiconCancel(ordNo, cpNo, trId)
		resultCd := callresultMap["RESULT_CD"]
		if resultCd != "0000" {
			dprintf(1, c, "기프티콘 취소 실패 -- %s \n", callresultMap["RESULT_MSG"])
			m := make(map[string]interface{})
			m["resultCode"] = "99"
			m["resultMsg"] = strings.Replace(callresultMap["RESULT_MSG"], "기프티콘", "교환권", -1)
			return c.JSON(http.StatusOK, m)
		}

	} else if media == "Wincube" {

		tokenId, resultcd := wincubes.GetWincubeAuth()
		if resultcd == "99" {
			lprintf(1, "[INFO] token recv fail \n")
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
		}

		orderNo := params["orderNo"]
		callresultMap := wincubes.CallWincubeCancel(tokenId, orderNo)
		resultCd := callresultMap["RESULT_CD"]
		if resultCd != "0" {
			dprintf(1, c, "기프티콘 취소 실패 -- %s \n", callresultMap["RESULT_MSG"])
			m := make(map[string]interface{})
			m["resultCode"] = "99"
			m["resultMsg"] = strings.Replace(callresultMap["RESULT_MSG"], "기프티콘", "교환권", -1)
			return c.JSON(http.StatusOK, m)
		}

	} else {
		dprintf(1, c, "기프티콘 취소 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

	}

	// 매장 충전  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	txErr := err

	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -기프티콘 주문 취소(SetGiftOrderCancel)  \n")
			tx.Rollback()
		}
	}()

	// transation exec
	// 파라메터 맵으로 쿼리 변환

	// 선불일 경우 금액 환불
	if payTy == "0" {

		linkInfo, err := cls.GetSelectData(ordersql.SelectLinkInfo, params, c)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		if linkInfo == nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "협약이 내용이 없습니다."))
		}
		prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
		linkId := linkInfo[0]["LINK_ID"]

		// 포인트 화불
		params["linkId"] = linkId
		params["prepaidAmt"] = strconv.Itoa(prepaidAmt + totalAmt)
		prepaidPoint, _ := strconv.Atoi(linkInfo[0]["PREPAID_POINT"])
		params["prepaidPoint"] = strconv.Itoa(prepaidPoint + pointUse)

		UpdateLinkQuery, err := cls.SetUpdateParam(ordersql.UpdateLink, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(UpdateLinkQuery)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}

	//지원금 환불

	params["orderAmt"] = strconv.Itoa(totalAmt)
	UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserCancelSupportBalance, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(UserSupportBalanceUpdateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 취소
	OrderCancelQuery, err := cls.SetUpdateParam(ordersql.UpdateOrderCancel, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "OrderCancel parameter fail"))
	}
	_, err = tx.Exec(OrderCancelQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//기프티콘 취소
	UpdateCouponCancelQuery, err := cls.GetQueryJson(ordersql.UpdateCouponCancel, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdateCouponCancelQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 유저 가입 TRNAN 종료

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

//  오늘 주문 내역
func GetOrderToday_V2(c echo.Context) error {

	dprintf(4, c, "call GetOrderToday\n")

	params := cls.GetParamJsonMap(c)

	resultList, err := cls.GetSelectType(ordersql.SelectTodayOrder_V2, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultList"] = []string{}
		return c.JSON(http.StatusOK, m)

	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}

func GetOrderInfo_V2(c echo.Context) error {

	dprintf(4, c, "call GetOrderInfo_V2\n")

	params := cls.GetParamJsonMap(c)

	orderInfo, err := cls.GetSelectData(ordersql.SelectOrderInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}
	if orderInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 정보 입니다."))
	}

	couPonInfo, err := cls.GetSelectType(ordersql.SelectCouponInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	qrOrderTy := orderInfo[0]["QR_ORDER_TYPE"]

	totalMenu, err := cls.GetSelectType(ordersql.SelectOrderDetail, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
	}

	//	userList := make([]map[string]string)

	if qrOrderTy == "2" {

		userSplitList, err := cls.GetSelectType(ordersql.SelectOrderUserSplitAmt, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
		}

		order := make(map[string]interface{})
		order["orderNo"] = orderInfo[0]["ORDER_NO"]
		order["restNm"] = orderInfo[0]["REST_NM"]
		order["grpNm"] = orderInfo[0]["GRP_NM"]
		totalAmt, _ := strconv.Atoi(orderInfo[0]["TOTAL_AMT"])
		order["totalAmt"] = totalAmt
		order["orderStat"] = orderInfo[0]["ORDER_STAT"]
		order["orderDate"] = orderInfo[0]["ORDER_DATE"]
		order["totalMenu"] = totalMenu
		order["usersList"] = userSplitList
		order["couponInfo"] = empty{}

		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = order

		return c.JSON(http.StatusOK, m)

	} else {

		userDetail, err := cls.GetSelectData(ordersql.SelectOrderUserDetail, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult(err.Error(), "DB fail"))
		}

		userList := make([]map[string]interface{}, len(userDetail))
		for i := range userDetail {

			params["userId"] = userDetail[i]["USER_ID"]
			userMenu, err := cls.GetSelectType(ordersql.SelectOrderUserMenu, params, c)
			if err != nil {
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

			pOrderAmt, _ := strconv.Atoi(userDetail[i]["ORDER_AMT"])
			order2 := make(map[string]interface{})
			order2["userNm"] = userDetail[i]["USER_NM"]
			order2["orderAmt"] = pOrderAmt
			order2["menus"] = userMenu
			order2["memo"] = userDetail[i]["MEMO"]
			userList[i] = order2

		}

		order := make(map[string]interface{})
		order["orderNo"] = orderInfo[0]["ORDER_NO"]
		order["restNm"] = orderInfo[0]["REST_NM"]
		order["grpNm"] = orderInfo[0]["GRP_NM"]
		totalAmt, _ := strconv.Atoi(orderInfo[0]["TOTAL_AMT"])
		order["totalAmt"] = totalAmt
		order["orderStat"] = orderInfo[0]["ORDER_STAT"]
		order["orderDate"] = orderInfo[0]["ORDER_DATE"]
		order["totalMenu"] = totalMenu
		order["usersList"] = userList
		if couPonInfo == nil {
			order["couponInfo"] = empty{}
		} else {
			order["couponInfo"] = couPonInfo[0]
		}

		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = order

		return c.JSON(http.StatusOK, m)

	}

}

//  기프티콘 불러오기
func GetGiftiInfo(c echo.Context) error {

	dprintf(4, c, "call GetGiftiInfo\n")

	params := cls.GetParamJsonMap(c)

	resultData, err := cls.GetSelectData(ordersql.SelectCoupon, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "사용 가능한 쿠폰이 없습니다."
		return c.JSON(http.StatusOK, m)
	}

	params["restId"] = resultData[0]["restId"]
	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	ordNo := resultData[0]["ordNo"]
	cpNo := resultData[0]["cpNo"]
	params["cpNo"] = cpNo

	orderNo := params["orderNo"]

	if media == "Benepicon" {

		chkResult := benepicons.CallBenepiconCheck(ordNo, cpNo, orderNo)
		cpnoStatusCd := chkResult["CPNO_STATUS_CD"]
		params["exchplc"] = chkResult["EXCHPLC"]
		params["exchcoNm"] = chkResult["EXCHCO_NM"]
		params["cpnoExchDt"] = chkResult["CPNO_EXCH_DT"]
		params["cpnoStatus"] = chkResult["CPNO_STATUS"]
		params["cpnoStatusCd"] = cpnoStatusCd
		params["balance"] = chkResult["BALANCE"]

		switch cpnoStatusCd {

		case "IAD10": //IAD10 발행완료
			params["cpStatus"] = "0"
			break
		case "IAD03": //IAD03 쿠폰 사용불가-주문취소
			params["cpStatus"] = "1"
			break
		case "IAD05": //쿠폰사용불가_유효기간 경과
			params["cpStatus"] = "9"
			break
		case "IAD06": //쿠폰사용불가_환불
			params["cpStatus"] = "1"
			break
		case "IAD09": //교환완료
			params["cpStatus"] = "2"
			break
		case "IAD08": //다운로드완료
			params["cpStatus"] = "0"
			break
		}

	} else if media == "Wincube" {

		tokenId, resultcd := wincubes.GetWincubeAuth()
		if resultcd == "99" {
			lprintf(1, "[INFO] token recv fail \n")
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
		}
		chkResult := wincubes.CallWincubeCouponStatus(tokenId, orderNo)
		cpnoResultCd := chkResult["RESULT_CD"]
		params["exchplc"] = chkResult["EXCHPLC"]
		params["exchcoNm"] = chkResult["EXCHCO_NM"]
		params["cpnoExchDt"] = chkResult["CPNO_EXCH_DT"]
		params["cpnoStatus"] = chkResult["CPNO_STATUS"]
		params["cpnoStatusCd"] = chkResult["CPNO_STATUS_CD"]
		params["balance"] = chkResult["BALANCE"]

		switch cpnoResultCd {

		case "0": //
			params["cpStatus"] = "0"
			break
		case "4005": //이미 취소된 상품입니다.
			params["cpStatus"] = "1"
			break
		case "4006": //교환된 상품으로 취소가 불가합니다.
			params["cpStatus"] = "2"
			break
		case "4007": //상품권의 기간이 만료되었습니다.
			params["cpStatus"] = "9"
			break
		case "3201": //매체아이디가 존재하지않습니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "매체아이디가 존재하지않습니다."))
			break
		case "3901": //고유아이디가 존재하지 않습니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "고유아이디가 존재하지 않습니다."))
			break
		case "3912": //유효한 매체코드가 아닙니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "유효한 매체코드가 아닙니다."))
			break
		case "2001": //유효한 구매번호가 아닙니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "유효한 구매번호가 아닙니다."))
			break
		case "4004": //상품권 조회 불가상태 기타
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품권 조회 불가상태 기타 ."))
			break
		}

	} else {
		dprintf(1, c, "기프티콘 주문 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))
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

	m := make(map[string]interface{})

	if params["cpStatus"] == "2" {
		m["resultCode"] = "99"
		m["resultMsg"] = "사용이 완료된 교환권입니다."
	} else if params["cpStatus"] == "1" {
		m["resultCode"] = "99"
		m["resultMsg"] = "취소된 교환권입니다."
	} else if params["cpStatus"] == "0" {
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		data := make(map[string]interface{})
		data["couponInfo"] = resultData[0]
		m["resultData"] = data
	}

	return c.JSON(http.StatusOK, m)

}

//  기프티콘 불러오기
func GetMyGiftList(c echo.Context) error {

	dprintf(4, c, "call GetMyGiftList\n")

	params := cls.GetParamJsonMap(c)

	resultList, err := cls.GetSelectTypeRequire(ordersql.SelectMyGiftList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {

		return c.JSON(http.StatusOK, []string{})
	}

	return c.JSON(http.StatusOK, resultList)

}

// 기프티콘 주문 2022.05.10 즉시결제 포함
func SetOrderGifticon_V2(c echo.Context) error {

	payTy := c.FormValue("payTy")

	unlink := c.FormValue("unlink")

	if unlink == "Y" {
		return SetOrderGifticon_UNLINK(c)
	}

	if payTy == "0" {
		//선불결제
		return SetOrderGifticon_PREPAY_V2(c)
	} else if payTy == "1" {
		//후불 결제
		return SetOrderGifticon_DEFERPAY_V2(c)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "99"
	m["resultMsg"] = "잘못된 호출입니다."
	return c.JSON(http.StatusOK, m)
}

// 주문하기(선불) - 기프티콘 즉시결제 포함
func SetOrderGifticon_PREPAY_V2(c echo.Context) error {

	dprintf(4, c, "call 주문하기(선불) - 기프티콘 즉시결제 포함 \n")

	params := cls.GetParamJsonMap(c)

	//1.장부 유효성 체크
	bookInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다."))
	}

	//userTel := bookInfo[0]["USER_TEL"]
	limitYn := bookInfo[0]["LIMIT_YN"]
	if limitYn == "Y" {
		newLayout := "15:04"
		now := time.Now().Format(newLayout)

		chkGrpUseTime, _ := time.Parse(newLayout, now)
		grpUseTimeStart, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_START"]+":00")
		grpUseTimeEnd, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_END"]+":00")
		chkTime := inTimeSpan(grpUseTimeStart, grpUseTimeEnd, chkGrpUseTime)

		if bookInfo[0]["LIMIT_USE_TIME_START"] == "0" && bookInfo[0]["LIMIT_USE_TIME_END"] == "0" {
			chkTime = true
		}

		if chkTime == false {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 사용 가능시간은 "+bookInfo[0]["LIMIT_USE_TIME_START"]+"시 부터 "+bookInfo[0]["LIMIT_USE_TIME_END"]+"시 까지입니다."))
		}

	}

	params["checkTime"] = bookInfo[0]["CHECK_TIME"]
	//2.선후불 체크
	grpPayTy := bookInfo[0]["GRP_PAY_TY"]
	if grpPayTy == "1" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "후불 결제만 가능한 장부 입니다."))
	}

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	limitAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_AMT"])
	limitDayAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_AMT"])
	limitDayCnt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_CNT"])
	supportYn := bookInfo[0]["SUPPORT_YN"]
	supportExeedYn := bookInfo[0]["SUPPORT_EXCEED_YN"]                  // 지원금 초과사용 여부
	mySupportBalance, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 내 지원금
	orderAmt, _ := strconv.Atoi(params["orderAmt"])                     // 결제 금액

	//4.장부 제한 사항 체크
	// 결제 타입

	myPrepay := 0
	myPoint := 0
	pointUse := 0

	//5.장부 한도 체크
	linkInfo, err := cls.GetSelectData(restsql.SelectStoreLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결된 가맹점 정보가 없습니다."))
	}
	prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
	prepaidPoint, _ := strconv.Atoi(linkInfo[0]["PREPAID_POINT"])
	agrmId := linkInfo[0]["AGRM_ID"]
	pointRate := linkInfo[0]["POINT_RATE"]

	if orderAmt > prepaidAmt {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불금액이 모자랍니다. 선불금액 충전 후 주문해주세요. "))
	} else {
		myPrepay = prepaidAmt - orderAmt
		if pointRate != "0" {
			pointUse = orderPoint(orderAmt, prepaidPoint, pointRate)
			myPoint = prepaidPoint - pointUse
		}

	}

	//아이템별 가격 체크
	itemInfo, err := cls.GetSelectData(restsql.SelectStoreItem, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if itemInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품 정보가 없습니다."))
	}

	itemPrice := itemInfo[0]["ITEM_PRICE"]
	itemName := itemInfo[0]["ITEM_NM"]
	prodId := itemInfo[0]["PROD_ID"]

	if params["itemPrice"] != itemPrice {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
	}
	if prodId == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품코드가 없습니다."))
	}

	params["prodId"] = prodId

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	if media == "Benepicon" {
		// 베네피콘 api 시작

		/*
			callresultMap := benepicons.CallBenepiconOrder(params["orderNo"], prodId, "1", userTel, params["userId"])
			if callresultMap["RESULT_CD"] != "0000" {
				dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
			}

			params["ordNo"] = callresultMap["ORD_NO"]
			params["cpNo"] = callresultMap["CPNO"]
			params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
			params["exchToDy"] = callresultMap["EXCH_TO_DY"]
			params["cpStatus"] = "0"
		*/

		params["ordNo"] = "11122"       //callresultMap["ORD_NO"]
		params["cpNo"] = "1231231231"   //callresultMap["CPNO"]
		params["exchFrDy"] = "20220510" // callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = "20221122" //callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 베네피콘 api 끝

	} else if media == "Wincube" {

		/*
			tokenId, resultcd := wincubes.GetWincubeAuth()
			if resultcd == "99" {
				lprintf(1, "[INFO] token recv fail \n")
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
			}

			itemCheckResult := wincubes.CallWincubeItemCheck(tokenId, prodId)
			if itemCheckResult["RESULT_CD"] != "0" {
				dprintf(1, c, "기프티콘 주문 상품 체크 실패 -- %s \n", itemCheckResult["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문(상품체크)에 실패하였습니다."))
			}

			callresultMap := wincubes.CallWincubeOrder(tokenId, prodId, params["orderNo"])

			if callresultMap["RESULT_CD"] != "1000" {
				dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
			}



			params["ordNo"] = callresultMap["ORD_NO"]
			params["cpNo"] = callresultMap["CPNO"]
			params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
			params["exchToDy"] = callresultMap["EXCH_TO_DY"]
			params["cpStatus"] = "0"

			// 쿠폰 일단 저장
			temp_param := make(map[string]string)

			temp_param["orderNo"] = params["orderNo"]
			temp_param["resultCd"] = callresultMap["RESULT_CD"]
			temp_param["reason"] = callresultMap["RESULT_MSG"]
			temp_param["trID"] = callresultMap["TR_ID"]
			temp_param["ctrID"] = callresultMap["ORD_NO"]
			temp_param["pinNo"] = callresultMap["CPNO"]
			temp_param["createDateTime"] = callresultMap["EXCH_FR_DY"]
			temp_param["expirationDate"] = callresultMap["EXCH_TO_DY"]

			wincubeCouponHistory, err := cls.GetQueryJson(ordersql.InsertWincubeCoupon, temp_param)
			if err != nil {
				dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
			}
			// 쿼리 실행
			_, err = cls.QueryDB(wincubeCouponHistory)
			if err != nil {
				dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
			}

		*/
		params["ordNo"] = "11122"       //callresultMap["ORD_NO"]
		params["cpNo"] = "1231231231"   //callresultMap["CPNO"]
		params["exchFrDy"] = "20220510" // callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = "20221122" //callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

	} else {
		dprintf(1, c, "기프티콘 주문 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

	}

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback

			//기프티콘 취소
			if media == "Benepicon" {
				benepicons.CallBenepiconCancel(params["ordNo"], params["cpNo"], params["orderNo"])
			} else if media == "Wincube" {

				tokenId, resultcd := wincubes.GetWincubeAuth()
				if resultcd == "99" {
					lprintf(1, "[INFO] token recv fail \n")
				}
				wincubes.CallWincubeCancel(tokenId, params["orderNo"])
			}

			dprintf(1, c, "do rollback - 기프티콘 선불 주문 (SetOrderGifticon_PREPAY_V2)  \n")
			tx.Rollback()
		}
	}()

	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	params["pointUse"] = strconv.Itoa(pointUse)

	remainAmt, _ := strconv.Atoi(params["remainAmt"])
	chargeAmt, _ := strconv.Atoi(params["chargeAmt"])

	qrOrderTy := params["qrOrderTy"]

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 상세 메뉴 등록
	//8.order detail insert

	params["orderSeq"] = "1"
	params["itemUserId"] = params["userId"]

	OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderDetailCreateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//========== 임시 코드==========//
	couponQuery := ""

	chkItemNo := params["restId"]
	if chkItemNo == "C0000000001" {
		couponQuery = ordersql.InsertOrderCoupon_expireUnlimit
	} else {
		couponQuery = ordersql.InsertOrderCoupon
	}
	//========== 임시 코드==========//

	InsertOrderCouponQuery, err := cls.SetUpdateParam(couponQuery, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(InsertOrderCouponQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//지원금 사용인경우
	//10.개인 지원금 차감
	if supportYn == "Y" {
		if qrOrderTy == "0" {

			//금액 제한 설정 체크
			if limitYn == "Y" {
				limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, remainAmt, params["userId"], params["grpId"])
				if limitChek == "N" {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
				}
			}

			if supportExeedYn == "N" {
				if orderAmt > mySupportBalance+chargeAmt {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금 잔액이 부족합니다."))
				}
			}
			params["orderAmt"] = strconv.Itoa(remainAmt)
			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 유형입니다."))
		}
	} else {
		if limitYn == "Y" {
			limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, remainAmt, params["userId"], params["grpId"])
			if limitChek == "N" {
				return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
			}
		}
	}

	//11. 충전금 차감
	params["agrmId"] = agrmId
	params["prepaidAmt"] = strconv.Itoa(myPrepay)
	params["prepaidPoint"] = strconv.Itoa(myPoint)

	LinkAmtUpdateQuery, err := cls.SetUpdateParam(restsql.UpdateLinkAmt, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(LinkAmtUpdateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	expireDate := ""
	itemImg := ""
	memo := ""
	// 주문 번호 생성
	storeEtc, err := cls.GetSelectData(restsql.SelectStoreEtc, params, c)
	if err != nil {
		//return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if storeEtc != nil {
		expireDate = storeEtc[0]["expireDate"]
		itemImg = storeEtc[0]["itemImg"]
		memo = storeEtc[0]["memo"]
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]
	data["cpNo"] = params["cpNo"]
	data["ordNo"] = params["ordNo"]
	data["itemPrice"] = itemPrice
	data["itemName"] = itemName
	data["expireDate"] = expireDate
	data["itemImg"] = itemImg
	data["itemDesc"] = memo

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

// 기프티콘 주문하기(후불) - 즉시결제 포함
func SetOrderGifticon_DEFERPAY_V2(c echo.Context) error {

	dprintf(4, c, "call 기프티콘 주문하기 후불 \n")

	// 상세 주문데이터 get 끝
	params := cls.GetParamJsonMap(c)

	//1.장부 유효성 체크
	bookInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "사용가능한 장부가 없습니다."))
	}

	//userTel := bookInfo[0]["USER_TEL"]
	limitYn := bookInfo[0]["LIMIT_YN"]
	if limitYn == "Y" {
		newLayout := "15:04"
		now := time.Now().Format(newLayout)

		chkGrpUseTime, _ := time.Parse(newLayout, now)
		grpUseTimeStart, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_START"]+":00")
		grpUseTimeEnd, _ := time.Parse(newLayout, bookInfo[0]["LIMIT_USE_TIME_END"]+":00")
		chkTime := inTimeSpan(grpUseTimeStart, grpUseTimeEnd, chkGrpUseTime)

		if bookInfo[0]["LIMIT_USE_TIME_START"] == "0" && bookInfo[0]["LIMIT_USE_TIME_END"] == "0" {
			chkTime = true
		}

		if chkTime == false {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 사용 가능시간은 "+bookInfo[0]["LIMIT_USE_TIME_START"]+"시 부터 "+bookInfo[0]["LIMIT_USE_TIME_END"]+"시 까지입니다."))
		}

	}

	params["checkTime"] = bookInfo[0]["CHECK_TIME"]
	//2.선후불 체크
	grpPayTy := bookInfo[0]["GRP_PAY_TY"]
	if grpPayTy == "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불 결제만 가능한 장부 입니다."))
	}

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	limitAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_AMT"])
	limitDayAmt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_AMT"])
	limitDayCnt, _ := strconv.Atoi(bookInfo[0]["LIMIT_DAY_CNT"])
	supportYn := bookInfo[0]["SUPPORT_YN"]
	supportExeedYn := bookInfo[0]["SUPPORT_EXCEED_YN"]                  // 지원금 초과사용 여부
	mySupportBalance, _ := strconv.Atoi(bookInfo[0]["SUPPORT_BALANCE"]) // 내 지원금
	orderAmt, _ := strconv.Atoi(params["orderAmt"])                     // 결제 금액

	//4.장부 제한 사항 체크
	// 결제 타입

	//myPrepay := 0
	//myPoint := 0
	pointUse := 0

	//5.장부 한도 체크
	linkInfo, err := cls.GetSelectData(restsql.SelectStoreLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결된 가맹점 정보가 없습니다."))
	}

	//아이템별 가격 체크
	itemInfo, err := cls.GetSelectData(restsql.SelectStoreItem, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if itemInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품 정보가 없습니다."))
	}

	itemPrice := itemInfo[0]["ITEM_PRICE"]
	itemName := itemInfo[0]["ITEM_NM"]
	prodId := itemInfo[0]["PROD_ID"]

	if params["itemPrice"] != itemPrice {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
	}
	if prodId == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품코드가 없습니다."))
	}

	params["prodId"] = prodId

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	if media == "Benepicon" {
		// 베네피콘 api 시작

		//callresultMap := benepicons.CallBenepiconOrder(params["orderNo"], prodId, "1", userTel, params["userId"])
		//if callresultMap["RESULT_CD"] != "0000" {
		//	dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
		//	return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
		//}

		params["ordNo"] = "11122"       //callresultMap["ORD_NO"]
		params["cpNo"] = "1231231231"   //callresultMap["CPNO"]
		params["exchFrDy"] = "20220510" // callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = "20221122" //callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 베네피콘 api 끝

	} else if media == "Wincube" {

		/*
			tokenId, resultcd := wincubes.GetWincubeAuth()
			if resultcd == "99" {
				lprintf(1, "[INFO] token recv fail \n")
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
			}

			itemCheckResult := wincubes.CallWincubeItemCheck(tokenId, prodId)
			if itemCheckResult["RESULT_CD"] != "0" {
				dprintf(1, c, "기프티콘 주문 상품 체크 실패 -- %s \n", itemCheckResult["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문(상품체크)에 실패하였습니다."))
			}

			callresultMap := wincubes.CallWincubeOrder(tokenId, prodId, params["orderNo"])

			if callresultMap["RESULT_CD"] != "1000" {
				dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
			}

			params["ordNo"] = callresultMap["ORD_NO"]
			params["cpNo"] = callresultMap["CPNO"]
			params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
			params["exchToDy"] = callresultMap["EXCH_TO_DY"]
			params["cpStatus"] = "0"


			// 쿠폰 일단 저장
			temp_param := make(map[string]string)

			temp_param["orderNo"] = params["orderNo"]
			temp_param["resultCd"] = callresultMap["RESULT_CD"]
			temp_param["reason"] = callresultMap["RESULT_MSG"]
			temp_param["trID"] = callresultMap["TR_ID"]
			temp_param["ctrID"] = callresultMap["ORD_NO"]
			temp_param["pinNo"] = callresultMap["CPNO"]
			temp_param["createDateTime"] = callresultMap["EXCH_FR_DY"]
			temp_param["expirationDate"] = callresultMap["EXCH_TO_DY"]

			wincubeCouponHistory, err := cls.GetQueryJson(ordersql.InsertWincubeCoupon, temp_param)
			if err != nil {
				dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
			}
			// 쿼리 실행
			_, err = cls.QueryDB(wincubeCouponHistory)
			if err != nil {
				dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
			}

		*/

		params["ordNo"] = "11122"       //callresultMap["ORD_NO"]
		params["cpNo"] = "1231231231"   //callresultMap["CPNO"]
		params["exchFrDy"] = "20220510" // callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = "20221122" //callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

	} else {
		dprintf(1, c, "기프티콘 주문 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

	}

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback

			//기프티콘 취소
			if media == "Benepicon" {
				benepicons.CallBenepiconCancel(params["ordNo"], params["cpNo"], params["orderNo"])
			} else if media == "Wincube" {

				tokenId, resultcd := wincubes.GetWincubeAuth()
				if resultcd == "99" {
					lprintf(1, "[INFO] token recv fail \n")
				}
				wincubes.CallWincubeCancel(tokenId, params["orderNo"])
			}

			dprintf(1, c, "do rollback - 기프티콘 후불 주문 (SetOrderGifticon_DEFERPAY_V2)  \n")
			tx.Rollback()
		}
	}()

	remainAmt, _ := strconv.Atoi(params["remainAmt"])
	//chargeAmt:= params["chargeAmt"]
	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	params["pointUse"] = strconv.Itoa(pointUse)

	qrOrderTy := params["qrOrderTy"]

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 상세 메뉴 등록
	//8.order detail insert

	params["orderSeq"] = "1"
	params["itemUserId"] = params["userId"]

	OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderDetailCreateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//========== 임시 코드==========//
	couponQuery := ""

	chkItemNo := params["restId"]
	if chkItemNo == "C0000000001" {
		couponQuery = ordersql.InsertOrderCoupon_expireUnlimit
	} else {
		couponQuery = ordersql.InsertOrderCoupon
	}
	//========== 임시 코드==========//

	InsertOrderCouponQuery, err := cls.SetUpdateParam(couponQuery, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(InsertOrderCouponQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//지원금 사용인경우
	//10.개인 지원금 차감
	if supportYn == "Y" {
		if qrOrderTy == "0" {

			//금액 제한 설정 체크
			if limitYn == "Y" {
				limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, remainAmt, params["userId"], params["grpId"])
				if limitChek == "N" {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
				}
			}

			if supportExeedYn == "N" {
				if orderAmt > mySupportBalance {
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원금 잔액이 부족합니다."))
				}
			}

			params["orderAmt"] = strconv.Itoa(remainAmt)
			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 주문 유형입니다."))
		}
	} else {
		if limitYn == "Y" {
			limitChek, eMsG := grpUseAmtCheck(limitAmt, limitDayAmt, limitDayCnt, remainAmt, params["userId"], params["grpId"])
			if limitChek == "N" {
				return c.JSON(http.StatusOK, controller.SetErrResult("99", eMsG))
			}
		}
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	expireDate := ""
	itemImg := ""
	memo := ""
	// 주문 번호 생성
	storeEtc, err := cls.GetSelectData(restsql.SelectStoreEtc, params, c)
	if err != nil {
		//return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if storeEtc != nil {
		expireDate = storeEtc[0]["expireDate"]
		itemImg = storeEtc[0]["itemImg"]
		memo = storeEtc[0]["memo"]
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]
	data["cpNo"] = params["cpNo"]
	data["ordNo"] = params["ordNo"]
	data["itemPrice"] = itemPrice
	data["itemName"] = itemName
	data["expireDate"] = expireDate
	data["itemImg"] = itemImg
	data["itemDesc"] = memo

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

// 미연결 주문  - 기프티콘
func SetOrderGifticon_UNLINK(c echo.Context) error {

	dprintf(4, c, "call 주문하기(미연결) - 기프티콘 \n")

	params := cls.GetParamJsonMap(c)

	//미연결 고정값
	params["checkTime"] = "5"
	params["orderTy"] = "6"
	params["grpId"] = "00000000000"

	//3.(중복) 시간내 같은 거래 주문 체크(기본 설정 5초)
	orderCheck, err := cls.GetSelectData(ordersql.SelectOrderCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderCheck != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "중복 주문입니다."))
	}

	pointUse := 0

	//아이템별 가격 체크
	itemInfo, err := cls.GetSelectData(restsql.SelectStoreItem, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if itemInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품 정보가 없습니다."))
	}

	itemPrice := itemInfo[0]["ITEM_PRICE"]
	itemName := itemInfo[0]["ITEM_NM"]
	prodId := itemInfo[0]["PROD_ID"]

	if params["itemPrice"] != itemPrice {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선택한 메뉴의 가격이 다릅니다."))
	}
	if prodId == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품코드가 없습니다."))
	}

	params["prodId"] = prodId

	// 주문 번호 생성
	orderSeq, err := cls.GetSelectData(ordersql.SelectCreateOrderSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderSeq == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문번호 생성 실패(2)"))
	}
	params["orderNo"] = orderSeq[0]["orderSeq"] + params["userId"]

	mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	media := mediaInfo[0]["REST_NM"]

	if media == "Benepicon" {
		// 베네피콘 api 시작

		/*
			callresultMap := benepicons.CallBenepiconOrder(params["orderNo"], prodId, "1", userTel, params["userId"])
			if callresultMap["RESULT_CD"] != "0000" {
				dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
			}

			params["ordNo"] = callresultMap["ORD_NO"]
			params["cpNo"] = callresultMap["CPNO"]
			params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
			params["exchToDy"] = callresultMap["EXCH_TO_DY"]
			params["cpStatus"] = "0"
		*/

		params["ordNo"] = "11122"       //callresultMap["ORD_NO"]
		params["cpNo"] = "1231231231"   //callresultMap["CPNO"]
		params["exchFrDy"] = "20220510" // callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = "20221122" //callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

		// 베네피콘 api 끝

	} else if media == "Wincube" {

		/*
			tokenId, resultcd := wincubes.GetWincubeAuth()
			if resultcd == "99" {
				lprintf(1, "[INFO] token recv fail \n")
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
			}

			itemCheckResult := wincubes.CallWincubeItemCheck(tokenId, prodId)
			if itemCheckResult["RESULT_CD"] != "0" {
				dprintf(1, c, "기프티콘 주문 상품 체크 실패 -- %s \n", itemCheckResult["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문(상품체크)에 실패하였습니다."))
			}

			callresultMap := wincubes.CallWincubeOrder(tokenId, prodId, params["orderNo"])

			if callresultMap["RESULT_CD"] != "1000" {
				dprintf(1, c, "기프티콘 주문 실패 -- %s \n", callresultMap["RESULT_MSG"])
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "기프티콘 주문에 실패하였습니다."))
			}



			params["ordNo"] = callresultMap["ORD_NO"]
			params["cpNo"] = callresultMap["CPNO"]
			params["exchFrDy"] = callresultMap["EXCH_FR_DY"]
			params["exchToDy"] = callresultMap["EXCH_TO_DY"]
			params["cpStatus"] = "0"

			// 쿠폰 일단 저장
			temp_param := make(map[string]string)

			temp_param["orderNo"] = params["orderNo"]
			temp_param["resultCd"] = callresultMap["RESULT_CD"]
			temp_param["reason"] = callresultMap["RESULT_MSG"]
			temp_param["trID"] = callresultMap["TR_ID"]
			temp_param["ctrID"] = callresultMap["ORD_NO"]
			temp_param["pinNo"] = callresultMap["CPNO"]
			temp_param["createDateTime"] = callresultMap["EXCH_FR_DY"]
			temp_param["expirationDate"] = callresultMap["EXCH_TO_DY"]

			wincubeCouponHistory, err := cls.GetQueryJson(ordersql.InsertWincubeCoupon, temp_param)
			if err != nil {
				dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
			}
			// 쿼리 실행
			_, err = cls.QueryDB(wincubeCouponHistory)
			if err != nil {
				dprintf(1, c, "wincube 기프티콘 저장 실패 -- %s \n")
			}

		*/
		params["ordNo"] = "11122"       //callresultMap["ORD_NO"]
		params["cpNo"] = "1231231231"   //callresultMap["CPNO"]
		params["exchFrDy"] = "20220510" // callresultMap["EXCH_FR_DY"]
		params["exchToDy"] = "20221122" //callresultMap["EXCH_TO_DY"]
		params["cpStatus"] = "0"

	} else {
		dprintf(1, c, "기프티콘 주문 실패 -- 업체가 설정되지 않았습니다. %s \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

	}

	// 주문등록  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}
	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback

			//기프티콘 취소
			if media == "Benepicon" {
				benepicons.CallBenepiconCancel(params["ordNo"], params["cpNo"], params["orderNo"])
			} else if media == "Wincube" {

				tokenId, resultcd := wincubes.GetWincubeAuth()
				if resultcd == "99" {
					lprintf(1, "[INFO] token recv fail \n")
				}
				wincubes.CallWincubeCancel(tokenId, params["orderNo"])
			}

			dprintf(1, c, "do rollback - 주문하기(미연결) - 기프티콘 (SetOrderGifticon_UNLINK)  \n")
			tx.Rollback()
		}
	}()

	params["creditAmt"] = params["orderAmt"]
	params["discount"] = "0"
	params["pointUse"] = strconv.Itoa(pointUse)

	params["moid"] = params["instantMoid"]

	//주문 생성
	//7.order Insert
	OrderCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrder, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", OrderCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//주문 상세 메뉴 등록
	//8.order detail insert

	params["orderSeq"] = "1"
	params["itemUserId"] = params["userId"]

	OrderDetailCreateQuery, err := cls.SetUpdateParam(ordersql.InsertOrderDetail, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(OrderDetailCreateQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	//========== 임시 코드==========//
	couponQuery := ""
	chkItemNo := params["restId"]
	if chkItemNo == "C0000000001" {
		couponQuery = ordersql.InsertOrderCoupon_expireUnlimit
	} else {
		couponQuery = ordersql.InsertOrderCoupon
	}
	//========== 임시 코드==========//

	InsertOrderCouponQuery, err := cls.SetUpdateParam(couponQuery, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(InsertOrderCouponQuery)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	expireDate := ""
	itemImg := ""
	memo := ""
	// 주문 번호 생성
	storeEtc, err := cls.GetSelectData(restsql.SelectStoreEtc, params, c)
	if err != nil {
		//return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if storeEtc != nil {
		expireDate = storeEtc[0]["expireDate"]
		itemImg = storeEtc[0]["itemImg"]
		memo = storeEtc[0]["memo"]
	}

	data := make(map[string]string)
	data["orderNo"] = params["orderNo"]
	data["cpNo"] = params["cpNo"]
	data["ordNo"] = params["ordNo"]
	data["itemPrice"] = itemPrice
	data["itemName"] = itemName
	data["expireDate"] = expireDate
	data["itemImg"] = itemImg
	data["itemDesc"] = memo

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = data

	return c.JSON(http.StatusOK, m)

}

func SetOrderGifticon_V2_Cancel(c echo.Context) error {

	dprintf(4, c, "call SetOrderGifticon_V2_Cancel\n")

	params := cls.GetParamJsonMap(c)

	orderInfo, err := cls.GetSelectData(ordersql.SelectOrder, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if orderInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "주문 내용이 없습니다."))
	}

	payTy := orderInfo[0]["PAY_TY"]
	orderStat := orderInfo[0]["ORDER_STAT"]
	totalAmt, _ := strconv.Atoi(orderInfo[0]["TOTAL_AMT"])
	bookId := orderInfo[0]["BOOK_ID"]
	storeId := orderInfo[0]["STORE_ID"]
	userId := orderInfo[0]["USER_ID"]
	pointUse, _ := strconv.Atoi(orderInfo[0]["POINT_USE"])
	//orderDate := orderInfo[0]["ORDER_DATE"]

	intantMoid := orderInfo[0]["INSTANT_MOID"]
	intantAmt, _ := strconv.Atoi(orderInfo[0]["INSTANT_AMT"])
	orderTy := orderInfo[0]["ORDER_TY"]

	params["bookId"] = bookId
	params["storeId"] = storeId
	params["userId"] = userId
	params["grpId"] = bookId
	params["restId"] = storeId

	if orderStat != "20" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소가 불가능한 주문입니다."))
	}

	params["moid"] = intantMoid
	paymentInfo, err := cls.GetSelectDataRequire(paymentsql.SelectPaymentInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if paymentInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "결제 정보가 없습니다."))
	}
	payInfo := paymentInfo[0]["PAY_INFO"]

	couPonInfo, err := cls.GetSelectData(ordersql.SelectCouponCancelInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if couPonInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "취소 가능한 쿠폰이 없습니다."
		return c.JSON(http.StatusOK, m)
	}

	//ordNo := couPonInfo[0]["ORD_NO"]
	//cpNo := couPonInfo[0]["CPNO"]
	//trId := params["orderNo"]

	/*
		mediaInfo, err := cls.GetSelectData(ordersql.SelectGiftMedia, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		media := mediaInfo[0]["REST_NM"]

		if media == "Benepicon" {
			callresultMap := benepicons.CallBenepiconCancel(ordNo, cpNo, trId)
			resultCd := callresultMap["RESULT_CD"]
			if resultCd != "0000" {
				dprintf(1, c, "기프티콘 취소 실패 -- %s \n", callresultMap["RESULT_MSG"])
				m := make(map[string]interface{})
				m["resultCode"] = "99"
				m["resultMsg"] = strings.Replace(callresultMap["RESULT_MSG"], "기프티콘", "교환권", -1)
				return c.JSON(http.StatusOK, m)
			}

		} else if media == "Wincube" {

			tokenId, resultcd := wincubes.GetWincubeAuth()
			if resultcd == "99" {
				lprintf(1, "[INFO] token recv fail \n")
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
			}

			orderNo := params["orderNo"]
			callresultMap := wincubes.CallWincubeCancel(tokenId, orderNo)
			resultCd := callresultMap["RESULT_CD"]
			if resultCd != "0" {
				dprintf(1, c, "기프티콘 취소 실패 -- %s \n", callresultMap["RESULT_MSG"])
				m := make(map[string]interface{})
				m["resultCode"] = "99"
				m["resultMsg"] = strings.Replace(callresultMap["RESULT_MSG"], "기프티콘", "교환권", -1)
				return c.JSON(http.StatusOK, m)
			}

		} else {
			dprintf(1, c, "기프티콘 취소 실패 -- 업체가 설정되지 않았습니다. %s \n")
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "업체가 설정되지 않았습니다."))

		}
	*/

	resultCode := ""
	resultMsg := ""
	CancelDate := ""
	CancelTime := ""
	Canceltid := ""

	if payInfo == "5" {
		tpayCancel := tpays.BillingPayCancel(intantMoid, strconv.Itoa(intantAmt), "사용자취소", paymentInfo[0]["TID"])
		resultCode = tpayCancel["resultCode"]
		resultMsg = tpayCancel["resultMsg"]
		CancelDate = tpayCancel["CancelDate"]
		CancelTime = tpayCancel["CancelTime"]
		Canceltid = tpayCancel["Canceltid"]

	} else {

		tpayCancel := tpays.PayCancel(intantMoid, strconv.Itoa(intantAmt), "사용자취소", "0", paymentInfo[0]["TID"])
		resultCode = tpayCancel["result_cd"]
		resultMsg = tpayCancel["result_msg"]
		CancelDate = tpayCancel["CancelDate"]
		CancelTime = tpayCancel["CancelTime"]
		Canceltid = tpayCancel["tid"]
	}

	if resultCode == "00" {
		tx, err := cls.DBc.Begin()
		if err != nil {
			//return "5100", errors.New("begin error")
		}

		txErr := err

		// 오류 처리
		defer func() {
			if txErr != nil {
				// transaction rollback
				dprintf(1, c, "do rollback -미연결 및 즉시결제 취소  - 기프티콘 (SetOrderGifticon_V2_Cancel)  \n")
				tx.Rollback()
			}
		}()

		params["resultcd"] = resultCode
		params["resultmsg"] = resultMsg
		params["canceldate"] = CancelDate
		params["canceltime"] = CancelTime
		params["statecd"] = "0"
		params["cancelamt"] = strconv.Itoa(intantAmt)
		params["cancelmsg"] = "사용자취소"
		params["canceltid"] = Canceltid
		params["ccIp"] = ""

		// 결제 취소
		UpdatePaymentCancelQuery, err := cls.SetUpdateParam(paymentsql.UpdatePaymentCancel, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(UpdatePaymentCancelQuery)
		if err != nil {
			txErr = err
			dprintf(1, c, "Query(%s) -> error (%s) \n", UpdatePaymentCancelQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		params["paymentTy"] = "6" // 즉시결제 취소
		params["addAmt"] = "0"
		params["creditAmt"] = strconv.Itoa(intantAmt)
		params["userTy"] = paymentInfo[0]["USER_TY"]
		params["searchTy"] = paymentInfo[0]["SEARCH_TY"]
		params["payInfo"] = paymentInfo[0]["PAY_INFO"]
		params["payChannel"] = paymentInfo[0]["PAY_CHANNEL"]

		InsertPaymentHistoryQuery, err := cls.GetQueryJson(paymentsql.InsertPaymentHistory, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(InsertPaymentHistoryQuery)
		if err != nil {
			txErr = err
			dprintf(1, c, "Query(%s) -> error (%s) \n", InsertPaymentHistoryQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		//주문 취소
		OrderCancelQuery, err := cls.SetUpdateParam(ordersql.UpdateOrderCancel, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "OrderCancel parameter fail"))
		}
		_, err = tx.Exec(OrderCancelQuery)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		//기프티콘 취소
		UpdateCouponCancelQuery, err := cls.GetQueryJson(ordersql.UpdateCouponCancel, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateCouponCancelQuery)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		if orderTy != "6" {

			if payTy == "0" {

				linkInfo, err := cls.GetSelectData(ordersql.SelectLinkInfo, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if linkInfo == nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("99", "협약이 내용이 없습니다."))
				}
				prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
				linkId := linkInfo[0]["LINK_ID"]

				// 포인트 화불
				params["linkId"] = linkId
				params["prepaidAmt"] = strconv.Itoa(prepaidAmt + totalAmt - intantAmt)
				prepaidPoint, _ := strconv.Atoi(linkInfo[0]["PREPAID_POINT"])
				params["prepaidPoint"] = strconv.Itoa(prepaidPoint + pointUse)

				UpdateLinkQuery, err := cls.SetUpdateParam(ordersql.UpdateLink, params)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
				}
				_, err = tx.Exec(UpdateLinkQuery)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
			}

			//지원금 환불
			params["orderAmt"] = strconv.Itoa(totalAmt - intantAmt)
			UserSupportBalanceUpdateQuery, err := cls.SetUpdateParam(booksql.UpdateBookUserCancelSupportBalance, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UserSupportBalanceUpdateQuery)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		}

		// transaction commit
		err = tx.Commit()
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "성공"
		return c.JSON(http.StatusOK, m)

	} else {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = resultMsg
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "성공"

	return c.JSON(http.StatusOK, m)

}
