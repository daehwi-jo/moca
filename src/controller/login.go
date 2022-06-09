package controller

import (
	"net/http"

	"mocaApi/src/controller/cls"

	"github.com/labstack/echo/v4"
)

/* log format */
// 로그 레벨(5~1:INFO, DEBUG, GUIDE, WARN, ERROR), 1인 경우 DB 롤백 필요하며, 에러 테이블에 저장
// darayo printf(로그레벨, 요청 컨텍스트, format, arg) => 무엇을(서비스, 요청), 어떻게(input), 왜(원인,조치)
var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

func Health(c echo.Context) error {
	return c.String(http.StatusOK, "health")
}

// 주간
func GetPost(c echo.Context) error {

	m := make(map[string]interface{})
	m["s_Siteid"] = ""

	//lprintf(4, "[INFO] pageNum : %s, siteId : %s\n", pageNum, siteId)

	return c.Render(http.StatusOK, "address.html", m)
}

/*
// login
func LoginWeb(c echo.Context) error {
	dprintf(4, c, "call Login Web\n")

	// request JSON body map으로 변환
	params := cls.GetParamJsonMap(c)
	// request parameter를 map으로 변환
	// params := cls.GetParamMap(c)

	// 파라메터 맵으로 쿼리 변환
	loginQuery, err := cls.GetQueryJson(Select_LoginInfo, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, SetErrResult(err.Error(), "query parameter fail"))
	}

	// 쿼리 실행 후 JSON 형태로 결과 받기
	var loginInfo LoginData
	resultList, err := cls.QueryJsonColumn(loginQuery, &loginInfo, c, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, SetErrResult(string(resultList), "로그인 실패"))
	}
	dprintf(4, c, "return scan : %s\n", resultList)

	// convert json 리스트로 변환
	var loginJson []LoginData
	err = json.Unmarshal(resultList, &loginJson)
	if err != nil {
		return c.JSON(http.StatusOK, SetErrResult("5000", err.Error()))
	}

	// 필요 파라메터 추출
	userPw := params["passwd"]
	userId := params["userid"]
	shaPass := sha256.Sum256([]byte(userPw))

	// login fail check
	passString := fmt.Sprintf("%x", shaPass)
	if loginJson == nil {
		dprintf(1, c, "login Data is null\n")
		return c.String(http.StatusInternalServerError, "error")
	}

	if loginJson[0].LoginPass != passString {
		dprintf(1, c, "pass error : %s != (%s)\n", passString, loginJson[0].LoginPass)
		return c.String(http.StatusInternalServerError, "error")
	}

	dprintf(4, c, "login success : %s\n", userId)

	m := make(map[string]interface{})
	m["result"] = "success"
	m["resultCode"] = "0000"
	m["msg"] = "로그인 성공"
	m["userId"] = userId

	// 세션 생성시 JWT 도 같이 생성된다.
	c = cls.SetLoginSession(c, userId)
	return c.String(http.StatusOK, "ok")
}

func InputPage(c echo.Context) error {
	i_selectCourse := c.QueryParam("courseid")

	m := make(map[string]interface{})
	m["SelectCourse"] = i_selectCourse
	return c.Render(http.StatusOK, "input.html", m)
}
func LoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "signin.html", echo.Map{})
}

func ListPage(c echo.Context) error {
	return c.JSONP(http.StatusOK, "", "")
}

func InfoPage(c echo.Context) error {

	courseid := c.QueryParam("courseid")

	m := make(map[string]interface{})
	m["courseid"] = courseid

	return c.Render(http.StatusOK, "course_desc.html", m)
}

// 사용(주문) 내역 조회
func OrderList(c echo.Context) error {
	// JSON body map으로 변환
	params := cls.GetParamJsonMap(c)

	// parameter로 쿼리 매핑
	orderQuery, err := cls.GetQueryJson(Query.Select_OrderInfo, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, SetErrResult(err.Error(), "query parameter fail"))
	}

	var myJson UseHistory
	resultList, err := cls.QueryJsonColumn(orderQuery, &myJson, c, false)
	if err != nil {
		return c.JSON(http.StatusOK, SetErrResult(string(resultList), err.Error()))
	}
	var dbJson []UseHistory
	err = json.Unmarshal(resultList, &dbJson)
	if err != nil {
		return c.JSON(http.StatusOK, SetErrResult("5000", err.Error()))
	}

	var resp ResponseHeader
	resp.Result = "true"
	resp.ResultCode = "0000"
	resp.Msg = "지원금 사용내역 조회 성공"
	resp.Data = dbJson
	return c.JSON(http.StatusOK, resp)
}

/*
func CourseList(c echo.Context) error {
	var myJson CourseIngList

	resultList, err := cls.GetResultByColumn(Query.Select_CourseList, &myJson, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult(string(resultList), err.Error()))
	}
	var dbJson []CourseIngList
	err = json.Unmarshal(resultList, &dbJson)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult("5000", err.Error()))
	}
	lprintf(4, "[INFO] result %v", dbJson)

	m := make(map[string]interface{})
	m["CourseList"] = dbJson
	return c.JSON(http.StatusOK, m)
}

func CourseDesc(c echo.Context) error {
	var myJson CourseInfoDesc

	resultList, err := cls.GetSelectName(Query.Select_CourseDesc, &myJson, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult(string(resultList), err.Error()))
	}
	var dbJson []CourseInfoDesc
	err = json.Unmarshal(resultList, &dbJson)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult("5000", err.Error()))
	}
	lprintf(4, "[INFO] result %v", dbJson)

	m := make(map[string]interface{})
	m["CourseDesc"] = dbJson
	return c.JSON(http.StatusOK, m)
}

func CourseUser(c echo.Context) error {
	var myJson CourseUserList

	resultList, err := cls.GetSelectName(Query.Select_CourseUserList, &myJson, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult(string(resultList), err.Error()))
	}
	var dbJson []CourseUserList
	err = json.Unmarshal(resultList, &dbJson)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult("5000", err.Error()))
	}
	lprintf(4, "[INFO] result %v", dbJson)

	m := make(map[string]interface{})
	m["CourseUser"] = dbJson
	return c.JSON(http.StatusOK, m)
}

// 강좌 목록 조회
func CourseData(c echo.Context) error {
	var myJson CourseInfoList
	i_selectCourse := c.QueryParam("selected")

	resultList, err := cls.GetSelectName(Query.Select_CourseInfo, &myJson, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult(string(resultList), err.Error()))
	}
	var dbJson []CourseInfoList
	err = json.Unmarshal(resultList, &dbJson)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, setErrResult("5000", err.Error()))
	}
	lprintf(4, "[INFO] result %v", dbJson)

	m := make(map[string]interface{})
	m["CourseList"] = dbJson
	m["SelectCourse"] = i_selectCourse
	return c.JSON(http.StatusOK, m)
}

// 강좌 사람 입력
func SetCourseMember(c echo.Context) error {
	i_name := c.FormValue("cc_name")
	i_phone := c.FormValue("cc_number")
	i_courseId := c.QueryParam("courseId")

	procName := fmt.Sprintf("CALL sp_course_insert_single('%s', '%s', '%s');", i_courseId, i_name, i_phone)
	lprintf(4, "[INFO] procName : %s\n", procName)
	rows, dberr := cls.QueryDB(procName)
	if dberr != nil {
		lprintf(1, "[ERR ] %s call error : %s\n", procName, dberr)
		return c.Render(http.StatusInternalServerError, "", nil)
	}
	defer rows.Close()

	// 응답코드
	result := cls.GetRespCode(rows, procName)
	lprintf(4, "[INFO] procName : %s result(%d)\n", procName, result)
	if result != 0 {
		if result == -2 {
			return c.JSON(http.StatusInternalServerError, setErrResult("5000", "동일 과정이나 교육기간이 겹치는 교육에 등록된 교육생입니다."))
		} else if result == -3 {
			return c.JSON(http.StatusInternalServerError, setErrResult("5000", "기존 회원 정보와 일치하지 않습니다.(이름, 휴대폰 확인)"))
		}
		return c.JSON(http.StatusInternalServerError, setErrResult("5000", "등록시 오류가 발생했습니다."))
	}

	m := make(map[string]interface{})
	m["result"] = result

	return c.JSON(http.StatusOK, nil)
}
*/
