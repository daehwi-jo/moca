package books

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	booksql "mocaApi/query/books"
	companysql "mocaApi/query/companys"
	"mocaApi/query/homes"
	ordersql "mocaApi/query/orders"
	restsql "mocaApi/query/rests"
	usersql "mocaApi/query/users"
	"mocaApi/src/controller"
	apiPush "mocaApi/src/controller/api/push"
	"mocaApi/src/controller/cls"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)


var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

type empty struct{
}

// 장부관리 - 장부 리스트
func GetBookList(c echo.Context) error {

	dprintf(4, c, "call GetBookList\n")

	params := cls.GetParamJsonMap(c)

	book := make(map[string]interface{})



	myBookList, err := cls.GetSelectType(booksql.SelectMyBookList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if myBookList == nil {
		book["myBookList"] = []string{}
	}else{
		book["myBookList"] = myBookList
	}


	memberBookList, err := cls.GetSelectType(booksql.SelectMemberBookList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if memberBookList == nil {
		book["memberBookList"] = []string{}
	}else{
		book["memberBookList"] = memberBookList
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = book

	return c.JSON(http.StatusOK, m)

}


// 장부편집 - 구성원, 가게관리, 장부 상세
func GetBookDesc(c echo.Context) error {

	dprintf(4, c, "call GetBookDesc\n")

	params := cls.GetParamJsonMap(c)


	params["grpAuth"]="0"
	BookDesc, err := cls.GetSelectType(booksql.SelectBookDesc, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if BookDesc == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}

	book := make(map[string]interface{})
	book["bookDesc"] = BookDesc[0]

	bookUserList, err := cls.GetSelectType(booksql.SelectBookUserList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookUserList == nil {
		book["bookUserList"] = []string{}
	}else{
		book["bookUserList"] = bookUserList
	}


	linkStoreList, err := cls.GetSelectType(booksql.SelectBookLinkStoreList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkStoreList == nil {
		book["linkStoreList"] = []string{}
	}else{
		book["linkStoreList"] = linkStoreList
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = book

	return c.JSON(http.StatusOK, m)

}

// 장부초대 링크
func GetBookInviteLink(c echo.Context) error {

	dprintf(4, c, "call GetBookInviteLink\n")

	params := cls.GetParamJsonMap(c)

	grpId :=params["grpId"]

	BookDesc, err := cls.GetSelectDataRequire(booksql.SelectBookInvteLink, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if BookDesc == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}

	iLink :=BookDesc[0]["INVITE_LINK"]

	if iLink=="N"{

	  rCode ,inviteLink:=makeInviteLink(grpId)

	  if rCode !="00"{
		  return c.JSON(http.StatusOK, controller.SetErrResult("99", inviteLink))
	  }

		params["inviteLink"] = inviteLink
		UpdateLinkQuery, err := cls.SetUpdateParam(booksql.UpdateBookIviteLink, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateLinkQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		iLink= inviteLink
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["inviteLink"] = iLink
	return c.JSON(http.StatusOK, m)

}







type DynamicLinkInfo struct {
	ShortLink      string `json:"shortLink"`    //
	PreviewLink    string  `json:"previewLink"` //

}


func makeInviteLink(grpId string)(string,string){
	url := "https://firebasedynamiclinks.googleapis.com/v1/shortLinks?key=AIzaSyB4jaVLPdwWU28SzyKIU8DngpPF5zbFLLg"

	var jsonStr = []byte(`{"dynamicLinkInfo":{
          "domainUriPrefix":"https://darayo.page.link",
          "link":"https://darayo.com/grpAdd?bookId=`+grpId+`",
          "androidInfo": {
            "androidPackageName": "com.fit.darayo"
          },
          "iosInfo": {
            "iosBundleId": "kr.co.darayo",
            "iosAppStoreId": "1158745361"
          },
          "analyticsInfo": {
            "googlePlayAnalytics": {
              "utmSource": "darayo",
              "utmMedium": "web",
              "utmCampaign": "inviteBook"
            }
          },
      },
      "suffix":{
        "option" : "SHORT"
      }}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var rMap DynamicLinkInfo
	err = json.Unmarshal(body, &rMap)
	if err != nil {
		return "99","링크 생성 실패"
	}
	inviteLink :=rMap.ShortLink
	return "00",inviteLink

}

// 장부 가입 승인 및 탈퇴 처리
func SetBookUser(c echo.Context) error {

	dprintf(4, c, "call SetBookUser\n")

	params := cls.GetParamJsonMap(c)

	authStat := params["authStat"]


	userInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if userInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "유저 정보가 없습니다."))
	}


	ugrpAuth :=userInfo[0]["GRP_AUTH"]



	//uAuth :=userInfo[0]["AUTH_STAT"]

	if authStat =="3"{

		if ugrpAuth=="0" {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부장은 탈퇴 시킬수 없습니다."))
		}

	}

	sqlQuery :=""
	if authStat=="1"{
		sqlQuery =booksql.UpdateBookUserReg
	}else if authStat=="3"{
		sqlQuery =booksql.UpdateBookUserLeave
	}else{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "요청 상태 정보가 잘못되었습니다."))
	}


	UpdateLinkQuery, err := cls.SetUpdateParam(sqlQuery, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdateLinkQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}



// 장부 정보 수정
func SetBookDesc(c echo.Context) error {

	dprintf(4, c, "call SetBookDesc\n")

	params := cls.GetParamJsonMap(c)


	supportAmt,_:= strconv.Atoi(params["supportAmt"])
	params["supportYn"] ="N"
	if supportAmt > 0 {
		params["supportYn"] ="Y"
	}


	UpdateLinkQuery, err := cls.SetUpdateParam(booksql.UpdateBookDesc, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdateLinkQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}


// 장부 생성
func SetMakeBook(c echo.Context) error {

	dprintf(4, c, "call setMakeBook\n")


	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectData(booksql.SelectCreatBookSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부아이디 생성 실패(2)"))
	}

	newGrpId := resultData[0]["newGrpId"]
	params["grpId"]=newGrpId

	grpTypeCd := params["grpTypeCd"]

	companyId := ""

	if grpTypeCd == "1" {

		companySeq, err := cls.GetSelectData(companysql.SelectCompanyId, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		companyId = companySeq[0]["companyId"]

	}

	// 장부 생성 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	// 오류 처리
	defer func() {
		if err != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -장부생성(SetMakeBook)  \n")
			tx.Rollback()
		}
	}()

	// transation exec
	// 파라메터 맵으로 쿼리 변환

	// 장부 생성
	supportAmt,_:= strconv.Atoi(params["supportAmt"])
	params["supportYn"] ="N"
	if supportAmt > 0 {
		params["supportYn"] ="Y"
	}

	bookCreateQuery, err := cls.SetUpdateParam(booksql.InsertCreateBook, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(bookCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", bookCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// 장부 인원 추가
	params["grpAuth"] = "0"
	params["authStat"] = "1"
	params["joinTy"] = "0"
	params["supportBalance"] = "0"

	bookUserAddQuery, err := cls.SetUpdateParam(booksql.InsertBookUser, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(bookUserAddQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", bookUserAddQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	if grpTypeCd == "1" {


		companyDupChek, err := cls.GetSelectData(companysql.SelectCompanyInfo, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		if companyDupChek == nil {
			params["companyId"] = companyId
			companyAddQuery, err := cls.SetUpdateParam(companysql.InsertCompany, params)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(companyAddQuery)
			if err != nil {
				dprintf(1, c, "Query(%s) -> error (%s) \n", companyAddQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else {
			companyId = companyDupChek[0]["COMPANY_ID"]

		}
		params["companyId"] = companyId
		companyAddBookQuery, err := cls.SetUpdateParam(companysql.InsertCompanyBook, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(companyAddBookQuery)
		if err != nil {
			dprintf(1, c, "Query(%s) -> error (%s) \n", companyAddBookQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// 장부 생성 TRNAN 종료

	grpData := make(map[string]interface{})
	grpData["grpId"] = newGrpId



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = grpData


	return c.JSON(http.StatusOK, m)

}



// 회원가입 기본 장부 생성
func SetMakeBookFirst(c echo.Context ,userId string)  {

	dprintf(4, c, "call SetMakeBookFirst - 기본 장부 생성\n")


	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectData(booksql.SelectCreatBookSeq, params, c)
	if err != nil {
		dprintf(1, c, "기본장부 생성 실패 \n",  err)
		return
	}
	if resultData == nil {
		dprintf(1, c, "기본장부 생성 실패 \n",  err)
		return
	}


	params["userId"]=userId
	userInfo, err := cls.GetSelectData(homes.SelectMyInfo, params, c)
	if err != nil {
		dprintf(1, c, "기본장부 생성 실패 \n",  err)
		return
	}
	if userInfo == nil {
		return
	}

	newGrpId := resultData[0]["newGrpId"]
	userNm := userInfo[0]["USER_NM"]
	params["grpId"]=newGrpId
	params["grpTypeCd"] ="4"
	params["grpNm"] =userNm +" 의 장부"


	// 장부 생성 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	// 오류 처리
	defer func() {
		if err != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -장부생성(SetMakeBook)  \n")
			tx.Rollback()
		}
	}()

	// transation exec
	// 파라메터 맵으로 쿼리 변환

	// 장부 생성
	params["supportYn"] ="N"
	params["supportAmt"] ="0"


	bookCreateQuery, err := cls.SetUpdateParam(booksql.InsertCreateBook, params)
	if err != nil {
		dprintf(1, c, "기본장부 생성 실패 \n",  err)
		return
	}
	_, err = tx.Exec(bookCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", bookCreateQuery, err)
		return
	}

	// 장부 인원 추가
	params["grpAuth"] = "0"
	params["authStat"] = "1"
	params["joinTy"] = "0"
	params["supportBalance"] = "0"

	bookUserAddQuery, err := cls.SetUpdateParam(booksql.InsertBookUser, params)
	if err != nil {
		dprintf(1, c, "기본장부 생성 실패 \n",  err)
		return
	}
	_, err = tx.Exec(bookUserAddQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", bookUserAddQuery, err)
		return
	}


	if params["recomCode"]=="중원대"{

		params["restId"]="R0000000338"
		params["agrmId"]=params["restId"]+newGrpId
		params["reqStat"]="1"
		params["reqTy"]="0"
		params["reqComment"]="사용자요청"
		params["payTy"]="0"
		params["prepaidAmt"]="0"
		// 파라메터 맵으로 쿼리 변환
		insertLinkQuery, _ := cls.GetQueryJson(restsql.InsertLink, params)
		if err != nil {
			dprintf(1, c, "중원대 자동 협약 실패 \n",  err)
		}
		// 쿼리 실행
		_, err = cls.QueryDB(insertLinkQuery)
		if err != nil {
			dprintf(1, c, "중원대 자동 협약 실패 Query(%s) -> error (%s) \n", insertLinkQuery, err)
		}
	}


	// 삭제 예정
	if params["channelCode"]=="S0000000562"{
		params["restId"]="S0000000562 "
		params["agrmId"]=params["restId"]+newGrpId
		params["reqStat"]="1"
		params["reqTy"]="0"
		params["reqComment"]="자동연결-청해동태탕"
		params["payTy"]="0"
		params["prepaidAmt"]="0"
		// 파라메터 맵으로 쿼리 변환
		insertLinkQuery, _ := cls.GetQueryJson(restsql.InsertLink, params)
		if err != nil {
			dprintf(1, c, "자동 연결 실패 \n",  err)
		}
		// 쿼리 실행
		_, err = cls.QueryDB(insertLinkQuery)
		if err != nil {
			dprintf(1, c, "자동 연결 실패 Query(%s) -> error (%s) \n", insertLinkQuery, err)
		}

	}


	// transaction commit
	err = tx.Commit()
	if err != nil {
		return
	}

	// 장부 생성 TRNAN 종료




}






// 정산하기 - 준비화면
func GetReadyCalculate(c echo.Context) error {

	dprintf(4, c, "call GetReadyCalculate\n")

	params := cls.GetParamJsonMap(c)


	userChk, err := cls.GetSelectDataRequire(booksql.SelectBookMyAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if userChk == nil{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "정산하기는 장부장만 가능합니다."))
	}

	grpAuth :=userChk[0]["GRP_AUTH"]
	grpNm :=userChk[0]["GRP_NM"]
	if grpAuth !="0"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "정산하기는 장부장만 가능합니다."))
	}




	linkChk, err := cls.GetSelectDataRequire(booksql.SelectLinkInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkChk == nil{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결 정보가 없습니다."))
	}
	reqStat :=linkChk[0]["REQ_STAT"]
	if reqStat !="1"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결 정보를 확인해주세요."))
	}



	resultData, err := cls.GetSelectDataRequire(ordersql.SelectUnpaidListCount, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	totalCount,_ := strconv.Atoi(resultData[0]["orderCnt"])
	totalAmt, _ := strconv.Atoi(resultData[0]["TOTAL_AMT"])



	unPaidList, err := cls.GetSelectType(ordersql.SelectUnpaidList , params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if unPaidList == nil {
		LinkData := make(map[string]interface{})
		LinkData["totalCount"] = totalCount
		LinkData["totalAmt"] = 0
		LinkData["accountList"] = []string{}
		LinkData["grpNm"] =grpNm
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = LinkData
		return c.JSON(http.StatusOK, m)
	}


	result := make(map[string]interface{})
	result["totalCount"] = totalCount
	result["totalAmt"] = totalAmt
	result["accountList"] = unPaidList
	result["grpNm"] = grpNm

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = result


	return c.JSON(http.StatusOK, m)

}


// 충전하기 - 준비 화면
func GetReadyCharging(c echo.Context) error {

	dprintf(4, c, "call GetReadyCharging\n")

	params := cls.GetParamJsonMap(c)


	userChk, err := cls.GetSelectDataRequire(booksql.SelectBookMyAuth, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if userChk == nil{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "충전하기는 장부장만 가능합니다."))
	}

	grpAuth :=userChk[0]["GRP_AUTH"]
	grpNm :=userChk[0]["GRP_NM"]
	if grpAuth !="0"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "충전하기는 장부장만 가능합니다."))
	}


	linkChk, err := cls.GetSelectDataRequire(booksql.SelectLinkInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkChk == nil{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결 정보가 없습니다."))
	}
	reqStat :=linkChk[0]["REQ_STAT"]
	if reqStat !="1"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결 정보를 확인해주세요."))
	}
	prepaidAmt,_ :=strconv.Atoi(linkChk[0]["PREPAID_AMT"])


	paymentUseYn:=linkChk[0]["PAYMENT_USE_YN"]


	if paymentUseYn=="N"{
		result := make(map[string]interface{})
		result["prepaidAmt"] = prepaidAmt
		result["amtList"] = []string{}
		result["grpNm"] =grpNm
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = result
		return c.JSON(http.StatusOK, m)
	}


	amtList, err := cls.GetSelectType(booksql.SelectChargeAmtList , params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if amtList == nil {
		result := make(map[string]interface{})
		result["prepaidAmt"] = prepaidAmt
		result["amtList"] = []string{}
		result["grpNm"] =grpNm
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultData"] = result
		return c.JSON(http.StatusOK, m)
	}


	result := make(map[string]interface{})
	result["prepaidAmt"] = prepaidAmt
	result["amtList"] = amtList
	result["grpNm"] = grpNm

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = result

	return c.JSON(http.StatusOK, m)

}


// 충전하기 가맹점 리스트
func GetReadyStoreList(c echo.Context) error {

	dprintf(4, c, "call GetReadyStoreList\n")

	params := cls.GetParamJsonMap(c)

	payTy :=params["payTy"]


	storeList, err := cls.GetSelectTypeRequire(booksql.SelectBookStoreList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if storeList == nil {
		eMsg :="충전 가능한 가맹점이 없습니다."
		if payTy=="1" {
			eMsg ="정산 가능한 가맹점이 없습니다."
		}
		return c.JSON(http.StatusOK, controller.SetErrResult("99",eMsg))
	}



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = storeList

	return c.JSON(http.StatusOK, m)

}




// 장부 사용자 리스트
func GetBookUserList(c echo.Context) error {

	dprintf(4, c, "call GetBookUserList\n")

	params := cls.GetParamJsonMap(c)


	bookUserList, err := cls.GetSelectType(booksql.SelectBookUserList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookUserList == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}




	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = bookUserList

	return c.JSON(http.StatusOK, m)

}




// 장부에 연결된 가맹점 리스트
func GetBookRestList(c echo.Context) error {

	dprintf(4, c, "call GetBookRestList\n")

	params := cls.GetParamJsonMap(c)


	resultList, err := cls.GetSelectType(booksql.SelectBookLinkStoreList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultList == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "연결된 가게 정보가 없습니다."))
	}




	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}




// 장부 초대 승인 처리
func SetUserInvite(c echo.Context) error {

	dprintf(4, c, "call SetUserInvite\n")

	params := cls.GetParamJsonMap(c)

	fname := cls.Fname
	bizCompanyEventUrl, _ := cls.GetTokenValue("BIZ.COMPANY_EVENT_URL", fname)

	comYn :="Y"
	companyInfo, err := cls.GetSelectDataRequire(booksql.SelectCompanyBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if companyInfo == nil {
		comYn="N"
	}else{
		params["companyId"]=companyInfo[0]["company_id"]
	}

	eGrpId := params["grpId"]
	eGrpId = strings.TrimSpace(eGrpId)


	userInfo, err := cls.GetSelectDataRequire(usersql.SelectUserInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if userInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 회원 정보 입니다."))
	}

	params["userNm"]= userInfo[0]["USER_NM"]
	params["hpNo"]=userInfo[0]["HP_NO"]

	rMSG :=""

	if comYn=="Y" {
		companyUserChk, err := cls.GetSelectDataRequire(booksql.SelectCompanyUserChk, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		if companyUserChk[0]["cnt"]=="0"{
			params["authStat"]="0"
			rMSG="장부 사용 요청이 완료되었습니다."
		}else{
			params["authStat"]="1"
			rMSG="장부 가입이 완료되었습니다."
		}


		//companyinfo userId 업데이트
		UpdateCompanyInfoQuery, err := cls.GetQueryJson(booksql.UpdateCompanyUser, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateCompanyInfoQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
	}else{
		rMSG="장부 사용 요청이 완료되었습니다."
		params["authStat"]="0"

	}

	bookChk, err := cls.GetSelectDataRequire(booksql.SelectBookDesc, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookChk == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}
	//params["supportBalance"]="0"


	//if eGrpId=="G0000003419"{
	params["supportBalance"]= bookChk[0]["SUPPORT_AMT"]
	//}
	// 지원금 초기금액 세팅 (장부연결시 최초 1회 FIRST_SUPPORT_AMT 만큼 부여)
	// ex) 쿠폰장부 SUPPORT_AMT 0으로 해놓아야 하므로
	if bookChk[0]["FIRST_SUPPORT_AMT"] != "0" {
		params["supportBalance"]= bookChk[0]["FIRST_SUPPORT_AMT"]
	}


	bookUser, err := cls.GetSelectDataRequire(booksql.SelectBookUserView, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	sqlQuery :=""
	if bookUser == nil {
		params["grpAuth"]="1"
		params["joinTy"]="1"
		sqlQuery =booksql.InsertBookUser
	}else{
		if bookUser[0]["AUTH_STAT"]=="1"{
			m := make(map[string]interface{})
			m["resultCode"] = "00"
			m["resultMsg"] = "장부 가입이 완료되었습니다."
			return c.JSON(http.StatusOK, m)
		}
		sqlQuery =booksql.UpdateBookUserReg
	}


	UpdateLinkQuery, err := cls.SetUpdateParam(sqlQuery, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdateLinkQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	// 기업 기념일 금액 충전
	if comYn == "Y" {
		uValue := url.Values{
			"companyId": {params["companyId"]},
			"userId": {params["userId"]},
		}

		_, err := http.PostForm(bizCompanyEventUrl, uValue)
		if err != nil {
			m := make(map[string]interface{})
			m["resultCode"] = "99"
			m["resultMsg"] = "기념일 금액 충전 오류"
			return c.JSON(http.StatusOK, m)
		}
	}


	if params["authStat"]=="0"{

		pushInfo, err := cls.GetSelectDataRequire(booksql.SelectUserBookAuth, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		if userInfo == nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "유저 정보가 없습니다."))
		}

		grpName :=pushInfo[0]["GRP_NAME"]
		managerId :=pushInfo[0]["MANAGER_ID"]
		userNm :=pushInfo[0]["USER_NM"]
		// 푸쉬 전송 시작
		pushMsg:= userNm+"님이 " +grpName+" 장부에 가입 요청 하였습니다."
		apiPush.SendPush_Msg_V1("장부",pushMsg,"M","0",managerId,params["grpId"],"book")
		// 푸쉬 전송 완료

	}




	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] =rMSG

	return c.JSON(http.StatusOK, m)

}


func SetUserInvite_OLD(c echo.Context) error {

	dprintf(4, c, "call SetUserInvite\n")

	params := cls.GetParamJsonMap(c)


	companyInfo, err := cls.GetSelectDataRequire(booksql.SelectCompanyBookInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if companyInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 초대링크 입니다."))
	}


	params["companyId"]=companyInfo[0]["company_id"]


	userInfo, err := cls.GetSelectDataRequire(usersql.SelectUserInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if userInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 회원 정보 입니다."))
	}

	params["userNm"]= userInfo[0]["USER_NM"]
	params["hpNo"]=userInfo[0]["HP_NO"]


	companyUserChk, err := cls.GetSelectDataRequire(booksql.SelectCompanyUserChk, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	if companyUserChk[0]["cnt"] =="0"{
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 회원 정보입니다."))
	}


	params["authStat"]="1"


	bookUser, err := cls.GetSelectDataRequire(booksql.SelectBookUserView, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	sqlQuery :=""
	if bookUser == nil {
		params["grpAuth"]="1"
		params["joinTy"]="1"
		sqlQuery =booksql.InsertBookUser
	}else{
		sqlQuery =booksql.UpdateBookUserReg
	}


	UpdateLinkQuery, err := cls.SetUpdateParam(sqlQuery, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdateLinkQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}


	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "장부에 가입되었습니다."

	return c.JSON(http.StatusOK, m)

}



type BookMakeInfo struct {
	UserId    			string 	`json:"userId"`   //
	GrpNm     			string  `json:"grpNm"`   //
	GrpTypeCd    		string  `json:"grpTypeCd"`    //
	Intro   			string  `json:"intro"`  //
	LimitYn 			string  `json:"limitYn"` //
	LimitAmt			string	`json:"limitAmt"`
	LimitDayAmt			string	`json:"limitDayAmt"`
	LimitDayCnt  		string  `json:"limitDayCnt"`    //
	SupportAmt  		string  `json:"supportAmt"`    //
	SupportForwardYn  	string  `json:"supportForwardYn"`    //
	LimitUseTimeStart	string  `json:"limitUseTimeStart"`    //
	LimitUseTimeEnd		string  `json:"limitUseTimeEnd"`    //
	CompanyNm			string  `json:"companyNm"`    //
	BizNum  			string  `json:"bizNum"`    //
	Addr  				string  `json:"addr"`    //
	Addr2  				string  `json:"addr2"`    //
	BaseStore  			[]BaseStore `json:"baseStore"`    //
}


type BaseStore struct {
	RestID string `json:"restId"`
	UseYn  string `json:"useYn"`
}


// 장부 생성- 회사 기존가맹점 추가
func SetMakeBook_V2(c echo.Context) error {

	dprintf(4, c, "call SetMakeBook_V2\n")


	bodyBytes, _ := ioutil.ReadAll(c.Request().Body)

	// 상세 주문데이터 get
	var bookInfo BookMakeInfo
	err2 := json.Unmarshal(bodyBytes, &bookInfo)
	if err2 != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err2.Error()))
	}


	c.Request().Body.Close() //  must close
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))


	params := cls.GetParamJsonMap(c)

	resultData, err := cls.GetSelectData(booksql.SelectCreatBookSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부아이디 생성 실패(2)"))
	}

	newGrpId := resultData[0]["newGrpId"]
	params["grpId"]=newGrpId

	grpTypeCd := bookInfo.GrpTypeCd

	companyId := ""

	if grpTypeCd == "1" {

		companySeq, err := cls.GetSelectData(companysql.SelectCompanyId, params, c)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		companyId = companySeq[0]["companyId"]

	}

	// 장부 생성 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}


	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -장부생성(SetMakeBook_V2)  \n")
			tx.Rollback()
		}
	}()

	// transation exec
	// 파라메터 맵으로 쿼리 변환

	// 장부 생성
	supportAmt,_:= strconv.Atoi(bookInfo.SupportAmt)
	params["supportYn"] ="N"
	if supportAmt > 0 {
		params["supportYn"] ="Y"
	}

	bookCreateQuery, err := cls.SetUpdateParam(booksql.InsertCreateBook, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(bookCreateQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", bookCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// 장부 인원 추가
	params["grpAuth"] = "0"
	params["authStat"] = "1"
	params["joinTy"] = "0"
	params["supportBalance"] = "0"

	bookUserAddQuery, err := cls.SetUpdateParam(booksql.InsertBookUser, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(bookUserAddQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", bookUserAddQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	if grpTypeCd == "1" {


		companyDupChek, err := cls.GetSelectData(companysql.SelectCompanyInfo, params, c)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		if companyDupChek == nil {
			params["companyId"] = companyId
			companyAddQuery, err := cls.SetUpdateParam(companysql.InsertCompany, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(companyAddQuery)
			if err != nil {
				txErr = err
				dprintf(1, c, "Query(%s) -> error (%s) \n", companyAddQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}
		} else {
			companyId = companyDupChek[0]["COMPANY_ID"]

		}
		params["companyId"] = companyId
		companyAddBookQuery, err := cls.SetUpdateParam(companysql.InsertCompanyBook, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(companyAddBookQuery)
		if err != nil {
			txErr = err
			dprintf(1, c, "Query(%s) -> error (%s) \n", companyAddBookQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}



		for i, _ := range bookInfo.BaseStore {

			restId := bookInfo.BaseStore[i].RestID
			useYn := bookInfo.BaseStore[i].UseYn

			if useYn=="Y"{
				params["restId"]=restId
				// 파라메터 맵으로 쿼리 변환
				checkLink, err := cls.GetSelectType(restsql.SelectLinkCheck, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if checkLink == nil {

					params["agrmId"]=params["restId"]+params["grpId"]
					params["reqStat"]="1"
					params["reqTy"]="0"
					params["reqComment"]="기본가맹점연결"
					params["payTy"]="0"
					params["prepaidAmt"]="0"

					// 파라메터 맵으로 쿼리 변환
					insertLinkQuery, err := cls.GetQueryJson(restsql.InsertLink, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(insertLinkQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
				}else{

					params["reqStat"]="1"
					params["payTy"]="0"
					UpdateLinkQuery, err := cls.GetQueryJson(restsql.UpdateLink, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(UpdateLinkQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}

				}

			}


		}


	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		txErr = err
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// 장부 생성 TRNAN 종료

	grpData := make(map[string]interface{})
	grpData["grpId"] = newGrpId



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = grpData


	return c.JSON(http.StatusOK, m)

}



// 장부 정보 수정
func SetBookDesc_V2(c echo.Context) error {

	dprintf(4, c, "call SetBookDesc_V2\n")

	bodyBytes, _ := ioutil.ReadAll(c.Request().Body)

	// 상세 주문데이터 get
	var bookInfo BookMakeInfo
	err2 := json.Unmarshal(bodyBytes, &bookInfo)
	if err2 != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err2.Error()))
	}


	c.Request().Body.Close() //  must close
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	params := cls.GetParamJsonMap(c)


	// 장부 생성 TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}


	txErr := err
	// 오류 처리
	defer func() {
		if txErr != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -장부생성(SetMakeBook_V2)  \n")
			tx.Rollback()
		}
	}()

	// transation exec
	// 파라메터 맵으로 쿼리 변환


	params["grpTypeCd"] = bookInfo.GrpTypeCd
	params["grpNm"]=bookInfo.GrpNm
	params["intro"]=bookInfo.Intro
	params["limitYn"]=bookInfo.LimitYn
	params["limitUseTimeStart"]=bookInfo.LimitUseTimeStart
	params["limitUseTimeEnd"]=bookInfo.LimitUseTimeEnd
	params["limitAmt"]= bookInfo.LimitAmt
	params["limitDayAmt"]= bookInfo.LimitDayAmt
	params["limitDayCnt"]= bookInfo.LimitDayCnt
	params["supportAmt"]=bookInfo.SupportAmt

	if params["supportAmt"] != "" {
		supportAmt, _ := strconv.Atoi(params["supportAmt"])
		params["supportYn"] = "N"
		if supportAmt > 0 {
			params["supportYn"] = "Y"
		}
	}
	
	UpdateLinkQuery, err := cls.SetUpdateParam(booksql.UpdateBookDesc, params)
	if err != nil {
		txErr = err
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
	}
	_, err = tx.Exec(UpdateLinkQuery)
	if err != nil {
		txErr = err
		dprintf(1, c, "Query(%s) -> error (%s) \n", UpdateLinkQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	grpTypeCd :=bookInfo.GrpTypeCd



	if grpTypeCd == "1" {

		for i, _ := range bookInfo.BaseStore {

			restId := bookInfo.BaseStore[i].RestID
			useYn := bookInfo.BaseStore[i].UseYn
			params["restId"]=restId
			if useYn=="Y"{

				// 파라메터 맵으로 쿼리 변환
				checkLink, err := cls.GetSelectType(restsql.SelectLinkCheck, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if checkLink == nil {

					params["agrmId"]=params["restId"]+params["grpId"]
					params["reqStat"]="1"
					params["reqTy"]="0"
					params["reqComment"]="기본가맹점연결"
					params["payTy"]="0"
					params["prepaidAmt"]="0"

					// 파라메터 맵으로 쿼리 변환
					insertLinkQuery, err := cls.GetQueryJson(restsql.InsertLink, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(insertLinkQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
				}else{

					params["reqStat"]="1"

					UpdateLinkQuery, err := cls.SetUpdateParam(restsql.UpdateLink, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(UpdateLinkQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}

				}

			}else{
				params["reqStat"]="4"
				UpdateLinkQuery, err := cls.SetUpdateParam(restsql.UpdateLink, params)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				// 쿼리 실행
				_, err = tx.Exec(UpdateLinkQuery)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}

			}


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
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}



// 장부편집 - 구성원, 가게관리, 장부 상세
func GetBookDesc_V2(c echo.Context) error {

	dprintf(4, c, "call GetBookDesc_V2\n")

	params := cls.GetParamJsonMap(c)


	params["grpAuth"]="0"
	BookDesc, err := cls.GetSelectType(booksql.SelectBookDesc, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if BookDesc == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "장부 정보가 없습니다."))
	}

	book := make(map[string]interface{})
	book["bookDesc"] = BookDesc[0]

	bookUserList, err := cls.GetSelectType(booksql.SelectBookUserList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if bookUserList == nil {
		book["bookUserList"] = []string{}
	}else{
		book["bookUserList"] = bookUserList
	}


	linkStoreList, err := cls.GetSelectType(booksql.SelectBookLinkStoreList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkStoreList == nil {
		book["linkStoreList"] = []string{}
	}else{
		book["linkStoreList"] = linkStoreList
	}


	baseStoreList, err := cls.GetSelectTypeRequire(booksql.SelectBookBaseStoreList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if baseStoreList == nil {
		book["baseStoreList"] = []string{}
	}else{
		book["baseStoreList"] = baseStoreList
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = book

	return c.JSON(http.StatusOK, m)

}



//전체 연결 가맹점 정산 정보
func GetBookLinkList(c echo.Context) error {

	dprintf(4, c, "call GetBookLinkList\n")

	params := cls.GetParamJsonMap(c)

	orderby :=params["orderby"]

	orderbyQuery :=" ORDER BY TOTAL_AMT DESC"

	if orderby=="0"{
		orderbyQuery =" ORDER BY TOTAL_AMT ASC"
	}else if orderby=="1"{
		orderbyQuery =" ORDER BY TOTAL_AMT DESC"
	}else if orderby=="2"{
		orderbyQuery =" ORDER BY REST_NM ASC"
	}

	linkStoreList, err := cls.GetSelectType(booksql.SelectGrpLinkList + orderbyQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	m := make(map[string]interface{})


	if linkStoreList == nil {
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		m["resultList"] = []string{}
		return c.JSON(http.StatusOK, m)
	}


	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = linkStoreList

	return c.JSON(http.StatusOK, m)

}