package books

var SelectUserBookInfo string = `SELECT A.AUTH_STAT
								,C.LIMIT_YN
								,IFNULL(C.LIMIT_AMT,0) AS LIMIT_AMT
								,C.SUPPORT_YN
								,C.CHECK_TIME
								,A.SUPPORT_BALANCE
								,C.SUPPORT_EXCEED_YN
								,CASE WHEN C.GRP_PAY_TY <> D.PAY_TY THEN D.PAY_TY ELSE C.GRP_PAY_TY END AS GRP_PAY_TY
								,IFNULL(C.LIMIT_DAY_AMT,0) AS LIMIT_DAY_AMT
								,IFNULL(LIMIT_USE_TIME_START,0) AS LIMIT_USE_TIME_START
								,IFNULL(LIMIT_USE_TIME_END,0) AS LIMIT_USE_TIME_END
								,IFNULL(LIMIT_DAY_CNT,0) AS LIMIT_DAY_CNT
								,B.USER_NM
								,B.HP_NO AS USER_TEL
								FROM PRIV_GRP_USER_INFO AS A
								INNER JOIN PRIV_USER_INFO AS B ON A.USER_ID = B.USER_ID
								INNER JOIN PRIV_GRP_INFO AS C ON A.GRP_ID = C.GRP_ID
								INNER JOIN org_agrm_info AS D ON C.GRP_ID =D.GRP_ID AND D.REQ_STAT='1'
								INNER JOIN priv_rest_info AS E on D.REST_ID = E.rest_id or (E.FRAN_YN = 'Y' and E.FRAN_ID = D.REST_ID)
								WHERE 
								A.AUTH_STAT='1'   
								AND  A.USER_ID = '#{userId}'
								AND  A.GRP_ID = '#{grpId}'
								AND  E.REST_ID ='#{restId}'
									`

var UpdateBookUserSupportBalance string = `UPDATE PRIV_GRP_USER_INFO SET  
											SUPPORT_BALANCE = SUPPORT_BALANCE - #{orderAmt},
											MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
											WHERE 
											GRP_ID = '#{grpId}'
											AND USER_ID ='#{userId}'
`

var UpdateBookUserCancelSupportBalance string = `UPDATE PRIV_GRP_USER_INFO SET  
											SUPPORT_BALANCE = SUPPORT_BALANCE + #{orderAmt},
											MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
											WHERE 
											GRP_ID = '#{grpId}'
											AND USER_ID ='#{userId}'
`

var SelectMyBookList string = `SELECT *
								FROM (SELECT  A.GRP_ID
									,A.GRP_NM
									,A.GRP_TYPE_CD
									,A.GRP_PAY_TY
									,A.SUPPORT_YN
									,FN_GET_GRP_MGR_NAME(A.GRP_ID) AS GRP_MANAGER
									,FN_GET_GRP_MEM_CNT(A.GRP_ID) AS GRP_USER_CNT
									,CASE WHEN A.GRP_TYPE_CD ='1' THEN  company_nm ELSE '' END AS COMPANY_NM
									,IFNULL(A.INTRO,'') AS INTRO
									,CASE WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='Y')  THEN 'D1'
											WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='N')  THEN 'D2'
											WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='Y')  THEN 'D3'
											WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='N')  THEN 'D4'
											WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='4')  THEN 'D5'
											WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='4')  THEN 'D6'																										
											ELSE 'D5' END AS D_TYPE
									,'' AS IMG_URL
									,IFNULL((SELECT IFNULL(SUM(TOTAL_AMT),0) 
													FROM dar_order_info  AS AA
													WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
													AND A.GRP_ID = AA.GRP_ID AND AA.ORDER_TY !='4'),0) AS THIS_MONTH_ORDER
									,IFNULL(B.SUPPORT_BALANCE,0) AS SUPPORT_BALANCE
									,IFNULL((SELECT SUM(PREPAID_AMT) FROM org_agrm_info AS AA WHERE A.GRP_ID = AA.GRP_ID AND  AA.REQ_STAT='1'),0) AS PREPAID_AMT
									,IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
										FROM dar_order_info AS AA 
										INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
										WHERE B.GRP_ID = AA.GRP_ID AND AA.ORDER_STAT='20' AND AA.PAID_YN='N' AND ORDER_TY <>'4' AND PAY_TY='1'),0) AS POST_PAID_AMT
								FROM  priv_grp_info AS A
								INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID  AND B.AUTH_STAT='1'
								LEFT OUTER JOIN b_company_book AS C ON A.GRP_ID = C.BOOK_ID 
								LEFT OUTER JOIN b_company AS D ON C.company_id = D.company_id
								WHERE 
									B.USER_ID='#{userId}'
									AND A.AUTH_STAT='1'
									AND B.GRP_AUTH='0' 
									AND A.USE_YN='Y') AS ZZ
								ORDER BY PREPAID_AMT DESC , THIS_MONTH_ORDER DESC
					`

var SelectMemberBookList string = `SELECT *
								FROM (SELECT  A.GRP_ID
									,A.GRP_NM
									,A.GRP_TYPE_CD
									,A.GRP_PAY_TY
									,A.SUPPORT_YN
									,FN_GET_GRP_MGR_NAME(A.GRP_ID) AS GRP_MANAGER
									,FN_GET_GRP_MEM_CNT(A.GRP_ID) AS GRP_USER_CNT
									,CASE WHEN A.GRP_TYPE_CD ='1' THEN  company_nm ELSE '' END AS COMPANY_NM
									,IFNULL(A.INTRO,'') AS INTRO
									,CASE WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='Y')  THEN 'D7'
											WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='N')  THEN 'D8'
											WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='Y')  THEN 'D7'
											WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='N')  THEN 'D8'
											WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='4')  THEN 'D9'
											WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='4')  THEN 'D9'																										
											ELSE 'D9' END AS D_TYPE
									,'' AS IMG_URL
									,IFNULL((SELECT IFNULL(SUM(BB.ORDER_AMT * BB.ORDER_QTY),0) 
													FROM dar_order_info  AS AA
													INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
													WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
													AND A.GRP_ID = AA.GRP_ID AND B.USER_ID = BB.USER_ID AND AA.ORDER_TY !='4' ),0) AS THIS_MONTH_ORDER
									,IFNULL(B.SUPPORT_BALANCE,0) AS SUPPORT_BALANCE
									,IFNULL((SELECT SUM(PREPAID_AMT) FROM org_agrm_info AS AA WHERE A.GRP_ID = AA.GRP_ID AND  AA.REQ_STAT='1'),0) AS PREPAID_AMT
									,IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
										FROM dar_order_info AS AA 
										INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
										WHERE B.GRP_ID = AA.GRP_ID AND AA.ORDER_STAT='20' AND AA.PAID_YN='N' AND ORDER_TY <>'4' AND PAY_TY='1' AND BB.USER_ID = B.USER_ID),0) AS POST_PAID_AMT
								FROM  priv_grp_info AS A
								INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID  AND B.AUTH_STAT='1'
								LEFT OUTER JOIN b_company_book AS C ON A.GRP_ID = C.BOOK_ID 
								LEFT OUTER JOIN b_company AS D ON C.company_id = D.company_id
								WHERE 
									B.USER_ID='#{userId}'
									AND A.AUTH_STAT='1'
									AND B.GRP_AUTH='1' 
									AND A.USE_YN='Y') AS ZZ
						   ORDER BY PREPAID_AMT DESC , THIS_MONTH_ORDER DESC
					`

var SelectBookUserList string = `SELECT A.USER_ID
										,B.USER_NM
										,B.HP_NO
										,A.AUTH_STAT
										,A.GRP_AUTH
								FROM priv_grp_user_info AS A
								INNER JOIN priv_user_info AS B ON A.USER_ID = B.USER_ID
								WHERE 
								A.GRP_ID='#{grpId}'
								AND A.AUTH_STAT !='3'
								ORDER BY A.AUTH_STAT ASC
								`

var SelectBookLinkStoreList string = `SELECT A.REST_ID
												 ,B.REST_NM
												 ,B.TEL
										FROM org_agrm_info AS A
										INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID AND REQ_STAT='1'
										WHERE 
										A.GRP_ID='#{grpId}'
										`
var SelectBookDesc string = `SELECT A.GRP_ID
									 ,A.GRP_NM
									 ,A.GRP_TYPE_CD
									 ,DATE_FORMAT(A.REG_DATE, '%Y년 %m월 %d일') AS REG_DATE
									 ,IFNULL(A.INTRO,'') AS INTRO
									 ,LIMIT_YN
									 ,IFNULL(LIMIT_USE_TIME_START,0) AS LIMIT_USE_TIME_START
									 ,IFNULL(LIMIT_USE_TIME_END,0) AS LIMIT_USE_TIME_END
									 ,LIMIT_AMT
									 ,LIMIT_DAY_AMT
									 ,LIMIT_DAY_CNT
									 ,IFNULL(SUPPORT_AMT,'0') AS SUPPORT_AMT
									 ,IFNULL(SUPPORT_FORWARD_YN,'N') AS SUPPORT_FORWARD_YN
									 ,IFNULL(FIRST_SUPPORT_AMT,'0') AS FIRST_SUPPORT_AMT
                                     ,FN_GET_GRP_MEM_CNT(A.GRP_ID) AS USER_CNT
									 ,(SELECT COUNT(*) FROM org_agrm_info AS AA WHERE A.GRP_ID = AA.GRP_ID AND AA.REQ_STAT='1') AS REST_CNT
							FROM priv_grp_info AS A
							WHERE
								A.GRP_ID='#{grpId}'
								`

var UpdateBookDesc string = `UPDATE priv_grp_info SET
								 MOD_DATE=DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								 ,GRP_TYPE_CD= '#{grpTypeCd}'
								 ,GRP_NM= '#{grpNm}'
								 ,INTRO= '#{intro}'
								 ,LIMIT_YN= '#{limitYn}'
								 ,LIMIT_USE_TIME_START= #{limitUseTimeStart}
								 ,LIMIT_USE_TIME_END= #{limitUseTimeEnd}
								 ,LIMIT_AMT= #{limitAmt}
								 ,LIMIT_DAY_AMT= #{limitDayAmt}
								 ,LIMIT_DAY_CNT= #{limitDayCnt}
								 ,SUPPORT_YN= '#{supportYn}'
								 ,SUPPORT_AMT= '#{supportAmt}'
							WHERE
								GRP_ID='#{grpId}'
								`
var UpdateBookUserReg string = `UPDATE priv_grp_user_info SET 
										 AUTH_STAT= '#{authStat}'
										,AUTH_DATE=DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										,SUPPORT_BALANCE='#{supportBalance}'
										,MOD_DATE=DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
									WHERE 
									GRP_ID='#{grpId}'
									AND USER_ID='#{userId}'	
								`

var UpdateBookUserLeave string = `UPDATE priv_grp_user_info SET AUTH_STAT='3'
										,MOD_DATE=DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										,LEAVE_TY='0'
									WHERE 
									GRP_ID='#{grpId}'
									AND USER_ID='#{userId}'	
								`

var UpdateBookIviteLink string = `UPDATE priv_grp_info SET 
										 INVITE_LINK='#{inviteLink}'
										,MOD_DATE=DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
									WHERE 
									GRP_ID='#{grpId}'
								`

var SelectCreatBookSeq string = `SELECT CONCAT('B',IFNULL(LPAD(MAX(SUBSTRING(GRP_ID, -10)) + 1, 10, 0), '0000000001')) as newGrpId
							FROM priv_grp_info
							`

var InsertCreateBook string = `INSERT INTO priv_grp_info 
							(
							GRP_ID,
							GRP_NM,
							GRP_TYPE_CD,
							AUTH_STAT,
							if #{limitYn} != '' then LIMIT_YN,
							if #{limitUseTimeStart} != '' then LIMIT_USE_TIME_START,
							if #{limitUseTimeEnd} != '' then LIMIT_USE_TIME_END,
							if #{limitAmt} != '' then LIMIT_AMT,
							if #{limitDayAmt} != '' then LIMIT_DAY_AMT,
							if #{limitDayCnt} != '' then LIMIT_DAY_CNT,
							if #{supportForwardYn} != '' then SUPPORT_FORWARD_YN,
							if #{intro} != '' then INTRO,
							if #{supportAmt} != '' then SUPPORT_AMT,
							SUPPORT_YN,
							REG_DATE
							)
							VALUES(
								'#{grpId}'
								,'#{grpNm}'
								,'#{grpTypeCd}'
								,'1'
								,'#{limitYn}'
								,'#{limitUseTimeStart}'
								,'#{limitUseTimeEnd}'
								,'#{limitAmt}'
								,'#{limitDayAmt}'
								,'#{limitDayCnt}'
								,'#{supportForwardYn}'
								,'#{intro}'
								,'#{supportAmt}'
								,'#{supportYn}'
								,DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
							)
							`

var InsertBookUser string = `INSERT INTO priv_grp_user_info
							(
								GRP_ID,
								USER_ID,
								GRP_AUTH,
								AUTH_STAT,
								SUPPORT_BALANCE,
								JOIN_TY,
								AUTH_DATE,
								REG_DATE
							)
							VALUES(
								'#{grpId}'
								, '#{userId}'
								, '#{grpAuth}'
								, '#{authStat}'
								, '#{supportBalance}'
								, '#{joinTy}'
								, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								)
								`

var SelectBookInvteLink string = `SELECT 
									IFNULL(INVITE_LINK,'N') AS INVITE_LINK
									FROM priv_grp_info AS A
									WHERE
										A.GRP_ID='#{grpId}'
								`

var SelectBookMyAuth string = `SELECT 
									A.GRP_NM
									,B.GRP_AUTH
									FROM priv_grp_info AS A
									INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID
									WHERE
										A.GRP_ID='#{grpId}'
										AND B.USER_ID='#{userId}'
								`

var SelectLinkInfo string = `SELECT  REQ_STAT
									,IFNULL(PREPAID_AMT,0) AS PREPAID_AMT
									,PAY_TY
									,PAYMENT_USE_YN
									FROM org_agrm_info AS A
									Inner join priv_rest_info as b on a.rest_id = b.rest_id
									WHERE
										A.GRP_ID='#{grpId}'
										AND A.REST_ID='#{restId}'
								`

var SelectBookStoreList string = `SELECT A.REST_ID
								,B.REST_NM
								FROM org_agrm_info AS a
								INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
								WHERE 
								grp_id='#{grpId}'
								AND PAY_TY='#{payTy}'
								AND req_stat='1'
								`

var SelectChargeAmtList string = `SELECT AMT
											,ADD_AMT
										,IFNULL(AMT+ADD_AMT,0) AS SUM_AMT
									FROM dar_prepayment_charge_info
									WHERE 
									REST_ID='#{restId}'
									AND USE_YN='Y'
									ORDER BY AMT asc
								`

var SelectUserBookAuth string = `SELECT A.GRP_AUTH
										,A.AUTH_STAT
                                        ,(SELECT USER_ID FROM priv_grp_user_info AS AA WHERE AA.GRP_ID = A.GRP_ID AND GRP_AUTH='0') AS MANAGER_ID
										,FN_GET_GRPNAME(A.GRP_ID) AS GRP_NAME
										,B.USER_NM
								FROM PRIV_GRP_USER_INFO AS A
								INNER JOIN PRIV_USER_INFO AS B ON A.USER_ID = B.USER_ID
								WHERE 
								A.USER_ID = '#{userId}'
								AND  A.GRP_ID = '#{grpId}'`

var SelectCompanyBookInfo string = `SELECT
                                             b.company_id
											,b.company_nm
									FROM b_company_book AS A
									INNER JOIN b_company AS b ON a.company_id = b.company_id
									WHERE a.book_id='#{grpId}'
									`

var SelectCompanyUserChk string = `		SELECT COUNT(*) as cnt
										FROM b_company_user AS A
										WHERE 
										COMPANY_ID= '#{companyId}'
										AND TRIM(USER_NM)= '#{userNm}'
										AND TRIM(HP_NO)= '#{hpNo}'
										AND BOOK_ID = '#{grpId}'
									`

var UpdateCompanyUser string = ` UPDATE b_company_user SET
													USER_ID ='#{userId}'
										WHERE 
										COMPANY_ID= '#{companyId}'
										AND TRIM(USER_NM)= '#{userNm}'
										AND TRIM(HP_NO)= '#{hpNo}'
										AND BOOK_ID = '#{grpId}'
									`

var SelectBookUserView string = `SELECT
										A.AUTH_STAT
								FROM PRIV_GRP_USER_INFO A
								WHERE 
										A.GRP_ID= '#{grpId}'
										AND A.USER_ID= '#{userId}'
									`
var SelectBookBaseStoreList string = `SELECT A.REST_ID
											,A.REST_NM
											,CASE B.REQ_STAT WHEN 1 THEN 'Y' ELSE 'N' END AS USE_YN
                                            ,IFNULL(CONCAT(C.FILE_PATH,'/',C.SYS_FILE_NM)
												,CONCAT('/public/img/',CASE A.BUETY WHEN '00' THEN '한식' 
														  WHEN '01' THEN '중식' 
														  WHEN '02' THEN '일식' 
														  WHEN '03' THEN '양식' 
														  WHEN '04' THEN '카페' 
														  WHEN '05' THEN '분식' 
														  WHEN '06' THEN '부페' 
														  WHEN 'CA' THEN '부페' 
														  WHEN '07' THEN '기타' 
														  WHEN '08' THEN '유통' 
														  WHEN '09' THEN '뷰티' 
														  WHEN NULL THEN '기타' 
														  WHEN '' THEN '기타' 
														  ELSE cc.CODE_NM
												 END,'.png')) AS restImg
									FROM priv_rest_info AS A
									LEFT OUTER JOIN (SELECT A.REST_ID,B.REQ_STAT
															FROM priv_rest_info AS A
															INNER JOIN org_agrm_info AS B ON A.REST_ID = B.REST_ID
															INNER JOIN PRIV_GRP_INFO AS C ON B.GRP_ID = C.GRP_ID 
															AND C.GRP_ID='#{grpId}'
															WHERE A.rest_type='G'
															AND A.USE_YN='Y' ) AS B ON A.REST_ID = B.REST_ID
									LEFT outer join priv_rest_file AS C ON a.REST_ID = C.rest_id AND C.FILE_TY='1'
									LEFT OUTER JOIN b_code AS cc ON A.BUETY = cc.CODE_ID  AND A.USE_YN='Y'
									WHERE A.rest_type='G'
									AND A.USE_YN='Y' 
									`

var SelectGrpLinkList string = `SELECT *
										FROM(
												SELECT B.REST_NM
														 ,B.REST_ID
														 ,A.PAY_TY
														 ,CASE A.PAY_TY WHEN '1'  THEN IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
																							FROM dar_order_info AS AA 
																							INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
																							WHERE A.REST_ID = AA.REST_ID AND AA.ORDER_STAT='20' AND AA.PAID_YN='N' AND ORDER_TY !='4' AND PAY_TY='1' 
																							AND AA.GRP_ID = A.GRP_ID),0)
																					ELSE A.PREPAID_AMT  END  AS TOTAL_AMT
												FROM org_agrm_info AS A
												INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID AND B.USE_YN='Y'
												WHERE 
													A.GRP_ID='#{grpId}'
												AND REQ_STAT='1' ) AS Z
										`
