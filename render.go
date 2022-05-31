/**
* @program: stock
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-31 21:48
**/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eiannone/keyboard"
)

var Stdout = os.Stdout

func init() {

	go func() {

		if err := keyboard.Open(); err != nil {
			panic(err)
		}

		defer func() { _ = keyboard.Close() }()

		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				panic(err)
			}

			switch key {
			case 0xffea:
				ctrlC(key)
			case 0xffeb:
				ctrlC(key)
			case 0xffec:
				ctrlC(key)
			case 0xffed:
				ctrlC(key)
			case 0x000d:
				ctrlC(key)
			default:
				switch char {
				case 'l':
					selectMenu()
				case 'd':
					changeModeToDay()
				case 'm':
					changeModeToMinute()
				case 'q':
					exit()
				default:
					// exit()
				}
			}

			time.Sleep(time.Millisecond * 50)
		}
	}()
}

func changeModeToDay() {
	if isSelectMenu {
		return
	}
	if mode == day {
		return
	}
	mode = day
	stop <- struct{}{}

	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

func changeModeToMinute() {
	if isSelectMenu {
		return
	}
	if mode == min {
		return
	}
	mode = min
	stop <- struct{}{}

	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

var isSelectMenu = false

func flush() {
	write("\x1bc")
	write("\033[H\033[3J")
}

func write(str string) {
	_, _ = fmt.Fprint(Stdout, str)
}

func exit() {
	if isSelectMenu {
		reset()
		return
	}
	flush()
	os.Exit(0)
}

type config struct {
	area string
	code string
	name string
}

var x, y = 2, 1
var oldX, oldY = x, y

var menu = []config{
	{"sh", "000001", "上证指数"},
	{"sh", "000300", "沪深300"},
	{"sh", "600519", "贵州茅台"},
	{"sz", "399006", "创业扳指"},
	{"sh", "603103", "横店影视"},
	{"sz", "000651", "格力电器"},
	{"sh", "601318", "中国平安"},
}

func selectMenu() {
	if isSelectMenu {
		return
	}

	oldX, oldY = x, y
	isSelectMenu = true
	stop <- struct{}{}
	flush()

	tips()

	for i := 0; i < len(menu); i++ {
		write(fmt.Sprintf(" %d %s %s %s\n", i+1, menu[i].area, menu[i].code, menu[i].name))
	}

	resetCursor()
}

func ctrlC(key keyboard.Key) {
	if !isSelectMenu {
		return
	}

	switch key {
	case 0xffea:
		// right()
	case 0xffeb:
		// left()
	case 0xffec:
		down()
	case 0xffed:
		up()
	case 0x000d:
		enter()
	}
}

func resetCursor() {
	write(fmt.Sprintf("\033[%d;%dH", oldX, oldY))
}

func reset() {
	isSelectMenu = false
	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

func enter() {
	isSelectMenu = false
	x, y = oldX, oldY
	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

func left() {}

func right() {}

func up() {
	if oldX > 1+1 {
		oldX--
		resetCursor()
	}
}

func down() {
	if oldX < len(menu)+1 {
		oldX++
		resetCursor()
	}
}
