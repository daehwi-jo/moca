package companys


var CreateCompanySeq string = `INSERT INTO b_company_seq VALUE()`

var InsertCompany string = `INSERT INTO b_company 
							(
								COMPANY_ID,
								COMPANY_NM,
								BUSID,
								if #{ceoNm} != '' then CEO_NM,
								if #{addr} != '' then ADDR,
								if #{addr2} != '' then ADDR2,
								if #{companyTel} != '' then TEL,
								if #{email} != '' then EMAIL,
								if #{hompage} != '' then homepage,
								if #{intro} != '' then intro,
								REG_DATE
							)
							VALUES(
								#{companyId}
								,'#{companyNm}'
								,'#{bizNum}'
								,'#{ceoNm}'
								,'#{addr}'
								,'#{addr2}'
								,'#{companyTel}'
								,'#{email}'
								,'#{hompage}'
								,'#{intro}'
								,DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
							)`

var InsertCompanyBook string = `INSERT INTO b_company_book 
							(
								company_id
								,BOOK_ID
								,REG_DATE
							)
							VALUES(
								#{companyId}
								,'#{grpId}'
								,DATE_FORMAT(SYSDATE(), '%Y%m%d%H%i%s') 
							)`


var SelectCompanyId string = `SELECT
								max(company_id)+1  AS companyId
							FROM b_company
							`

var SelectCompanyInfo string = `SELECT BUSID
								,COMPANY_ID
							FROM b_company
							WHERE 
							BUSID='#{bizNum}'`