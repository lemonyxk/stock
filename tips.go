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
	"unicode/utf8"

	"github.com/lemonyxk/console"
)

func tips() {
	var sm = mode
	if mode == day {
		sm += " 365"
	}
	var str = fmt.Sprintf("[Mode: %s] [Q: Quit] [B: Back] [M: Min K] [D: Day K]\r\n", sm)
	var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(str))/2)
	write(console.FgYellow.Sprint(s + str))
}

func menuTips() {
	var str = fmt.Sprintf("[Q: Quit] [↵: Enter] [↑↓: Move] [E: Edit] [%d]\r\n", runningProcess)
	var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(str))/2)
	write(console.FgYellow.Sprint(s + str))
}
