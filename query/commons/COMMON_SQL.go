package commons



var PagingQuery string =`
					LIMIT #{pageSize}
					OFFSET #{offSet}
`
var SelectNotice string = `SELECT
							BOARD_ID
							,TITLE
							,BOARD_TYPE
							,LINK_URL
							,DATE_FORMAT(REG_DATE, '%Y.%m.%d') AS regDate	
							FROM sys_boards
							WHERE 
							BOARD_TYPE='0'
							AND START_DATE <= NOW()
							AND END_DATE >= NOW()
							AND B_KIND IN ('A','0')
							AND MAIN_YN='Y'
							ORDER BY REG_DATE desc
							LIMIT 1
							`
var SelectCompanyBizNumCheck string = `SELECT COUNT(*) as bizCnt
							FROM b_company
							WHERE
							BUSID='#{bizNum}'
							`



var SelectBoardList string = `SELECT
							BOARD_ID
							,TITLE
							,BOARD_TYPE
							,LINK_URL
							,DATE_FORMAT(REG_DATE, '%Y.%m.%d') AS regDate	
							FROM sys_boards
							WHERE 
							START_DATE <= NOW() AND END_DATE >=NOW()
							AND B_KIND IN ('A','0')
							`

var SelectCategoryList string = `SELECT CATEGORY_ID
									,CATEGORY_NM 
								FROM b_category
								WHERE 
								CATEGORY_GRP_CODE='#{grpCode}'
								AND USE_YN='Y'
							`
var SelectCodeist string = `SELECT CODE_ID
								,CODE_NM
								FROM b_code
								WHERE
								CATEGORY_ID='#{categoryId}'
								AND USE_YN='Y'
							`

// 버전
var SelectVersion string = `SELECT 
							  VERSION AS versionCode 
							, CASE WHEN AUTO_YN ='Y' THEN 'true' ELSE 'false' END AS  isRequireUpdate
							FROM sys_version_info
							WHERE 
							USE_YN = 'Y' 
							AND OS_TY = '#{osTy}'
							AND APP_TY = '#{appTy}'
							ORDER BY VERSION_ID DESC
							LIMIT 1
							`


var InsertPushLog string = `INSERT INTO sys_log_push
						(
						NAME
						, TITLE
						, BODY
						, APP_TY
						, REG_DATE
						, REG_ID
						)
						VALUES
						(
                          '#{name}'
						, '#{title}'
						, '#{body}'
						, '#{appTy}'
						, NOW()
						, '#{regId}'
						)
						`

var SelectPushGrp string = `SELECT
								A.USER_ID
								, A.REG_ID
								, A.OS_TY
								, A.LOGIN_YN
								, C.PUSH_YN
							FROM 
								SYS_REG_INFO A,
								PRIV_USER_INFO C
							WHERE 
								EXISTS (SELECT USER_ID FROM PRIV_GRP_USER_INFO B
										WHERE 
										A.USER_ID = B.USER_ID 
										AND B.GRP_ID = '#{grpId}'
										AND B.GRP_AUTH = '0'
								)
							AND A.LOGIN_YN = 'Y'
							AND A.USER_ID = C.USER_ID
							`


var SelectPushUser string = `SELECT
							A.USER_ID
							, A.REG_ID
							, A.OS_TY
							, A.REG_DATE
							, A.LOGIN_YN
							, B.PUSH_YN
							FROM
							SYS_REG_INFO A,
							PRIV_USER_INFO B
							WHERE
							A.USER_ID = B.USER_ID
							AND B.USE_YN = 'Y'
							AND A.USER_ID = '#{userId}'
							`

var SelectPushBizNum string = `SELECT C.USER_ID
							, C.REG_ID
							, C.OS_TY
							, C.LOGIN_YN
							, B.PUSH_YN
							FROM PRIV_REST_INFO AS A
							INNER JOIN PRIV_REST_USER_INFO AS B ON A.REST_ID = B.REST_ID
							INNER JOIN SYS_REG_INFO AS C ON B.USER_ID = C.USER_ID
							WHERE
							BUSID='#{bizNum}'
							AND B.REST_AUTH=0
							`


var SelectPushRest string = `SELECT C.USER_ID
							, C.REG_ID
							, C.OS_TY
							, C.LOGIN_YN
							, B.PUSH_YN
							FROM PRIV_REST_INFO AS A
							INNER JOIN PRIV_REST_USER_INFO AS B ON A.REST_ID = B.REST_ID
							INNER JOIN SYS_REG_INFO AS C ON B.USER_ID = C.USER_ID
							WHERE
							A.REST_ID='#{restId}'
							AND B.REST_AUTH=0
							`


var SelectPushMsgInfo string = `SELECT 	MSG
									,TITLE
								FROM dar_msg_info
								WHERE 
								MSG_CODE='#{msgCode}'
								LIMIT 1;
							`
