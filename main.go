/**
* @program: stock
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-31 14:39
**/
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/guptarohit/asciigraph"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/console"
	"github.com/lemonyxk/kitty/v2/socket/http/client"
	"github.com/olekukonko/ts"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// 分钟
var minURL = `https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=`

// 日线
// sh000001,day,,,365,hfq
var dayURL = `https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?param=`

// 实时
var realURL = `https://web.sqt.gtimg.cn/q=`

func GetFlagAndArgs(flag []string, args []string) (string, string) {
	for i := 0; i < len(args); i++ {
		for j := 0; j < len(flag); j++ {
			if args[i] == flag[j] {
				if i+1 < len(args) {
					return flag[j], args[i+1]
				}
			}
		}
	}
	return "", ""
}

const (
	// 分钟
	min = "min"
	// 日线
	day = "day"
)

var stop = make(chan struct{})

var mode = min

var size, _ = ts.GetSize()
var termWidth, termHeight = size.Col(), size.Row()

var minWidth = utf8.RuneCountInString(`[Mode: day 365] [Q: Quit] [L: List] [M: Min K] [D: Day K]`) + 8
var minHeight = 6 + 3

func main() {

	var _, code = GetFlagAndArgs([]string{"code", "--code", "-c"}, os.Args[1:])
	if code == "" {
		code = "000001"
	}

	var _, area = GetFlagAndArgs([]string{"area", "--area", "-a"}, os.Args[1:])
	if area == "" {
		area = "sh"
	}

	// width
	var _, w = GetFlagAndArgs([]string{"width", "--width", "-w"}, os.Args[1:])
	var width, _ = strconv.Atoi(w)
	if width > 0 && width < 100 {
		termWidth = termWidth * width / 100
		if termWidth < minWidth {
			termWidth = minWidth
		}
	}

	// height
	var _, h = GetFlagAndArgs([]string{"height", "--height", "-h"}, os.Args[1:])
	var height, _ = strconv.Atoi(h)
	if height > 0 && height < 100 {
		termHeight = termHeight * height / 100
		if termHeight < minHeight {
			termHeight = minHeight
		}
	}

	var _, name = GetFlagAndArgs([]string{"name", "--name", "-n"}, os.Args[1:])

	if code != "000001" && area != "sh" {
		menu = append([]config{{area, code, name, "", ""}}, menu...)
	}

	renderStockByCodeAndArea(area, code)

	select {}
}

func tips() {
	var sm = mode
	if mode == day {
		sm += " 365"
	}
	var str = "[Mode: " + sm + "] [Q: Quit] [L: List] [M: Min K] [D: Day K] \r\n"
	var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(str))/2)
	write(console.FgYellow.Sprint(s + str))
}

func renderStockByCodeAndArea(area, code string) {
	if mode == min {
		minRender(area, code)
	} else {
		dayRender(area, code)
	}
}

func dayRender(area, code string) {
	var index = 0

	var fn = func() {

		if index%60 == 0 {
			getDayData(area, code)
		}

		var realStr, _, _ = realData(area, code)

		graph := asciigraph.Plot(
			priceData,
			asciigraph.Width(termWidth-8),
			asciigraph.Height(termHeight-3),
			// asciigraph.Caption(),
		)

		flush()

		tips()

		write(graph)

		var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(realStr))/2)

		write("\n" + s + realStr)

		index++
	}

	fn()

	var timer = time.NewTicker(time.Second * 3)

	go func() {

		for {
			select {
			case <-timer.C:
				fn()
			case <-stop:
				timer.Stop()
				return
			}
		}
	}()
}

func minRender(area, code string) {
	var index = 0

	var fn = func() {

		if index%60 == 0 {
			getMinData(area, code)
		}

		var realStr, _, _ = realData(area, code)

		graph := asciigraph.Plot(
			priceData,
			asciigraph.Width(termWidth-8),
			asciigraph.Height(termHeight-3),
			// asciigraph.Caption(realStr),
		)

		if isSelectMenu {
			return
		}

		flush()

		tips()

		write(graph)

		var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(realStr))/2)

		write("\n" + s + realStr)

		index++
	}

	go fn()

	var timer = time.NewTicker(time.Second * 3)

	go func() {

		for {
			select {
			case <-timer.C:
				go fn()
			case <-stop:
				return
			}
		}
	}()

}

var timeData []float64
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
		timeData = append(timeData, StringToFloat(a[0]))
		priceData = append(priceData, StringToFloat(a[1]))
	}

	// the lib is not support the nil value
	if len(timeData) == 0 || len(priceData) == 0 {
		timeData = append(timeData, 930)
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
		timeData = append(timeData, StringToFloat(dayArr[0]))
		priceData = append(priceData, StringToFloat(dayArr[2]))
	}

	// the lib is not support the nil value
	if len(timeData) == 0 || len(priceData) == 0 {
		timeData = append(timeData, 930)
		priceData = append(priceData, 0)
	}

}

func StringToFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return f
}

func realData(area, code string) (string, string, float64) {

	var now = time.Now()

	var res = client.Get(realURL + area + code).Query().Send()
	var bts, _ = GbkToUtf8(res.Bytes())
	var str = string(bts)

	var sub = time.Now().Sub(now)

	var arr = strings.Split(str, "=")
	var data = arr[1]
	if len(data) < 2 {
		return "", "", 0
	}

	data = strings.Replace(data, "\"", "", -1)
	data = strings.Replace(data, ";", "", -1)
	var dataArr = strings.Split(data, "~")
	if len(dataArr) < 6 {
		return "", "", 0
	}

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
		percentStr = console.FgRed.Sprintf("+%s +%.2f%%", absoluteChange, percent)
	} else {
		percentStr = console.FgGreen.Sprintf("%s %.2f%%", absoluteChange, percent)
	}

	var st = fmt.Sprintf(
		"%s %s %s N: %s ( %s ) L: %s H: %s O: %s [%s %.2fms]",
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

	return st, absoluteChange, percent
}
