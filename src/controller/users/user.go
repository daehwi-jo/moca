package users

import (
	"github.com/labstack/echo/v4"
	usersql "mocaApi/query/users"
	"mocaApi/src/controller"
	books "mocaApi/src/controller/books"
	"mocaApi/src/controller/cls"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

func BaseUrl(c echo.Context) error {

	resultData, err := cls.SelectHealthData("SELECT health FROM health_check LIMIT 1", nil)
	if err != nil {
		return c.JSONP(http.StatusInternalServerError, "", err.Error())
	}

	return c.JSONP(http.StatusOK, "", resultData[0]["health"])
}

func LoginDarayo(c echo.Context) error {

	dprintf(4, c, "call LoginDarayo\n")
	ip, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectDataRequire(usersql.SelectUserLoginCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		// 접속 로그
		params["logInOut"] = "I"
		params["ip"] = ip
		params["succYn"] = "N"
		params["type"] = "id"
		go LoginAcceesLog(c, params)
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "아이디 또는 비밀번호가 잘못되었습니다."))
	}

	useYn := resultData[0]["USE_YN"]
	userId := resultData[0]["userId"]
	loginId := resultData[0]["LOGIN_ID"]
	userNm := resultData[0]["USER_NM"]
	userTel := resultData[0]["HP_NO"]

	params["userId"] = userId
	if useYn == "N" {
		// 접속 로그
		params["logInOut"] = "I"
		params["ip"] = ip
		params["succYn"] = "N"
		params["type"] = "id"
		go LoginAcceesLog(c, params)
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "로그인 실패."))
	}

	userTy := resultData[0]["USER_TY"]
	if userTy != "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "일반 사용자가 아닙니다."))
	}

	// 푸쉬 아이디 등록
	regId := resultData[0]["regId"]
	PushQuery := ""

	params["loginYn"] = "Y"
	if regId == "null" {
		// insert
		PushQuery = usersql.InserPushData
	} else {
		// update
		PushQuery = usersql.UpdatePushData
	}
	InsertPushQuery, err := cls.GetQueryJson(PushQuery, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "push reg query parameter fail"))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(InsertPushQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail(push reg)"))
	}

	//토큰 발행
	c = cls.SetLoginJWT(c, userId)

	// 접속 로그
	params["logInOut"] = "I"
	params["ip"] = ip
	params["succYn"] = "Y"
	params["type"] = "id"

	go LoginAcceesLog(c, params)

	userData := make(map[string]interface{})
	userData["userId"] = userId
	userData["loginId"] = loginId
	userData["userNm"] = userNm
	userData["userTel"] = userTel

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = userData

	return c.JSON(http.StatusOK, m)

}

// 소셜 로그인
func LoginDarayoSocial(c echo.Context) error {

	dprintf(4, c, "call GetUserInfo\n")
	ip, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
	socialType := c.FormValue("socialType")

	LoginSqlQuery := ""
	if socialType == "kakao" {
		LoginSqlQuery = usersql.SelectKakaoUserLoginCheck
	} else if socialType == "naver" {
		LoginSqlQuery = usersql.SelectNaverUserLoginCheck
	} else if socialType == "apple" {
		LoginSqlQuery = usersql.SelectAppleUserLoginCheck
	}

	if LoginSqlQuery == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원하지 않는 로그인 방식 입니다."))
	}
	params := cls.GetParamJsonMap(c)

	socialToken := params["socialToken"]
	resultData, err := cls.GetSelectDataRequire(LoginSqlQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		// 접속 로그
		params["logInOut"] = "I"
		params["ip"] = ip
		params["succYn"] = "N"
		params["type"] = socialType
		params["loginId"] = socialToken
		go LoginAcceesLog(c, params)
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "아이디 또는 비밀번호가 잘못되었습니다."))
	}

	useYn := resultData[0]["USE_YN"]
	userId := resultData[0]["userId"]
	userBirth := resultData[0]["USER_BIRTH"]

	params["userId"] = userId
	if useYn == "N" {

		// 접속 로그
		params["logInOut"] = "I"
		params["ip"] = ip
		params["succYn"] = "N"
		params["type"] = socialType
		params["loginId"] = socialToken
		go LoginAcceesLog(c, params)
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "로그인 실패."))
	}

	userTy := resultData[0]["USER_TY"]
	if userTy != "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "일반 사용자가 아닙니다."))
	}

	// 푸쉬 아이디 등록
	regId := resultData[0]["regId"]
	params["loginYn"] = "Y"
	PushQuery := ""
	if regId == "null" {
		// insert
		PushQuery = usersql.InserPushData
	} else {
		// update
		PushQuery = usersql.UpdatePushData
	}
	InsertPushQuery, err := cls.GetQueryJson(PushQuery, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "push reg query parameter fail"))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(InsertPushQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail(push reg)"))
	}

	c = cls.SetLoginJWT(c, userId)

	// 접속 로그
	//	dprintf(4, c, "call GetUserInfo\n", resultData)
	loginId := resultData[0]["LOGIN_ID"]
	userNm := resultData[0]["USER_NM"]
	userTel := resultData[0]["HP_NO"]

	params["loginId"] = loginId
	params["logInOut"] = "I"
	params["ip"] = ip
	params["succYn"] = "Y"
	params["type"] = socialType
	params["loginId"] = socialToken
	go LoginAcceesLog(c, params)

	userData := make(map[string]interface{})
	userData["userId"] = userId
	userData["loginId"] = loginId
	userData["userNm"] = userNm
	userData["userTel"] = userTel
	userData["userBirth"] = userBirth

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = userData

	return c.JSON(http.StatusOK, m)

}

// 이메일 중복 체크 (아이디 체크)
func GetEmailDupCheck(c echo.Context) error {

	dprintf(4, c, "call GetEmailDupCheck\n")
	params := cls.GetParamJsonMap(c)

	var validEmail, _ = regexp.Compile(
		"^[_a-z0-9+-.]+@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$",
	)
	email := params["email"]

	chkResult := validEmail.MatchString(email)

	if chkResult == false {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "잘못된 email 입니다."))
	}

	resultData, err := cls.GetSelectData(usersql.SelectEmailDupCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "no data"))
	}

	emailCnt := resultData[0]["emailCnt"]

	if emailCnt != "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("01", "Email 중복"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

//소셜 로그인 토큰 중복체크
func GetSocialTokenDupCheck(c echo.Context) error {

	dprintf(4, c, "call GetSocialTokenDupCheck\n")

	socialType := c.FormValue("socialType")

	SqlQuery := ""
	if socialType == "kakao" {
		SqlQuery = usersql.SelectKakaoTokenDupCheck
	} else if socialType == "naver" {
		SqlQuery = usersql.SelectNaverTokenDupCheck
	} else if socialType == "apple" {
		SqlQuery = usersql.SelectAppleTokenDupCheck
	}

	if SqlQuery == "" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "지원하지 않는 로그인 방식 입니다."))
	}

	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectData(SqlQuery, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "no data"))
	}

	emailCnt := resultData[0]["tokenCnt"]

	if emailCnt != "0" {
		return c.JSON(http.StatusOK, controller.SetErrResult("01", "토큰 중복"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

// 푸쉬 정보 업데이트
func PushInfo(c echo.Context) error {

	dprintf(4, c, "call PushInfoUpdate \n")
	params := cls.GetParamJsonMap(c)
	params["loginYn"] = "Y"

	InsertPushQuery, err := cls.GetQueryJson(usersql.UpdatePushData, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "push reg query parameter fail"))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(InsertPushQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail(push reg)"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

func LoginOut(c echo.Context) error {
	params := cls.GetParamJsonMap(c)

	userId := params["userId"]
	params["regId"] = " "
	params["loginYn"] = "N"

	selectQuery, err := cls.SetUpdateParam(usersql.UpdatePushData, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(selectQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	cls.ClearLoginSession(c, userId)

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)
}

func LoginAcceesLog(c echo.Context, params map[string]string) {

	dprintf(4, c, "call LoginAcceesLog\n")
	// 파라메터 맵으로 쿼리 변환
	selectQuery, err := cls.GetQueryJson(usersql.InsertLoginAccess, params)
	if err != nil {
		dprintf(4, c, "LoginAcceesLog query parameter fail\n")
	}
	// 쿼리 실행
	_, err = cls.QueryDB(selectQuery)
	if err != nil {
		dprintf(4, c, "LoginAcceesLog DB fail\n")
	}

}

// 회원 가입
func SetUserJoin(c echo.Context) error {

	dprintf(4, c, "call setUserJoin\n")

	params := cls.GetParamJsonMap(c)

	resultData, err := cls.GetSelectData(usersql.SelectCreatUserSeq, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "유저아이디 생성 실패(2)"))
	}
	userId := resultData[0]["newUserId"]
	params["userId"] = userId
	// 유저 가입  TRNAN 시작
	tx, err := cls.DBc.Begin()
	if err != nil {
		//return "5100", errors.New("begin error")
	}

	// 오류 처리
	defer func() {
		if err != nil {
			// transaction rollback
			dprintf(4, c, "do rollback -유저 가입(setUserJoin)  \n")
			tx.Rollback()
		}
	}()

	// transation exec
	// 파라메터 맵으로 쿼리 변환

	socialType := params["socialType"]
	socialToken := params["socialToken"]
	loginPw := params["loginPw"]

	params["loginId"] = params["email"]

	if socialType == "kakao" {
		params["kakaoKey"] = socialToken
		params["kakaoPw"] = loginPw
		params["loginPw"] = "bcb15f821479b4d5772bd0ca866c00ad5f926e3580720659cc80d39c9d09802a"
	} else if socialType == "naver" {
		params["naverKey"] = socialToken
		params["naverPw"] = loginPw
		params["loginPw"] = "bcb15f821479b4d5772bd0ca866c00ad5f926e3580720659cc80d39c9d09802a"
	} else if socialType == "apple" {
		params["appleKey"] = socialToken
		params["applePw"] = loginPw
		params["loginPw"] = "bcb15f821479b4d5772bd0ca866c00ad5f926e3580720659cc80d39c9d09802a"

		if params["userNm"] == "" {
			params["userNm"] = strings.Replace(userId, "U", "A", -1)
		}
		if params["loginId"] == "" {
			params["loginId"] = strings.Replace(userId, "U", "A", -1)
		}
	}

	termsOfBenefit := params["termsOfBenefit"]
	pushYn := "N"
	if termsOfBenefit == "Y" {
		pushYn = "Y"
	}

	// 유저 생성
	params["userTy"] = "0"
	params["atLoginYn"] = "Y"
	params["pushYn"] = pushYn
	UserCreateQuery, err := cls.SetUpdateParam(usersql.InserCreateUser, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "UserCreateQuery parameter fail"))
	}

	_, err = tx.Exec(UserCreateQuery)
	dprintf(4, c, "call set Query (%s)\n", UserCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", UserCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	TermsCreateQuery, err := cls.SetUpdateParam(usersql.InsertTermsUser, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "TermsCreateQuery parameter fail"))
	}

	_, err = tx.Exec(TermsCreateQuery)
	dprintf(4, c, "call set Query (%s)\n", TermsCreateQuery)
	if err != nil {
		dprintf(1, c, "Query(%s) -> error (%s) \n", TermsCreateQuery, err)
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// transaction commit
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	// 유저 가입 TRNAN 종료

	//기본 장부 생성 시작
	books.SetMakeBookFirst(c, userId)
	//기본 장부 생성 끝

	c = cls.SetLoginJWT(c, userId)

	userData := make(map[string]interface{})
	userData["userId"] = userId

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = userData

	return c.JSON(http.StatusOK, m)

}

func GetSearchId(c echo.Context) error {

	dprintf(4, c, "call GetSearchId\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectTypeRequire(usersql.SelectUserIdSearch, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultList == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "이름, 전화번호에 해당하는 고객을 찾을 수 없습니다."))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultList"] = resultList

	return c.JSON(http.StatusOK, m)

}

func GetSearchPw(c echo.Context) error {

	dprintf(4, c, "call GetSearchPw\n")

	params := cls.GetParamJsonMap(c)
	resultData, err := cls.GetSelectDataRequire(usersql.SelectUserPwSearch, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if resultData == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "로그인 아이디, 전화번호에 해당하는 고객을 찾을 수 없습니다."))
	}

	loginType := resultData[0]["LOGIN_TYPE"]

	if loginType != "ID" {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", loginType+" 로그인을 사용하는 계정입니다. 비밀번호 없이 해당 소셜 로그인을 기능을 이용하면 됩니다."))
	}

	userData := make(map[string]interface{})
	userData["userId"] = resultData[0]["USER_ID"]

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = userData

	return c.JSON(http.StatusOK, m)

}

func SetUserNotice(c echo.Context) error {

	dprintf(4, c, "call SetUserNotice\n")

	params := cls.GetParamJsonMap(c)

	sqlQuery := ""
	storeNotice, err := cls.GetSelectData(usersql.SelectUserNoticeInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if storeNotice == nil {
		sqlQuery = usersql.InsertUserNotice
	} else {
		sqlQuery = usersql.UpdateUserNotice
	}

	exQuery, err := cls.SetUpdateParam(sqlQuery, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(exQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}

func SetChPw(c echo.Context) error {

	dprintf(4, c, "call SetChPw\n")

	params := cls.GetParamJsonMap(c)

	UpdatePwChQuery, err := cls.GetQueryJson(usersql.UpdateUserPassWd, params)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	// 쿼리 실행
	_, err = cls.QueryDB(UpdatePwChQuery)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"

	return c.JSON(http.StatusOK, m)

}
