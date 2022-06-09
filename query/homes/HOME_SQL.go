package homes




var SelectUserAmountInfo string = `SELECT SUM(IFNULL((SELECT SUM(PREPAID_AMT) FROM org_agrm_info AS AA WHERE B.GRP_ID = AA.GRP_ID AND AA.REQ_STAT='1'),0)) AS PRE_PAID
										,SUM(IFNULL((SELECT SUM(BB.TOTAL_AMT)  FROM dar_order_info AS BB WHERE B.GRP_ID = BB.GRP_ID  AND ORDER_TY <>'4' AND BB.PAY_TY='1' AND BB.ORDER_STAT='20' AND BB.PAID_YN='N'),0)) AS POST_PAID
									FROM priv_user_info AS A
									INNER JOIN priv_grp_user_info AS B ON A.USER_ID = B.USER_ID AND B.GRP_AUTH='0'
									WHERE 
									a.user_id='#{userId}'
									`

var SelectMyGrpList string = ` SELECT *
							FROM (
								SELECT A.GRP_ID
									,A.GRP_NM
									,A.GRP_TYPE_CD
									,A.GRP_PAY_TY
									,A.SUPPORT_YN
									,IFNULL(B.SUPPORT_BALANCE,0)  AS SUPPORT_BALANCE  
									,IFNULL((SELECT IFNULL(SUM(ORDER_QTY*ORDER_AMT),0) 
													FROM dar_order_info  AS AA
											      INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
													WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
													AND B.GRP_ID = AA.GRP_ID AND BB.USER_ID = B.USER_ID),0) AS thisMonthOrder
								   ,IFNULL((SELECT SUM(PREPAID_AMT) FROM org_agrm_info AS AA WHERE A.GRP_ID = AA.GRP_ID AND  AA.REQ_STAT='1'),0) AS prepaidAmt 
								   ,IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
													FROM dar_order_info AS AA 
													INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
													WHERE B.GRP_ID = AA.GRP_ID AND AA.ORDER_STAT='20' AND ORDER_TY <>'4' AND PAY_TY='1' AND AA.PAID_YN='N' AND BB.USER_ID = B.USER_ID),0) AS postPaidAmt 	 
									,CASE 
										    WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='Y')  THEN 'D1'
										 	 WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='20' AND A.SUPPORT_YN='Y') THEN 'D1'
										    WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='N')  THEN 'D2'
										    WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='20' AND A.SUPPORT_YN='N')  THEN 'D2'
										 	 WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='Y')  THEN 'D1'
										 	 WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='20' AND A.SUPPORT_YN='Y')  THEN 'D1'
										 	 WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='1' AND A.SUPPORT_YN='N')  THEN 'D2'
										 	 WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='20' AND A.SUPPORT_YN='N')  THEN 'D2'
										 	 WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='4')  THEN 'D3'
										 	 WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='4')  THEN 'D4'
											 WHEN (A.GRP_PAY_TY='0' AND A.GRP_TYPE_CD='10')  THEN 'D3'
										 	 WHEN (A.GRP_PAY_TY='1' AND A.GRP_TYPE_CD='10')  THEN 'D4'																											
										  	 ELSE 'D3' END AS D_TYPE
										 ,(SELECT MAX(BB.ORDER_DATE) 
										 	FROM dar_order_info AS AA 
										 	INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO			
										 	WHERE A.GRP_ID = AA.GRP_ID AND BB.USER_ID = B.USER_ID) AS ORDER_DATE
							FROM  priv_grp_info AS A
							INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID  AND B.AUTH_STAT='1'
							WHERE 
							B.USER_ID='#{userId}'
							AND A.USE_YN='Y'
							AND A.AUTH_STAT='1') AS AA
							ORDER BY ORDER_DATE DESC ,SUPPORT_BALANCE DESC
					`

var SelectMyTickeGrpList string = `SELECT  B.GRP_NM
											, C.GRP_ID
									FROM org_agrm_info AS A
									INNER JOIN priv_grp_info AS B ON A.GRP_ID = B.GRP_ID
									INNER JOIN priv_grp_user_info AS C ON B.GRP_ID = C.GRP_ID
									WHERE 
									C.USER_ID='#{userId}'
									AND A.LINK_TY='T' 
									AND A.REQ_STAT='1'
									AND C.AUTH_STAT='1'
									GROUP BY  B.GRP_NM , C.GRP_ID
							`



var SelectHomeStoreList string = `SELECT REST_ID
												,REST_NM
												,INTRO
												,REST_IMG
												,GRP_PAY_TY
												,THIS_MONTH_ORDER
												,PREPAID_AMT
												,POST_PAID_AMT
												,freqYn
												,CASE 
													WHEN (GRP_PAY_TY='0' AND GRP_TYPE_CD='1' AND SUPPORT_YN='Y')  THEN 'S1' -- 이달 사용액(지원금 기준)
													WHEN (GRP_PAY_TY='0' AND GRP_TYPE_CD='1' AND SUPPORT_YN='N')  THEN 'S2' -- 이달 사용액
													WHEN (GRP_PAY_TY='0' AND GRP_TYPE_CD='20' AND SUPPORT_YN='Y')  THEN 'S1' -- 이달 사용액(지원금 기준)
													WHEN (GRP_PAY_TY='0' AND GRP_TYPE_CD='20' AND SUPPORT_YN='N')  THEN 'S2' -- 이달 사용액
													WHEN (GRP_PAY_TY='0' AND GRP_TYPE_CD='4')  					  THEN 'S3' -- 충전잔액
													WHEN (GRP_PAY_TY='0' AND GRP_TYPE_CD='10')  				  THEN 'S3' -- 충전잔액
													WHEN (GRP_PAY_TY='1' AND GRP_TYPE_CD='1' AND SUPPORT_YN='Y')  THEN 'S4' -- 이달 사용액(지원금 기준)
													WHEN (GRP_PAY_TY='1' AND GRP_TYPE_CD='1' AND SUPPORT_YN='N')  THEN 'S5' -- 이달 사용액
													WHEN (GRP_PAY_TY='1' AND GRP_TYPE_CD='20' AND SUPPORT_YN='Y')  THEN 'S4' -- 이달 사용액(지원금 기준)
													WHEN (GRP_PAY_TY='1' AND GRP_TYPE_CD='20' AND SUPPORT_YN='N')  THEN 'S5' -- 이달 사용액
													WHEN (GRP_PAY_TY='1' AND GRP_TYPE_CD='4')                     THEN 'S6'	-- 미정산 사용금액		
													WHEN (GRP_PAY_TY='1' AND GRP_TYPE_CD='10')                     THEN 'S6'	-- 미정산 사용금액	
													ELSE 'S3' 
												END AS S_TYPE
												,OPEN_MSG
												,CASE WHEN  OPENING_HOUR='Y' AND LEFT(CURTIME(),5) >=LEFT(OPEN_TIME,5)  AND LEFT(CURTIME(),5) <=RIGHT(OPEN_TIME,5) THEN 'Y'
														WHEN 	OPENING_HOUR='N' THEN 'Y'	
												      ELSE 'N' END AS OPEN_YN
								  		FROM(  
										  SELECT C.REST_ID
											,C.REST_NM
											,IFNULL(C.INTRO,'') AS INTRO
									 		,IFNULL(CONCAT(D.FILE_PATH,'/',D.SYS_FILE_NM)
									 		,CONCAT('/public/img/',CASE c.BUETY WHEN '00' THEN '한식' 
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
												 END,'.png')) AS REST_IMG		
									   ,CASE WHEN a.GRP_PAY_TY <> B.PAY_TY THEN B.PAY_TY ELSE a.GRP_PAY_TY END AS GRP_PAY_TY
										,IFNULL((SELECT IFNULL(SUM(ORDER_AMT*ORDER_QTY),0) 
												FROM dar_order_info  AS AA
												INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
												WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
												AND C.REST_ID = AA.REST_ID 
												AND AA.ORDER_TY !='4'
												AND AA.GRP_ID = A.GRP_ID
												AND BB.USER_ID ='#{userId}' ),0) AS THIS_MONTH_ORDER
										,IFNULL(PREPAID_AMT,0) AS PREPAID_AMT
										,IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
											FROM dar_order_info AS AA 
											INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
											WHERE C.REST_ID = AA.REST_ID AND AA.ORDER_STAT='20' AND AA.PAID_YN='N' AND ORDER_TY !='4' AND PAY_TY='1' 
											AND AA.GRP_ID = A.GRP_ID
											AND BB.USER_ID ='#{userId}'),0) AS POST_PAID_AMT
									,(SELECT MAX(ORDER_DATE) FROM dar_order_info AS AA WHERE C.REST_ID = AA.REST_ID) AS ORDER_DATE
									,IFNULL((SELECT CASE WHEN USER_ID IS NULL THEN 'N' ELSE 'Y' END 
										FROM dar_freq_rest_info AS AA 
										WHERE 
										C.REST_ID = AA.REST_ID 
										AND USER_ID ='#{userId}'
	                         ),'N') AS freqYn
	                         ,A.SUPPORT_YN
	                         ,A.GRP_TYPE_CD
	                         ,IFNULL(ROUND((6371*acos(cos(radians('#{lat}'))*cos(radians(C.LAT))
														*cos(radians(C.LNG)-radians('#{lng}'))
														+sin(radians('#{lat}'))*sin(radians(C.LAT)))), 3),'9999') AS distance
                             ,IFNULL(E.USE_YN,'N') AS OPENING_HOUR
									,case  WEEKDAY(NOW())
										    when '0' then D0
										    when '1' then D1
										    when '2' then D2
										    when '3' then D3
										    when '4' then D4
										    when '5' then D5
										    when '6' then D6
									 end AS OPEN_TIME
							 ,IFNULL(OPEN_MSG,'') AS OPEN_MSG
						FROM priv_grp_info AS A
						INNER JOIN org_agrm_info AS B ON A.GRP_ID =B.GRP_ID AND B.REQ_STAT='1'
						INNER JOIN priv_rest_info AS C on B.REST_ID = C.rest_id or (C.FRAN_YN = 'Y' and C.FRAN_ID = B.REST_ID) 
						LEFT OUTER JOIN priv_rest_file AS D ON C.REST_ID = D.REST_ID AND D.FILE_TY='1'
						LEFT OUTER JOIN b_category AS bb ON c.category = bb.CATEGORY_ID AND A.USE_YN='Y'
						LEFT OUTER JOIN b_code AS cc ON c.BUETY = cc.CODE_ID  AND A.USE_YN='Y'
						LEFT OUTER JOIN priv_rest_hour AS E ON C.REST_ID = E.REST_ID
						WHERE 
							A.GRP_ID='#{grpId}'
							AND C.USE_YN='Y'
						) AS ZZ
						ORDER BY freqYn DESC, ORDER_DATE DESC, distance ASC
					    `

var SelectMyInfo string = `	SELECT login_id
										,USER_BIRTH
										,hp_no
										,USER_NM
										,RECOM_CODE
										,PUSH_YN
									FROM priv_user_info AS A
									WHERE 
									a.user_id='#{userId}'
									`



var UpdateMyInfo string = `	UPDATE  priv_user_info SET  
											MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
											,USER_BIRTH = '#{userBirth}'	
											,HP_NO = '#{hpNo}'
											,PUSH_YN = '#{pushYn}'
											,LOGIN_PW = '#{loginPw}'
							WHERE 
									user_id='#{userId}'
							`

