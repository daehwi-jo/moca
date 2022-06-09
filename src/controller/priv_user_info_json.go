package controller

var test string = "abcd"

// 로그인 데이터 조회
var Select_LoginInfo string = `select a.Login_id as LoginId, a.Login_pw as LoginPass 
                                   from priv_user_info  as a
                                   inner join priv_course_user_info as b on a.user_id = b.USER_ID
                                   where b.course_id=1 and a.login_id ='#{userid}'`

// 로그인 정보 조회 JSON
type LoginData struct {
	//	LoginId    string `json:"LoginId"`    // user id
	LoginPass string `json:"LoginPass"` // user name
	//	LoginPass2 string `json:"LoginPass2"` // user name
}
