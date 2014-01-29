package main

import (
	"fmt"
	"github.com/mgutz/ansi"
)

/* some terminal helper functions */
var (
	cgreen   func(string) string
	cred     func(string) string
	credu    func(string) string
	cmagenta func(string) string
)

func init() {
	cgreen = ansi.ColorFunc("green")
	cred = ansi.ColorFunc("red")
	credu = ansi.ColorFunc("red+u")
	cmagenta = ansi.ColorFunc("magenta")
}

/* prints either green or red text to the screen, depending
 * on decision. */
func binary(text string, decision bool) string {
	if decision {
		return cgreen(text)
	} else {
		return cred(text)
	}
}

func binaryf(f float64, decision bool) string {
	if decision {
		return greenf(f)
	} else {
		return redf(f)
	}
}

func binaryfp(f float64, decision bool) string {
	if decision {
		return greenfp(f)
	} else {
		return redfp(f)
	}
}

func green(format string, a ...interface{}) string {
	return cgreen(fmt.Sprintf(format, a...))
}

func greenf(f float64) string {
	return green("%.2f", f)
}

func greenfp(f float64) string {
	return green("%.2f%%", f)
}

func red(format string, a ...interface{}) string {
	return cred(fmt.Sprintf(format, a...))
}

func redu(format string, a ...interface{}) string {
	return credu(fmt.Sprintf(format, a...))
}

func redf(f float64) string {
	return red("%.2f", f)
}

func redfp(f float64) string {
	return red("%.2f%%", f)
}

func number(format string, a ...interface{}) string {
	return cmagenta(fmt.Sprintf(format, a...))
}

func numberf(f float64) string {
	return number("%.2f", f)
}

func numberfp(f float64) string {
	return number("%.2f%%", f)
}

func arrow(decision bool) string {
	if decision {
		return "↑"
	} else {
		return "↓"
	}
}
