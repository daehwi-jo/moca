package rests

import (
	"github.com/labstack/echo/v4"
	commons "mocaApi/query/commons"
	restsql "mocaApi/query/rests"
	usersql "mocaApi/query/users"
	"mocaApi/src/controller"
	"mocaApi/src/controller/cls"
	"net/http"
	"strconv"
	"strings"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf
type empty struct{
}



// 장부장 권한가진 장부 리스트
func GetMyAuthGrpList(c echo.Context) error {

	dprintf(4, c, "call GetMyAuthGrpList\n")

	params := cls.GetParamJsonMap(c)


	restId := params["restId"]

	grpQuery :=""
	if restId ==""{
		grpQuery=restsql.SelectAuthMyGrpList
	}else{
		grpQuery=restsql.SelectUnlinkMyGrpList
	}

	resultList, err := cls.GetSelectType(grpQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
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


func GetSearchStoreList(c echo.Context) error {

	dprintf(4, c, "call GetSearchStoreList\n")
	lprintf(4, "[INFO] GetSearchStoreList \n")

	params := cls.GetParamJsonMap(c)


	pageSize,_ := strconv.Atoi(params["pageSize"])
	pageNo,_ := strconv.Atoi(params["pageNo"])
	offset := strconv.Itoa((pageNo-1) * pageSize)
	if pageNo == 1 {
		offset = "0"
	}
	params["offSet"] = offset

	searchKey :=params["searchKey"]

	if searchKey==""{
		params["searchKey"] = ""
	}else{
		params["searchKey"] = searchKey
	}


	myGrpArray, err := cls.GetSelectDataRequire(usersql.SelectMyGrpArray, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	myGrp := strings.ReplaceAll(myGrpArray[0]["myGrp"] ,",","','")
	params["myGrp"]=myGrp
	linkQuery :=""
	searchType := params["searchType"]
	if searchType=="1"{
		linkQuery = "WHERE linkGrpId <>'N' "
	}else if searchType=="2"{
		linkQuery = "WHERE linkGrpId ='N' "
	}else if searchType=="3"{
		linkQuery = "WHERE freqYn ='Y' "
	}else{
		linkQuery = "WHERE linkGrpId ='N' "
	}




	resultCnt, err := cls.GetSelectDataRequire(restsql.SelectSearchRestCnt + linkQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	totalCount,_ :=strconv.Atoi(resultCnt[0]["totalCount"])
	totalPage := strconv.Itoa((totalCount/pageSize) )
	if totalPage =="0"{
		totalPage="1"
	}


	linkQuery ="ORDER BY DISTANCE ASC"
	if searchType=="1"{
		linkQuery = "WHERE linkGrpId <>'N' ORDER BY DISTANCE ASC"
	}else if searchType=="2"{
		linkQuery = "WHERE linkGrpId ='N'  ORDER BY DISTANCE ASC"
	}else if searchType=="3"{
		linkQuery = "WHERE freqYn ='Y' ORDER BY DISTANCE ASC"
	}else {
		linkQuery = "WHERE linkGrpId ='N'  ORDER BY DISTANCE ASC"
	}

	//HAVING distance <= 10


	resultList, err := cls.GetSelectDataRequire(restsql.SelectSearchRest + linkQuery + commons.PagingQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultList == nil {
		Data := make(map[string]interface{})
		Data["totalPage"] = totalPage
		Data["currentPage"] = pageNo
		Data["searchList"] =  []string{}
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = Data
		return c.JSON(http.StatusOK, m)
	}


	// 연결된 장부정보 불러오기
	for i := range resultList {
		params["restId"] = resultList[i]["restId"]

		grpId := resultList[i]["linkGrpId"]
		if grpId=="N"{
			resultList[i]["linkGrpNm"]= ""
			resultList[i]["linkGrpPayType"]= ""
			resultList[i]["linkGrpAmt"]= ""
			resultList[i]["linkRestOpenYn"]= ""
			resultList[i]["linkRestOpenMsg"]= ""
		}else{
			params["grpId"] = resultList[i]["linkGrpId"]
			myGrpData, err := cls.GetSelectData(restsql.SelectSearchGrpInfo, params, c)
			if err != nil {
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
			resultList[i]["linkRestOpenYn"]= myGrpData[0]["OPEN_YN"]
			resultList[i]["linkRestOpenMsg"]= myGrpData[0]["OPEN_MSG"]
			resultList[i]["linkGrpNm"]= myGrpData[0]["grpNm"]
			payTy :=""
			if resultList[i]["payTy"] != myGrpData[0]["grpPayTy"] {
				payTy = resultList[i]["payTy"]
			}else{
				payTy = myGrpData[0]["grpPayTy"]
			}
			resultList[i]["linkGrpPayType"]= payTy
			resultList[i]["linkGrpAmt"]=  myGrpData[0]["grpAmt"]
		}

	}



	Data := make(map[string]interface{})
	Data["searchList"] = resultList
	Data["totalPage"] = totalPage
	Data["currentPage"] = pageNo

	if searchType=="0"{
		list :=GetSearchRandStore(params["lat"],params["lng"],myGrp)
		if list==nil{
			Data["recomList"] = []string{}
		}else{
			Data["recomList"] = list
		}

	}else{
		Data["recomList"] = []string{}
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = Data

	return c.JSON(http.StatusOK, m)

}



func GetSearchRandStore(lat string,lng string,myGrp string) []map[string]string {



	params := make(map[string]string)
	params["lat"] = lat
	params["lng"] = lng
	params["myGrp"] = myGrp

	var errMap []map[string]string

	resultList, err := cls.SelectData(restsql.SelectSearchRandStore, params)
	if err != nil {
		lprintf(1, "[ERROR] GetSearchRandStore error(%s) \n", err.Error())
		return errMap
	}
	if resultList == nil{
		return errMap
	}

	// 연결된 장부정보 불러오기
	for i := range resultList {
		params["restId"] = resultList[i]["restId"]
		params["grpId"] = resultList[i]["linkGrpId"]
		myGrpData, err := cls.SelectData(restsql.SelectSearchGrpInfo, params)
		if err != nil {
			return errMap
		}

		payTy :=""
		if resultList[i]["payTy"] != myGrpData[0]["grpPayTy"] {
			payTy = resultList[i]["payTy"]
		}else{
			payTy = myGrpData[0]["grpPayTy"]
		}

		resultList[i]["linkGrpNm"]= myGrpData[0]["grpNm"]
		resultList[i]["linkGrpPayType"]= payTy
		resultList[i]["linkGrpAmt"]=  myGrpData[0]["grpAmt"]
		resultList[i]["linkRestOpenYn"]= myGrpData[0]["OPEN_YN"]
		resultList[i]["linkRestOpenMsg"]= myGrpData[0]["OPEN_MSG"]

	}
	return resultList

}


func SetStoreLinkCancel(c echo.Context) error {

	dprintf(4, c, "call SetStoreLinkCancel\n")

	params := cls.GetParamJsonMap(c)


	resultData, err := cls.GetSelectData(restsql.SelectLinkCancelCheckInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결 정보가 없습니다."))
	}
	preAmt, _ := strconv.Atoi(resultData[0]["PREPAID_AMT"])
	unpaidAmt, _ := strconv.Atoi(resultData[0]["UNPAID_AMT"])
	if preAmt > 0 {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "선불 충전금액이 남아 있습니다."))
	}

	if unpaidAmt > 0 {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "미정산 주문금액이 남아있습니다."))
	}

	// 파라메터 맵으로 쿼리 변환
	selectQuery, err := cls.GetQueryJson(restsql.UpdateLinkCancel, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(selectQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	// m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}



func SetStoreLink(c echo.Context) error {

	dprintf(4, c, "call SetStoreLink\n")

	params := cls.GetParamJsonMap(c)

	grpPayTy :=params["grpPayTy"]

	if grpPayTy=="1"{
		m := make(map[string]interface{})
		m["resultCode"] = "01"
		m["resultMsg"] = "후불 장부 연결은 관리자에게 문의 해주세요."
		return c.JSON(http.StatusOK, m)
	}


	storeChk, err := cls.GetSelectData(restsql.SelectLinkStoreCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	restType :=storeChk[0]["REST_TYPE"]

	// 기프티콘 가맹점은 회사장부만 연결가능
	if restType=="G" {
		grpChk, err := cls.GetSelectData(restsql.SelectLinkGrpCheck, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		grpTypeCd :=grpChk[0]["GRP_TYPE_CD"]
		if grpTypeCd !="1"{
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "회사 장부만 연결 가능한 가맹점입니다."))
		}

	}


	checkLink, err := cls.GetSelectType(restsql.SelectLinkCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if checkLink == nil {

		params["agrmId"]=params["restId"]+params["grpId"]
		params["reqStat"]="1"
		params["reqTy"]="0"
		params["reqComment"]="사용자요청"
		params["payTy"]="0"
		params["prepaidAmt"]="0"

		// 파라메터 맵으로 쿼리 변환
		insertLinkQuery, err := cls.GetQueryJson(restsql.InsertLink, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(insertLinkQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
	}else{

		params["reqStat"]="1"
		params["payTy"]="0"
		UpdateLinkQuery, err := cls.GetQueryJson(restsql.UpdateLink, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateLinkQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	// m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}



func SetStoreReq(c echo.Context) error {

	dprintf(4, c, "call SetStoreReq\n")

	params := cls.GetParamJsonMap(c)

	insertStoreReqQuery, err := cls.GetQueryJson(restsql.InsertStoreReq, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(insertStoreReqQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}


func GetStoreInfo(c echo.Context) error {

	dprintf(4, c, "call GetStoreInfo\n")

	params := cls.GetParamJsonMap(c)


	UnlinkGrpCnt, err := cls.GetSelectDataRequire(restsql.SelectStoreUnLinkBookCnt, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}




	// 기본 정보
	storeInfo,err := cls.GetSelectDataRequire(restsql.SelectStoreInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] =  []string{}
		return c.JSON(http.StatusOK, m)
	}

	storeInfo[0]["unlinkMyGrpCnt"] = UnlinkGrpCnt[0]["unlinkGrpCnt"]
	store := make(map[string]interface{})
	store["storeInfo"] = storeInfo[0]

	// 부대 시설 정보
	storeService ,err := cls.GetSelectType(restsql.SelectStoreServiceList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeService == nil {
		store["storeService"] = []string{}

	}else{
		store["storeService"] = storeService
	}

	// 메뉴정보
	storeMenu ,err := cls.GetSelectType(restsql.SelectStoreMenu, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeMenu == nil {
		store["storeMenu"] = []string{}

	}else{
		store["storeMenu"] = storeMenu
	}
	grpId :=params["grpId"]

	if grpId !=""{
		linkInfo ,err := cls.GetSelectType(restsql.SelectStoreLinkInfo, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}
		if linkInfo == nil {
			store["linkInfo"] = empty{}

		}else{
			store["linkInfo"] = linkInfo[0]
		}
	}else{
		store["linkInfo"] =  empty{}
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = store

	return c.JSON(http.StatusOK, m)

}


// 가맹점 즐겨찾기 설정
func SetStoreFavorite(c echo.Context) error {

	dprintf(4, c, "call SetFavorite\n")

	params := cls.GetParamJsonMap(c)

	menuChk, err := cls.GetSelectDataRequire( restsql.SelectFreqStoreChk, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	storeCnt,_ :=strconv.Atoi(menuChk[0]["storeCnt"])


	favoriteYn := params["favoriteYn"]

	var sqlquery string
	if favoriteYn == "Y" {
		sqlquery = restsql.InsertFreqStore
	} else {
		sqlquery = restsql.DeleteFreqStore
	}


	if (storeCnt > 0 && favoriteYn=="Y") {
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





func GetStoreInfo_V2(c echo.Context) error {

	dprintf(4, c, "call GetStoreInfo_V2\n")

	params := cls.GetParamJsonMap(c)


	UnlinkGrpCnt, err := cls.GetSelectDataRequire(restsql.SelectStoreUnLinkBookCnt, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}



	// 기본 정보
	storeInfo,err := cls.GetSelectDataRequire(restsql.SelectStoreInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] =  []string{}
		return c.JSON(http.StatusOK, m)
	}

	storeInfo[0]["unlinkMyGrpCnt"] = UnlinkGrpCnt[0]["unlinkGrpCnt"]
	store := make(map[string]interface{})
	store["storeInfo"] = storeInfo[0]

	// 부대 시설 정보
	storeService ,err := cls.GetSelectType(restsql.SelectStoreServiceList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeService == nil {
		store["storeService"] = []string{}

	}else{
		store["storeService"] = storeService
	}

	// 메뉴정보
	storeMenu ,err := cls.GetSelectType(restsql.SelectStoreMenu, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeMenu == nil {
		store["storeMenu"] = []string{}

	}else{
		store["storeMenu"] = storeMenu
	}
	grpId :=params["grpId"]

	if grpId !=""{
		linkInfo ,err := cls.GetSelectType(restsql.SelectStoreLinkInfo, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
		}
		if linkInfo == nil {
			store["linkInfo"] = empty{}

		}else{
			store["linkInfo"] = linkInfo[0]
		}
	}else{
		store["linkInfo"] =  empty{}
	}



	// 개인 알림 정보
	params["grpId"]=""
	params["notice"]="storeInfo"
	storeNotice ,err := cls.GetSelectData(usersql.SelectUserNoticeInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeNotice == nil {
		store["storeNotice"] ="Y"
	}else{
		store["storeNotice"] = storeNotice[0]["VALUE"]
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = store

	return c.JSON(http.StatusOK, m)

}



func GetBaseStore(c echo.Context) error {

	dprintf(4, c, "call GetBaseStore\n")

	params := cls.GetParamJsonMap(c)


	// 기본 정보
	storeInfo,err := cls.GetSelectType(restsql.SelectBaseStore, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultList"] =  []string{}
		return c.JSON(http.StatusOK, m)
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = storeInfo

	return c.JSON(http.StatusOK, m)

}





func GetBarcodeStoreList(c echo.Context) error {

	dprintf(4, c, "call GetBarcodeStoreList\n")

	params := cls.GetParamJsonMap(c)


	orderby:= params["orderby"]
	orderbyQuery:=""

	if orderby=="r" {
		orderbyQuery="order by orderCnt desc"
	}else{
		orderbyQuery="order by REST_NM asc"
	}

	// 기본 정보
	storeInfo,err := cls.GetSelectType(restsql.SelectBaseStore+orderbyQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultList"] =  []string{}
		return c.JSON(http.StatusOK, m)
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = storeInfo

	return c.JSON(http.StatusOK, m)

}



func GetSearchStoreList_v2(c echo.Context) error {

	dprintf(4, c, "call GetSearchStoreList_v2\n")


	params := cls.GetParamJsonMap(c)


	pageSize,_ := strconv.Atoi(params["pageSize"])
	pageNo,_ := strconv.Atoi(params["pageNo"])
	offset := strconv.Itoa((pageNo-1) * pageSize)
	if pageNo == 1 {
		offset = "0"
	}
	params["offSet"] = offset

	searchKey :=params["searchKey"]

	if searchKey==""{
		params["searchKey"] = ""
	}else{
		params["searchKey"] = searchKey
	}




	resultCnt, err := cls.GetSelectDataRequire(restsql.SelectSearchRestCnt_v2 , params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	totalCount,_ :=strconv.Atoi(resultCnt[0]["totalCount"])
	totalPage := strconv.Itoa((totalCount/pageSize) )
	if totalPage =="0"{
		totalPage="1"
	}


	//HAVING distance <= 10


	resultList, err := cls.GetSelectDataRequire(restsql.SelectSearchRest_v2  + commons.PagingQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultList == nil {
		Data := make(map[string]interface{})
		Data["totalPage"] = totalPage
		Data["currentPage"] = pageNo
		Data["searchList"] =  []string{}
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = Data
		return c.JSON(http.StatusOK, m)
	}


	Data := make(map[string]interface{})
	Data["searchList"] = resultList
	Data["totalPage"] = totalPage
	Data["currentPage"] = pageNo


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = Data

	return c.JSON(http.StatusOK, m)

}