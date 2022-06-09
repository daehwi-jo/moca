package rests



var SelectSearchRestCnt string = `
					SELECT count(*) as totalCount
					FROM (
							SELECT A.REST_ID AS restId
									,A.REST_NM as restNm
									,IFNULL(A.INTRO,'') AS intro
									,IFNULL(ROUND((6371*acos(cos(radians('#{lat}'))*cos(radians(LAT))
														*cos(radians(LNG)-radians('#{lng}'))
														+sin(radians('#{lat}'))*sin(radians(LAT)))), 3),'9999') AS distance
								   ,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM),'') AS restImg
								   ,IFNULL((SELECT group_concat(SERVICE_NM) FROM priv_rest_service AS AA WHERE A.REST_ID = AA.REST_ID AND AA.USE_YN=1),'') AS serviceNm
								   ,IFNULL(C.GRP_ID,'N') AS linkGrpId
								   ,CASE WHEN D.USER_ID IS NULL THEN 'N' ELSE 'Y' END AS freqYn
							FROM priv_rest_info AS A
							LEFT OUTER JOIN priv_rest_file AS B ON A.REST_ID = B.REST_ID AND B.FILE_TY='1'
							LEFT OUTER JOIN org_agrm_info AS C ON A.REST_ID = C.REST_ID  AND C.REQ_STAT='1'
							AND C.GRP_ID IN('#{myGrp}') 
							LEFT OUTER JOIN dar_freq_rest_info AS D ON A.REST_ID = D.REST_ID  
							AND D.USER_ID='#{userId}'
							WHERE 
							A.USE_YN='Y'
							AND A.TEST_YN='N'
							AND A.REST_TYPE='N'
							AND (A.REST_NM LIKE '%#{searchKey}%' 
							or A.ADDR LIKE '%#{searchKey}%')
							
					) Z
					
					`

var SelectSearchRest string = `
					SELECT *
					FROM (
							SELECT A.REST_ID AS restId
									,A.REST_NM as restNm
									,IFNULL(A.INTRO,'') AS intro
									,IFNULL(ROUND((6371*acos(cos(radians('#{lat}'))*cos(radians(LAT))
														*cos(radians(LNG)-radians('#{lng}'))
														+sin(radians('#{lat}'))*sin(radians(LAT)))), 3),'9999') AS distance
								   ,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM)
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
								   ,IFNULL((SELECT group_concat(SERVICE_NM) FROM priv_rest_service AS AA WHERE A.REST_ID = AA.REST_ID AND AA.USE_YN=1),'') AS serviceNm
								   ,IFNULL(C.GRP_ID,'N') AS linkGrpId
								   ,CASE WHEN D.USER_ID IS NULL THEN 'N' ELSE 'Y' END AS freqYn
								   ,IFNULL(C.PAY_TY,'') AS payTy
							FROM priv_rest_info AS A
							LEFT OUTER JOIN priv_rest_file AS B ON A.REST_ID = B.REST_ID AND B.FILE_TY='1'
							LEFT OUTER JOIN org_agrm_info AS C ON A.REST_ID = C.REST_ID  AND C.REQ_STAT='1'
							AND C.GRP_ID IN('#{myGrp}') 
							LEFT OUTER JOIN dar_freq_rest_info AS D ON A.REST_ID = D.REST_ID  
							AND D.USER_ID='#{userId}'
							LEFT OUTER JOIN b_category AS bb ON a.category = bb.CATEGORY_ID AND A.USE_YN='Y'
							LEFT OUTER JOIN b_code AS cc ON A.BUETY = cc.CODE_ID  AND A.USE_YN='Y'
							WHERE 
							A.USE_YN='Y'
							AND A.TEST_YN='N'
							AND A.REST_TYPE='N'
							AND (A.REST_NM LIKE '%#{searchKey}%' 
							or A.ADDR LIKE '%#{searchKey}%')
							
					) Z
					`




var SelectSearchGrpInfo =`SELECT grpNm
								,grpPayTy
								,grpId
								,grpAmt
								,CASE WHEN  OPENING_HOUR='Y' AND LEFT(CURTIME(),5) >=LEFT(OPEN_TIME,5)  AND LEFT(CURTIME(),5) <=RIGHT(OPEN_TIME,5) THEN 'Y'
														WHEN 	OPENING_HOUR='N' THEN 'Y'	
												      ELSE 'N' END AS OPEN_YN
								,OPEN_MSG
						  FROM (
						    SELECT A.GRP_NM as grpNm
								,A.GRP_PAY_TY as grpPayTy
								,A.GRP_ID as grpId
								,CASE  WHEN A.GRP_PAY_TY='0'  THEN IFNULL(B.PREPAID_AMT,0)
										 WHEN A.GRP_PAY_TY='1'  THEN IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
																									FROM dar_order_info AS AA 
																									INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
																									WHERE AA.REST_ID = B.REST_ID AND A.GRP_ID = AA.GRP_ID 
																									AND AA.ORDER_STAT='20' AND AA.PAID_YN='N'),0)																										
											ELSE IFNULL(B.PREPAID_AMT,0)	END AS grpAmt
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
								FROM org_agrm_info AS B
								INNER JOIN priv_grp_info AS A ON A.GRP_ID = B.GRP_ID
								LEFT OUTER JOIN priv_rest_hour AS E ON B.REST_ID = E.REST_ID
								WHERE 
										A.GRP_ID ='#{grpId}'
										AND B.REST_ID='#{restId}'
										AND  B.REQ_STAT='1'
								) AS ZZ
						`



var SelectSearchRandStore =`
						SELECT A.REST_ID AS restId
									,A.REST_NM as restNm
									,IFNULL(A.INTRO,'') AS intro
									,IFNULL(ROUND((6371*acos(cos(radians('#{lat}'))*cos(radians(LAT))
														*cos(radians(LNG)-radians('#{lng}'))
														+sin(radians('#{lat}'))*sin(radians(LAT)))), 3),'9999') AS distance
								   ,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM),'') AS restImg
								   ,IFNULL((SELECT group_concat(SERVICE_NM) FROM priv_rest_service AS AA WHERE A.REST_ID = AA.REST_ID AND AA.USE_YN=1),'') AS serviceNm
								   ,IFNULL(C.GRP_ID,'N') AS linkGrpId
                                   ,IFNULL(C.PAY_TY,'') AS payTy
							FROM priv_rest_info AS A
							INNER JOIN priv_rest_file AS B ON A.REST_ID = B.REST_ID AND B.FILE_TY='1'
							INNER JOIN org_agrm_info AS C ON A.REST_ID = C.REST_ID  AND C.REQ_STAT='1'
							AND C.GRP_ID IN('#{myGrp}') 
							WHERE A.USE_YN='Y'
							ORDER BY RAND() LIMIT 2
						`


var SelectAuthMyGrpList string=`
								SELECT A.GRP_ID
								,A.GRP_NM
								,A.GRP_PAY_TY
								FROM  priv_grp_info AS A
								INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID  AND B.AUTH_STAT='1'
								WHERE 
								B.USER_ID='#{userId}'
								AND A.AUTH_STAT='1'
								AND B.GRP_AUTH='0'
								AND A.GRP_TYPE_CD ='#{grpTypeCd}'
								
								`

var SelectUnlinkMyGrpList string=`
							

								SELECT A.GRP_ID
								,A.GRP_NM
								,A.GRP_PAY_TY
								FROM priv_grp_info AS A
								INNER JOIN priv_grp_user_info AS B ON A.GRP_ID = B.GRP_ID AND B.GRP_AUTH='0'
								LEFT OUTER JOIN org_agrm_info AS C ON B.GRP_ID = C.GRP_ID 
								AND  C.REST_ID ='#{restId}'
								WHERE 
								B.USER_ID='#{userId}'
								AND C.AGRM_ID IS NULL
								AND A.GRP_TYPE_CD ='#{grpTypeCd}'

								
								`


var SelectLinkCancelCheckInfo string = `SELECT PREPAID_AMT
											,REQ_STAT
											,PAY_TY
											,IFNULL((SELECT SUM(TOTAL_AMT) FROM DAR_ORDER_INFO AS AA 
													WHERE 
														AA.REST_ID = A.REST_ID AND AA.GRP_ID = A.GRP_ID 
														AND  ORDER_STAT='20' AND PAID_YN='N' AND PAY_TY='1' ),0) AS UNPAID_AMT
									FROM org_agrm_info AS A
									WHERE 
									REST_ID ='#{restId}'
									AND GRP_ID='#{grpId}'
										`


var UpdateLinkCancel string = `	UPDATE org_agrm_info SET REQ_STAT = '4'		
														,MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
									WHERE 
									REST_ID ='#{restId}'
									AND GRP_ID ='#{grpId}'
									`



var SelectLinkStoreCheck string=`
								SELECT REST_TYPE
								FROM priv_rest_info as a
								WHERE 
								A.rest_id='#{restId}'
								`

var SelectLinkGrpCheck string=`
								SELECT GRP_TYPE_CD
								FROM priv_grp_info as a
								WHERE 
								A.grp_id='#{grpId}'
								`

var SelectLinkCheck string=`
								SELECT REQ_STAT
								,IFNULL(A.PREPAID_AMT, 0) AS PREPAID_AMT
								,AGRM_ID
								FROM org_agrm_info as a
								WHERE 
								A.rest_id='#{restId}' 
								AND A.grp_id='#{grpId}'
								`

var InsertLink =`
					INSERT INTO ORG_AGRM_INFO
					(
						AGRM_ID
						, GRP_ID
						, REST_ID
						, REQ_STAT
						, REQ_TY
						, REQ_DATE
						, REQ_COMMENT
						, REJ_COMMENT
						, AUTH_DATE
						, PAY_TY
						, PREPAID_AMT
					)
					VALUES
					(
						  '#{agrmId}'
						, '#{grpId}'
						, '#{restId}'
						, '#{reqStat}'
						, '#{reqTy}'
						, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
						, '#{reqComment}'
						, NULL
						, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
						, '#{payTy}'
						, '#{prepaidAmt}'
					)
				`

var UpdateLink =`UPDATE ORG_AGRM_INFO SET
								 REQ_STAT = '#{reqStat}'
								,AUTH_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								,PAY_TY = '#{payTy}'
								,MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
					WHERE 
						rest_id='#{restId}' 
						AND grp_id='#{grpId}'
				`

var InsertStoreReq =`
					 INSERT INTO b_store_req
						(
						STORE_NM
						,ADDR
						,USER_ID
						)
					VALUES (  
						'#{storeNm}'
						,'#{addr}' 
						,'#{userId}' 
						)
					`

var SelectStoreInfo string = `SELECT A.REST_ID AS restId
										,A.REST_NM as restNm
										,IFNULL(A.OPEN_WEEK,'') AS openWeek
										,IFNULL(A.OPEN_WEEKEND,'') AS openWeekend
										,IFNULL(A.INTRO,'') AS intro
										,IFNULL(A.ADDR,'') as addr
										,IFNULL(A.ADDR2,'') AS addr2
										,IFNULL(A.TEL,'') AS tel
										,IFNULL(A.LAT,'') AS lat
										,IFNULL(A.LNG,'') AS lng
										,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM)
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
										,CASE WHEN C.REST_ID IS NULL THEN 'N' ELSE 'Y'  END AS restFavoritesYn
										,'' as unlinkMyGrpCnt
										,A.REST_TYPE AS restType
										,IFNULL(D.NOTICE,'') AS restNotice
								FROM priv_rest_info AS A
								LEFT OUTER JOIN priv_rest_file AS B ON A.REST_ID = B.REST_ID AND B.FILE_TY='1'
								LEFT OUTER JOIN dar_freq_rest_info AS C ON A.REST_ID = C.REST_ID AND C.USER_ID='#{userId}'
								LEFT OUTER JOIN b_category AS bb ON a.category = bb.CATEGORY_ID AND A.USE_YN='Y'
							    LEFT OUTER JOIN b_code AS cc ON A.BUETY = cc.CODE_ID  AND A.USE_YN='Y'
								LEFT OUTER JOIN priv_rest_etc AS D ON A.REST_ID = D.REST_ID
								WHERE 
									A.REST_ID = '#{restId}'
									`




var SelectStoreUnLinkBookCnt string = ` SELECT  COUNT(*) AS unlinkGrpCnt
									FROM priv_grp_user_info AS A
									LEFT OUTER JOIN org_agrm_info AS B ON A.GRP_ID = B.GRP_ID 
									AND B.REST_ID ='#{restId}'
									WHERE 
									A.USER_ID='#{userId}'
									AND A.GRP_AUTH='0'
									AND A.AUTH_STAT='1'
									AND B.AGRM_ID IS NULL
										`


var SelectStoreServiceList string = `SELECT SERVICE_ID
										, SERVICE_NM
										, SERVICE_INFO
										, USE_YN
										, NOTICE_YN
										FROM priv_rest_service
										WHERE 
										REST_ID = '#{restId}'
										`


var SelectStoreMenu string = `SELECT ITEM_NO
												,ITEM_NM
												,ITEM_PRICE
												,BEST_YN
										FROM dar_sale_item_info
										WHERE 
										REST_ID='#{restId}'
										AND USE_YN='Y'
										AND ITEM_MENU !='99999'
										AND sales_end_Date >  DATE_FORMAT(NOW(),'%Y%m%d')
										ORDER BY BEST_YN DESC
										LIMIT 4
										`
var SelectStoreLinkInfo string = `SELECT GRP_NM
									,GRP_MANAGER
									,GRP_TYPE_CD
									,mocaOrderType
									,GRP_PAY_TY
									,THIS_MONTH_ORDER
									,PREPAID_AMT
									,SUPPORT_BALANCE
									,POST_PAID_AMT
								    ,CASE 
											WHEN (GRP_PAY_TY='0')  THEN 'S3' -- 충전 잔액
											WHEN (GRP_PAY_TY='1' AND GRP_AUTH='0')  THEN 'S6' -- 미정산 잔액 
											WHEN (GRP_PAY_TY='1' AND GRP_AUTH='1')  THEN 'S1' -- 이달 사용액(지원금 기준)
											ELSE 'S3' 
									  END AS S_TYPE
									,SUB_REST_ID
									,SUPPORT_YN
									FROM(
									SELECT A.GRP_NM
									,CASE WHEN  C.GRP_AUTH='0' THEN 'Y' ELSE 'N'  END AS GRP_MANAGER
									,A.SUPPORT_YN
									,A.GRP_TYPE_CD
									,0 AS mocaOrderType 
									,CASE WHEN a.GRP_PAY_TY <> B.PAY_TY THEN B.PAY_TY ELSE a.GRP_PAY_TY END AS GRP_PAY_TY
									,IFNULL((SELECT IFNULL(SUM(TOTAL_AMT),0) 
											FROM dar_order_info  AS AA
											INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
											WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
											AND D.REST_ID = AA.REST_ID 
											AND AA.ORDER_TY !='4'
											AND BB.USER_ID ='#{userId}' ),0) AS THIS_MONTH_ORDER
									,IFNULL(PREPAID_AMT,0) AS PREPAID_AMT
									,IFNULL(c.SUPPORT_BALANCE,0) AS SUPPORT_BALANCE
									,IFNULL((SELECT SUM(BB.ORDER_QTY*BB.ORDER_AMT)  
										FROM dar_order_info AS AA 
										INNER JOIN dar_order_detail AS BB ON AA.ORDER_NO = BB.ORDER_NO
										WHERE D.REST_ID = AA.REST_ID AND AA.ORDER_STAT='20' AND AA.PAID_YN='N' AND ORDER_TY <>'4' AND PAY_TY='1' 
										AND BB.USER_ID ='#{userId}'),0) AS POST_PAID_AMT
									,B.REST_ID AS SUB_REST_ID
									,IFNULL(C.GRP_AUTH,'1') AS GRP_AUTH
									FROM priv_grp_info AS A 
									INNER JOIN org_agrm_info AS B ON A.GRP_ID = B.GRP_ID
									INNER join priv_rest_info AS D on B.REST_ID = D.rest_id or (D.FRAN_YN = 'Y' and D.FRAN_ID = B.REST_ID)
									LEFT OUTER JOIN priv_grp_user_info AS C ON A.GRP_ID = C.GRP_ID 
									AND C.USER_ID='#{userId}'
									WHERE 
									A.GRP_ID= '#{grpId}'
									AND D.REST_ID ='#{restId}'
									) AS ZZ
                                   `

var UpdateLinkInfo string = `UPDATE org_agrm_info SET  
											PREPAID_AMT = #{prepaidAmt},
											MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
									WHERE
									 AGRM_ID ='#{agrmId}'
									`



var SelectFreqMenuChk string =`SELECT count(*) as itemCnt
								FROM dar_freq_menu_info
								WHERE 
								USER_ID ='#{userId}' 
								AND REST_ID='#{restId}'
								AND ITEM_NO='#{itemNo}'
								`

var InsertFreqStoreMenu string = `INSERT INTO dar_freq_menu_info (
									ITEM_NO
									, USER_ID
									, REST_ID
									, REG_DATE
									)
								VALUES
								(
									'#{itemNo}'
									,'#{userId}'
									,'#{restId}'
									, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								)`

var DeleteFreqStoreMenu string = `DELETE FROM dar_freq_menu_info
								WHERE 
								USER_ID ='#{userId}' 
								AND REST_ID='#{restId}'
								AND ITEM_NO='#{itemNo}'`


var SelectStoreLinkCheck string = `SELECT 
								  A.AGRM_ID
								, A.REQ_TY
								, DATE_FORMAT(A.REQ_DATE, '%Y-%m-%d') AS REQ_DATE	
								, DATE_FORMAT(A.AUTH_DATE, '%Y-%m-%d') AS AUTH_DATE
								, C.GRP_PAY_TY AS PAY_TY
								, IFNULL(A.PREPAID_AMT, 0) AS PREPAID_AMT
								, IFNULL(A.PREPAID_POINT, 0) AS PREPAID_POINT
								, IFNULL((SELECT POINT_RATE FROM priv_rest_etc AS AA WHERE A.REST_ID = AA.REST_ID),0) AS POINT_RATE
								, B.REST_NM
								FROM org_agrm_info AS A
								INNER join PRIV_REST_INFO AS B on A.REST_ID = B.REST_ID or (B.FRAN_YN = 'Y' and B.FRAN_ID = A.REST_ID)
								INNER JOIN PRIV_GRP_INFO AS C ON A.GRP_ID = C.GRP_ID
								WHERE 
								A.REQ_STAT='1'
								AND A.GRP_ID = '#{grpId}' 
								AND B.REST_ID = '#{restId}' 
								`


var SelectStoreItemList string = `SELECT 
									  ITEM_NO
									, ITEM_NM
									, ITEM_PRICE
									, ITEM_MENU
									FROM dar_sale_item_info 
									WHERE 
									USE_YN = 'Y'
									AND REST_ID = '#{restId}'`


var SelectStoreItem string = `SELECT 
									  ITEM_NO
									, ITEM_NM
									, ITEM_PRICE
									, ITEM_MENU
									, PROD_ID
									FROM dar_sale_item_info 
									WHERE 
									USE_YN = 'Y'
									AND REST_ID = '#{restId}'
									AND ITEM_NO = '#{itemNo}'

`

var UpdateLinkAmt string = `UPDATE org_agrm_info SET
										  PREPAID_AMT = #{prepaidAmt},
										  PREPAID_POINT = #{prepaidPoint},
										  MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										WHERE
										AGRM_ID = '#{agrmId}'`


var InsertFreqStore string = `INSERT INTO dar_freq_rest_info (
														USER_ID
														, REST_ID
														, REG_DATE
														)
													VALUES
													(
														'#{userId}'
														,'#{restId}'
														, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
													)`

var DeleteFreqStore string = `DELETE FROM dar_freq_rest_info
								WHERE 
								USER_ID ='#{userId}' 
								AND REST_ID='#{restId}'`


var SelectFreqStoreChk string =`SELECT count(*) as storeCnt
								FROM dar_freq_rest_info
								WHERE 
								USER_ID ='#{userId}' 
								AND REST_ID='#{restId}'
								`




var SelectBaseStore string = ` 	SELECT A.REST_ID
										,REST_NM
										,'Y' AS USE_YN
										,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM)
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
										,IFNULL(ORDER_CNT,0) AS orderCnt
								FROM  priv_rest_info AS a
								LEFT outer join priv_rest_file AS b ON a.REST_ID = b.rest_id AND B.FILE_TY='1'
								LEFT OUTER JOIN b_code AS cc ON A.BUETY = cc.CODE_ID  AND A.USE_YN='Y'
								LEFT OUTER JOIN (SELECT AA.REST_ID,COUNT(*) AS ORDER_CNT
														FROM dar_order_info AS AA
														INNER JOIN priv_rest_info AS BB ON AA.REST_ID = BB.REST_ID AND BB.REST_TYPE='G'
														WHERE DATE_FORMAT(AA.ORDER_DATE, '%Y-%m-%d %H:%i:%s') >= DATE_ADD(NOW(), INTERVAL -30 DAY) 
														AND DATE_FORMAT(AA.ORDER_DATE, '%Y-%m-%d %H:%i:%s') <= NOW()
														AND AA.ORDER_STAT='20'
														GROUP BY AA.REST_ID ) AS D ON A.REST_ID = D.REST_ID
								WHERE rest_type='G'
								AND A.USE_YN='Y'
								AND (A.REST_NM LIKE '%#{searchKey}%' 
								or A.ADDR LIKE '%#{searchKey}%')
											`

var SelectStoreEtc string =`	SELECT IFNULL(MEMO,'') AS memo
										,DATE_FORMAT(NOW(), '~%Y년 %m월 %d일') AS expireDate
										,(SELECT IFNULL(AA.ITEM_IMG,'') FROM  dar_sale_item_info AS AA WHERE A.REST_ID = AA.REST_ID 
											AND  ITEM_NO='#{itemNo}' ) 	AS itemImg
								FROM priv_rest_etc AS a
								WHERE 
								REST_ID ='#{restId}'

`

var SelectStoreEtcSeq string = `SELECT CONCAT('C',IFNULL(LPAD(MAX(SUBSTRING(A.REST_ID, -10)) + 1, 10, 0), '0000000001')) as storeSeq
								FROM priv_rest_info AS A
								INNER JOIN priv_rest_etc AS B ON A.REST_ID = B.REST_ID
								WHERE 
								A.REST_TYPE='G'
							`


var SelectStoreEtcCheck string = `SELECT B.REST_ID
								  FROM priv_rest_info AS A
								  INNER JOIN priv_rest_etc AS B ON A.REST_ID = B.REST_ID
								  WHERE 
								 	B.MEDIA_CODE='#{swapCd}'
							`
var SelectStoreBiznum string = `SELECT A.REST_ID AS MAIN_REST_ID
								 	FROM b_rest_combine AS a
								 	INNER JOIN priv_rest_info AS b ON a.rest_id = b.rest_id
								 	WHERE 
									  b.busid='#{busId}'
							`

var InsertWincubeStore string = `INSERT INTO priv_rest_info
								(
								 REST_ID
								,REST_NM
								,CEO_NM
								,BUSID
								,BUETY
								,AUTH_STAT
								,USE_YN
								,REG_DATE
								,REST_TYPE
							)
							VALUES (
							'#{restId}'
							,'#{restNm}'
							,'#{ceoNm}'
							,'#{busId}'
							,'04'
							,'1'
							,'Y'
							,DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')		
							,'G'				
							)
							`


var InsertWincubeStoreEtc string = `INSERT INTO priv_rest_etc
								(
								 REST_ID
								,MEMO
								,MEDIA_CODE
								,MOD_DATE
								)
								VALUES (
								'#{restId}'
								,'#{memo}'
								,'#{swapCd}'
								,DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
							)
							`

var UpdateWincubeStoreEtc string = `UPDATE priv_rest_etc SET
										  MEMO = '#{memo}',
										  MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										WHERE
										rest_id = '#{restId}'
`
var InsertWincubeStoreCombine string = `INSERT INTO b_rest_combine_sub
										(
										 REST_ID
										,SUB_REST_ID
										,USE_YN
										,REG_DATE
										)
										VALUES (
										'#{mainRestId}'
										,'#{restId}'
										,'Y'
										,DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
									)
							`

var SelectWincubeItemCheck string = `SELECT ITEM_NO
								 	FROM dar_sale_item_info
								 	WHERE 
								 		REST_ID='#{restId}'
								 		AND PROD_ID='#{goodsId}'
									`

var SelectStoreMenuSeq string = `SELECT
								IFNULL(LPAD(MAX(ITEM_NO) + 1, 10, 0), '0000000001') as itemNo
								FROM dar_sale_item_info	
								`


var InsertStroeMenu string = ` INSERT INTO dar_sale_item_info(
												ITEM_NO
												, ITEM_NM
												, REST_ID
												, ITEM_STAT
												, ITEM_PRICE
												, REG_DATE
												, ITEM_MENU
												, ITEM_IMG
												, PROD_ID
												, SALES_END_DATE
												,NOMAL_SALE_PRICE
												,NOMAL_SALE_VAT
												,SALE_PRICE
												,SALE_VAT
												,TOTAL_PRICE
											)
											VALUES (
												'#{itemNo}'
												,'#{itemNm}'
												,'#{restId}'
												,'1'
												,#{itemPrice}
												,DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
												,'#{codeId}'
												,'#{itemImg}'
												,'#{goodsId}'
												,'#{periodEnd}'
												,'#{normalSalePrice}'
											 	,'#{normalSaleVat}'
											 	,'#{salePrice}'
											 	,'#{saleVat}'
											 	,'#{totalPrice}'
											)`

var UpdateStroeMenu string = `UPDATE  dar_sale_item_info SET
											SALES_END_DATE = '#{periodEnd}',
											MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										WHERE 
										ITEM_NO='#{itemNo}'
										`




var SelectStoreItemCodeSeq string = `SELECT 
										IFNULL(CATEGORY_ID,FN_GET_SEQUENCE('CATEGORY_ID')) AS CATEGORY_ID
										,CONCAT(IFNULL(LPAD(MAX(SUBSTRING(CODE_ID, -5)) + 1, 5, 0), '00001')) as cateSeq
										FROM DAR_CATEGORY_INFO
										WHERE 
										REST_ID='#{restId}'
										AND CODE_ID <> '99999'
										`

var InsertStoreCategories string = `INSERT INTO DAR_CATEGORY_INFO
										(
											CATEGORY_ID
											, REST_ID
											, CODE_ID
											, CODE_NM
											, USE_YN
										)
										VALUES
										(
											 '#{categoryId}'
											, '#{restId}'
											, '#{cateSeq}'
											, '#{codeNm}'
											, 'Y'
										)`




var SelectSearchRestCnt_v2 string = `
					SELECT count(*) as totalCount
					FROM (
							SELECT A.REST_ID AS restId
									,A.REST_NM as restNm
									,IFNULL(A.INTRO,'') AS intro
									,IFNULL(ROUND((6371*acos(cos(radians('#{lat}'))*cos(radians(LAT))
														*cos(radians(LNG)-radians('#{lng}'))
														+sin(radians('#{lat}'))*sin(radians(LAT)))), 3),'9999') AS distance
								   ,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM),'') AS restImg
								   ,IFNULL((SELECT group_concat(SERVICE_NM) FROM priv_rest_service AS AA WHERE A.REST_ID = AA.REST_ID AND AA.USE_YN=1),'') AS serviceNm
							FROM priv_rest_info AS A
							LEFT OUTER JOIN priv_rest_file AS B ON A.REST_ID = B.REST_ID AND B.FILE_TY='1'
							WHERE 
							A.USE_YN='Y'
							AND A.TEST_YN='N'
							AND A.REST_TYPE='N'
							AND (A.REST_NM LIKE '%#{searchKey}%' 
							or A.ADDR LIKE '%#{searchKey}%')
							
					) Z
					
					`

var SelectSearchRest_v2 string = `
					SELECT *
					FROM (
							SELECT A.REST_ID AS restId
									,A.REST_NM as restNm
									,IFNULL(A.INTRO,'') AS intro
									,IFNULL(ROUND((6371*acos(cos(radians('#{lat}'))*cos(radians(LAT))
														*cos(radians(LNG)-radians('#{lng}'))
														+sin(radians('#{lat}'))*sin(radians(LAT)))), 3),'9999') AS distance
								   ,IFNULL(CONCAT(B.FILE_PATH,'/',B.SYS_FILE_NM)
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
								   ,IFNULL((SELECT group_concat(SERVICE_NM) FROM priv_rest_service AS AA WHERE A.REST_ID = AA.REST_ID AND AA.USE_YN=1),'') AS serviceNm
							FROM priv_rest_info AS A
							LEFT OUTER JOIN priv_rest_file AS B ON A.REST_ID = B.REST_ID AND B.FILE_TY='1'
							LEFT OUTER JOIN b_category AS bb ON a.category = bb.CATEGORY_ID AND A.USE_YN='Y'
							LEFT OUTER JOIN b_code AS cc ON A.BUETY = cc.CODE_ID  AND A.USE_YN='Y'
							WHERE 
							A.USE_YN='Y'
							AND A.TEST_YN='N'
							AND A.REST_TYPE='N'
							AND (A.REST_NM LIKE '%#{searchKey}%' 
							or A.ADDR LIKE '%#{searchKey}%')
							
					) Z
					ORDER BY DISTANCE ASC
					`
