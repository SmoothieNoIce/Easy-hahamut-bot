package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	//https://haha.gamer.com.tw/bot_detail.php?id=<YOUR-BOTID>
	APP_SECRET     = "<YOUR_APP_SECRET>"
	ACCESS_TOKEN   = "<YOUR_ACCESS_TOKEN>"
	POST_URL       = "https://us-central1-hahamut-8888.cloudfunctions.net/messagePush?access_token="
	POST_IMAGE_URL = "https://us-central1-hahamut-8888.cloudfunctions.net/ImgMessagePush?access_token="
)

type Text struct {
	Text string `json:"text"`
}

type Messaging struct {
	SenderID string `json:"sender_id"`
	Message  Text   `json:"message"`
}

type Message struct {
	BotID     string      `json:"botid"`
	Time      int         `json:"time"`
	Messaging []Messaging `json:"messaging"`
}

type Recipient struct {
	ID string `json:"id"`
}

type MessageingSend struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MessageSend struct {
	MyRecipient      Recipient      `json:"recipient"`
	MyMessageingSend MessageingSend `json:"message"`
}

type StickerSend struct {
	Type         string `json:"type"`
	StickerGroup string `json:"sticker_group"`
	StickerID    string `json:"sticker_id"`
}

type StickerMessageSend struct {
	MyRecipient Recipient   `json:"recipient"`
	StickerSend StickerSend `json:"message"`
}

type VerifyError struct {
	When time.Time
	What string
}

func (e *VerifyError) Error() string {
	fmt.Println("Error:", e.What)
	return fmt.Sprintf("at %v, %s", e.When, e.What)
}

/*
http.HandleFunc的handler
*/
func handler(w http.ResponseWriter, r *http.Request) {

	// Do: 結束後執行 關閉httpRequest
	defer r.Body.Close()

	// Do: Read body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Do: Decode body 的東西
	var msg Message
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Do: 傳回200
	request200(w)

	// Do: 驗證簽章
	err = verifyWebhook(r, b)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Do: 剩下靠自己
	fmt.Println("bot_id=", msg.BotID)
	fmt.Println("time=", msg.Time)
	fmt.Println("senderid=", msg.Messaging[0].SenderID)
	fmt.Println("text=", msg.Messaging[0].Message.Text)

	handleMessage(w, msg.Messaging[0].Message.Text, msg.Messaging[0].SenderID)

}

/*
處理使用者傳入的訊息
*/
func handleMessage(w http.ResponseWriter, msg string, senderid string) {

	switch msg {

	case "打招呼":
		sendMsg(w, senderid, "嗨嗨OwO")

	case "發貼圖":
		sendSticker(w, senderid, "28", "09")

	case "發圖片":
		sendImg(w, senderid, "src/img/ralsei.png")

	default:
		sendMsg(w, senderid, "我聽不懂==")

	}

}

/*
發送文字訊息
*/
func sendMsg(w http.ResponseWriter, senderID string, msg string) {

	// 包成json
	r := Recipient{senderID}
	m := MessageingSend{"text", msg}
	ms := MessageSend{r, m}
	json, err := json.Marshal(ms)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 請求Post
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", POST_URL, ACCESS_TOKEN), bytes.NewBuffer(json))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 設置header
	req.Header.Set("X-Custom-Header", "badfox-sender")
	req.Header.Set("Content-Type", "application/json")

	// For control over HTTP client headers
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	//取得哈哈的伺服器取得狀態
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Headers:", resp.Header)
	fmt.Println("Response Body:", string(body))

}

func sendSticker(w http.ResponseWriter, senderID string, stickerGroup string, stickerID string) {

	r := Recipient{senderID}
	s := StickerSend{"sticker", stickerGroup, stickerID}
	sms := StickerMessageSend{r, s}
	json, err := json.Marshal(sms)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 請求Post
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", POST_URL, ACCESS_TOKEN), bytes.NewBuffer(json))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 設置header
	req.Header.Set("X-Custom-Header", "badfox-sender")
	req.Header.Set("Content-Type", "application/json")

	// For control over HTTP client headers
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	//取得哈哈的伺服器取得狀態
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Headers:", resp.Header)
	fmt.Println("Response Body:", string(body))

}

func sendImg(w http.ResponseWriter, senderID string, img string) {

	fullPath, _ := os.Getwd()
	fullPath += "/" + img
	fmt.Println("fullPath:" + fullPath)

	extraParams := map[string]string{
		"message":   `{"type":"img"}`,
		"recipient": fmt.Sprintf(`{"id":"%s"}`, senderID),
	}

	req, err := newfileUploadRequest(fmt.Sprintf("%s%s", POST_IMAGE_URL, ACCESS_TOKEN), extraParams, "filedata", img)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// For control over HTTP client headers
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		//取得哈哈的伺服器取得狀態
		fmt.Println("Response Status:", resp.Status)
		fmt.Println("Response Headers:", resp.Header)
		fmt.Println("Response Body:", string(body))
		resp.Body.Close()
	}

}

/*
	-需要使用multipart/form-data格式的表單
	-不能用golang默認的http.POST
	-自行實現body參數邏輯
*/
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {

	//打開檔案
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close() //函式執行完關閉檔案

	body := &bytes.Buffer{}                                            //建立一個容納byte的緩衝器
	writer := multipart.NewWriter(body)                                //建立一個新的writter
	part, err := writer.CreateFormFile(paramName, filepath.Base(path)) //建立一個form 參數名稱是"filedata" ///檔案名稱是path
	if err != nil {
		return nil, err
	}

	//把file的數據複製到建立的form
	_, err = io.Copy(part, file)
	//key=i,val等於params參數map的數值
	//寫入 message="type":"img"
	//寫入"recipient"="senderID"
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	//把自己編好的body送出去
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func request200(w http.ResponseWriter) {
	w.WriteHeader(200)
}

func verifyWebhook(r *http.Request, body []byte) error {

	//Note: 取得x-baha-data-signature開頭的Header
	headerget := r.Header.Get("x-baha-data-signature")
	key := []byte(headerget[5:])
	fmt.Println("key=", key)

	//Note: 使用HMAC-SHA1雜湊APP_SECRET和傳來的Body
	mac := hmac.New(sha1.New, []byte(APP_SECRET))
	mac.Write(body)
	fmt.Printf("%x\n", mac.Sum(nil))

	//轉為字串比較
	str1 := string(key[:])
	str2 := fmt.Sprintf("%x", mac.Sum(nil))

	fmt.Println("str1=", str1)
	fmt.Println("str2=", str2)

	if str1 != str2 {
		return &VerifyError{
			time.Now(),
			"Verify error",
		}
	} else {
		return nil
	}

}

func main() {

	//開啟伺服器
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	address := ":" + port
	log.Println("Starting server on address", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}

}
