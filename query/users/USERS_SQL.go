package users


var InsertLoginAccess string =`	INSERT INTO SYS_LOG_ACCESS (
											USER_ID
											,ADDED_DATE
											,LOG_IN_OUT
											,IP
											,SUCC_YN 
											,SERVICE 
											,TYPE
								) VALUES (
									'#{loginId}'
									,SYSDATE()
									,'#{logInOut}'
									,'#{ip}'
									,'#{succYn}'
									,'#{osTy}'
									,'#{type}'
								)
								`



var SelectUserLoginCheck string = `SELECT A.USER_ID as userId
										,ATLOGIN_YN as atloginYn
										,A.USE_YN
										,ifnull(B.REG_ID,'null') AS regId
										,A.USER_TY
										,ifnull(A.USER_BIRTH,'') AS USER_BIRTH
										,A.LOGIN_ID
										,A.USER_NM
										,A.HP_NO
								    FROM priv_user_info AS A
									LEFT OUTER JOIN sys_reg_info AS B ON A.USER_ID = B.USER_ID
									WHERE 
									LOGIN_ID ='#{loginId}'
									AND LOGIN_PW='#{password}'
									ORDER BY A.USER_TY ASC
									`



var InserPushData string = `INSERT INTO SYS_REG_INFO
							(
							USER_ID
							, REG_ID
							, OS_TY
							, REG_DATE
							, LOGIN_YN
							)
							VALUES
							(
							'#{userId}'
							, '#{regId}'
							, '#{osTy}'
							, DATE_FORMAT(now(), '%Y%m%d%H%i%s')
							, '#{loginYn}'
							)
							`
var UpdatePushData string = `UPDATE SYS_REG_INFO
							SET
								REG_ID = '#{regId}',
								OS_TY = '#{osTy}',
								LOGIN_YN = '#{loginYn}',
								REG_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
							WHERE 
							USER_ID = '#{userId}'
							`



var SelectKakaoUserLoginCheck string = `SELECT A.USER_ID as userId
										,ATLOGIN_YN as atloginYn
										,A.USE_YN
										,ifnull(B.REG_ID,'null') AS regId
										,A.USER_TY
										,ifnull(A.USER_BIRTH,'') AS USER_BIRTH
										,A.LOGIN_ID
										,A.USER_NM
										,A.HP_NO
								    FROM priv_user_info AS A
									LEFT OUTER JOIN sys_reg_info AS B ON A.USER_ID = B.USER_ID
									WHERE 
									KAKAO_KEY = '#{socialToken}'
									`

var SelectNaverUserLoginCheck string = `SELECT A.USER_ID as userId
										,ATLOGIN_YN as atloginYn
										,A.USE_YN
										,ifnull(B.REG_ID,'null') AS regId
										,A.USER_TY
										,ifnull(A.USER_BIRTH,'') AS USER_BIRTH
										,A.LOGIN_ID
										,A.USER_NM
										,A.HP_NO
								    FROM priv_user_info AS A
									LEFT OUTER JOIN sys_reg_info AS B ON A.USER_ID = B.USER_ID
									WHERE 
									NAVER_KEY ='#{socialToken}'
									`

var SelectAppleUserLoginCheck string = `SELECT A.USER_ID as userId
										,ATLOGIN_YN as atloginYn
										,A.USE_YN
										,ifnull(B.REG_ID,'null') AS regId
										,A.USER_TY
										,ifnull(A.USER_BIRTH,'') AS USER_BIRTH
										,A.LOGIN_ID
										,A.USER_NM
										,A.HP_NO
								    FROM priv_user_info AS A
									LEFT OUTER JOIN sys_reg_info AS B ON A.USER_ID = B.USER_ID
									WHERE 
									APPLE_KEY ='#{socialToken}'
									`


var SelectEmailDupCheck string = `SELECT count(*) as emailCnt
								FROM priv_user_info
								WHERE 
								LOGIN_ID ='#{email}'`


var SelectKakaoTokenDupCheck string = `SELECT count(*) as tokenCnt
								FROM priv_user_info
								WHERE 
								KAKAO_KEY ='#{socialToken}'
								AND USER_TY ='0'
								`

var SelectNaverTokenDupCheck string = `SELECT count(*) as tokenCnt
								FROM priv_user_info
								WHERE 
								NAVER_KEY ='#{socialToken}'
								AND USER_TY ='0'
								`

var SelectAppleTokenDupCheck string = `SELECT count(*) as tokenCnt
								FROM priv_user_info
								WHERE 
								APPLE_KEY ='#{socialToken}'
								AND USER_TY ='0'
								`


var SelectCreatUserSeq string = `SELECT CONCAT('U',IFNULL(LPAD(MAX(SUBSTRING(USER_ID, -10)) + 1, 10, 0), '0000000001')) as newUserId
								 FROM priv_user_info`



var InserCreateUser string = `INSERT INTO priv_user_info
									(
										USER_ID,
										USER_NM,
										LOGIN_ID,
										LOGIN_PW,
										USER_TY,
										if #{email} != '' then EMAIL,
										HP_NO,
										ATLOGIN_YN,
										GEOLOC_YN,
										PUSH_YN,
										if #{kakaoPw} != '' then KAKAO_PW,
										if #{kakaoKey} != '' then KAKAO_KEY,
										if #{applePw} != '' then APPLE_PW,
										if #{appleKey} != '' then APPLE_KEY,
										if #{naverPw} != '' then NAVER_PW,
										if #{naverKey} != '' then NAVER_KEY,
										if #{recomCode} != '' then RECOM_CODE,
										if #{channelCode} != '' then CHANNEL_CODE,
										if #{userBirth} != '' then USER_BIRTH,
										USE_YN,
										JOIN_DATE
									)
									VALUES
									(
										'#{userId}'
										, '#{userNm}'
										, '#{loginId}'
										, '#{loginPw}'
										, '#{userTy}'
										, '#{email}'
										, '#{userTel}'
										, '#{atLoginYn}'
										, 'Y'
										, '#{pushYn}'
										, '#{kakaoPw}'
										, '#{kakaoKey}'
										, '#{applePw}'
										, '#{appleKey}'
										, '#{naverPw}'
										, '#{naverKey}'
										, '#{recomCode}'
										, '#{channelCode}'
										, '#{userBirth}'
										, 'Y'
										, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										)`



var InsertTermsUser string = `INSERT INTO b_user_terms
							(USER_ID, 
							TERMS_OF_SERVICE, 
							TERMS_OF_PERSONAL, 
							TERMS_OF_PAYMENT, 
							TERMS_OF_BENEFIT, 
							REG_DATE
							)
							VALUES (
							'#{userId}'
							,'#{termsOfService}'
							,'#{termsOfPersonal}'
							,'#{termsOfPayment}'
							,'#{termsOfBenefit}'
							,DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')						
							)
							`

var SelectMyGrpArray string = `SELECT group_concat(GRP_ID)  as myGrp
									FROM priv_grp_user_info
									WHERE 
									USER_ID='#{userId}'
									AND AUTH_STAT='1'
									`

var SelectMyGrpAuth0Array string = `SELECT group_concat(GRP_ID)  as myGrp
									FROM priv_grp_user_info
									WHERE 
									USER_ID='#{userId}'
									AND AUTH_STAT='1'
									And GRP_AUTH ='0'
									`


var SelectMyGrpList string=`
								SELECT A.GRP_ID AS grpId
								,B.GRP_AUTH as grpAuth
								,A.GRP_NM as grpNm
								,'' as orderNo
								,'' as orderDate
								,'' as restNm
								,'' as orderUserCnt
								,'' as totalAmt
								,'' as userNm
								,'' AS orderStat
								,'' AS grpTypeCd
								,CASE B.GRP_AUTH WHEN '0' THEN  IFNULL((SELECT COUNT(*)
																	FROM dar_order_info  AS AA
																	WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
																	AND A.GRP_ID = AA.GRP_ID ),0) 
								 ELSE  	(SELECT COUNT(*)
											FROM dar_order_info AS AA
											INNER JOIN (SELECT ORDER_NO,AA.USER_ID
													FROM dar_order_detail AS AA
												WHERE  
												LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
												AND AA.USER_ID='#{userId}'
											  ) AS CC ON AA.ORDER_NO  =CC.ORDER_NO 
											WHERE AA.ORDER_STAT='20')
								 END AS thisMonthOrder
								FROM  priv_grp_info AS A
								INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID  AND B.AUTH_STAT='1'
								WHERE 
								B.USER_ID='#{userId}'
								AND A.AUTH_STAT='1'
								`





var SelectUserInfo string = `SELECT USER_NM
                                   ,HP_NO
                             FROM PRIV_USER_INFO 
							 WHERE 
								USER_ID='#{userId}'
									`




var SelectUserIdSearch string = `SELECT 
									CASE 	WHEN kakao_key IS NOT NULL THEN ''
											WHEN apple_key IS NOT NULL THEN ''
											WHEN naver_key IS NOT NULL THEN ''
									ELSE CONCAT(LEFT(CONCAT(SUBSTRING(LOGIN_ID,1,2),'*********************')
											,CASE INSTR(LOGIN_ID,"@")-1 WHEN  -1 THEN LENGTH(LOGIN_ID)  ELSE INSTR(LOGIN_ID,"@")-1 END) 
											,SUBSTRING(LOGIN_ID,INSTR(LOGIN_ID,"@"),30))
									 END AS LOGIN_ID
								   ,DATE_FORMAT(JOIN_DATE, '%Y.%m.%d') AS JOIN_DATE
								   ,CASE 	WHEN kakao_key IS NOT NULL THEN 'KAKAO'
												WHEN apple_key IS NOT NULL THEN 'APPLE'
												WHEN naver_key IS NOT NULL THEN 'NAVER'
									ELSE 'ID' END AS LOGIN_TYPE
								FROM priv_user_info
								WHERE 
								USER_NM='#{userNm}'
								AND HP_NO='#{userTel}'
								AND USER_TY='0'
								`

var SelectUserPwSearch string = `SELECT USER_ID
									 ,CASE 	WHEN kakao_key IS NOT NULL THEN 'KAKAO'
												WHEN apple_key IS NOT NULL THEN 'APPLE'
												WHEN naver_key IS NOT NULL THEN 'NAVER'
										ELSE 'ID' END AS LOGIN_TYPE
									FROM priv_user_info
									WHERE
									LOGIN_ID = '#{loginId}'
									AND HP_NO='#{userTel}'
									AND USER_TY='0'
								`

var UpdateUserPassWd string = `UPDATE priv_user_info
							SET
								LOGIN_PW = '#{loginPw}',
								MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
							WHERE 
							USER_ID = '#{userId}'
								`



var SelectUserNoticeInfo string = `SELECT VALUE
									FROM priv_user_notice
									WHERE 
									NOTICE ='#{notice}'
									AND USER_ID='#{userId}'
									AND REST_ID='#{restId}'
									AND GRP_ID='#{grpId}'
								`



var InsertUserNotice string = `INSERT INTO priv_user_notice
							(	
								NOTICE,
								USER_ID,
								if #{restId} != '' then REST_ID,
								if #{grpId} != '' then GRP_ID,
								VALUE
							)
							VALUES (
							'#{notice}'
							,'#{userId}'
							,'#{restId}'
							,'#{grpId}'	
							,'#{value}'
							)
								`

var UpdateUserNotice string = `UPDATE priv_user_notice
							SET
								VALUE = '#{value}',
							WHERE 
							NOTICE = '#{notice}'
							AND USER_ID='#{userId}'
							AND REST_ID='#{restId}'
							AND GRP_ID='#{grpId}'
							`