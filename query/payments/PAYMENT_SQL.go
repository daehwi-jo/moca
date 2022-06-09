package payments

var UpdatePaymentCancel string = `UPDATE 	DAR_PAYMENT_REPORT  SET 	
                                     RESULTCD= '#{resultcd}'
									,RESULTMSG= '#{resultmsg}'
									,CANCELDATE= '#{canceldate}'
									,CANCELTIME= '#{canceltime}'
									,STATECD= '#{statecd}'
									,CANCELAMT= #{cancelamt}
									,CANCELMSG= '#{cancelmsg}'
									,CANCELTID= '#{canceltid}'
									,CC_IP= #{ccIp}
									WHERE
									MOID = '#{moid}'
									`

var SelctLinkInfo string = `SELECT AGRM_ID
										,PREPAID_AMT
										,PREPAID_POINT
										FROM org_agrm_info
										WHERE 
										REST_ID='#{restId}'
										AND GRP_ID='#{grpId}'
										`

var SelectPaymentHist string = `SELECT
									A.GRP_ID
									, A.REST_ID
									, A.PAY_TY
									, C.SEARCH_TY
									, C.USER_ID
									, C.PAY_INFO
									, C.PAY_CHANNEL
									, IFNULL(C.ADD_AMT,0) AS ADD_AMT
									, C.PAYMENT_TY
									, IFNULL(C.CREDIT_AMT,0) AS CREDIT_AMT
									, (SELECT TID FROM dar_payment_report AS AA WHERE C.MOID = AA.MOID LIMIT 1) AS TID
									, LEFT(C.REG_DATE,8) AS REG_DATE
									, DATE_FORMAT(NOW(), '%Y%m%d') AS TO_DATE
									FROM ORG_AGRM_INFO AS A
									left join priv_rest_info AS B on A.REST_ID = B.rest_id or (B.FRAN_YN = 'Y' and B.FRAN_ID = A.REST_ID)
									left join dar_payment_hist AS C ON  A.GRP_ID = C.GRP_ID AND  B.REST_ID = C.REST_ID
									WHERE  
									C.MOID = '#{moid}'
									AND C.PAYMENT_TY IN('0','3')
									`

var UpdateAgrm string = `UPDATE ORG_AGRM_INFO SET MOD_DATE = DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
													,PREPAID_AMT = #{prepaidAmt}
													,PREPAID_POINT = #{prepaidPoint}
												WHERE 
													AGRM_ID = '#{agrmId}'
													`

var UpdatePostPay string = `UPDATE DAR_ORDER_INFO SET PAID_YN = 'N'
													,PAY_DATE = NULL
										WHERE 
										   REST_ID = '#{restId}'
										   AND GRP_ID = '#{grpId}'
										   AND MOID = '#{moid}'
										   AND ORDER_STAT = '20'
										   AND PAY_TY = '1'
									`

var InsertPaymentHistory string = `INSERT INTO DAR_PAYMENT_HIST
										(
										HIST_ID
										,REST_ID
										,GRP_ID
										,USER_ID
										,CREDIT_AMT
										,USER_TY
										,SEARCH_TY
										,PAYMENT_TY
										,PAY_INFO
										,REG_DATE
										,PAY_CHANNEL
										,ADD_AMT
										,MOID
										)
										VALUES(
										( SELECT FN_GET_SEQUENCE('PAYMENT_HIST_ID') AS TMP )
										, '#{restId}'
										, '#{grpId}'
										, '#{userId}'
										, #{creditAmt}
										, '#{userTy}'
										, '#{searchTy}'
										, '#{paymentTy}'
										, '#{payInfo}'
										, DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										, '#{payChannel}'
										, #{addAmt}
										, '#{moid}'
										)
									`

var InsertPrepaid string = `INSERT INTO DAR_PREPAID_INFO
										(
										PREPAID_NO
										,GRP_ID
										,REST_ID
										,JOB_TY
										,PREPAID_AMT
										,REG_DATE
										)
										VALUES(
										 CONCAT(DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '#{grpId}' )
										,'#{grpId}'
										,'#{restId}'
										,'#{jobTy}'
										,'#{prepaidAmt}'
										,DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')
										)
									`

var SelectTpayBillingKey string = `SELECT SEQ
									,CARD_NUM
									FROM b_tpay_billing_key
									WHERE
									USER_ID='#{userId}'
										`

var SelectTpayBillingCardInfo string = `SELECT USER_ID
										, SEQ
										, CARD_NAME
										, CARD_CODE
										, CARD_NUM
										, CARD_TOKEN
										, CARD_TYPE
										, USE_YN
									FROM b_tpay_billing_key
									WHERE
									USER_ID='#{userId}'
									AND SEQ='#{seq}'
										`

var SelectTpayBillingCardPwq string = `SELECT IFNULL(BILLING_PWD,'NONE') AS BILLING_PWD
									FROM priv_user_info
									WHERE
									USER_ID='#{userId}'
										`

var InsertTpayBillingKey string = `INSERT INTO b_tpay_billing_key
										(
										USER_ID, 
										SEQ, 
										CARD_NAME, 
										CARD_CODE, 
										CARD_NUM, 
										CARD_TOKEN, 
										CARD_TYPE
										)
										VALUES(
										'#{userId}'
										,'#{seq}'
										,'#{cardName}'
										,'#{cardCode}'
										,'#{cardNum}'
										,'#{cardToken}'
										,'#{cardType}'
										)
									`

var UpdateTpayBillingKey string = `UPDATE b_tpay_billing_key SET CARD_TOKEN = ''
													,USE_YN = 'N'
													,CARD_NUM=''
										WHERE
											USER_ID='#{userId}'
											AND SEQ='#{seq}'
									`

var UpdateTpayBillingPwd string = `UPDATE priv_user_info SET 
												BILLING_PWD = '#{billPwd}'
										WHERE
											USER_ID='#{userId}'
									`

var SelectTpayBillingCardList string = `SELECT USER_ID
										, SEQ
										, CARD_NAME
										, CARD_CODE
										, CARD_NUM
										, CARD_TYPE
									FROM b_tpay_billing_key
									WHERE
									USER_ID='#{userId}'
									AND USE_YN='Y'
										`

var SelectPaymentInfo string = `SELECT A.PAYMENT_TY
										,A.PAY_INFO
										,A.SEARCH_TY
										,B.TID
										,A.CREDIT_AMT
										,A.PAY_CHANNEL
										,A.USER_TY
										FROM dar_payment_hist AS a
										INNER JOIN dar_payment_report AS B ON A.MOID = B.MOID
										WHERE 
										A.MOID='#{moid}'
`
