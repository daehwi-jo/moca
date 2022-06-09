package wincubes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	ordersql "mocaApi/query/orders"
	restsql "mocaApi/query/rests"
	controller "mocaApi/src/controller"
	"mocaApi/src/controller/cls"
	commons "mocaApi/src/controller/commons"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf



type AuthReqData struct {
	CustId       string `json:"custId"`       //
	Pwd      	 string `json:"pwd"`      //
	AutKey 	 	 string `json:"autKey"` //
	AesKey	 	 string `json:"aesKey"` //
	AesIv	 	 string `json:"aesIv"` //
}

type AuthResult struct {
	ResultCode       string `json:"resultCode"`       //
	CodeId      	 string `json:"codeId"`      //
	ExpireDate 	 	 string `json:"expireDate"` //
	ExpireTime	 	 string `json:"expireTime"` //
	Message	 		 string `json:"message"` //
}

type TokenReqData struct {
	CodeId       string `json:"codeId"`       //
}

type TokenResult struct {
	ResultCode       string `json:"resultCode"`       //
	TokenId       	 string `json:"tokenId"`       //
	ExpireDate       string `json:"expireDate"`       //
	ExpireTime       string `json:"expireTime"`       //
	Message       	 string `json:"message"`       //
}

type ItemListResult struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		Code     string `xml:"code"`
		Goodsnum string `xml:"goodsnum"`
	} `xml:"result"`
	Value struct {
		Goodslist []struct {
			GoodsID           string `xml:"goods_id"`
			Category1         string `xml:"category1"`
			Category2         string `xml:"category2"`
			Affiliate         string `xml:"affiliate"`
			AffiliateCategory string `xml:"affiliate_category"`
			HeadSwapCd        string `xml:"head_swap_cd"`
			SwapCd            string `xml:"swap_cd"`
			Desc              string `xml:"desc"`
			GoodsNm           string `xml:"goods_nm"`
			GoodsImg          string `xml:"goods_img"`
			NormalSalePrice   string `xml:"normal_sale_price"`
			NormalSaleVat     string `xml:"normal_sale_vat"`
			SalePrice         string `xml:"sale_price"`
			SaleVat           string `xml:"sale_vat"`
			TotalPrice        string `xml:"total_price"`
			PeriodEnd         string `xml:"period_end"`
			LimitDate         string `xml:"limit_date"`
			EndDate           string `xml:"end_date"`
		} `xml:"goodslist"`
	} `xml:"value"`
}


type OrderResult struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		Code   string `xml:"code"`
		Reason string `xml:"reason"`
	} `xml:"result"`
	Value struct {
		PinNo          string `xml:"pin_no"`
		TrID           string `xml:"tr_id"`
		CtrID          string `xml:"ctr_id"`
		ExpirationDate string `xml:"ExpirationDate"`
		CreateDateTime string `xml:"createDateTime"`
		PinSubInfo     string `xml:"pin_sub_info"`
	} `xml:"value"`
}



type CouponStatusResult struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		TrID           string `xml:"trID"`
		PurchNo        string `xml:"PurchNo"`
		StatusCode     string `xml:"StatusCode"`
		StatusText     string `xml:"StatusText"`
		SwapDt         string `xml:"SwapDt"`
		CpnNo          string `xml:"CpnNo"`
		TotAmt         string `xml:"TotAmt"`
		UseAmt         string `xml:"UseAmt"`
		ExpirationDate string `xml:"ExpirationDate"`
		CpnStatus      string `xml:"CpnStatus"`
		PinSubInfo     string `xml:"pin_sub_info"`
	} `xml:"result"`
}

type CancelResult struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		TrID           string `xml:"trID"`
		StatusCode     string `xml:"StatusCode"`
		StatusText     string `xml:"StatusText"`
		CancelDateTime string `xml:"cancelDateTime"`
	} `xml:"result"`
}


type ItemCheckResult struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		Code     string `xml:"code"`
		Reason   string `xml:"reason"`
		StockCnt string `xml:"stock_cnt"`
	} `xml:"result"`
}

type CheckBalance struct {
	Result_code			string `json:"result_code"`
	Available_balance   int `json:"available_balance"`
}



// 선물 내역
func GetTest(c echo.Context) error {

	dprintf(4, c, "call GetTest\n")

	//params := cls.GetParamJsonMap(c)

	custId :="fitdarayo"
	pwd := "ekfdkdyahzk!@#"
	autKey := "892d63afddbf03c57e45bbf116a7271f26cc79c0123ce560995d0873ee85d679"
	aesKey :="d2022a2022r2022a2022y2022o202201"
	aesIv :="fitdarayowincube"

	var reqData AuthReqData
	reqData.CustId 	=Ase256(custId,aesKey,aesIv,aes.BlockSize)
	reqData.Pwd	=  Ase256(pwd,aesKey,aesIv,aes.BlockSize)
	reqData.AutKey	=  Ase256(autKey,aesKey,aesIv,aes.BlockSize)
	reqData.AesKey 	= RsaEncrypt(aesKey)
	reqData.AesIv 	= RsaEncrypt(aesIv)


	pUrl:="http://dev.giftting.co.kr:48081/auth/code/issue"
	pbytes, _ := json.Marshal(reqData)
	buff := bytes.NewBuffer(pbytes)
	req, err := http.NewRequest("POST", pUrl, buff)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println(err.Error())
	}
	defer resp.Body.Close()

	result := make(map[string]string)

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return c.JSON(http.StatusOK, result)
	}

	var authResult AuthResult
	err = json.Unmarshal(respBody, &authResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "인증 요청 실패"
		return c.JSON(http.StatusOK, result)
	}

	if authResult.ResultCode=="200"{
		println(authResult.CodeId)
	}else{

		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "인증 요청 실패"
		return c.JSON(http.StatusOK, result)

	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["result"] = result

	return c.JSON(http.StatusOK, m)

}
func SetWincubeOrder(c echo.Context) error {


	params := cls.GetParamJsonMap(c)

	tokenId,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}

	goods_id :=params["goods_id"]
	order_no :=params["order_no"]


	orderResult := CallWincubeOrder(tokenId,goods_id,order_no)

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["ttt"] = orderResult

	return c.JSON(http.StatusOK, m)
}


// 윈큐브 상품 업데이트
func SetWincubeItemUpdate(c echo.Context) error {

	dprintf(4, c, "call SetWincubeItemUpdate\n")

	params := cls.GetParamJsonMap(c)


	tokenId,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}

	gooodsId:=params["gooodsId"]
	itemCheckResult := CallWincubeItemCheck(tokenId,gooodsId)



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = itemCheckResult

	return c.JSON(http.StatusOK, m)

}


// 윈큐브 잔여한도 체크
func GetWincubeCheckBalance(c echo.Context) error {

	dprintf(4, c, "call GetWincubeCheckBalance\n")



	tokenId,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}

	itemCheckResult := CallWincubeCheckBalance(tokenId)



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = itemCheckResult

	return c.JSON(http.StatusOK, m)

}



// 기프티콘 상태 취소
func SetWincubeCancel(c echo.Context) error {

	dprintf(4, c, "call SetWincubeCancel\n")

	params := cls.GetParamJsonMap(c)


	tokenId,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}

	orderNo:=params["orderNo"]
	couponStatusResult := CallWincubeCancel(tokenId,orderNo)



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = couponStatusResult

	return c.JSON(http.StatusOK, m)

}



// 기프티콘 상태 업데이트
func SetWincubeGifticonStatus1(c echo.Context) error {

	dprintf(4, c, "call SetWincubeGifticonStatus\n")

	params := cls.GetParamJsonMap(c)


	tokenId,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}

	orderNo:=params["orderNo"]
	couponStatusResult := CallWincubeCouponStatus(tokenId,orderNo)



	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["resultData"] = couponStatusResult

	return c.JSON(http.StatusOK, m)

}


// 기프티콘 사용체크
func SetGifticonCheck(c echo.Context) error {

	dprintf(4, c, "call SetGifticonCheck\n")

	params := cls.GetParamJsonMap(c)

	tokenId,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}



	couPonInfo,err := cls.GetSelectData(ordersql.SelectCouponCheckList, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", "DB fail"))
	}
	if couPonInfo == nil {
		m := make(map[string]interface{})
		m["resultCode"] = "99"
		m["resultMsg"] = "미사용 쿠폰이 없습니다."
		return c.JSON(http.StatusOK, m)
	}

	for i, _ := range couPonInfo {

		ordNo:=couPonInfo[i]["ORD_NO"]
		cpNo:=couPonInfo[i]["CPNO"]
		orderNo:=couPonInfo[i]["ORDER_NO"]

		params["ordNo"] =ordNo
		params["cpNo"] =cpNo
		params["orderNo"] =orderNo



		chkResult := CallWincubeCouponStatus(tokenId,orderNo)
		cpnoResultCd	:=chkResult["RESULT_CD"]
		params["exchplc"]		=chkResult["EXCHPLC"]
		params["exchcoNm"]		=chkResult["EXCHCO_NM"]
		params["cpnoExchDt"]	=chkResult["CPNO_EXCH_DT"]
		params["cpnoStatus"]	=chkResult["CPNO_STATUS"]
		params["cpnoStatusCd"]	=chkResult["CPNO_STATUS_CD"]
		params["balance"]		=chkResult["BALANCE"]


		switch cpnoResultCd {

		case "0": //
			params["cpStatus"]="0"
			break
		case "4005": //이미 취소된 상품입니다.
			params["cpStatus"]="1"
			break
		case "4006": //교환된 상품으로 취소가 불가합니다.
			params["cpStatus"]="2"
			break
		case "4007": //상품권의 기간이 만료되었습니다.
			params["cpStatus"]="9"
			break
		case "3201":  //매체아이디가 존재하지않습니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "매체아이디가 존재하지않습니다."))
			break
		case "3901": //고유아이디가 존재하지 않습니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "고유아이디가 존재하지 않습니다."))
			break
		case "3912": //유효한 매체코드가 아닙니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "유효한 매체코드가 아닙니다."))
			break
		case "2001": //유효한 구매번호가 아닙니다.
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "유효한 구매번호가 아닙니다."))
			break
		case "4004": //상품권 조회 불가상태 기타
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "상품권 조회 불가상태 기타 ."))
			break
		}

		UpdateCouponQuery, err := cls.GetQueryJson(ordersql.UpdateCoupon, params)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}
		// 쿼리 실행
		_, err = cls.QueryDB(UpdateCouponQuery)
		if err != nil {
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	}

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"


	return c.JSON(http.StatusOK, m)

}


func GetWincubeItemList(c echo.Context) error {



	tokenid,resultcd :=GetWincubeAuth()
	if resultcd=="99"{
		lprintf(1, "[INFO] token recv fail \n")
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "token 생성 실패"))
	}

	//println(tokenid)


	params := make(map[string]string)



	result,itemList := CallWincubeItemList(tokenid)
	if result=="00"{
		for i := range itemList.Value.Goodslist {

			//  TRNAN 시작
			tx, err := cls.DBc.Begin()
			if err != nil {
				//return "5100", errors.New("begin error")
			}
			txErr := err

			// 오류 처리
			defer func() {
				if txErr != nil {
					// transaction rollback
					dprintf(4, c, "do rollback -윈큐브 가맹점 및 상품 등록 (GetWincubeItemList)  \n")
					tx.Rollback()
				}
			}()

				swapCd := itemList.Value.Goodslist[i].SwapCd
				restNm := itemList.Value.Goodslist[i].Affiliate
				memo := itemList.Value.Goodslist[i].Desc
				goodsId := itemList.Value.Goodslist[i].GoodsID
				params["swapCd"] = swapCd



			//가맹점 체크
				storeChk, err := cls.GetSelectData(restsql.SelectStoreEtcCheck, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				if storeChk == nil {

					//가맹점 아이디 생성
					storeSeqData, err := cls.GetSelectData(restsql.SelectStoreEtcSeq, params, c)
					if err != nil {
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					params["restId"] = storeSeqData[0]["storeSeq"]
					params["restNm"] = restNm
					params["busId"] = "2148878503"
					params["ceoNm"] = "Wincube"

					insertStore, err := cls.GetQueryJson(restsql.InsertWincubeStore, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(insertStore)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}


					params["memo"] = strings.Replace(memo, "'", "", -1)
					insertStoreEtcQuery, err := cls.GetQueryJson(restsql.InsertWincubeStoreEtc, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(insertStoreEtcQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}

					restCobimeInfo, err := cls.GetSelectData(restsql.SelectStoreBiznum, params, c)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}

					params["mainRestId"]=restCobimeInfo[0]["MAIN_REST_ID"]

					InsertWincubeStoreCombineQuery, err := cls.GetQueryJson(restsql.InsertWincubeStoreCombine, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(InsertWincubeStoreCombineQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}


					cateSeq, err := cls.GetSelectData(restsql.SelectStoreItemCodeSeq, params, c)
					if err != nil {
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					if cateSeq == nil {
						return c.JSON(http.StatusOK, controller.SetErrResult("99", "카테고리 ID  생성 실패."))
					}
					params["cateSeq"] = cateSeq[0]["cateSeq"]
					params["categoryId"] = cateSeq[0]["CATEGORY_ID"]
					params["codeNm"] = "전체"

					insertCategoryQuery, err := cls.GetQueryJson(restsql.InsertStoreCategories, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(insertCategoryQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}


				}else{
						params["restId"]= storeChk[0]["REST_ID"]
						params["memo"] = strings.Replace(memo, "'", "", -1)
						insertStore, err := cls.GetQueryJson(restsql.UpdateWincubeStoreEtc, params)
						if err != nil {
							txErr = err
							return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
						}
						// 쿼리 실행
						_, err = tx.Exec(insertStore)
						if err != nil {
							txErr = err
							return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
						}
				}


				params["goodsId"]= goodsId

			//상품 체크



				params["itemNm"]= strings.Replace(itemList.Value.Goodslist[i].GoodsNm, "'", "", -1)
				params["itemPrice"]=itemList.Value.Goodslist[i].NormalSalePrice // 달아요 판매가격
				params["itemImg"]=itemList.Value.Goodslist[i].GoodsImg
				params["periodEnd"]=itemList.Value.Goodslist[i].PeriodEnd


				params["normalSalePrice"]=itemList.Value.Goodslist[i].NormalSalePrice // 소비자 가격
				params["normalSaleVat"]=itemList.Value.Goodslist[i].NormalSaleVat // 소비자 부가세
				params["salePrice"]=itemList.Value.Goodslist[i].SalePrice // 윈큐브 판매가격
				params["saleVat"]=itemList.Value.Goodslist[i].SaleVat // 윈큐브 부개
				params["totalPrice"]=itemList.Value.Goodslist[i].TotalPrice //윈큐브 판매 + 부가세


				itemChk, err := cls.GetSelectData(restsql.SelectWincubeItemCheck, params, c)
				if err != nil {
					txErr = err
					return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
				}
				//insert
				if itemChk == nil {

					menuSeq, err := cls.GetSelectData(restsql.SelectStoreMenuSeq, params, c)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					if menuSeq == nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("99", "메뉴 ID  생성 실패."))
					}


					params["itemNo"]=menuSeq[0]["itemNo"]
					params["codeId"]="00001"
					InsertStroeMenuQuery, err := cls.GetQueryJson(restsql.InsertStroeMenu, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(InsertStroeMenuQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}

				}else{

					params["itemNo"]=itemChk[0]["ITEM_NO"]

					UpdateStroeMenuQuery, err := cls.GetQueryJson(restsql.UpdateStroeMenu, params)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}
					// 쿼리 실행
					_, err = tx.Exec(UpdateStroeMenuQuery)
					if err != nil {
						txErr = err
						return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
					}

				}
			// transaction commit
			err = tx.Commit()
			if err != nil {
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}


		}


	}








	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = "응답 성공"
	m["itemList"] = itemList

	return c.JSON(http.StatusOK, m)

}



func GetWincubeAuth() (string,string) {


	tokenRequestYn := false

	tokenInfo := make(map[string]string)
	rst,tokenId :=commons.RedisGet("tokenId")
	//tokenExpire:=""
	if rst > 0{
		tokenInfo["tokenId"] = tokenId
		_,tokenExpireDate :=commons.RedisGet("tokenExpireDate")
		tokenInfo["tokenExpireDate"] = tokenExpireDate
		//tokenExpire= tokenExpireDate
		_,tokenExpireTime :=commons.RedisGet("tokenExpireTime")
		tokenInfo["tokenExpireTime"] = tokenExpireTime
	}else{
		tokenRequestYn=true
	}

	//======토큰 만료 체크
	tokenTime := tokenInfo["tokenExpireDate"]+tokenInfo["tokenExpireTime"]
	nowTime := time.Now().Format("20060102150405")

	intTokenTime,_ := strconv.Atoi(tokenTime)
	intNowTime,_ := strconv.Atoi(nowTime)
	TimeCompare := intTokenTime-intNowTime
	if TimeCompare < 0{
		tokenRequestYn=true
	}
	//======토큰 만료 체크


	if tokenRequestYn==true{
		wincubeAUTH :=CallWincubeCode()
		if wincubeAUTH["ResultCode"]=="200" {
			wincubeToken :=CallWincubeToken(wincubeAUTH["CodeId"])
			if wincubeToken["ResultCode"]=="200"{
				commons.RedisSet("tokenId",wincubeToken["TokenId"])
				commons.RedisSet("tokenExpireDate",wincubeToken["ExpireDate"])
				commons.RedisSet("tokenExpireTime",wincubeToken["ExpireTime"])
				tokenId=wincubeToken["TokenId"]
			}else{
				return "","99"
			}
		}else{
			return "","99"
		}
	}

	return tokenId,"00"
}


func CallWincubeCode() (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeCode \n")

	burl:= controller.WINCUBE_URL
	autKey := controller.WINCUBE_AUTKEY

	custId :="fitdarayo"
	pwd := "ekfdkdyahzk!@#"
	//autKey := "892d63afddbf03c57e45bbf116a7271f26cc79c0123ce560995d0873ee85d679"
	aesKey :="d2022a2022r2022a2022y2022o202201"
	aesIv :="fitdarayowincube"
	pURL:=burl+"/auth/code/issue"

	result := make(map[string]string)


	var reqData AuthReqData
	reqData.CustId 	=Ase256(custId,aesKey,aesIv,aes.BlockSize)
	reqData.Pwd	=  Ase256(pwd,aesKey,aesIv,aes.BlockSize)
	reqData.AutKey	=  Ase256(autKey,aesKey,aesIv,aes.BlockSize)
	reqData.AesKey 	= RsaEncrypt(aesKey)
	reqData.AesIv 	= RsaEncrypt(aesIv)


	pbytes, _ := json.Marshal(reqData)
	buff := bytes.NewBuffer(pbytes)
	req, err := http.NewRequest("POST", pURL, buff)
	if err != nil {
		result["ResultCode"] = "99"
		result["Message"] = "인증 요청 실패"
		return result
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		result["ResultCode"] = "99"
		result["Message"] = "코드 요청 실패"
		return result
	}
	defer resp.Body.Close()


	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["ResultCode"] = "99"
		result["Message"] = "코드 요청 실패"
		return result
	}

	var authResult AuthResult
	err = json.Unmarshal(respBody, &authResult)
	if err != nil {
		result["ResultCode"] = "99"
		result["Message"] = "코드 요청 실패"
		return result
	}
	lprintf(4, "[INFO] Call CallWincubeCode resultCode \n", authResult.ResultCode)
	lprintf(4, "[INFO] Call CallWincubeCode resultMessage \n", authResult.Message)


	if authResult.ResultCode=="200"{
		result["ResultCode"] = authResult.ResultCode
		result["CodeId"] = authResult.CodeId
		result["ExpireDate"] = authResult.ExpireDate
		result["ExpireTime"] = authResult.ExpireTime
		result["Message"] = authResult.Message
	}else{
		result["ResultCode"] = authResult.ResultCode
		result["Message"] = authResult.Message
		return result
	}
	return result

}



func CallWincubeToken(codeId string) (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeToken \n")

	burl:= controller.WINCUBE_URL
	pURL:=burl+"/auth/token/issue"

	result := make(map[string]string)


	var reqData TokenReqData
	reqData.CodeId 	= codeId


	pbytes, _ := json.Marshal(reqData)
	buff := bytes.NewBuffer(pbytes)
	req, err := http.NewRequest("POST", pURL, buff)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()


	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["ResultCode"] = "99"
		result["Message"] = "인증 요청 실패"
		return result
	}

	var tokenResult TokenResult
	err = json.Unmarshal(respBody, &tokenResult)
	if err != nil {
		result["ResultCode"] = "99"
		result["Message"] = "인증 요청 실패"
		return result
	}

	if tokenResult.ResultCode=="200"{
		result["ResultCode"] = tokenResult.ResultCode
		result["TokenId"] = tokenResult.TokenId
		result["ExpireDate"] = tokenResult.ExpireDate
		result["ExpireTime"] = tokenResult.ExpireTime
		result["Message"] = tokenResult.Message
	}else{
		result["ResultCode"] = tokenResult.ResultCode
		result["Message"] = tokenResult.Message
		return result
	}
	return result

}


func CallWincubeItemCheck(tokenId,goodsId string) (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeItemCheck \n")

	burl:= controller.WINCUBE_MEDIA_URL
	mdcode:= controller.WINCUBE_MDCODE
	pURL:=burl+"/check_goods.do"

	result := make(map[string]string)

	var itemCheckResult ItemCheckResult

	uValue := url.Values{
		"mdcode": {mdcode},
		"goods_id": {goodsId},
		"token": {tokenId},
	}

	resp, err := http.PostForm(pURL,uValue)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	err = xml.Unmarshal(respBody, &itemCheckResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	if itemCheckResult.Result.Code=="0"{
		result["RESULT_CD"] = itemCheckResult.Result.Code
		result["RESULT_MSG"] =itemCheckResult.Result.Reason
		result["STOCK_CNT"] = itemCheckResult.Result.StockCnt
	}else{
		result["RESULT_CD"] = itemCheckResult.Result.Code
		result["RESULT_MSG"] =itemCheckResult.Result.Reason
		return result
	}

	return result

}


func CallWincubeCheckBalance(tokenId string) (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeCheckBalance \n")

	burl:= controller.WINCUBE_MEDIA_URL
	mdcode:= controller.WINCUBE_MDCODE
	pURL:=burl+"/check_balance.do"

	result := make(map[string]string)

	var checkBalance CheckBalance

	uValue := url.Values{
		"mdcode": {mdcode},
		"token": {tokenId},
	}

	resp, err := http.PostForm(pURL,uValue)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lprintf(4, "[INFO] CallWincubeCheckBalance  1 \n" , err.Error())
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	err = json.Unmarshal(respBody, &checkBalance)
	if err != nil {
		lprintf(4, "[INFO] CallWincubeCheckBalance 2 \n" , err.Error())
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(2)"
		return result
	}

	if checkBalance.Result_code=="0"{
		result["RESULT_CD"] = checkBalance.Result_code
		result["Available_balance"] =  strconv.Itoa(checkBalance.Available_balance)
	}else{
		result["RESULT_CD"] = checkBalance.Result_code
		//result["RESULT_MSG"] =checkBalance.Result.Reason
		return result
	}

	return result

}

func CallWincubeCancel(tokenId,order_no string) (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeCancel \n")

	burl:= controller.WINCUBE_MEDIA_URL
	mdcode:= controller.WINCUBE_MDCODE
	pURL:=burl+"/coupon_cancel.do"

	result := make(map[string]string)

	var cancelResult CancelResult

	uValue := url.Values{
		"mdcode": {mdcode},
		"tr_id": {order_no},
		"token": {tokenId},
	}

	resp, err := http.PostForm(pURL,uValue)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	err = xml.Unmarshal(respBody, &cancelResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	if cancelResult.Result.StatusCode=="0"{
		result["RESULT_CD"] = cancelResult.Result.StatusCode
		result["RESULT_MSG"] =cancelResult.Result.StatusText
		result["TR_ID"] = cancelResult.Result.TrID
	}else{
		result["RESULT_CD"] = cancelResult.Result.StatusCode
		result["RESULT_MSG"] =cancelResult.Result.StatusText
		return result
	}


	return result

}

func CallWincubeCouponStatus(tokenId,order_no string) (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeCouponStatus \n")

	burl:= controller.WINCUBE_MEDIA_URL
	mdcode:= controller.WINCUBE_MDCODE
	pURL:=burl+"/coupon_status.do"

	result := make(map[string]string)

	var couponStatusResult CouponStatusResult

	uValue := url.Values{
		"mdcode": {mdcode},
		"tr_id": {order_no},
		"token": {tokenId},
	}

	resp, err := http.PostForm(pURL,uValue)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	err = xml.Unmarshal(respBody, &couponStatusResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	result["RESULT_CD"] = couponStatusResult.Result.StatusCode
	result["RESULT_MSG"] =couponStatusResult.Result.StatusText
	result["EXCHPLC"] = ""  									//교환처
	result["EXCHCO_NM"] = ""										//교환처 - 매장명
	result["CPNO_EXCH_DT"] = couponStatusResult.Result.SwapDt 	// 교환일시
	result["CPNO_STATUS_CD"] = couponStatusResult.Result.StatusCode //쿠폰상태 코드
	result["BALANCE"] = ""



	return result

}

func CallWincubeOrder(tokenId,goods_id,order_no string) (map[string]string)  {

	lprintf(4, "[INFO] Call CallWincubeOrder \n")

	burl:= controller.WINCUBE_MEDIA_URL
	mdcode:= controller.WINCUBE_MDCODE
	pURL:=burl+"/request.do"

	result := make(map[string]string)

	var orderResult OrderResult
	//result := make(map[string]string)
	uValue := url.Values{
		"mdcode": {mdcode},
		"msg": {""},
		"title": {""},
		"callback": {""},
		"goods_id": {goods_id},
		"phone_no": {""},
		"tr_id": {order_no},
		"gubun": {"Y"},
		"token": {tokenId},
	}
	resp, err := http.PostForm(pURL,uValue)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	err = xml.Unmarshal(respBody, &orderResult)
	if err != nil {
		result["RESULT_CD"] = "99"
		result["RESULT_MSG"] = "요청 결과 반영 실패(1)"
		return result
	}

	if orderResult.Result.Code=="1000"{
		result["RESULT_CD"] = orderResult.Result.Code
		result["RESULT_MSG"] = orderResult.Result.Reason
		result["CAMP_ID"] = ""
		result["TR_ID"] = orderResult.Value.TrID
		result["ORD_NO"] = orderResult.Value.CtrID
		result["RCVR_MDN"] = ""
		result["CPNO"] = orderResult.Value.PinNo
		result["CPNO_SEQ"] = ""
		result["EXCH_FR_DY"] = orderResult.Value.CreateDateTime
		result["EXCH_TO_DY"] = orderResult.Value.ExpirationDate
	}else{
		result["RESULT_CD"] = orderResult.Result.Code
		result["RESULT_MSG"] =orderResult.Result.Reason
		return result
	}


	return result

}

func CallWincubeItemList(tokenId string) (string,ItemListResult)  {

	lprintf(4, "[INFO] Call CallWincubeItemList \n")



	burl:= controller.WINCUBE_MEDIA_URL
	mdcode:= controller.WINCUBE_MDCODE
	pURL:=burl+"/salelist.do"

	var itemResult ItemListResult
	//result := make(map[string]string)
	uValue := url.Values{
		"mdcode": {mdcode},
		"token": {tokenId},
	}
	resp, err := http.PostForm(pURL,uValue)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()


	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return "99",itemResult
	}


	err = xml.Unmarshal(respBody, &itemResult)
	if err != nil {
		return "99",itemResult
	}
	return "00",itemResult

}






func RsaEncrypt(input string) string {

	rsaPublicKey :="MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAjjBCNJgYRpAY/qOWFiwL5in17x1SD1vLd8y3d4OIzKT2RaBI/qanLVm+CuT/Kctt/RQYvtFydWETgUckPmbT+0t7v6ifRGu50w3oJgPawG44mYIali/PjaeCln+Le6RWVjlWsjClZPI4lnAodTOriQYud3RF73/3yr509jjCwCZW615/z11ZJ/hKwZTa4MFKLkncAvgzp8H9OLEIEKSCDkvjW42nNGFIvhTo9djPbEGJotm/hBZHJ52ehMhOWekBoymYMIFuhA38vgOLLJxRR4ov328Of6+s17M+JFegomBs8ztnmsRuNrXlLkb6bDS8P9CAXOx7Xrh5Q7Ntr7o8PwIDAQAB"
	tt,_:=PEMtoPublicKey(rsaPublicKey)

	encText, err := rsa.EncryptPKCS1v15(rand.Reader,tt , []byte(input))
	if err != nil {
		lprintf(4, "[INFO] RsaEncrypt error :  \n" , err)
	}
	encTextStr := base64.StdEncoding.EncodeToString(encText)
	return encTextStr
}


func Ase256(plaintext string, key string, iv string, blockSize int) string {
	bKey := []byte(key)
	bIV := []byte(iv)
	bPlaintext := PKCS5Padding([]byte(plaintext), blockSize, len(plaintext))
	block, _ := aes.NewCipher(bKey)
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, bIV)
	mode.CryptBlocks(ciphertext, bPlaintext)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}


func PEMtoPublicKey(pubPEM string) (*rsa.PublicKey, error) {
	tmpStr, err := base64.StdEncoding.DecodeString(pubPEM)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKIXPublicKey(tmpStr)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, nil
}



func Ase256Decode(cipherText string, encKey string, iv string) (decryptedString string) {
	bKey := []byte(encKey)
	bIV := []byte(iv)
	cipherTextDecoded, err := hex.DecodeString(cipherText)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(bKey)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCDecrypter(block, bIV)
	mode.CryptBlocks([]byte(cipherTextDecoded), []byte(cipherTextDecoded))
	return string(cipherTextDecoded)
}




