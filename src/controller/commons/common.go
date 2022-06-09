package commons

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	commonsql "mocaApi/query/commons"
	"mocaApi/src/controller"
	"mocaApi/src/controller/cls"
	"net/http"
	"strconv"
	"strings"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf





// 카테고리
func GetCategoryList(c echo.Context) error {

	dprintf(4, c, "call GetCategoryList\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectType(commonsql.SelectCategoryList, params, c)
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

func GetCodeList(c echo.Context) error {

	dprintf(4, c, "call GetCodeList\n")

	params := cls.GetParamJsonMap(c)
	resultList, err := cls.GetSelectType(commonsql.SelectCodeist, params, c)
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


type VerisonData struct {
	VersionCode     string `json:"versionCode"`     //
	IsRequireUpdate bool   `json:"isRequireUpdate"` //
}

//앱 최신 버전 호출
func GetVersionsLatest(c echo.Context) error {

	dprintf(4, c, "call GetVersionsLatest\n")

	params := cls.GetParamJsonMap(c)
	// 파라메터 맵으로 쿼리 변환
	selectQuery, err := cls.GetQueryJson(commonsql.SelectVersion, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "query parameter fail"))
	}
	// 쿼리 실행 후 JSON 형태로 결과 받기
	dbData, err := cls.QueryJsonColumn(selectQuery, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", "DB fail"))
	}
	var resultData VerisonData
	err = json.Unmarshal(dbData, &resultData)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "Unmarshal fail"))
	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = resultData

	return c.JSON(http.StatusOK, m)

}




type ResultMsgMap struct {
	XMLName  xml.Name `xml:"map"`
	Id string `xml:"id,attr"`
	DetailMsg  string   `xml:"detailMsg"`
	Msg  string   `xml:"msg"`
	Code  string   `xml:"code"`
	Result  string   `xml:"result"`
}

type Map struct {
	XMLName  xml.Name `xml:"map"`
	ResultMsgMap ResultMsgMap  `xml:"map"`
	TrtEndCd string   `xml:"trtEndCd"`
	SmpcBmanEnglTrtCntn  string   `xml:"smpcBmanEnglTrtCntn"`
	NrgtTxprYn  string   `xml:"nrgtTxprYn"`
	SmpcBmanTrtCntn  string   `xml:"smpcBmanTrtCntn"`
	TrtCntn  string   `xml:"trtCntn"`
}


// 사업자 번호 조회
func BizNumCheck(c echo.Context) error {

	params := cls.GetParamJsonMap(c)

	bizNum :=params["bizNum"]


	//테스트용
	if strings.Contains(bizNum,"12345678") {
		m := make(map[string]interface{})
		m["resultCode"] = "00"
		m["resultMsg"] = "응답 성공"
		return c.JSON(http.StatusOK, m)
	}


	url :="https://teht.hometax.go.kr/wqAction.do?actionId=ATTABZAA001R08&screenId=UTEABAAA13&popupYn=false&realScreenId="
	xmlData :="<map id='ATTABZAA001R08'>" +
		"<pubcUserNo/>" +
		"<mobYn>N</mobYn>" +
		"<inqrTrgtClCd>1</inqrTrgtClCd>" +
		"<txprDscmNo>"+bizNum+"</txprDscmNo>" +
		"<dongCode>05</dongCode" +
		"><psbSearch>Y</psbSearch>" +
		"<map id='userReqInfoVO'/>" +
		"</map>"
	buf :=bytes.NewBufferString(xmlData)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var rMap Map
	xml.Unmarshal(body, &rMap)

	checkResult := rMap.SmpcBmanTrtCntn

	if strings.Replace(checkResult, " ", "", -1) != "등록되어있는사업자등록번호입니다." {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", checkResult))
	}


	bizNumCnt, err := cls.GetSelectData(commonsql.SelectCompanyBizNumCheck, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}

	bizCnt,_:= strconv.Atoi(bizNumCnt[0]["bizCnt"])
	if bizCnt > 0 {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "이미 가입된 사업자 번호입니다."))
	}



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	return c.JSON(http.StatusOK, m)
}

func redisConnect(addr string) redis.Conn{
	c, err := redis.Dial("tcp", addr)
	if err != nil{
		lprintf(1, "[ERROR] redis con err(%s) \n", err.Error())
		return nil
	}

	pong, err := redis.String(c.Do("PING"))
	if err != nil{
		lprintf(1, "[ERROR] redis ping pong err(%s) \n", err.Error())
		c.Close()
		return nil
	}

	lprintf(4, "[INFO] redis con ping pong(%s)\n", pong)
	return c
}

func RedisSet(key,value string){

	for _, addr := range controller.RedisAddr{
		c := redisConnect(addr)
		if c == nil{
			return
		}
		defer c.Close()
		reply, err := c.Do("SET", key, value)
		if err != nil{
			lprintf(1, "[ERROR] redis con do err(%s) \n", err.Error())
		}else{
			lprintf(4, "[INFO] redis do reply(%v) \n", reply)
		}

	}
}

func RedisGet(key string) (int, string){

	for _,addr := range controller.RedisAddr{
		c := redisConnect(addr)
		if c == nil{
			continue
		}

		reply, err := redis.String(c.Do("GET", key))
		if err != nil{
			lprintf(1, "[ERROR] redis get err(%s)\n", err.Error())
			c.Close()
			continue
		}

		c.Close()
		return 1, reply
	}

	return -1, ""
}
