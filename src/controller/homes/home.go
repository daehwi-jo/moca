package homes

import (
	"github.com/labstack/echo/v4"
	commonsql "mocaApi/query/commons"
	giftsql "mocaApi/query/gifts"
	homessql "mocaApi/query/homes"
	ordersql "mocaApi/query/orders"
	"mocaApi/src/controller"
	"mocaApi/src/controller/benepicons"
	"mocaApi/src/controller/cls"
	"mocaApi/src/controller/wincubes"
	"net/http"
)


var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

type empty struct{
}



// 홈화면 데이터
func GetHomeData(c echo.Context) error {

	dprintf(4, c, "call GetHomeData\n")

	params := cls.GetParamJsonMap(c)

	homeData := make(map[string]interface{})

	notice, err := cls.GetSelectType(commonsql.SelectNotice, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if notice == nil {
		homeData["notice"]= empty{}
	}else{
		homeData["notice"] = notice[0]
	}


	amountData, err := cls.GetSelectType(homessql.SelectUserAmountInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if amountData == nil {
		homeData["amountData"]= empty{}
	}else{
		homeData["amountData"] = amountData[0]
	}

	gift, err := cls.GetSelectType(giftsql.SelectGift, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if gift == nil {
		giftData := make(map[string]interface{})
		giftData["giftMsg"]= "충전금을 선물해보세요."
		giftData["giftDate"]= ""
		giftData["giftId"]= ""
		giftData["giftStat"]="link"
		giftData["giftLink"]=""

		homeData["gift"]= giftData
	}else{
		homeData["gift"] = gift[0]
	}







	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = homeData

	return c.JSON(http.StatusOK, m)

}

//내 장부 잔액 정보
func GetUserAmountInfo(c echo.Context) error {

	dprintf(4, c, "call GetUserAmountInfo\n")

	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectType(homessql.SelectUserAmountInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "응답 실패"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultData[0]

	return c.JSON(http.StatusOK, m)

}


func GetMyGrpList(c echo.Context) error {

	dprintf(4, c, "call GetMyGrpList\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectType(homessql.SelectMyGrpList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	m := make(map[string]interface{})
	if resultList == nil {

		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = []string{}
		return c.JSON(http.StatusOK, m)
	}
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultList

	return c.JSON(http.StatusOK, m)

}



func GetMyTicketGrpList(c echo.Context) error {

	dprintf(4, c, "call GetMyTicketGrpList\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectType(homessql.SelectMyTickeGrpList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	m := make(map[string]interface{})
	if resultList == nil {

		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = []string{}
		return c.JSON(http.StatusOK, m)
	}
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultList

	return c.JSON(http.StatusOK, m)

}
func GetStoreList(c echo.Context) error {

	dprintf(4, c, "call GetStoreList\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectType(homessql.SelectHomeStoreList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	m := make(map[string]interface{})
	if resultList == nil {

		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = []string{}
		return c.JSON(http.StatusOK, m)
	}
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultList

	return c.JSON(http.StatusOK, m)

}

func GetBoardList(c echo.Context) error {

	dprintf(4, c, "call GetBoardList\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectType(commonsql.SelectBoardList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultList == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "no data"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}



func GetMyInfo(c echo.Context) error {

	dprintf(4, c, "call GetMyInfo\n")

	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectTypeRequire(homessql.SelectMyInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "개인정보 불러오기 오류"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultData[0]

	return c.JSON(http.StatusOK, m)

}

func GetMyAlim(c echo.Context) error {

	dprintf(4, c, "call GetMyAlim\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectData(ordersql.SelectRemainCouponList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultList == nil {

		result := make(map[string]string)
		result["Alim"]=""

		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] =  result
		return c.JSON(http.StatusOK, m)
	}


	chkCoupon :=0

	//IAD10 발행완료
	//IAD03 쿠폰 사용불가-주문취소
	//IAD05 쿠폰사용불가_유효기간 경과
	//IAD06 쿠폰사용불가_환불
	//IAD07 쿠폰사용불가_기간연장
	//IAD09 교환완료
	for _, item := range resultList {

	    orderNo:=item["ORDER_NO"]
	    ordNo:=item["ORD_NO"]
	    cpNo:=item["CPNO"]
		media:=item["REST_NM"]

		params["cpNo"]		= cpNo
		params["orderNo"]	= orderNo

		if media=="Benepicon"{
			chkResult := benepicons.CallBenepiconCheck(ordNo,cpNo,orderNo)

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
				chkCoupon = chkCoupon+1
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
			}

		}else if media=="Wincube"{

			tokenId,resultcd :=wincubes.GetWincubeAuth()
			if resultcd=="99"{
				lprintf(1, "[INFO] token recv fail \n")
				return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
			}
			chkResult := wincubes.CallWincubeCouponStatus(tokenId,orderNo)
			cpnoResultCd	:=chkResult["RESULT_CD"]
			params["exchplc"]		=chkResult["EXCHPLC"]
			params["exchcoNm"]		=chkResult["EXCHCO_NM"]
			params["cpnoExchDt"]	=chkResult["CPNO_EXCH_DT"]
			params["cpnoStatus"]	=chkResult["CPNO_STATUS"]
			params["cpnoStatusCd"]	=chkResult["CPNO_STATUS_CD"]
			params["balance"]		=chkResult["BALANCE"]

			switch cpnoResultCd {

			case "0": //
				params["cpStatus"]="0"
				chkCoupon = chkCoupon+1
				break
			case "4005": //이미 취소된 상품입니다.
				params["cpStatus"]="1"
				break
			case "4006": //교환된 상품으로 취소가 불가합니다.
				params["cpStatus"]="2"
				break
			case "4007": //상품권의 기간이 만료되었습니다.
				params["cpStatus"]="9"
				break
			}
		}else{
			m := make(map[string]interface{})
			m["resultCode"] = "00"
			m["resultMsg"] = "응답 성공"
			return c.JSON(http.StatusOK, m)
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

	msg :=""
	if chkCoupon >0 {
		msg="오늘 소멸예정인 교환권이 있습니다."
	}

	result := make(map[string]string)
	result["Alim"]=msg


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = result

	return c.JSON(http.StatusOK, m)

}




// 개인정보 설정 업데이트
func SetMyInfo(c echo.Context) error {

	dprintf(4, c, "call SetMyInfo\n")

	params := cls.GetParamJsonMap(c)

	// 파라메터 맵으로 쿼리 변환
	selectQuery, err := cls.SetUpdateParam(homessql.UpdateMyInfo, params)
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


