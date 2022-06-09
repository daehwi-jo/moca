package tpays

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"math"
	paymentsql "mocaApi/query/payments"
	"mocaApi/src/controller"
	"mocaApi/src/controller/cls"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var dprintf func(int, echo.Context, string, ...interface{}) = cls.Dprintf
var lprintf func(int, string, ...interface{}) = cls.Lprintf

type Tpay struct {
	MerchantKey string
	EncKey      string
	EdiDate     string
	Mid         string
}

// tpay decrypt
func AesHandlerDecrypt(input, ediDate, merchantKey string) string {

	key := fmt.Sprintf("%s%s", ediDate, merchantKey)
	encKey := getMD5HashHandler(key)
	decode, _ := base64.StdEncoding.DecodeString(input)

	result := AesDecrypt(byteTostring(decode), encKey, merchantKey)
	fmt.Println("decrypt : ", string(result))

	return string(result)
}

func byteTostring(buf []byte) string {

	var tmp string

	for i := 0; i < len(buf); i++ {
		if buf[i]&0xff < 0x10 {
			tmp += "0"
		}

		tmp += strconv.FormatInt(int64(buf[i]&0xff), 16)
	}

	return tmp
}

func AesDecrypt(input, key, merchantKey string) []byte {

	ips := []byte(merchantKey[:16])
	keyBytes := hexToByteArray(key)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		fmt.Println("key error1", err)
		return nil
	}

	blockMode := cipher.NewCBCDecrypter(block, ips)

	crypted := make([]byte, len(input))
	blockMode.CryptBlocks(crypted, hexToByteArray(input))

	return PKCS5Trimming(crypted)
}

func PKCS5Trimming(encrypt []byte) []byte {

	var index int

	for idx, v := range encrypt {
		if v == uint8(10) {
			index = idx
			break
		}
	}

	//padding := encrypt[len(encrypt)-1]
	//return encrypt[:len(encrypt)-int(padding)]

	return encrypt[:index]
}

// tpay encrypt
func AesHandlerEncrypt(input, key, merchantKey string) []byte {

	inbytes := []byte(merchantKey[:16])

	return AesEncrypt([]byte(input), hexToByteArray(key), inbytes)
}

func hexToByteArray(hex string) []byte {

	bb := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		pi, err := strconv.ParseInt(hex[i:i+2], 16, 16)
		if err != nil {
			fmt.Println(err.Error())
		}
		bb[int(math.Floor(float64(i/2)))] = uint8(pi)
	}

	return bb
}

func AesEncrypt(origData, key, IV []byte) []byte {

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("key error1", err)
		return nil
	}

	//fmt.Println("blockSize : ", block.BlockSize())

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCEncrypter(block, IV[:blockSize])

	origData = PKCS5Padding(origData, blockSize)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)

	return crypted
}

func getMD5HashHandler(key string) string {

	hash := md5.Sum([]byte(key))
	hash2 := md5.Sum([]byte(key))
	var hexString string

	//fmt.Println(text, "  hash len : ", len(hash))

	for i := 0; i < len(hash); i++ {
		var tmp []byte

		//fmt.Println("hash i : ", hash[i], " 0xFF : ", 0xFF&hash[i])

		tmp = append(tmp, uint8(hash[i]&0xFF))
		hexEncode := hex.EncodeToString(tmp)
		if len(hexEncode) == 1 {
			hexEncode = "0" + hexEncode
		}

		hexString += hexEncode
	}

	//fmt.Println("hex    ", hex.EncodeToString(hash2[:]))
	//fmt.Println("encKey  ", hexString, " len ", len(hexString))

	if hex.EncodeToString(hash2[:]) == hexString {
		//fmt.Println("hex true")
	} else {
		//fmt.Println("hex false")
	}

	return hexString
}

func getEdiDate() (string, string) {
	now := time.Now()

	bankDate := now.AddDate(0, 0, 1).Format("20060102")

	return now.Format("20060102150405"), bankDate
}

func getMoid(userId string) string {

	now := time.Now()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	return fmt.Sprintf("%d%s", millis, userId)
}

func (t *Tpay) TEncryptor(merchantKey string, ediDate string) string {
	s := t.EdiDate + t.MerchantKey

	dataArr := md5.Sum([]byte(s))
	data := dataArr[:]
	fmt.Println(data)
	//output := make([]byte, len(data)/4)
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {

		temp := strconv.FormatInt(int64(data[i]), 16)
		if len(temp) == 1 {
			temp = "0" + temp
		}
		//	result[i]= fmt.Sprintf("%x", md5.Sum([]byte(s)))

	}

	return string(result)
}

func (t *Tpay) Encryptor(merchantKey string, ediDate string) string {
	s := t.EdiDate + t.MerchantKey

	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func (t *Tpay) EncData(input string) string {

	encryptedData := AESEncrypt(input, []byte(t.EncKey), t.MerchantKey)
	encryptedString := base64.StdEncoding.EncodeToString(encryptedData)
	return encryptedString
}

func AESEncrypt(src string, key []byte, merchantKey string) []byte {

	initialVector := merchantKey[0:16]
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("key error1", err)
	}
	if src == "" {
		fmt.Println("plain content empty")
	}
	ecb := cipher.NewCBCEncrypter(block, []byte(initialVector))
	content := []byte(src)
	content = PKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	return crypted
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

type TpayRecvData struct {
	Result_cd  string `json:"result_cd"`  //
	Result_msg string `json:"result_msg"` //
	PayMethod  string `json:"PayMethod"`  //
	CancelDate string `json:"CancelDate"` //
	CancelTime string `json:"CancelTime"` //
	Tid        string `json:"tid"`        //
	Moid       string `json:"moid"`       //
}

// 티페이 취소
func SetTpayCancel(c echo.Context) error {

	params := cls.GetParamJsonMap(c)

	moid := params["moid"]
	cancelAmt := params["cancelAmt"]
	cancelMsg := "사용자취소"
	partialCancelCode := params["partialCancelCode"]

	//println(moid,partialCancelCode)

	///////// 취소 가능 체크 로직 넣어야함

	paymentInfo, err := cls.GetSelectDataRequire(paymentsql.SelectPaymentHist, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if paymentInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "결제 정보가 없습니다."))
	}

	params["restId"] = paymentInfo[0]["REST_ID"]
	params["grpId"] = paymentInfo[0]["GRP_ID"]
	payTy := paymentInfo[0]["SEARCH_TY"]
	regDate := paymentInfo[0]["REG_DATE"]
	toDate := paymentInfo[0]["TO_DATE"]

	if regDate != toDate {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "당일 이외의 취소는 가맹점과 통화 승인이 필요합니다."))
	}

	creditAmt, _ := strconv.Atoi(paymentInfo[0]["CREDIT_AMT"])
	chkcancelAmt, _ := strconv.Atoi(cancelAmt)

	if chkcancelAmt != creditAmt {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "취소금액과 결제 금액이 다릅니다."))
	}

	tid := paymentInfo[0]["TID"]
	params["tid"] = tid

	linkInfo, err := cls.GetSelectDataRequire(paymentsql.SelctLinkInfo, params, c)
	if err != nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
	}
	if linkInfo == nil {
		return c.JSON(http.StatusOK, controller.SetErrResult("99", "협약 정보가 없습니다."))
	}
	prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])

	if payTy == "1" {
		if chkcancelAmt > prepaidAmt {
			return c.JSON(http.StatusOK, controller.SetErrResult("99", "충전금액을 사용하신경우에는 취소가 불가능합니다."))
		}
	}

	// Tpay 취소 요청
	// 결제취소 결과가 2001(카드 - 취소성공), 2211(계좌이체 - 환불성공) 성공이 아닐경우
	tpay := PayCancel(moid, cancelAmt, cancelMsg, partialCancelCode, tid)

	tpayResultCd := tpay["result_cd"]
	tpayResultMsg := tpay["result_msg"]
	//PayMethod := tpay["PayMethod"]
	CancelDate := tpay["CancelDate"]
	CancelTime := tpay["CancelTime"]
	canceltid := tpay["tid"]

	if tpayResultCd == "2001" {

		params["resultcd"] = tpayResultCd
		params["resultmsg"] = tpayResultMsg
		params["canceldate"] = CancelDate
		params["canceltime"] = CancelTime
		params["statecd"] = "0"
		params["cancelamt"] = cancelAmt
		params["cancelmsg"] = cancelMsg
		params["canceltid"] = canceltid

		// 결제 취소  TRNAN 시작
		tx, err := cls.DBc.Begin()
		if err != nil {
			//return "5100", errors.New("begin error")
		}

		txErr := err

		// 오류 처리
		defer func() {
			if txErr != nil {
				// transaction rollback
				dprintf(4, c, "do rollback -결제 취소(SetTpayCancel)  \n")
				tx.Rollback()
			}
		}()

		UpdatePaymentCancelQuery, err := cls.SetUpdateParam(paymentsql.UpdatePaymentCancel, params)
		if err != nil {
			txErr = err
			return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
		}
		_, err = tx.Exec(UpdatePaymentCancelQuery)
		if err != nil {
			txErr = err
			dprintf(1, c, "Query(%s) -> error (%s) \n", UpdatePaymentCancelQuery, err)
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

		params["userId"] = paymentInfo[0]["USER_ID"]
		params["creditAmt"] = cancelAmt
		params["searchTy"] = paymentInfo[0]["SEARCH_TY"]
		params["payInfo"] = paymentInfo[0]["PAY_INFO"]
		params["payChannel"] = paymentInfo[0]["PAY_CHANNEL"]
		params["userTy"] = "0"
		params["histId"] = "0"

		//선불 취소 처리
		if payTy == "1" {

			prepaidAmt, _ := strconv.Atoi(linkInfo[0]["PREPAID_AMT"])
			prepaidPoint, _ := strconv.Atoi(linkInfo[0]["PREPAID_POINT"])

			addAmt, _ := strconv.Atoi(paymentInfo[0]["ADD_AMT"])
			TcancelAmt, _ := strconv.Atoi(cancelAmt)
			params["prepaidAmt"] = strconv.Itoa(prepaidAmt - TcancelAmt - addAmt)
			params["prepaidPoint"] = strconv.Itoa(prepaidPoint - addAmt)
			params["agrmId"] = linkInfo[0]["AGRM_ID"]
			params["addAmt"] = paymentInfo[0]["ADD_AMT"]

			// ORG_AGRM_INFO 테이블(협약 정보)에 수정일시, 선불금액 update
			UpdateAgrmQuery, err := cls.GetQueryJson(paymentsql.UpdateAgrm, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UpdateAgrmQuery)
			if err != nil {
				txErr = err
				dprintf(1, c, "Query(%s) -> error (%s) \n", UpdateAgrmQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

			// 선금 결제 정보 테이블에 환불 정보 insert
			params["jobTy"] = "1" // 환불

			InsertPrepaidQuery, err := cls.GetQueryJson(paymentsql.InsertPrepaid, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(InsertPrepaidQuery)
			if err != nil {
				txErr = err
				dprintf(1, c, "Query(%s) -> error (%s) \n", InsertPrepaidQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

			params["paymentTy"] = "1" // 선불 취소
			InsertPaymentHistoryQuery, err := cls.GetQueryJson(paymentsql.InsertPaymentHistory, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(InsertPaymentHistoryQuery)
			if err != nil {
				txErr = err
				dprintf(1, c, "Query(%s) -> error (%s) \n", InsertPaymentHistoryQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

		} else if payTy == "2" {

			// DAR_ORDER_INFO에 있는 기존 주문 정보 update
			// 결제 여부(PAID_YN)를 N으로 PAID_DATE 를 null
			UpdatePostPayQuery, err := cls.GetQueryJson(paymentsql.UpdatePostPay, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(UpdatePostPayQuery)
			if err != nil {
				txErr = err
				dprintf(1, c, "Query(%s) -> error (%s) \n", UpdatePostPayQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

			params["paymentTy"] = "4" // 후불 취소
			params["addAmt"] = "0"
			InsertPaymentHistoryQuery, err := cls.GetQueryJson(paymentsql.InsertPaymentHistory, params)
			if err != nil {
				txErr = err
				return c.JSON(http.StatusInternalServerError, controller.SetErrResult("98", err.Error()))
			}
			_, err = tx.Exec(InsertPaymentHistoryQuery)
			if err != nil {
				txErr = err
				dprintf(1, c, "Query(%s) -> error (%s) \n", InsertPaymentHistoryQuery, err)
				return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
			}

		}

		// transaction commit
		err = tx.Commit()
		if err != nil {
			txErr = err
			return c.JSON(http.StatusOK, controller.SetErrResult("98", err.Error()))
		}

	} else {

		m := make(map[string]interface{})

		m["resultCode"] = "01"
		m["resultMsg"] = tpayResultMsg
		return c.JSON(http.StatusOK, m)

	}

	// 성공 시 Tpay 로 결과 전송

	//ResultConfirm(canceltid,"000")
	sendCancelResultConfirm(canceltid, "000")

	//

	m := make(map[string]interface{})
	m["resultCode"] = "00"
	m["resultMsg"] = tpayResultMsg
	return c.JSON(http.StatusOK, m)

}

func sendCancelResultConfirm(tid string, result string) {

	//fname := cls.Cls_conf(os.Args)
	fname := cls.Fname
	mid, _ := cls.GetTokenValue("TPAY.TPAY_MID", fname)
	apikey, _ := cls.GetTokenValue("TPAY.TPAY_MERCHANT_KEY", fname)

	payload := strings.NewReader("api_key=" + apikey + "&mid=" + mid + "&tid=" + tid + "&result_code=" + result + "&result_cd=" + result + "&result=" + result)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", "https://webtx.tpay.co.kr/api/v1/result_confirm", payload)
	if err != nil {
		lprintf(1, "[ERROR] Tpay Cancel resultConfirm Send1 : %s\n", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		lprintf(1, "[ERROR] Tpay Cancel resultConfirm Send2 : %s\n", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		lprintf(1, "[ERROR] Tpay Cancel resultConfirm Send3 : %s\n", err)
	}
	lprintf(4, "[INFO] Tpay Cancel resultConfirm resp(%s)\n", string(body))
	println(string(body))

}

func PayCancel(moid string, cancelAmt string, cancelMsg string, partialCancelCode string, tid string) map[string]string {

	lprintf(4, "call 결제 취소요청  \n")
	lprintf(4, "moid:  \n", moid)
	lprintf(4, "tid:  \n", tid)

	TpayMap := make(map[string]string)

	//fname := cls.Cls_conf(os.Args)
	fname := cls.Fname
	mid, _ := cls.GetTokenValue("TPAY.TPAY_MID", fname)
	merchantKey, _ := cls.GetTokenValue("TPAY.TPAY_MERCHANT_KEY", fname)
	cancelPw, _ := cls.GetTokenValue("TPAY.TPAY_CANCEL_PW", fname)
	restfulCancelUrl, _ := cls.GetTokenValue("TPAY.RESTFUL_CANCEL_URL", fname)

	uValue := url.Values{
		"api_key":        {merchantKey},
		"mid":            {mid},
		"moid":           {moid},
		"cancel_pw":      {cancelPw},
		"cancel_amt":     {cancelAmt},
		"cancel_msg":     {cancelMsg},
		"partial_cancel": {partialCancelCode},
		"tid":            {tid},
	}

	resp, err := http.PostForm(restfulCancelUrl, uValue)
	if err != nil {
		lprintf(4, "결제 취소 요청 오류 :  ", err)
		panic(err)
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		TpayMap["result_cd"] = "9999"
		TpayMap["result_msg"] = "전송 요청 오류"
		lprintf(4, "결제 취소 요청 오류 : ", err)
		return TpayMap
	}

	var result TpayRecvData
	err = json.Unmarshal(respBody, &result)
	if err != nil {

		TpayMap["result_cd"] = "9999"
		TpayMap["result_msg"] = "전송 요청 오류"
		return TpayMap
	}

	TpayMap["result_cd"] = result.Result_cd
	TpayMap["result_msg"] = result.Result_msg
	TpayMap["PayMethod"] = result.PayMethod
	TpayMap["CancelDate"] = result.CancelDate
	TpayMap["CancelTime"] = result.CancelTime
	TpayMap["tid"] = result.Tid
	TpayMap["moid"] = result.Moid

	return TpayMap
}

func ResultConfirm(tid string, result string) {

	lprintf(4, "call Tpay 확인  \n")
	lprintf(4, "tid:  \n", tid)

	//fname := cls.Cls_conf(os.Args)
	fname := cls.Fname
	merchantKey, _ := cls.GetTokenValue("TPAY.TPAY_MERCHANT_KEY", fname)

	uValue := url.Values{
		"api_key": {merchantKey},
		"tid":     {tid},
		"result":  {result},
	}
	resp, err := http.PostForm("https://webtx.tpay.co.kr/resultConfirm", uValue)
	if err != nil {
		lprintf(4, "tpay 확인  요청 오류 :  ", err)
		panic(err)
	}
	defer resp.Body.Close()
	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lprintf(4, "tpay 확인  요청 오류 : ", err)
	}
	lprintf(4, "tpay 확인  : ", string(respBody))

}
