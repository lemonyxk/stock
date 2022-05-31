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

	"github.com/guptarohit/asciigraph"
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/console"
	"github.com/lemonyxk/kitty/v2/socket/http/client"
	"golang.org/x/crypto/ssh/terminal"
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

func main() {

	var _, code = GetFlagAndArgs([]string{"code", "--code", "-c"}, os.Args[1:])
	if code == "" {
		code = "000001"
	}

	var _, area = GetFlagAndArgs([]string{"area", "--area", "-a"}, os.Args[1:])
	if area == "" {
		area = "sh"
	}

	renderStockByCodeAndArea(area, code)

	select {}
}

var termWidth, termHeight, _ = terminal.GetSize(int(os.Stdin.Fd()))

func tips() {
	var sm = mode
	if mode == day {
		sm += " 365"
	}
	var str = "[Mode: " + sm + "] [Q:Quit] [L:List] [M:Change to min K] [D: Change to day K] \r\n"
	var s = strings.Repeat(" ", (termWidth-8-len(str))/2)
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

		graph := asciigraph.Plot(
			priceData,
			asciigraph.Width(termWidth-8),
			asciigraph.Height(termHeight-3),
			asciigraph.Caption(realData(area, code)),
		)

		flush()

		tips()

		write(graph)

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

		graph := asciigraph.Plot(
			priceData,
			asciigraph.Width(termWidth-8),
			asciigraph.Height(termHeight-3),
			asciigraph.Caption(realData(area, code)),
		)

		flush()

		tips()

		write(graph)

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

func realData(area, code string) string {

	var now = time.Now()

	var res = client.Get(realURL + area + code).Query().Send()
	var bts, _ = GbkToUtf8(res.Bytes())
	var str = string(bts)

	var sub = time.Now().Sub(now)

	var arr = strings.Split(str, "=")
	var data = arr[1]
	if len(data) < 2 {
		return ""
	}

	data = strings.Replace(data, "\"", "", -1)
	data = strings.Replace(data, ";", "", -1)
	var dataArr = strings.Split(data, "~")
	if len(dataArr) < 6 {
		return ""
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

	if percent > 0 {
		percentStr = console.FgRed.Sprintf("+%s +%.2f%%", absoluteChange, percent)
	} else {
		percentStr = console.FgGreen.Sprintf("%s %.2f%%", absoluteChange, percent)
	}

	var ts = fmt.Sprintf(
		"%s %s %s current: %s ( %s ) lowest: %s hightest: %s open: %s [ %s %.2fms ]",
		date, title, co, currentPrice, percentStr,
		lowestPrice, highestPrice, openPrice, now.Format("15:04:05"),
		float64(sub.Milliseconds())/1000.0*1000,
	)

	return ts
}
