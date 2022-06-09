package mocaApi

import (
	apiPush "mocaApi/src/controller/api/push"
	smss "mocaApi/src/controller/api/sms"
	benepicons "mocaApi/src/controller/benepicons"
	books "mocaApi/src/controller/books"
	commons "mocaApi/src/controller/commons"
	gifts "mocaApi/src/controller/gifts"
	homes "mocaApi/src/controller/homes"
	"mocaApi/src/controller/orders"
	rests "mocaApi/src/controller/rests"
	tpays "mocaApi/src/controller/tpays"
	users "mocaApi/src/controller/users"
	wincubes "mocaApi/src/controller/wincubes"

	//공통 api
	"mocaApi/src/controller/cls"

	"github.com/labstack/echo/v4"
)

var lprintf func(int, string, ...interface{}) = cls.Lprintf

func SvcSetting(e *echo.Echo, fname string) *echo.Echo {
	lprintf(4, "[INFO] sql start \n")

	// biz manager
	moca := e.Group("/api/moca")

	moca.GET("/PushTest", apiPush.PushTest) // 푸쉬 테스트

	moca.HEAD("/health", users.BaseUrl) // 헬스체크
	moca.GET("/health", users.BaseUrl)  // 헬스체크

	// login 및 회원가입
	moca.POST("/login", users.LoginDarayo)                      // 로그인
	moca.POST("/logOut", users.LoginOut)                        // 로그아웃
	moca.POST("/socialLogin", users.LoginDarayoSocial)          // 소셜 로그인
	moca.POST("/smsRequest", smss.SendSmsConfirm)               // 문자 인증 요청 // sms api 개발 필요
	moca.POST("/smsCheck", smss.ConfirmCheck)                   // 문자 인증 확인
	moca.GET("/emailCheck", users.GetEmailDupCheck)             // 이메일 중복 체크 (아이디 체크)
	moca.GET("/socialTokenCheck", users.GetSocialTokenDupCheck) // 소셜 로그인 토근 중복확인
	moca.PUT("/pushInfo", users.PushInfo)                       // 푸쉬 정보 업데이트
	moca.POST("/join", users.SetUserJoin)                       // 회원 가입 - 사용자 정보 입력
	moca.POST("/socialJoin", users.SetUserJoin)                 // 소셜 회원 가입

	common := moca.Group("/commons")
	common.GET("/category/:grpCode", commons.GetCategoryList) //공통코드
	common.GET("/code/:categoryId", commons.GetCodeList)      //코드
	common.GET("/versions/latest", commons.GetVersionsLatest) //앱 최신 버전 호출
	common.GET("/bizNumCheck", commons.BizNumCheck)           //사업자 등록증 번호 조회

	//홈
	home := moca.Group("/home")
	home.GET("/:userId", homes.GetHomeData) //홈화면 데이터
	//home.GET("/:userId/amounts", homes.GetUserAmountInfo)      	// 내 장부 잔액 정보
	home.GET("/:userId/grp", homes.GetMyGrpList)   // 내 장부
	home.GET("/:userId/store", homes.GetStoreList) // 장부에 연결된 가맹점
	home.GET("/board", homes.GetBoardList)         // 공지사항 && 이벤트 리스트
	home.GET("/:userId/myInfo", homes.GetMyInfo)   // 개인정보 설정
	home.PUT("/:userId/myInfo", homes.SetMyInfo)   // 개인정보 설정 업데이트

	home.GET("/:userId/ticket", homes.GetMyTicketGrpList) // 장부 QR 리스트 (식권 장부 리스트만)

	//검색
	search := moca.Group("/search")
	search.GET("/:userId/store", rests.GetSearchStoreList)   // 스토어 검색
	search.GET("/:userId/myAuthGrp", rests.GetMyAuthGrpList) // 장부연결시 - 장부장권한 가진 리스트
	search.POST("/:userId/storeReq", rests.SetStoreReq)      // 스토어 등록 요청

	store := moca.Group("/store")
	store.GET("/:restId/storeInfo", rests.GetStoreInfo)      // 스토어 상세
	store.POST("/:restId/favorites", rests.SetStoreFavorite) // 스토어 즐겨찾기
	store.POST("/LinkCancel", rests.SetStoreLinkCancel)      // 장부연결 취소 요청
	store.POST("/storeLink", rests.SetStoreLink)             // 장부연결

	gift := moca.Group("/gift")
	gift.GET("/:userId", gifts.GetGiftHistory)  // 선물함 내역
	gift.GET("/info/:giftId", gifts.GetGiftMsg) // 선물내역 상세
	gift.GET("/ready", gifts.GetGiftReady)      // 선물 보내기 정보
	gift.POST("/send", gifts.SetGiftSend)       // 선물하기
	gift.POST("/recv", gifts.SetGiftRecv)       // 선물받기
	gift.POST("/cancel", gifts.SetGiftCancel)   // 선물취소

	gift.POST("/reject", gifts.SetGiftRecv) // 선물거부

	order := moca.Group("/order")

	order.POST("/:restId/pay", orders.SetOrderPay) // 주문하기

	order.GET("/:restId/category", orders.GetStoreCategory)            // 카테고리 리스트
	order.GET("/:restId/menu", orders.GetStoreMenu)                    // 메뉴 리스트
	order.POST("/:restId/menu/favorites", orders.SetStoreMenuFavorite) // 메뉴 즐겨찾기

	order.GET("/:userId/todayOrder", orders.GetOrderToday)       // 오늘 주문
	order.GET("/:userId/books", orders.GetBooksOrder)            // 장부별 주문
	order.GET("/:userId/booksOrders", orders.GetBookOrdersList)  // 장부별 주문 상세 리스트
	order.GET("/:orderNo", orders.GetOrderInfo)                  // 주문 내역서
	order.GET("/:userId/payment", orders.GetBooksPayment)        // 결제 내역 - 장부별
	order.GET("/:grpId/paymentList", orders.GetBooksPaymentList) // 결제 내역 - 장부별 상세 리스트
	order.GET("/payment", orders.GetPaymentInfo)                 // 결제 내역서
	order.GET("/paymentCancel", orders.GetPaymentCancelInfo)     // 취소 내역서

	book := moca.Group("/books")
	book.GET("/:userId", books.GetBookList)                 // 장부관리 - 장부 리스트
	book.GET("/:grpId/desc", books.GetBookDesc)             // 장부편집 - 구성원, 가게관리, 장부 상세
	book.PUT("/:grpId/desc", books.SetBookDesc)             // 장부 정보 수정
	book.PUT("/:grpId/user", books.SetBookUser)             // 장부 가입 승인 및 탈퇴 처리
	book.POST("/make", books.SetMakeBook)                   // 장부 만들기
	book.GET("/:grpId/inviteLink", books.GetBookInviteLink) //  장부초대 링크
	book.POST("/:grpId/invite", books.SetUserInvite)        //  장부 초대 승인 처리

	book.GET("/:grpId/rest", books.GetBookRestList) // 장부에 연결된 가맹점
	book.GET("/:grpId/user", books.GetBookUserList) // 장부 인원 리스트

	book.GET("/calculate", books.GetReadyCalculate) // 정산하기
	book.GET("/charging", books.GetReadyCharging)   // 충전하기

	book.POST("/paymentCancel", tpays.SetTpayCancel) // tpay 정산&충전 취소
	book.GET("/storeList", books.GetReadyStoreList)  // 가맹점 리스트

	/////////////////////////////////////모카 v2 시작  2021.10.11
	moca_v2 := e.Group("/api/moca/v2")

	moca_v2.GET("/idSearch", users.GetSearchId)         // 회원 - 아이디 찾기
	moca_v2.GET("/pwSearch", users.GetSearchPw)         // 회원 - 비밀번호  찾기
	moca_v2.PUT("/pwCh", users.SetChPw)                 // 회원 - 비밀번호 변경
	moca_v2.POST("/SetUserNotice", users.SetUserNotice) // 개인 알림 정보 설정

	order_v2 := moca_v2.Group("/order")
	order_v2.GET("/:userId/booksOrders", orders.GetBookOrdersList_V2)  // 장부별 주문 상세 리스트
	order_v2.GET("/:grpId/paymentList", orders.GetBooksPaymentList_V2) // 결제 내역 - 장부별 상세 리스트
	order_v2.PUT("/cancel/pay_gift", orders.SetGiftOrderCancel)        // 기프티콘 취소
	order_v2.GET("/:userId/todayOrder", orders.GetOrderToday_V2)       // 오늘 주문
	order_v2.GET("/:orderNo", orders.GetOrderInfo_V2)                  // 주문 내역서
	order_v2.GET("/gifticon", orders.GetGiftiInfo)                     // 기프티콘 불러오기

	order_v2.POST("/:restId/pay_gift", orders.SetOrderGifticon)                // 기프티콘 주문
	order_v2.POST("/:restId/instantPayOrder", orders.SetOrderGifticon_V2)      // 기프티콘 주문 2022.05.10 즉시결제 포함
	order_v2.PUT("/cancel/instantPayOrder", orders.SetOrderGifticon_V2_Cancel) // 미연결 및 즉시결제 취소  - 기프티콘

	store_v2 := moca_v2.Group("/store")
	store_v2.GET("/:restId/storeInfo", rests.GetStoreInfo_V2) // 스토어 상세
	store_v2.GET("/baseStore", rests.GetBaseStore)            // 기본 가맹점

	book_v2 := moca_v2.Group("/books")
	book_v2.POST("/make", books.SetMakeBook_V2)        // 장부 만들기
	book_v2.GET("/:grpId/desc", books.GetBookDesc_V2)  // 장부편집 - 구성원, 가게관리, 장부 상세
	book_v2.PUT("/:grpId/desc", books.SetBookDesc_V2)  // 장부 정보 수정
	book_v2.GET("/:grpId/link", books.GetBookLinkList) // 장부 - 전체 연결 가맹점 정산 정보

	home_v2 := moca_v2.Group("/home")
	home_v2.GET("/myAlim", homes.GetMyAlim) // 홈화면 알림 체크

	//검색
	search_v2 := moca_v2.Group("/search")
	search_v2.GET("/store/barcode", rests.GetBarcodeStoreList) // 바코드 매장 검색
	search_v2.GET("/store", rests.GetSearchStoreList_v2)       // 스토어 검색

	benepicon := moca_v2.Group("/benepicons")
	benepicon.GET("/test", benepicons.GetTest)              // 베네피콘 테스트
	benepicon.GET("/order", benepicons.SetBeneficonOrder)   // 기프티콘 주문
	benepicon.PUT("/cancel", benepicons.SetBeneficonCancel) // 기프티콘 취소
	benepicon.GET("/status", benepicons.SetGifticonStatus)  // 기프티콘 상태 업데이트
	benepicon.GET("/check", benepicons.SetGifticonCheck)    // 기프티콘 사용체크

	wincube := moca_v2.Group("/wincubes")
	wincube.GET("/test", wincubes.GetTest)                        // 윈큐브 테스트
	wincube.GET("/itemList", wincubes.GetWincubeItemList)         // 아이템,가맹점  등록 및 업데이트
	wincube.GET("/order", wincubes.SetWincubeOrder)               // 윈큐브 교환권 주문
	wincube.GET("/status", wincubes.SetGifticonCheck)             // 윈큐브 기프티콘 사용 체크
	wincube.GET("/cancel", wincubes.SetWincubeCancel)             // 윈큐브 기프티콘 취소
	wincube.GET("/check", wincubes.SetWincubeItemUpdate)          // 윈큐브 상품상태 확인
	wincube.GET("/checkBalance", wincubes.GetWincubeCheckBalance) // 윈큐브 잔여한도 체크

	sub_p := moca_v2.Group("/sub")
	sub_p.GET("/myCouponList", orders.GetMyGiftList) // 보유 쿠폰 리스트

	simplePay := moca_v2.Group("/simplePay")
	simplePay.GET("/my", tpays.GetBillingCardList)            // 간편결제 카드 리스트
	simplePay.POST("/genBillikey", tpays.TpayGenBillikey)     // 간편결제 빌링키 생성
	simplePay.POST("/delBillkey", tpays.TpayDelbillkey)       // 간편결제 빌링키 삭제
	simplePay.GET("/billingPwdYn", tpays.TpayBillingPwdYn)    // 간편결제 비밀 설정 여부 확인
	simplePay.POST("/billingPwdChk", tpays.TpayBillingPwdChk) // 간편결제 비밀 번호 확인
	simplePay.PUT("/billingPwd", tpays.TpayBillingPwd)        // 간편결제 비밀번호 등록 및 수정

	//simplePay.POST("/billingPay", tpays.TpayBillingPay)             // 간편결제 빌링키 결제
	//simplePay.POST("/billingPayCancel", tpays.TpayBillingPayCancel) // 간편결제 빌링키 결제 취소

	cls.SetNotLoginUrl("/")
	lprintf(4, "[INFO] page start \n")

	return e

}
