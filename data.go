/**
* @program: stock
*
* @description:
*
* @author: lemo
*
* @create: 2022-06-01 19:26
**/

package main

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/console"
	"github.com/lemonyxk/kitty/v2/socket/http/client"
)

var timeData []string
var priceData []float64

func getMinData(area, code string) {

	timeData = timeData[:0]
	priceData = priceData[:0]

	// var m = make(map[float64]float64)

	var res = client.Get(minURL + area + code).Query().Send()

	var arrStr = jsoniter.Get(res.Bytes(), "data", area+code, "data", "data").ToString()

	var arr []string

	_ = jsoniter.Unmarshal([]byte(arrStr), &arr)

	for _, v := range arr {
		var a = strings.Split(v, " ")
		timeData = append(timeData, a[0])
		priceData = append(priceData, StringToFloat(a[1]))
	}

	// the lib is not support the nil value
	if len(timeData) == 0 || len(priceData) == 0 {
		timeData = append(timeData, "0930")
		priceData = append(priceData, 0)
	}

}

func getDayData(area, code string) {

	timeData = timeData[:0]
	priceData = priceData[:0]

	// var m = make(map[float64]float64)

	var res = client.Get(dayURL + area + code + `,day,,,365,`).Query().Send()

	var arrStr = jsoniter.Get(res.Bytes(), "data", area+code, "day").ToString()

	var arr [][]string

	_ = jsoniter.Unmarshal([]byte(arrStr), &arr)

	for _, dayArr := range arr {
		var t = dayArr[0][0:2] + ":" + dayArr[0][2:]
		timeData = append(timeData, t)
		priceData = append(priceData, StringToFloat(dayArr[2]))
	}

	// the lib is not support the nil value
	if len(timeData) == 0 || len(priceData) == 0 {
		timeData = append(timeData, time.Now().AddDate(0, 0, -365).Format("2006-01-02"))
		priceData = append(priceData, 0)
	}

}

var now = time.Now()

func realData(params []string) [][]string {

	now = time.Now()

	var res = client.Get(realURL + strings.Join(params, ",")).Query().Send()
	var bts, _ = GbkToUtf8(res.Bytes())
	var str = string(bts)

	var resData [][]string

	var arr = strings.Split(str, ";")

	for i := 0; i < len(arr); i++ {
		if strings.TrimSpace(arr[i]) == "" {
			continue
		}

		var as = strings.Split(arr[i], "=")
		var data = as[1]
		if len(data) < 2 {
			return resData
		}

		data = strings.Replace(data, "\"", "", -1)
		var dataArr = strings.Split(data, "~")
		if len(dataArr) < 35 {
			return resData
		}

		resData = append(resData, dataArr)

	}

	return resData
}

func renderRealData(data [][]string) []string {

	var str []string

	for i := 0; i < len(data); i++ {

		var dataArr = data[i]

		var sub = time.Now().Sub(now)

		var title = dataArr[1]
		var co = dataArr[2]
		var currentPrice = dataArr[3]
		// var startPrice = dataArr[4]
		var openPrice = dataArr[5]
		var date = strings.Replace(dataArr[30], "-", "", -1)[:8]
		date = date[:4] + "-" + date[4:6] + "-" + date[6:]
		var absoluteChange = dataArr[31]
		var percentChange = dataArr[32]
		var highestPrice = dataArr[33]
		var lowestPrice = dataArr[34]

		// var percent = (StringToFloat(currentPrice) - StringToFloat(startPrice)) / StringToFloat(startPrice) * 100
		var percent = StringToFloat(percentChange)

		var percentStr string

		if percent >= 0 {
			currentPrice = console.FgRed.Sprintf("↑ %s", currentPrice)
			percentStr = console.FgRed.Sprintf("+%s +%.2f%%", absoluteChange, percent)
		} else {
			currentPrice = console.FgGreen.Sprintf("↓ %s", currentPrice)
			percentStr = console.FgGreen.Sprintf("%s %.2f%%", absoluteChange, percent)
		}

		highestPrice = console.FgHiRed.Sprintf("%s", highestPrice)
		lowestPrice = console.FgHiGreen.Sprintf("%s", lowestPrice)

		var st = fmt.Sprintf(
			"%s %s %s N: %s ( %s ) L: %s H: %s O: %s T: %s %.fms",
			date, title, co, currentPrice, percentStr,
			lowestPrice, highestPrice, openPrice, now.Format("15:04:05"),
			float64(sub.Milliseconds())/1000.0*1000,
		)

		if utf8.RuneCountInString(st) > termWidth {
			st = fmt.Sprintf(
				"%s %s %s N: %s ( %s )",
				date, title, co, currentPrice, percentStr,
			)
		}

		str = append(str, st)
	}

	return str
}
