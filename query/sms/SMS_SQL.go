package sms



var InsertConfirmData string = `INSERT INTO biz_confirm_info 
									(
										CONFIRM_DIV
										,TEL_NUM 
										,EMAIL 
										,SND_DTM 
										,CONFIRM_NUM
										,CONFIRM_YN
										,CONFIRM_DTM
										,MOD_DTM
									)
									SELECT 
										IFNULL('#{confirmDiv}','')               
										,IFNULL('#{telNum}','') 
										,''                  
										,DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
										,IFNULL('#{confirmNum}','')               
										,'N'                                    
										, ''                                    
										, DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
									FROM DUAL`

var SelectConfirmCheck string = `SELECT 
									g.CONFIRM_DIV
									,g.TEL_NUM 
									,g.EMAIL 
									,g.SND_DTM 
									,g.CONFIRM_NUM 
									,DATE_FORMAT(DATE_ADD(SYSDATE(),INTERVAL -3 MINUTE), '%Y%m%d%H%i%s') AS NOW_DATE
									FROM biz_confirm_info g
									WHERE 
									g.CONFIRM_DIV = '#{confirmDiv}'
									AND  G.TEL_NUM='#{telNum}'
									`

var UpdateConfirmData string = `UPDATE biz_confirm_info SET CONFIRM_YN='Y'
															,CONFIRM_DTM=DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
								WHERE 
								CONFIRM_DIV = '#{confirmDiv}'
								AND TEL_NUM='#{telNum}'
								 `

var UpdateConfirmDataReset string = `UPDATE biz_confirm_info SET CONFIRM_YN='N'
								,SND_DTM=DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
								,CONFIRM_NUM ='#{confirmNum}'
								,MOD_DTM=DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
								,CONFIRM_DTM=''
								WHERE 
								CONFIRM_DIV = '#{confirmDiv}' 
								AND TEL_NUM='#{telNum}'
								`
