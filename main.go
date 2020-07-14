package main

import (
	"encoding/json"
	"github.com/go-echarts/go-echarts/charts"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type klineData struct {
	date string
	data [4]float32
}

type BiboxResponse struct {
	Result []struct {
		Time  int64  `json:"time"`
		Open  string `json:"open"`
		High  string `json:"high"`
		Low   string `json:"low"`
		Close string `json:"close"`
		Vol   string `json:"vol"`
	} `json:"result"`
	Cmd string `json:"cmd"`
	Ver string `json:"ver"`
}

// 或者使用 net/http，同上，后面也不就列出
func handler(w http.ResponseWriter, _ *http.Request) {

	response := parseBody(callBiboxApi())

	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([][4]float32, 0)
	for i := 0; i < len(response); i++ {
		x = append(x, response[i].date)
		y = append(y, response[i].data)
	}
	kline.AddXAxis(x).AddYAxis("kline", y)
	kline.SetGlobalOptions(
		charts.TitleOpts{Title: "BIBOX BTC"},
		charts.XAxisOpts{SplitNumber: 20},
		charts.YAxisOpts{Scale: true},
		charts.DataZoomOpts{Type:"inside", XAxisIndex: []int{0}, Start: 50, End: 100},
		charts.DataZoomOpts{Type:"slider", XAxisIndex: []int{0}, Start: 50, End: 100},
	)

	f, err := os.Create("kline.html")
	if err != nil {
		log.Println(err)
	}
	kline.Render(w, f) // Render 可接收多个 io.writer 接口
}

func callBiboxApi() []byte {
	// call api
	response, err := http.Get("https://api.bibox.com/v1/mdata?cmd=kline&pair=4BTC_USDT&period=day")
	if err != nil {
		log.Fatalln(err)
	}

	// decode body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return body
}

func parseBody(body []byte) []klineData {
	var kline klineData
	klines := make([]klineData, 0)

	var biboxResponse BiboxResponse
	err := json.Unmarshal(body, &biboxResponse)
	if err != nil {
		panic(err)
	}
	for _, currentKlineData := range biboxResponse.Result {
		// format date
		klineDate := strings.Split(time.Unix(currentKlineData.Time/1000, 0).String(), " ")
		formattedDate := strings.ReplaceAll(klineDate[0], "-", "/")

		// parse prices
		open, _ := strconv.ParseFloat(currentKlineData.Open, 32)
		high, _ := strconv.ParseFloat(currentKlineData.High, 32)
		low, _ := strconv.ParseFloat(currentKlineData.Low, 32)
		closePrice, _ := strconv.ParseFloat(currentKlineData.Close, 32)

		// create kline and append to slice
		kline.date = formattedDate
		kline.data = [4]float32{float32(open), float32(closePrice), float32(high), float32(low)}
		klines = append(klines, kline)
	}

	return klines
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
