// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.680
package view

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "strings"

import (
	"fmt"
)

func flexPercent(num int) templ.CSSClass {
	var templ_7745c5c3_CSSBuilder strings.Builder
	templ_7745c5c3_CSSBuilder.WriteString(string(templ.SanitizeCSS(`flex`, fmt.Sprint(num))))
	templ_7745c5c3_CSSID := templ.CSSID(`flexPercent`, templ_7745c5c3_CSSBuilder.String())
	return templ.ComponentCSSClass{
		ID:    templ_7745c5c3_CSSID,
		Class: templ.SafeCSS(`.` + templ_7745c5c3_CSSID + `{` + templ_7745c5c3_CSSBuilder.String() + `}`),
	}
}
