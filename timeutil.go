package main

import (
	"fmt"
	db "myapp/DB"
	"net/url"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/labstack/echo/v4"
)

const defaultAppTimezone = "America/Los_Angeles"

func mustLoadDefaultLocation() *time.Location {
	loc, err := time.LoadLocation(defaultAppTimezone)
	if err != nil {
		panic(err)
	}
	return loc
}

var defaultAppLocation = mustLoadDefaultLocation()
var timeNow = time.Now

func browserTimezone(c echo.Context) string {
	cookie, err := c.Cookie("tz")
	if err != nil {
		return ""
	}
	raw, err := url.QueryUnescape(strings.TrimSpace(cookie.Value))
	if err != nil {
		return ""
	}
	if raw == "" {
		return ""
	}
	if _, err := time.LoadLocation(raw); err != nil {
		return ""
	}
	return raw
}

func loadUserLocation(c echo.Context) *time.Location {
	if userID, ok := c.Get(ctxUserID).(int); ok {
		if timezone, err := db.GetUserTimezone(userID); err == nil && timezone != "" {
			if loc, err := time.LoadLocation(timezone); err == nil {
				return loc
			}
		}
	}
	if timezone := browserTimezone(c); timezone != "" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			return loc
		}
	}
	return defaultAppLocation
}

func currentLocalDay(loc *time.Location) time.Time {
	return anchorDay(timeNow().In(loc), loc)
}

func anchorDay(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 12, 0, 0, 0, loc)
}

func formatDayKey(t time.Time, loc *time.Location) string {
	return anchorDay(t, loc).Format("2006-01-02")
}

func parseDayValue(raw string, loc *time.Location) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if day, err := time.ParseInLocation("2006-01-02", raw, loc); err == nil {
		return anchorDay(day, loc), true
	}
	if unix, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return anchorDay(time.Unix(unix, 0).In(loc), loc), true
	}
	return time.Time{}, false
}

func canonicalDayValue(raw string, loc *time.Location) string {
	if day, ok := parseDayValue(raw, loc); ok {
		return formatDayKey(day, loc)
	}
	return ""
}

func parseRequestedDay(c echo.Context) time.Time {
	loc := loadUserLocation(c)
	if day, ok := parseDayValue(c.Param("date"), loc); ok {
		return day
	}
	if day, ok := parseDayValue(c.QueryParam("date"), loc); ok {
		return day
	}
	return currentLocalDay(loc)
}

func sameLocalDay(a, b time.Time, loc *time.Location) bool {
	aa := anchorDay(a, loc)
	bb := anchorDay(b, loc)
	return aa.Year() == bb.Year() && aa.YearDay() == bb.YearDay()
}

func overviewPathForDay(day time.Time, loc *time.Location) string {
	if sameLocalDay(day, currentLocalDay(loc), loc) {
		return "/"
	}
	return "/" + formatDayKey(day, loc)
}

func dayLabel(day time.Time, loc *time.Location) string {
	if sameLocalDay(day, currentLocalDay(loc), loc) {
		return "Today"
	}
	return anchorDay(day, loc).Format("Mon, Jan 2")
}

func queryDayValue(c echo.Context) string {
	loc := loadUserLocation(c)
	if day, ok := parseDayValue(c.QueryParam("date"), loc); ok {
		return formatDayKey(day, loc)
	}
	return ""
}

func querySuffixForDay(c echo.Context) string {
	if day := queryDayValue(c); day != "" {
		return "?date=" + day
	}
	return ""
}

func dayQueryValue(day time.Time, loc *time.Location) string {
	if sameLocalDay(day, currentLocalDay(loc), loc) {
		return ""
	}
	return formatDayKey(day, loc)
}

func addDateQuery(path string, dayValue string) string {
	if dayValue == "" {
		return path
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "date=" + dayValue
}

func persistInitialUserTimezone(c echo.Context, userID int) {
	timezone := browserTimezone(c)
	if timezone == "" {
		return
	}
	savedTimezone, err := db.GetUserTimezone(userID)
	if err != nil || savedTimezone != "" {
		return
	}
	_ = db.UpdateUserTimezone(userID, timezone)
}

func requireTimezone(raw string) (string, error) {
	timezone := strings.TrimSpace(raw)
	if timezone == "" {
		return "", fmt.Errorf("Timezone is required")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "", fmt.Errorf("Timezone must be a valid IANA timezone")
	}
	return timezone, nil
}
