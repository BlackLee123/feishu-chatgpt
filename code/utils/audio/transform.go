package audio

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
)

type Req struct {
	Common   CommonArgs   `json:"common"`
	Business BusinessArgs `json:"business"`
	Data     ReqData      `json:"data"`
}

type BusinessArgs struct {
	Aue string `json:"aue"`
	Sfl int    `json:"sfl"`
	Auf string `json:"auf"`
	Vcn string `json:"vcn"`
	Tte string `json:"tte"`
}

type ReqData struct {
	Status int    `json:"status"`
	Text   string `json:"text"`
}

type CommonArgs struct {
	AppId string `json:"app_id"`
}

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
	Data    Data   `json:"data"`
}

type Data struct {
	Audio  string `json:"audio"`
	Ced    string `json:"ced"`
	Status int    `json:"status"`
}

var hosturl = "wss://tts-api.xfyun.cn/v2/tts"

// var APPID = "18c9a5ec"
// var APISecret = "NzJiOTNlZjk2ZjExMjgxMDg0MWEwMGE2"
// var APIKey = "0333602a164a2e425f1d413722586925"
var vcn_en = []string{"x2_enus_catherine", "x2_engam_laura", "x2_engam_lindsay"}
var vcn_cn = []string{"xiaoyan", "aisjinger", "aisxping"}

func containsChinese(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return true
		}
	}
	return false
}

func assembleAuthUrl(hosturl string, apiKey, apiSecret string) string {
	ul, err := url.Parse(hosturl)
	if err != nil {
		fmt.Println(err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	// fmt.Println(sgin)
	//签名结果
	sha := HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	// fmt.Println(sha)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	callurl := hosturl + "?" + v.Encode()
	return callurl
}

func HmacWithShaTobase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

func TextToAudio(msgId, msgContent string, APPID, APISecret, APIKey string) error {
	index := rand.Intn(len(vcn_en))
	vcn := vcn_en[index]
	if containsChinese(msgContent) {
		index = rand.Intn(len(vcn_cn))
		vcn = vcn_cn[index]
	}
	wssurl := assembleAuthUrl(hosturl, APIKey, APISecret)
	c, _, err := websocket.DefaultDialer.Dial(wssurl, nil)
	if err != nil {
		return fmt.Errorf("dial: %v", err)
	}
	defer c.Close()
	var req = Req{
		Common: CommonArgs{
			AppId: APPID,
		},
		Business: BusinessArgs{
			Aue: "lame",
			Sfl: 1,
			Auf: "audio/L16;rate=16000",
			Vcn: vcn,
			Tte: "utf8",
		},
		Data: ReqData{
			Status: 2,
			Text:   base64.StdEncoding.EncodeToString([]byte(msgContent)),
		},
	}

	msg, _ := json.Marshal(req)
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("write:%v", err)
	}

	f, err := os.OpenFile(fmt.Sprintf("./%s.mp3", msgId), os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("read fail")
	}
	defer f.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %v", err)
		}
		v := &Resp{}
		json.Unmarshal(message, v)
		if v.Data.Status == 2 {
			data, _ := base64.StdEncoding.DecodeString(v.Data.Audio)
			f.Write(data)
			fmt.Println("ws is closed")
			return nil
		}
		if v.Code != 0 {
			return fmt.Errorf("sid:%s call error:%s code is:%d", v.Sid, v.Message, v.Code)
		} else {
			data, _ := base64.StdEncoding.DecodeString(v.Data.Audio)
			f.Write(data)
		}

	}
}
