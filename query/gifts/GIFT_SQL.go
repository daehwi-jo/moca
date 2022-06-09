package gifts


var SelectGift string = `SELECT  CASE RESULT WHEN 'S0' THEN CONCAT(' "',REST_NM,'" 선물을 보내는중입니다.') 
															  WHEN 'S1' THEN CONCAT(USER_NM,'님께서 "',REST_NM,'" 선물을 수령하였습니다.') 
															  WHEN 'S2' THEN CONCAT(USER_NM,'님께서 "',REST_NM,'" 선물을 거절하였습니다.') 
															  WHEN 'R1' THEN CONCAT(USER_NM,'님께서 주신 "',REST_NM,'"  선물을 받았습니다.') 
															  WHEN 'R2' THEN CONCAT(USER_NM,'님께서 주신 "',REST_NM,'"  선물을 거절하였습니다.') 
															  WHEN 'SR' THEN CONCAT(' "',REST_NM,'"  선물이 상대방이 받지 않아 회수되었습니다.') 
												END AS giftMsg
												,DATE_FORMAT(GIFT_DATE, '%Y.%m.%d') AS GIFT_DATE
												,RESULT AS GIFT_STAT
												,GIFT_ID
												,'' as giftLink
									FROM (
											SELECT A.GIFT_ID 
													,B.REST_NM
													,C.USER_NM 
													,GIFT_SND_DTM AS GIFT_DATE
													,CASE WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='0') THEN 'S0'   
													      WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='1') THEN 'S1'   
													      WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='2') THEN 'S2'   
													      WHEN GIFT_STS_CD='2' THEN 'SC'
													      WHEN GIFT_STS_CD='3' THEN 'SR'
													      END  AS RESULT
											from dar_gift_info AS A
											INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
											LEFT OUTER JOIN priv_user_info AS C ON A.RCV_USER_ID = C.USER_ID
											WHERE 
											A.SND_USER_ID='#{userId}'
											AND A.GIFT_STS_CD IN('1','3')
											UNION ALL
											SELECT A.GIFT_ID 
													,B.REST_NM
													,C.USER_NM 
													,GIFT_RCV_DTM AS GIFT_DATE
													,CASE WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='1') THEN 'R1'  
													      WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='2') THEN 'R2'   
													      END AS RESULT
											from dar_gift_info AS A
											INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
											INNER JOIN priv_user_info AS C ON A.SND_USER_ID = C.USER_ID
											WHERE 
											A.RCV_USER_ID='#{userId}'
											AND A.GIFT_STS_CD='1'
									) AA
									WHERE LEFT(GIFT_DATE,8) >= DATE_FORMAT(DATE_ADD(NOW(), INTERVAL -1 MONTH), '%Y%m%d')
									ORDER BY GIFT_DATE DESC
									LIMIT 1
									`

var SelectGiftHist string = `SELECT  CASE RESULT WHEN 'S0' THEN CONCAT(' "',REST_NM,'" 선물을 보내는중입니다.') 
															  WHEN 'S1' THEN CONCAT(USER_NM,'님께서 "',REST_NM,'" 선물을 수령하였습니다.') 
															  WHEN 'S2' THEN CONCAT(USER_NM,'님께서 "',REST_NM,'" 선물을 거절하였습니다.') 
															  WHEN 'R1' THEN CONCAT(USER_NM,'님께서 주신 "',REST_NM,'"  선물을 받았습니다.') 
															  WHEN 'R2' THEN CONCAT(USER_NM,'님께서 주신 "',REST_NM,'"  선물을 거절하였습니다.') 
															  WHEN 'SR' THEN CONCAT(' "',REST_NM,'"  선물이 상대방이 받지 않아 회수되었습니다.') 
												END AS giftMsg
												,DATE_FORMAT(GIFT_DATE, '%Y.%m.%d') AS GIFT_DATE
												,RESULT AS GIFT_STAT
												,GIFT_ID
									FROM (
											SELECT A.GIFT_ID 
													,B.REST_NM
													,C.USER_NM 
													,GIFT_SND_DTM AS GIFT_DATE
													,GIFT_SND_DTM AS GIFT_DATE2
													,CASE WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='0') THEN 'S0'   
													      WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='1') THEN 'S1'   
													      WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='2') THEN 'S2'   
													      WHEN GIFT_STS_CD='2' THEN 'SC'
													      WHEN GIFT_STS_CD='3' THEN 'SR'
													      END  AS RESULT
											from dar_gift_info AS A
											INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
											LEFT OUTER JOIN priv_user_info AS C ON A.RCV_USER_ID = C.USER_ID
											WHERE 
											A.SND_USER_ID='#{userId}'
											AND A.GIFT_STS_CD IN('1','3')
											UNION ALL
											SELECT A.GIFT_ID 
													,B.REST_NM
													,C.USER_NM 
													,GIFT_RCV_DTM AS GIFT_DATE
													,GIFT_RCV_DTM AS GIFT_DATE2
													,CASE WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='1') THEN 'R1'  
													      WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='2') THEN 'R2'   
													      END AS RESULT
											from dar_gift_info AS A
											INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
											INNER JOIN priv_user_info AS C ON A.SND_USER_ID = C.USER_ID
											WHERE 
											A.RCV_USER_ID='#{userId}'
											AND A.GIFT_STS_CD='1'
									) AA
									ORDER BY GIFT_DATE2 DESC
									`


var SelectGetGiftSeq string = `SELECT RIGHT(FN_GET_SEQUENCE('GIFT_MOID'),4) as GIFT_SEQ `




var InsertGift string = `INSERT INTO dar_gift_info 
							(
								GIFT_ID  
								,MOID 
								,GIFT_OWNER_DIV 
								,GIFT_METHOD_DIV 
								,GIFT_STS_CD 
								,REST_ID 
								,SND_GRP_ID
								,SND_USER_ID 
								,GIFT_AMT 
								,GIFT_SND_DTM 
								,RCV_USER_ID 
								,RCV_STS_CD 
								,ORDER_NO
								if #{giftMsg} != '' then ,SND_MSG
							)
							VALUES( 
								(SELECT FN_GET_SEQUENCE('GIFT_ID')) 
								,IFNULL('#{moid}','')              
								,IFNULL('#{giftOwnerDiv}','')       
								,IFNULL('#{giftMethodDiv}','')   
								,'1'          
								,IFNULL('#{restId}','')           
								,IFNULL('#{grpId}','')           
								,IFNULL('#{sndUserId}','')        
								,IFNULL('#{giftAmt}',0)            
								,DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
								,''                   
								,IFNULL('#{rcvUserId}','')          
								,'0'           
								,IFNULL('#{orderNo}','') 
								,'#{giftMsg}'
								)`


var SelectGiftRecvInfo string =`SELECT GIFT_STS_CD
										,RCV_STS_CD
										,GIFT_AMT
										,REST_ID
										,MOID
										,GIFT_ID
										,SND_USER_ID
										,FN_GET_USERNAME('#{rcvUserId}') as rcvUserNm
								FROM dar_gift_info
								WHERE 
								MOID='#{giftNum}'
								`


var UpdateGiftInfo string = `UPDATE dar_gift_info SET  
							 GIFT_RCV_DTM = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
							 ,RCV_GRP_ID  = '#{grpId}'
							 ,RCV_USER_ID = '#{rcvUserId}'
							 ,RCV_STS_CD  = '#{rcvStsCd}'
							WHERE 
							GIFT_STS_CD = '1'  
							AND RCV_STS_CD IN ('0', '2') 
							AND GIFT_ID ='#{giftId}'
							`


var SelectGiftLinkCheck string=`
								SELECT REQ_STAT
								,IFNULL(A.PREPAID_AMT, 0) AS PREPAID_AMT
								,AGRM_ID
								,GRP_PAY_TY
								,(SELECT USER_ID FROM priv_grp_user_info AS AA WHERE A.GRP_ID = AA.GRP_ID AND AA.GRP_AUTH='0') AS GRP_USER_ID
								FROM org_agrm_info as a
								INNER JOIN priv_grp_info AS b ON a.grp_id= b.grp_id
								WHERE 
								a.rest_id='#{restId}' 
								AND a.grp_id='#{grpId}'
								`

var SelectGiftGrpCheck string=`SELECT A.GRP_PAY_TY
									,(SELECT USER_ID FROM priv_grp_user_info AS AA WHERE A.GRP_ID = AA.GRP_ID AND AA.GRP_AUTH='0') AS GRP_USER_ID
									FROM PRIV_GRP_INFO AS A
									WHERE 
									A.GRP_ID='#{grpId}'
								`


var SelectGiftInfo =`SELECT GIFT_STS_CD
							,RCV_STS_CD
							,GIFT_AMT
							,ORDER_NO
							,REST_ID
							,SND_GRP_ID
							,GIFT_ID
						FROM DAR_GIFT_INFO
						WHERE 
						MOID = '#{giftNum}'
					`

var UpdateGiftCancel string = `UPDATE dar_gift_info SET  
								 GIFT_CAN_DTM = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								 ,GIFT_STS_CD  = '#{giftStsCd}'
							WHERE
							GIFT_ID ='#{giftId}'
							AND GIFT_STS_CD = '1'  
							AND RCV_STS_CD IN ('0', '2')
							`


var SelectGiftDesc =`SELECT GIFT_AMT
								,IFNULL(SND_MSG,'') AS SND_MSG 
								,MOID
								,REST_NM
								,CASE WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='0') THEN 'S0'   
																				  WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='1') THEN 'S1'   
																				  WHEN (GIFT_STS_CD='1' AND RCV_STS_CD='2') THEN 'S2'   
																				  WHEN GIFT_STS_CD='2' THEN 'SC'
																				  WHEN GIFT_STS_CD='3' THEN 'SR'
																				  END  AS GIFT_STAT
						FROM dar_gift_info AS A
						INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
						WHERE 
						gift_id='#{giftId}'
					`


var SelectGiftReady =` SELECT B.GRP_NM
					  		,C.REST_NM
					  		,IFNULL(A.PREPAID_AMT, 0) AS PREPAID_AMT
					FROM org_agrm_info AS A
					INNER JOIN priv_grp_info AS b ON a.grp_id= b.grp_id
					INNER JOIN priv_rest_info AS C ON A.REST_ID = C.REST_ID
					WHERE 
						a.rest_id = '#{restId}'
					AND a.grp_id = '#{grpId}'
					`