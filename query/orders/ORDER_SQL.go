package orders

var SelectCreateOrderSeq string = `SELECT  FN_GET_SEQUENCE('ORDER_NO') AS orderSeq;`

var InsertOrder string = `INSERT INTO dar_order_info
								(
									ORDER_NO
									if #{moid} != '' then ,MOID
									, REST_ID
									, USER_ID
									, GRP_ID
									, ORDER_STAT
									, ORDER_DATE
									, TOTAL_AMT
									, CREDIT_AMT
									, POINT_USE
									, DISCOUNT
									if #{chargeAmt} != '' then ,INSTANT_AMT
									if #{instantMoid} != '' then ,INSTANT_MOID
									, PAID_YN
									, PAY_DATE
									, ORDER_TY
									, PAY_TY
									, QR_ORDER_TYPE
									
								)
								VALUES
								(
									'#{orderNo}'
									, '#{moid}'
									, '#{restId}'
									, '#{userId}'
									, '#{grpId}'
									, '20'
									, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
									, #{orderAmt}
									, #{creditAmt}
									, #{pointUse}
									, #{discount}
									, #{chargeAmt}
									, '#{instantMoid}'
									, 'N'
									, NULL
									, '#{orderTy}'
									, '#{payTy}'
									, '#{qrOrderTy}'
									
								)`

var InsertOrderMemo string = `INSERT INTO dar_order_memo
							(
								  ORDER_NO
								, USER_ID
								, MEMO
							)
							VALUES
							(
								 '#{orderNo}'
								,'#{memoUserId}'
								,'#{userMemo}'
							)`

var InsertOrderDetail string = `INSERT INTO dar_order_detail
							(
								  ORDER_NO
								, ORDER_SEQ
								, ITEM_NO
								, USER_ID
								, ORDER_DATE
								, ORDER_QTY
								, ORDER_AMT
							)
							VALUES
							(
								 '#{orderNo}'
								, #{orderSeq}
								, '#{itemNo}'
								, '#{itemUserId}'
								, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								, #{itemCount}
								, #{itemPrice}
							)`

var InsertOrderPickupData string = `INSERT INTO dar_order_pickup
							(
								  ORDER_NO
								, P_STATUS
								, WAITING_DATE
							)
							VALUES
							(
								 '#{orderNo}'
								, '30'
								, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
							)`

var InsertOrderSplitAmt string = `INSERT INTO dar_order_split
							(
								ORDER_NO
								, USER_ID
								, SPLIT_AMT
							)
								VALUES
							(
								 '#{orderNo}'
								,'#{splitUserId}'
								, #{splitAmt}
							)
							`

var UpdateOrder string = `UPDATE dar_order_info SET 
								ORDER_STAT = '#{orderStat}'	
								,ORDER_CANCEL_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
						 WHERE 
							ORDER_NO = '#{orderNo}'
							`

var SelectStoreCategory string = `SELECT CODE_ID
											,CODE_NM
									FROM dar_category_info
									WHERE 
									REST_ID= '#{restId}'
									AND USE_YN='Y'
									`
var SelectStoreMenuList string = `SELECT A.ITEM_NO
											,A.ITEM_NM
											,A.ITEM_PRICE
											,CASE WHEN B.USER_ID IS NULL THEN 'N' ELSE 'Y'  END AS favoritesYn
											,A.BEST_YN
									FROM dar_sale_item_info AS A
									LEFT OUTER JOIN dar_freq_menu_info AS B ON A.REST_ID = B.REST_ID AND A.ITEM_NO = B.ITEM_NO 
									AND B.USER_ID='#{userId}'
									WHERE 
									ITEM_STAT ='1' 
									AND USE_YN='Y'
									AND A.REST_ID='#{restId}'
									AND A.sales_end_Date >  DATE_FORMAT(NOW(),'%Y%m%d')
									ORDER BY B.ITEM_NO DESC , A.BEST_YN DESC ,A.ITEM_MENU ASC,A.ITEM_NO ASC, A.ITEM_NM ASC
									`

var SelectStoreMenuListSearch string = `SELECT A.ITEM_NO
											,A.ITEM_NM
											,A.ITEM_PRICE
											,CASE WHEN B.USER_ID IS NULL THEN 'N' ELSE 'Y'  END AS favoritesYn
											,A.BEST_YN
									FROM dar_sale_item_info AS A
									LEFT OUTER JOIN dar_freq_menu_info AS B ON A.REST_ID = B.REST_ID AND A.ITEM_NO = B.ITEM_NO 
									AND B.USER_ID='#{userId}'
									WHERE 
									ITEM_STAT ='1' 
									AND USE_YN='Y'
									AND A.REST_ID='#{restId}'
									AND A.ITEM_MENU ='#{codeId}'
									AND A.sales_end_Date >  DATE_FORMAT(NOW(),'%Y%m%d')
									ORDER BY B.ITEM_NO DESC , A.BEST_YN DESC ,A.ITEM_MENU ASC, A.ITEM_NM ASC
									`

var SelectOrderCheck string = `SELECT	  A.ORDER_NO
									FROM dar_order_info A
									WHERE
									 A.REST_ID = '#{restId}'
									AND   A.USER_ID = '#{userId}'
									AND   A.GRP_ID = '#{grpId}'
									AND   A.ORDER_STAT = '20'
									AND   A.TOTAL_AMT = #{orderAmt}
									AND   A.ORDER_DATE >  DATE_FORMAT(DATE_ADD(SYSDATE(), INTERVAL -#{checkTime} SECOND), '%Y%m%d%H%i%s')`

var SelectTodayOrderAmt string = `
								SELECT  sum(TODAY_ORDER_AMT) as TODAY_ORDER_AMT
									,COUNT(a.order_no) AS TODAY_COUNT
								FROM dar_order_info AS a
								INNER JOIN (SELECT aa.ordeR_no,IFNULL(SUM(aa.ORDER_AMT*aa.ORDER_QTY),0) AS TODAY_ORDER_AMT
												FROM dar_order_detail AS aa
												WHERE 
										  AA.USER_ID= '#{userId}'
												AND DATE_FORMAT(AA.ORDER_DATE, '%Y%m%d')=DATE_FORMAT(NOW(), '%Y%m%d')
												GROUP BY AA.ORDER_NO
												) AS B ON A.ORDER_NO = B.ORDER_NO
								WHERE 
								 A.ORDER_STAT='20'
                                 AND A.GRP_ID= '#{grpId}'
								`

//SELECT IFNULL(SUM(B.ORDER_AMT*B.ORDER_QTY),0) AS TODAY_ORDER_AMT
//,COUNT(*) AS TODAY_COUNT
//FROM dar_order_info AS A
//INNER JOIN dar_order_detail AS B ON A.ORDER_NO = B.ORDER_NO
//WHERE DATE_FORMAT(A.ORDER_DATE, '%Y%m%d')=DATE_FORMAT(NOW(), '%Y%m%d')
//AND A.ORDER_STAT='20'
//AND A.GRP_ID= '#{grpId}'
//AND B.USER_ID= '#{userId}'

var SelectTodayOrder string = `SELECT A.ORDER_NO
									,DATE_FORMAT(A.ORDER_DATE,'%Y년 %m월 %d일')  AS ORDER_DATE
									,DATE_FORMAT(A.ORDER_DATE,'%p %h:%i')  AS ORDER_TIME
									,FN_GET_GRPNAME(A.GRP_ID) AS GRP_NM
									,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
									,FN_GET_ORDER_USER_CNT(A.ORDER_NO) AS ORDER_USER_CNT
									,B.TOTAL_AMT
									,C.USER_NM
									,(SELECT GRP_TYPE_CD FROM PRIV_GRP_INFO AS AA WHERE A.GRP_ID = AA.GRP_ID) AS GRP_TYPE_CD
									,A.GRP_ID
							FROM dar_order_info AS A
							INNER JOIN priv_user_info AS C ON A.USER_ID= C.USER_ID
							INNER JOIN (SELECT ORDER_NO,SUM(AA.ORDER_AMT * AA.ORDER_QTY) AS TOTAL_AMT
											FROM dar_order_detail AS AA
											WHERE  
							AA.USER_ID='#{userId}'
											AND LEFT(AA.ORDER_DATE,8)=DATE_FORMAT(NOW(), '%Y%m%d')
											GROUP BY ORDER_NO  ) AS B ON A.ORDER_NO  = B.ORDER_NO
							WHERE A.ORDER_STAT='20' 
							AND A.ORDER_TY !='4'
							ORDER BY A.ORDER_DATE DESC
							`

var SelectTodayOrder_V2 string = `SELECT A.ORDER_NO
									,DATE_FORMAT(A.ORDER_DATE,'%Y년 %m월 %d일')  AS ORDER_DATE
									,DATE_FORMAT(A.ORDER_DATE,'%p %h:%i')  AS ORDER_TIME
									,FN_GET_GRPNAME(A.GRP_ID) AS GRP_NM
									,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
									,FN_GET_ORDER_USER_CNT(A.ORDER_NO) AS ORDER_USER_CNT
									,B.TOTAL_AMT
									,C.USER_NM
									,(SELECT GRP_TYPE_CD FROM PRIV_GRP_INFO AS AA WHERE A.GRP_ID = AA.GRP_ID) AS GRP_TYPE_CD
									,A.GRP_ID
									,A.ORDER_STAT
							FROM dar_order_info AS A
							INNER JOIN priv_user_info AS C ON A.USER_ID= C.USER_ID
							INNER JOIN (SELECT ORDER_NO,SUM(AA.ORDER_AMT * AA.ORDER_QTY) AS TOTAL_AMT
											FROM dar_order_detail AS AA
											WHERE  
							AA.USER_ID='#{userId}'
											AND LEFT(AA.ORDER_DATE,8)=DATE_FORMAT(NOW(), '%Y%m%d')
											GROUP BY ORDER_NO  ) AS B ON A.ORDER_NO  = B.ORDER_NO
							WHERE A.ORDER_STAT IN('20','21') 
							AND A.ORDER_TY !='4'
							ORDER BY A.ORDER_DATE DESC
							`

var SelectGrpLastOrder0 string = `
								SELECT  A.ORDER_NO
										,DATE_FORMAT(A.ORDER_DATE,'%Y년 %m월 %d일 %p %h:%i')  AS ORDER_DATE
										,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
										,FN_GET_ORDER_USER_CNT(A.ORDER_NO) AS ORDER_USER_CNT
										,A.TOTAL_AMT
										,A.GRP_ID
										,B.USER_NM
										,A.ORDER_STAT
										,(SELECT GRP_TYPE_CD FROM PRIV_GRP_INFO AS AA WHERE A.GRP_ID = AA.GRP_ID) AS GRP_TYPE_CD
										,IFNULL((SELECT IFNULL(SUM(TOTAL_AMT),0) 
											FROM dar_order_info  AS AA
											WHERE AA.ORDER_STAT='20' AND LEFT(AA.ORDER_DATE,6)=DATE_FORMAT(NOW(), '%Y%m')
											AND AA.ORDER_TY !='4'
											AND A.GRP_ID = AA.GRP_ID ),0) AS THIS_MONTH_ORDER
								FROM dar_order_info AS A
								INNER JOIN priv_user_info AS B ON A.USER_ID = B.USER_ID
								WHERE 
								GRP_ID='#{grpId}'
								AND A.ORDER_TY !='4'
								ORDER BY A.ORDER_DATE DESC
								LIMIT 1
								`

var SelectGrpLastOrder1 string = `
								SELECT  A.ORDER_NO
										,DATE_FORMAT(A.ORDER_DATE,'%Y년 %m월 %d일 %p %h:%i')  AS ORDER_DATE
										,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
										,FN_GET_ORDER_USER_CNT(A.ORDER_NO) AS ORDER_USER_CNT
										,B.TOTAL_AMT
										,A.GRP_ID
                                        ,C.USER_NM
										,A.ORDER_STAT
										,(SELECT GRP_TYPE_CD FROM PRIV_GRP_INFO AS AA WHERE A.GRP_ID = AA.GRP_ID) AS GRP_TYPE_CD
								FROM dar_order_info AS A
								INNER JOIN (SELECT ORDER_NO,SUM(Z.ORDER_AMT * Z.ORDER_QTY) AS TOTAL_AMT
											FROM dar_order_detail AS Z
											WHERE  
											Z.USER_ID='#{userId}'
											GROUP BY ORDER_NO ) AS B ON A.ORDER_NO = B.ORDER_NO 
								INNER JOIN priv_user_info AS C ON A.USER_ID = C.USER_ID
								WHERE 
								GRP_ID='#{grpId}'
								AND A.ORDER_TY !='4'
								ORDER BY A.ORDER_DATE DESC
								LIMIT 1
								`

var SelectBookOrderCount string = `SELECT
									count(*) as totalCount
									FROM dar_order_info AS A
									WHERE
									A.GRP_ID='#{grpId}' 
									AND LEFT(A.ORDER_DATE,8) >= '#{startDate}' 
									AND LEFT(A.ORDER_DATE,8) <= '#{endDate}'
									AND A.REST_ID='#{restId}'
									AND A.USER_ID ='#{searchUserId}' 
									AND A.ORDER_TY !='4'
									`

var SelectBookOrderList string = `SELECT 	A.ORDER_NO
												,DATE_FORMAT(A.ORDER_DATE,'%Y.%m.%d %p %h:%i')  AS ORDER_DATE
												,A.TOTAL_AMT
												,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
												,B.USER_NM 
												,CASE WHEN A.ORDER_TY ='1' THEN 'pay'
													WHEN A.ORDER_TY ='2' THEN 'delivery'
													WHEN A.ORDER_TY ='3' THEN 'takeout'
													WHEN A.ORDER_TY ='5' THEN 'pay'
													END AS ORDER_TY
												,A.ORDER_STAT
									FROM dar_order_info AS A
									INNER JOIN priv_user_info AS B ON A.USER_ID = B.USER_ID
									WHERE
									A.GRP_ID='#{grpId}' 
									AND LEFT(A.ORDER_DATE,8) >= '#{startDate}' 
									AND LEFT(A.ORDER_DATE,8) <= '#{endDate}' 
									AND A.REST_ID='#{restId}'
									AND A.USER_ID ='#{searchUserId}' 
									AND A.ORDER_TY !='4'
									ORDER BY A.ORDER_DATE DESC
									`

var SelectBookOrderTotal string = `SELECT
									IFNULL(SUM(CASE WHEN A.ORDER_STAT='20' THEN A.TOTAL_AMT ELSE 0 END),0) as totalAmt
									FROM dar_order_info AS A
									WHERE
									A.GRP_ID='#{grpId}' 
									AND LEFT(A.ORDER_DATE,8) >= '#{startDate}' 
									AND LEFT(A.ORDER_DATE,8) <= '#{endDate}'
									AND A.REST_ID='#{restId}'
									AND A.USER_ID ='#{searchUserId}' 
									AND A.ORDER_TY !='4'
									`

var SelectBookOrderCount_AUTH1 string = `SELECT
									count(*) as totalCount
									FROM dar_order_info AS A
									INNER JOIN (SELECT ORDER_NO, SUM(ORDER_QTY*ORDER_AMT) AS TOTAL_AMT 
												,AA.USER_ID
											FROM dar_order_detail AS AA
											WHERE  
											
											LEFT(AA.ORDER_DATE,8) >= '#{startDate}' 
											AND LEFT(AA.ORDER_DATE,8) <= '#{endDate}' 
											AND AA.USER_ID='#{searchUserId}'
										GROUP BY ORDER_NO  ) AS B ON A.ORDER_NO  =B.ORDER_NO
									INNER JOIN priv_user_info AS C ON B.USER_ID = C.USER_ID
									WHERE
									A.GRP_ID='#{grpId}'
									AND A.REST_ID='#{restId}'
									AND A.ORDER_TY !='4'
									`

var SelectBookOrderList_AUTH1 string = `SELECT 	A.ORDER_NO
												,DATE_FORMAT(A.ORDER_DATE,'%Y.%m.%d %p %h:%i')  AS ORDER_DATE
												,B.TOTAL_AMT
												,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
												,C.USER_NM 
												,CASE WHEN A.ORDER_TY ='1' THEN 'pay'
													WHEN A.ORDER_TY ='2' THEN 'delivery'
													WHEN A.ORDER_TY ='3' THEN 'takeout'
													WHEN A.ORDER_TY ='5' THEN 'pay'
													END AS ORDER_TY
												,A.ORDER_STAT
									FROM dar_order_info AS A
									INNER JOIN (SELECT ORDER_NO, SUM(ORDER_QTY*ORDER_AMT) AS TOTAL_AMT 
												,AA.USER_ID
											FROM dar_order_detail AS AA
											WHERE
											LEFT(AA.ORDER_DATE,8) >= '#{startDate}' 
											AND LEFT(AA.ORDER_DATE,8) <= '#{endDate}' 
											AND AA.USER_ID='#{searchUserId}'
										GROUP BY ORDER_NO ,AA.USER_ID ) AS B ON A.ORDER_NO  =B.ORDER_NO
									INNER JOIN priv_user_info AS C ON B.USER_ID = C.USER_ID
									WHERE
									A.GRP_ID='#{grpId}'
									AND A.ORDER_TY !='4'
									AND A.REST_ID='#{restId}'
									ORDER BY A.ORDER_DATE DESC
									`

var SelectBookOrderTotal_AUTH1 string = `SELECT
									IFNULL(SUM(CASE WHEN A.ORDER_STAT='20' THEN B.TOTAL_AMT ELSE 0 END),0) as totalAmt
									FROM dar_order_info AS A
									INNER JOIN (SELECT ORDER_NO, SUM(ORDER_QTY*ORDER_AMT) AS TOTAL_AMT 
												,AA.USER_ID
											FROM dar_order_detail AS AA
											WHERE  
											 LEFT(AA.ORDER_DATE,8) >= '#{startDate}' 
											AND LEFT(AA.ORDER_DATE,8) <= '#{endDate}' 
											AND AA.USER_ID='#{searchUserId}'
										GROUP BY ORDER_NO  ) AS B ON A.ORDER_NO  =B.ORDER_NO
									INNER JOIN priv_user_info AS C ON B.USER_ID = C.USER_ID
									WHERE
									A.GRP_ID='#{grpId}'
									AND A.REST_ID='#{restId}'
									AND A.ORDER_TY !='4'
									`

var SelectOrderInfo string = `SELECT A.ORDER_NO
											,B.REST_NM
											,C.GRP_NM
											,A.TOTAL_AMT
											,A.ORDER_STAT
											,DATE_FORMAT(A.ORDER_DATE,'%Y.%m.%d %p%h:%i') AS ORDER_DATE
											,ifnull(A.ORDER_COMMENT,'') AS ORDER_COMMENT
											,A.QR_ORDER_TYPE
									FROM dar_order_info AS A
									INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
									INNER JOIN priv_grp_info AS C ON A.GRP_ID = C.GRP_ID
									WHERE 
									A.ORDER_NO = '#{orderNo}'
									`

var SelectOrderDetail string = `SELECT CASE WHEN  A.ITEM_NO='9999999999'  THEN '금액권'  ELSE B.ITEM_NM END  as menuNm
										 ,SUM(ORDER_AMT* ORDER_QTY) as menuPrice
										 ,SUM(ORDER_QTY) as menuQty
								FROM dar_order_detail AS A 
								LEFT OUTER JOIN dar_sale_item_info AS B ON A.ITEM_NO = B.ITEM_NO
								WHERE 
								A.ORDER_NO= '#{orderNo}'
                                GROUP BY A.ITEM_NO 
								`
var SelectOrderMemo string = `SELECT A.USER_ID as memoUserId
										 ,A.MEMO as memo
										,C.USER_NM
								FROM dar_order_memo AS A
								INNER JOIN priv_user_info AS C ON A.USER_ID = C.USER_ID
								WHERE 
								A.ORDER_NO= '#{orderNo}'
								AND A.USER_ID= '#{userId}'
								`

var SelectOrderUserDetail string = `SELECT C.USER_NM
										,SUM(A.ORDER_QTY * A.ORDER_AMT) AS ORDER_AMT
										,C.USER_ID
										,IFNULL(D.MEMO,'') AS MEMO
								FROM dar_order_detail AS A
								INNER JOIN priv_user_info AS C ON A.USER_ID = C.USER_ID
								LEFT OUTER JOIN dar_order_memo AS D ON A.ORDER_NO=D.ORDER_NO AND C.USER_ID = D.USER_ID
								WHERE 
								A.ORDER_NO= '#{orderNo}'
								GROUP BY  C.USER_NM,C.USER_ID
								
								`

var SelectOrderUserSplitAmt string = `SELECT C.USER_NM
											,SPLIT_AMT AS ORDER_AMT
											,C.USER_ID
											,IFNULL(D.MEMO,'') AS MEMO
										FROM dar_order_split AS A
										INNER JOIN priv_user_info AS C ON A.USER_ID = C.USER_ID
										LEFT OUTER JOIN dar_order_memo AS D ON A.ORDER_NO=D.ORDER_NO AND C.USER_ID = D.USER_ID
										WHERE 
											A.ORDER_NO= '#{orderNo}'
								      `

var SelectOrderUserMenu string = `SELECT A.ORDER_QTY 
										,A.ORDER_AMT
										,CASE A.ITEM_NO WHEN '9999999999' THEN '금액권' ELSE B.ITEM_NM END AS ITEM_NM
								FROM dar_order_detail AS A
								LEFT OUTER JOIN dar_sale_item_info AS B ON A.ITEM_NO = B.ITEM_NO
								WHERE 
								A.ORDER_NO= '#{orderNo}'
								AND A.USER_ID = '#{userId}'
								`

var SelectBooksPayment string = `SELECT A.GRP_ID  as grpId
										,FN_GET_GRPNAME(A.GRP_ID) AS grpNm
										,A.PAYMENT_TY as paymenTy
										,FN_GET_RESTNAME(A.REST_ID ) AS restNm
										,A.REG_DATE as rDate
										,DATE_FORMAT(A.REG_DATE,'%Y년 %m월 %d일') AS regDate
										,0 AS creditAmt
										,0 AS payCnt
										,(SELECT GRP_TYPE_CD FROM PRIV_GRP_INFO AS AA WHERE A.GRP_ID = AA.GRP_ID) AS grpTypeCd
								FROM dar_payment_hist AS A
								WHERE (GRP_ID,REG_DATE) IN (
										SELECT GRP_ID,MAX(REG_DATE) AS REG_DATE
										FROM dar_payment_hist AS A
										INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID
									   	WHERE 
										A.PAYMENT_TY IN ('0','3')
										GROUP BY GRP_ID)
								AND A.GRP_ID IN('#{myGrp}') 
								ORDER BY A.REG_DATE DESC
								
								`
var SelectLastPaymentData string = `SELECT SUM(A.CREDIT_AMT) AS creditAmt
									,COUNT(*) AS payCnt
									FROM dar_payment_hist AS A
									INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
									WHERE 
									A.GRP_ID ='#{grpId}'
									AND LEFT(REG_DATE,8)= LEFT('#{regDate}',8)
									AND A.PAYMENT_TY IN ('0','3')
								`

var SelectBookPaymentListCount string = `SELECT  
										count(*) as totalCount
									,SUM(A.CREDIT_AMT-IFNULL(B.CANCELAMT,0)) AS totalAmt
									FROM dar_payment_hist  AS A
									INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
									WHERE 
									A.GRP_ID ='#{grpId}'
									AND A.REST_ID='#{restId}'
									AND LEFT(A.REG_DATE,8) >= '#{startDate}' 
									AND LEFT(A.REG_DATE,8) <= '#{endDate}' 
								    AND A.PAYMENT_TY IN ('0','3')
								`

var SelectBookPaymentList string = `SELECT A.MOID
										,DATE_FORMAT(A.REG_DATE,'%Y.%m.%d %p%h:%i') AS REG_DATE
										,FN_GET_RESTNAME(A.REST_ID) AS REST_NM
										,CASE  WHEN A.PAYMENT_TY='0' AND CANCELTID IS NOT NULL THEN '1'
										       WHEN A.PAYMENT_TY='3' AND CANCELTID IS NOT NULL THEN '4'
										  ELSE A.PAYMENT_TY END  AS PAYMENT_TY
										,A.CREDIT_AMT
								FROM dar_payment_hist  AS A
								INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
								WHERE 
								A.GRP_ID ='#{grpId}'
								AND A.REST_ID='#{restId}'
								AND LEFT(A.REG_DATE,8) >= '#{startDate}' 
								AND LEFT(A.REG_DATE,8) <= '#{endDate}' 
								AND A.PAYMENT_TY IN ('0','3')
								ORDER BY A.REG_DATE DESC
								`

var SelectPaymentInfo string = `SELECT IFNULL(PAYMETHOD,'STORE') AS PAYMETHOD
										,FN_GET_GRPNAME(A.GRP_ID) AS GRP_NM
										,A.CREDIT_AMT
										,A.ADD_AMT
										,A.PAYMENT_TY
										,FN_GET_RESTNAME(A.REST_ID ) AS REST_NM
										,DATE_FORMAT(A.REG_DATE,'%Y.%m.%d %p%h:%i') AS REG_DATE
										,IFNULL(ACC_ST_DAY,'') AS ACC_ST_DAY
										,(SELECT COUNT(*) FROM dar_order_info AS AA WHERE AA.MOID = A.MOID) AS PAID_CNT
								FROM dar_payment_hist AS a
								LEFT OUTER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
								WHERE 
									A.moid='#{moid}' 
								`

var SelectPaymentCancelChk string = `SELECT A.MOID
								FROM dar_payment_hist AS a
								INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
								WHERE 
									A.moid='#{moid}' 
									AND A.PAYMENT_TY in ('1','4')
								`

var SelectPaymentCancelInfo string = `SELECT PAYMETHOD
										,FN_GET_GRPNAME(A.GRP_ID) AS GRP_NM
										,A.CREDIT_AMT
										,A.ADD_AMT
										,(SELECT PAYMENT_TY FROM dar_payment_hist AS AA WHERE AA.MOID = B.MOID AND AA.PAYMENT_TY IN('1','4') ) AS PAYMENT_TY
										,FN_GET_RESTNAME(A.REST_ID ) AS REST_NM
										,DATE_FORMAT(A.REG_DATE,'%Y.%m.%d %p%h:%i') AS REG_DATE
										,IFNULL(ACC_ST_DAY,'') AS ACC_ST_DAY
										,0 AS PAID_CNT
										,(SELECT DATE_FORMAT(CONCAT(CANCELDATE,CANCELTIME),'%Y.%m.%d %p%h:%i') 
												FROM dar_payment_report AS AA WHERE AA.MOID = B.MOID) AS CANCEL_DATE
								FROM dar_payment_hist AS a
								INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
								WHERE 
									A.moid='#{moid}' 
									AND A.PAYMENT_TY in('0','3')
								`

var SelectUnpaidListCount string = `SELECT COUNT(*) AS orderCnt
									, SUM(total_amt) AS TOTAL_AMT
								FROM  dar_order_info AS A
								INNER JOIN priv_user_info AS B ON A.USER_ID = B.USER_ID
								WHERE 
								A.REST_ID='#{restId}'
								AND A.order_ty IN ('1','2','3','5')
								AND A.GRP_ID ='#{grpId}'
								AND PAY_TY='1' 
								AND PAID_YN='N'
								AND order_stat = '20'
								AND LEFT(A.ORDER_DATE,8) <='#{accStDay}'
								`

var SelectUnpaidList string = `SELECT
									DATE_FORMAT(A.ORDER_DATE,'%Y.%m.%d')  AS ORDER_DATE
									,SUM(A.TOTAL_AMT) AS TOTAL_AMT
									,COUNT(*) AS ORDER_CNT
								FROM  dar_order_info AS A
								WHERE 
								A.REST_ID='#{restId}'
								AND A.order_ty IN ('1','2','3','5')
								AND A.GRP_ID ='#{grpId}'
								AND PAY_TY='1' 
								AND PAID_YN ='N'
								AND order_stat = '20'
								AND LEFT(A.ORDER_DATE,8) <='#{accStDay}'
								GROUP BY DATE_FORMAT(A.ORDER_DATE,'%Y.%m.%d')
								ORDER BY DATE_FORMAT(A.ORDER_DATE,'%Y.%m.%d') DESC
								`

var SelectLastOrderDate string = `SELECT DATE_FORMAT(MAX(A.ORDER_DATE),'%Y-%m-%d')  AS LAST_ORDER_DATE
									FROM dar_order_detail AS A
									INNER JOIN dar_order_info AS B ON A.ORDER_NO = B.ORDER_NO
									WHERE
									B.GRP_ID='#{grpId}'
									AND B.ORDER_TY !='4'
									AND A.USER_ID='#{searchUserId}'
									
								`
var SelectLastPayDate string = `SELECT  DATE_FORMAT(MAX(A.REG_DATE),'%Y-%m-%d')  AS LAST_PAY_DATE
									FROM dar_payment_hist  AS A
									INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID 
									WHERE 
									A.GRP_ID ='#{grpId}'
									AND A.REST_ID='#{restId}'
								    AND A.PAYMENT_TY IN ('0','3')
								`

var InsertOrderCoupon string = `INSERT INTO dar_order_coupon
								(
								ORDER_NO
								, PROD_ID
								, EXPIRE_DATE
								, ORD_NO
								, CPNO
								, EXCH_FR_DY
								, EXCH_TO_DY
								, CP_STATUS
								)
								VALUES
								(
									 '#{orderNo}'
									, '#{prodId}'
									, DATE_FORMAT(NOW(), '%Y%m%d')
									, '#{ordNo}'
									, '#{cpNo}'
									, '#{exchFrDy}'
									, '#{exchToDy}'
									, '#{cpStatus}'
								)`

var InsertOrderCoupon_expireUnlimit string = `INSERT INTO dar_order_coupon
								(
								ORDER_NO
								, PROD_ID
								, EXPIRE_DATE
								, ORD_NO
								, CPNO
								, EXCH_FR_DY
								, EXCH_TO_DY
								, CP_STATUS
								)
								VALUES
								(
									 '#{orderNo}'
									, '#{prodId}'
									, '#{exchToDy}'
									, '#{ordNo}'
									, '#{cpNo}'
									, '#{exchFrDy}'
									, '#{exchToDy}'
									, '#{cpStatus}'
								)`

var SelectRemainCouponList string = `	SELECT A.ORDER_NO
										,A.ORD_NO
										,A.CPNO
										,A.CP_STATUS
										,C.REST_NM
								FROM dar_order_coupon AS A
								INNER JOIN dar_order_info AS B ON A.ORDER_NO = B.ORDER_NO
								INNER JOIN (SELECT BB.REST_NM,AA.SUB_REST_ID
												FROM b_rest_combine_sub AS AA
												INNER JOIN priv_rest_info AS BB ON AA.REST_ID = BB.REST_ID ) AS C ON B.REST_ID = C.SUB_REST_ID
								WHERE 
									 B.USER_ID='#{userId}'
									AND A.CP_STATUS='0'
									AND A.expire_date = DATE_FORMAT(NOW(), '%Y%m%d')
								`

var UpdateCoupon string = `UPDATE dar_order_coupon SET
										CP_STATUS=		'#{cpStatus}',
										EXCHPLC=		'#{exchplc}',
										EXCHCO_NM=		'#{exchcoNm}',
										CPNO_EXCH_DT=	'#{cpnoExchDt}',
										CPNO_STATUS=	'#{cpnoStatus}',
										CPNO_STATUS_CD= '#{cpnoStatusCd}',
										BALANCE=		'#{balance}'
									WHERE 
										CPNO = '#{cpNo}'
										AND ORDER_NO = '#{orderNo}'

								`

var SelectCoupon string = `SELECT B.CPNO AS cpNo
									, DATE_FORMAT(B.EXPIRE_DATE, '~%Y년 %m월 %d일') AS expireDate
									,D.ITEM_NM AS itemName
									,C.ORDER_AMT AS itemPrice
									,B.ORD_NO AS ordNo
									,IFNULL(D.ITEM_IMG,'') AS itemImg
									,IFNULL(E.MEMO,'') AS itemDesc
									,B.CP_STATUS as cpStatus
									,A.REST_ID as restId
							FROM DAR_ORDER_INFO AS A
							INNER JOIN dar_order_coupon AS B ON A.ORDER_NO = B.ORDER_NO
							INNER JOIN dar_order_detail AS C ON B.ORDER_NO = C.order_no
							INNER JOIN dar_sale_item_info AS D ON C.ITEM_NO = D.ITEM_NO
							LEFT OUTER JOIN priv_rest_etc AS E ON A.REST_ID = E.REST_ID 
							WHERE 
								A.order_no='#{orderNo}'
								`

var SelectCouponCancelInfo string = `SELECT A.ORDER_NO
										,A.ORD_NO
										,A.CPNO
										,A.CP_STATUS
										,A.EXPIRE_DATE
								FROM dar_order_coupon AS A
								WHERE 
									 A.ORDER_NO='#{orderNo}'
									AND A.CP_STATUS='0'
								`

var SelectCouponInfo string = `SELECT 	A.CPNO
										,A.CP_STATUS
										,A.EXPIRE_DATE
								FROM dar_order_coupon AS A
								WHERE 
									 A.ORDER_NO='#{orderNo}'
								`

var SelectCouponCheckList string = `SELECT A.ORDER_NO
										,A.CPNO
										,A.ORD_NO
									FROM dar_order_coupon AS a
									INNER JOIN dar_order_info AS b ON a.ORDER_NO = b.order_no AND b.ORDER_STAT='20'
									INNER JOIN b_rest_combine_sub AS c ON b.REST_ID = c.sub_rest_id
									WHERE CP_STATUS='0' 
									AND C.REST_ID='#{restId}'
								`

var UpdateCouponCancel string = `UPDATE dar_order_coupon SET
										CP_STATUS='1'
										,CPNO_STATUS = '취소'
									WHERE 
										 ORDER_NO = '#{orderNo}'
								`

var SelectOrder string = `SELECT A.ORDER_NO
							,A.TOTAL_AMT
							,A.ORDER_STAT
							,A.PAY_TY
							,A.GRP_ID AS BOOK_ID
							,A.REST_ID AS STORE_ID
							,A.USER_ID
							,A.POINT_USE
							,A.ORDER_DATE
							,A.INSTANT_MOID
							,A.INSTANT_AMT
							,A.ORDER_TY
					FROM dar_order_info AS A
					WHERE 
					A.ORDER_NO = '#{orderNo}'
					`

var SelectLinkInfo string = `SELECT 
							A.AGRM_ID AS LINK_ID
							, A.GRP_ID
							, FN_GET_GRPNAME(A.GRP_ID) AS BOOK_NM
							, A.REST_ID
							, FN_GET_RESTNAME(A.REST_ID) AS STORE_NM
							, A.REQ_STAT
							, FN_GET_CODENAME('AGRM_STAT', A.REQ_STAT) AS REQ_STAT_NM
							, A.REQ_TY
							, DATE_FORMAT(A.REQ_DATE, '%Y-%m-%d') AS REQ_DATE	
							, DATE_FORMAT(A.AUTH_DATE, '%Y-%m-%d') AS AUTH_DATE
							, A.PAY_TY
							, IFNULL(A.PREPAID_AMT, 0) AS PREPAID_AMT
							, IFNULL(A.PREPAID_POINT, 0) AS PREPAID_POINT
							FROM org_agrm_info  AS a
							INNER join PRIV_REST_INFO AS B on A.REST_ID = B.REST_ID or (B.FRAN_YN = 'Y' and B.FRAN_ID = A.REST_ID)
							WHERE
							B.REST_ID ='#{storeId}'
							AND A.GRP_ID='#{bookId}'
							`
var UpdateLink string = `UPDATE org_agrm_info SET 
							PREPAID_AMT = '#{prepaidAmt}'
							,PREPAID_POINT = #{prepaidPoint}
							,MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
							WHERE 
							AGRM_ID ='#{linkId}'
						`

var UpdateOrderCancel string = ` UPDATE dar_order_info SET ORDER_STAT='21'
												, ORDER_CANCEL_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
								WHERE
								ORDER_NO = '#{orderNo}'
									`

var SelectGiftMedia string = ` SELECT B.REST_NM
								FROM b_rest_combine_sub AS A
								INNER JOIN priv_rest_info AS B ON A.REST_ID = B.REST_ID
								WHERE 
									A.SUB_REST_ID='#{restId}'
									`

var InsertWincubeCoupon string = ` INSERT INTO t_wincube_coupon
									(
									ORDER_NO
									,result_code
									,result_reason
									,tr_id
									,ctr_id
									,ExpirationDate
									,pin_no
									)
									VALUES (
									'#{orderNo}', 
									'#{resultCd}', 
									'#{reason}', 
									'#{trID}', 
									'#{ctrID}', 
									'#{expirationDate}', 
									'#{pinNo}'
									)
									`

var SelectMyGiftList string = `SELECT B.CPNO 
									,DATE_FORMAT(B.EXCH_TO_DY, '~%Y년 %m월 %d일') AS expireDateStr
									,D.ITEM_NM 
									,C.ORDER_AMT 
									,IFNULL(D.ITEM_IMG,'') as ITEM_IMG
									,B.CP_STATUS 
									,B.EXPIRE_DATE
									,EXCH_TO_DY
									,E.REST_NM
									,A.ORDER_NO
							FROM DAR_ORDER_INFO AS A
							INNER JOIN dar_order_coupon AS B ON A.ORDER_NO = B.ORDER_NO
							INNER JOIN dar_order_detail AS C ON B.ORDER_NO = C.order_no
							INNER JOIN dar_sale_item_info AS D ON C.ITEM_NO = D.ITEM_NO
							INNER JOIN PRIV_REST_INFO AS E ON D.REST_ID =E.REST_ID
							WHERE 
							CP_STATUS='0'
							AND A.USER_ID ='#{userId}'
								`
