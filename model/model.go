package model

import (
	"gorm.io/gorm"
	"time"
)

type DateInfo struct {
	gorm.Model
	Date    time.Time
	Holiday bool
	Name    string
	Type    int
}

type AutoGenerated struct {
	Code    int                    `json:"code"`
	Holiday map[string]HolidayInfo `json:"holiday"`
}

type HolidayInfo struct {
	Holiday bool   `json:"holiday"`
	After   bool   `json:"after"`
	Wage    int    `json:"wage"`
	Name    string `json:"name"`
	Target  string `json:"target"`
	Date    string `json:"date"`
}

type HolidayData struct {
	Code    int                       `json:"code"`
	Holiday map[string]*HolidayDetail `json:"holiday"`
}

type HolidayDetail struct {
	Holiday bool   `json:"holiday"`
	Name    string `json:"name"`
	Wage    int    `json:"wage"`
	Date    string `json:"date"`
	Rest    int    `json:"rest"`
}

type ResponseInfo struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data *ResponseData `json:"data"`
}

type ResponseData struct {
	Date    string `json:"date"`
	Holiday bool   `json:"holiday"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
}
