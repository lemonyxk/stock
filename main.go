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
	"os"
	"path/filepath"
	"strconv"

	"github.com/jedib0t/go-pretty/text"
	"github.com/lemonyxk/utils/v3"
	"github.com/olekukonko/ts"
)

// 分钟
var minURL = `https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=`

// 日线
// sh000001,day,,,365,hfq
var dayURL = `https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?param=`

// 实时
var realURL = `https://web.sqt.gtimg.cn/q=`

const (
	// 分钟
	min = "min"
	// 日线
	day = "day"
)

var stopData = make(chan struct{})

var isDataRun = false

var mode = min

var runningProcess int32 = 0

var size, _ = ts.GetSize()
var termWidth, termHeight = size.Col(), size.Row()

var minWidth = text.RuneCount(`[Mode: day 365] [Q: Quit] [B: Back] [M: Min K] [D: Day K]`)
var minHeight = 6 + 3

var home = HomeDir()

var configPath = filepath.Join(home, ".stock", "config.json")

func main() {

	_ = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)

	if !utils.File.IsExist(configPath) {
		var err = utils.File.ReadFromBytes([]byte(`[{"Area":"sh","Code":"000001","Name":"上证指数"}]`)).WriteToPath(configPath)
		if err != nil {
			println(err.Error())
			return
		}
	}

	var res = utils.File.ReadFromPath(configPath)
	if res.LastError() != nil {
		println(res.LastError().Error())
		return
	}

	var err = utils.Json.Decode(res.Bytes(), &menu)
	if err != nil {
		println(err.Error())
		return
	}

	if len(menu) == 0 {
		println("No stock in config")
		return
	}

	var _, code = GetFlagAndArgs([]string{"code", "--code", "-c"}, os.Args[1:])
	if code == "" {
		code = menu[0].Code
	}

	var _, area = GetFlagAndArgs([]string{"area", "--area", "-a"}, os.Args[1:])
	if area == "" {
		area = menu[0].Area
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

	if code != menu[0].Code && area != menu[0].Area {
		menu = append([]config{{area, code, name}}, menu...)
	}

	// realData([]string{area + code})
	//
	// renderStockByCodeAndArea(area, code)

	selectMenu()

	select {}
}
