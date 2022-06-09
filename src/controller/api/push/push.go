package pushs

import (
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	commonsql "mocaApi/query/commons"
	"mocaApi/src/controller/cls"
	"net/http"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/labstack/echo/v4"
)

/* log format */
// 로그 레벨(5~1:INFO, DEBUG, GUIDE, WARN, ERROR), 1인 경우 DB 롤백 필요하며, 에러 테이블에 저장
// darayo printf(로그레벨, 요청 컨텍스트, format, arg) => 무엇을(서비스, 요청), 어떻게(input), 왜(원인,조치)
var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

type SendData struct {
	Data  Sdata	  `json:"data"`    // /
	To    string `json:"to"`    // /
}

type Sdata struct {
	PageId   string `json:"pageId"`   //
	Title string `json:"title"` // /
	Content string `json:"content"` // /
}





//푸쉬 보내기 (메세지 설정 타입) -- 2021.09.03 버전
func SendPush_Msg_V1(title,content,pushType,userType,typeValue,param,menu string)  {

	pushParam := make(map[string]string)


	//유저 타입 과 파라미터
	//0 : 사용자 	   userId
	//1 : 가맹점(사장님) restId
	//2 : 가맹점(사장님) bizNum
	//3 : 사용자(장부장) grpId
	pushQuery :=""
	if userType =="0" {
		pushParam["userId"]=typeValue
		pushQuery =commonsql.SelectPushUser
	}else if userType =="1"{
		pushParam["restId"]=typeValue
		pushQuery =commonsql.SelectPushRest
	}else if userType =="2"{
		pushParam["bizNum"]=typeValue
		pushQuery =commonsql.SelectPushBizNum
	}else if userType =="3"{
		pushParam["grpId"]=typeValue
		pushQuery =commonsql.SelectPushGrp
	}else{
		lprintf(1, "[ERROR] error SendPush_V1: 잘못된 유저 타입입니다")
		return
	}


	resultData, err := cls.SelectData(pushQuery, pushParam)
	if err != nil {
		lprintf(1, "[ERROR]  error SendPush_V1: : %s\n", err)
		return
	}
	if resultData == nil {
		lprintf(1, "[ERROR] error SendPush_V1: 유저 데이터가 없습니다", err)
		return
	}

	rtoken := resultData[0]["REG_ID"]
	osTy := resultData[0]["OS_TY"]

	SendTotal(title,content,rtoken,pushType,menu,param,osTy)

}



func PushTest(c echo.Context) error {


	params := cls.GetParamJsonMap(c)

	rtoken:=params["regId"]
	title:=params["title"]
	content:=params["content"]
	menu:=params["menu"]
	osTy:=params["osTy"]
	pushType:="M"
	param:=params["param"]

	SendTotal(title,content,rtoken,pushType,menu,param,osTy)

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] ="전송완료 "

	return c.JSON(http.StatusOK, m)

}

// 모든기기 전송
func SendTotal(title, content,rtoken ,pushType,menu,param,osTy string) {

	var app *firebase.App

	opt := option.WithCredentialsFile("conf/darayo_push_key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		lprintf(1, "[ERROR] error getting Messaging: : %s\n", err)
	}

	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		//log.Fatalf("error getting Messaging client: %v\n", err)
		lprintf(1, "[ERROR] error getting Messaging: : %s\n", err)
	}

	oneHour := time.Duration(1) * time.Hour
	//badge := 42
	registrationToken := rtoken

	/// pushType  N 일반, D 중복처리, M  이동
	message := &messaging.Message{}
	if osTy=="A" {
		message = &messaging.Message{

			Data: map[string]string{
				"Title": title,
				"Body":  content,
				"Type" : pushType,
				"menu" : menu,
				"param" : param,
			},
			Android: &messaging.AndroidConfig{
				TTL: &oneHour,
			},
			Token: registrationToken,
		}
	}else{

		message = &messaging.Message{
			Notification: &messaging.Notification{
				Title: title,
				Body:  content,
			},
			Data: map[string]string{
				"title": title,
				"body":  content,
				"type" : pushType,
				"menu" : menu,
				"param" : param,
			},
			//	APNS: &messaging.APNSConfig{
			//		Payload: &messaging.APNSPayload{
			//			Aps: &messaging.Aps{
			//				Badge: &badge,
			//			},
			//		},
			//	},
			Token: registrationToken,
		}

	}


	params := make(map[string]string)
	params["name"] = "푸쉬"
	params["title"] = title
	params["body"] = content
	params["appTy"] = osTy
	params["regId"] = registrationToken
	// 파라메터 맵으로 쿼리 변환
	insertMenuQuery, err := cls.GetQueryJson(commonsql.InsertPushLog, params)
	if err != nil {
		lprintf(4, "insert push log error : %s\n", err)
	}
	// 쿼리 실행
	_, err = cls.QueryDB(insertMenuQuery)
	if err != nil {
		lprintf(4, "insert push log error : %s\n", err)
	}


	response, err := client.Send(ctx, message)
	if err != nil {
		lprintf(4, "send push error client: : %s\n", err)
	}
	lprintf(4, "send push result : %s\n", response)
	//fmt.Println("Successfully sent message:", response)
}




