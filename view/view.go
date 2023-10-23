package view

import (
	"fmt"

	"github.com/a-h/templ"
)

// ------------------------
func ProgressCssFlexPercent(percent int) templ.CSSClass {
	templCSSID := fmt.Sprintf("Flex-Percent-%d", percent)

	cls := fmt.Sprintf(`.%s { 
		flex : %d;
		}`, templCSSID, percent)

	return templ.ComponentCSSClass{
		ID:    templCSSID,
		Class: templ.SafeCSS(cls),
	}
}

func ProgressCssColor(color string) templ.CSSClass {
	templCSSID := fmt.Sprintf("Color-%s", color)

	cls := fmt.Sprintf(`.%s { 
		--progress-color : #%s;
		}`, templCSSID, color)

	return templ.ComponentCSSClass{
		ID:    templCSSID,
		Class: templ.SafeCSS(cls),
	}
}
