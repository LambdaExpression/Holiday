package main

import (
	"Holiday/model"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/golang-module/carbon/v2"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var gdb *gorm.DB
var prot, path, logLevel, logEncoding string
var client = &http.Client{}
var printVersion bool
var log *zap.SugaredLogger

var Version = "v0.1"
var GoVersion = "not set"
var GitCommit = "not set"
var BuildTime = "not set"

func init() {
	initFlag()
	if printVersion {
		version()
		os.Exit(0)
	}
	var err error
	log, err = initLogger()
	if err != nil {
		fmt.Println("initDB error ", err)
		os.Exit(0)
	}
	gdb, err = initDB(path + "holiday.db")
	if err != nil {
		log.Error("initDB error ", err)
		os.Exit(0)
	}
}

func main() {
	go func() {
		year := carbon.Now(getLocationStr()).Year()
		var d model.DateInfo
		gdb.Where("date = ?", time.Date(year, 1, 1, 0, 0, 0, 0, getLocation())).First(&d)
		if d.ID == 0 {
			initYear(int64(carbon.Now(getLocationStr()).Year()))
		}
	}()

	irisApp := iris.New()
	irisApp.Use(middleware)
	irisApp.Handle(iris.MethodGet, "/holiday/yesterday", yesterday)
	irisApp.Handle(iris.MethodGet, "/holiday/today", today)
	irisApp.Handle(iris.MethodGet, "/holiday/tomorrow", tomorrow)
	irisApp.Handle(iris.MethodGet, "/info/{date:string}", dataInfo)
	irisApp.Handle(iris.MethodGet, "/update/{year:int}", updateYear)

	err := irisApp.Run(iris.Addr(":" + prot))
	if err != nil {
		log.Error("InitIris error", err)
	}
}

// 初始化配置
func initFlag() {
	flag.StringVar(&prot, "prot", "8282", "--prot 8282")
	flag.StringVar(&path, "path", "./", "--path '/data/'")

	flag.StringVar(&logLevel, "logLevel", "info", "--logLevel info # 日志等级")
	flag.StringVar(&logEncoding, "logEncoding", "console", "--logEncoding console # 日志输出格式 console 或 json")

	flag.BoolVar(&printVersion, "version", false, "--version 打印程序构建版本")

	flag.Parse()
}

func version() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Go Version: %s\n", GoVersion)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Build Time: %s\n", BuildTime)
}

func initYear(year int64) error {
	dateInfos, err := initYearData(year)
	if err != nil {
		return err
	}
	for _, info := range dateInfos {
		var d model.DateInfo
		gdb.Where("date = ?", info.Date).First(&d)
		if d.ID != 0 {
			gdb.Model(d).Where("type != ?", 3).Updates(info)
		} else {
			gdb.Create(&info)
		}
	}
	return nil
}

func initDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&model.DateInfo{})

	return db, err
}

// 初始化 zap日志框架
func initLogger() (*zap.SugaredLogger, error) {
	level := getLevel()
	encoding := logEncoding
	// 保留两个变量，但设置成同一个文件
	//stdout := configs.DataPath + "/log/stdout.log"
	//stderr := configs.DataPath + "/log/stderr.log"

	stdout := &lumberjack.Logger{
		Filename:   path + "stdout.log",
		MaxSize:    10, // 每个日志文件的最大大小，单位为MB
		MaxBackups: 10, // 保留的旧日志文件的最大数量
		MaxAge:     60, // 保留的旧日志文件的最大天数
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
	}
	atom := zap.NewAtomicLevelAt(level)
	config := zap.Config{
		Level:         atom,          // 日志级别
		Development:   true,          // 开发模式，堆栈跟踪
		Encoding:      encoding,      // 输出格式 console 或 json
		EncoderConfig: encoderConfig, // 编码器配置
		//InitialFields:    map[string]interface{}{"serviceName": "gogs-backup"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout", stdout.Filename}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr", stdout.Filename},
	}
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	log := logger.Sugar()
	log.Info("logger 初始化成功")
	return log, err
}

// 获取 日志等级
func getLevel() zapcore.Level {
	levelStr := strings.TrimSpace(strings.ToLower(logLevel))
	var level zapcore.Level
	switch levelStr {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}
	return level
}

func initYearData(year int64) ([]model.DateInfo, error) {
	dateInfos := make([]model.DateInfo, 0)
	startYear := strconv.FormatInt(year, 10)
	endYear := strconv.FormatInt(year+1, 10)
	start := carbon.Parse(startYear)
	end := carbon.Parse(endYear)
	holidayInfo, err := getHolidayInfo(startYear)
	if err != nil {
		return nil, err
	}
	for ; start.Timestamp() < end.Timestamp(); start = start.AddDays(1) {
		d := start.StdTime()
		t := 0
		holiday := start.IsSaturday() || start.IsSunday()
		name := getDayName(start)

		dateStr := start.Format("m-d", getLocationStr())

		holidayDetail := holidayInfo[dateStr]
		if holidayDetail != nil {
			holiday = holidayDetail.Holiday
			name = holidayDetail.Name
			t = 1
		}

		dateInfos = append(dateInfos, model.DateInfo{Date: d, Holiday: holiday, Name: name, Type: t})
	}
	return dateInfos, nil
}

func getDayName(carbon carbon.Carbon) string {
	if carbon.IsMonday() {
		return "周一"
	} else if carbon.IsTuesday() {
		return "周二"
	} else if carbon.IsWeekday() {
		return "周三"
	} else if carbon.IsThursday() {
		return "周四"
	} else if carbon.IsFriday() {
		return "周五"
	} else if carbon.IsSaturday() {
		return "周六"
	} else if carbon.IsSunday() {
		return "周日"
	}
	return ""
}

func getHolidayInfo(year string) (map[string]*model.HolidayDetail, error) {
	url := "http://timor.tech/api/holiday/year/" + year
	// 创建一个GET请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// 发起GET请求
	resp, err := client.Do(req)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body) // 读取响应 body, 返回为 []byte
	if err != nil {
		return nil, err
	}
	var holidayData model.HolidayData
	err = json.Unmarshal(body, &holidayData)
	if err != nil {
		return nil, err
	}
	return holidayData.Holiday, nil
}

func middleware(ctx iris.Context) {
	log.Infof("iris access Path: %s | IP: %s", ctx.Path(), ctx.RemoteAddr())
	ctx.Next()
}

func getDateInfo(date string) (*model.ResponseData, error) {
	c := carbon.ParseByFormat(date, "yy-m-d", getLocationStr())
	t := c.StdTime()

	var d model.DateInfo
	gdb.Where("date = ?", t).First(&d)
	if d.ID == 0 {
		err := initYear(int64(c.Year()))
		if err != nil {
			return nil, err
		}
		gdb.Where("date = ?", t).First(&d)
	}
	return &model.ResponseData{Date: d.Date.Format(time.DateOnly), Holiday: d.Holiday, Name: d.Name, Type: d.Type}, nil
}

func getLocation() *time.Location {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return location
}
func getLocationStr() string {
	return getLocation().String()
}

func yesterday(ctx iris.Context) {
	ctx.Text(getIsHoliday(carbon.Yesterday(getLocationStr())))
}

func today(ctx iris.Context) {
	ctx.Text(getIsHoliday(carbon.Now(getLocationStr())))
}

func tomorrow(ctx iris.Context) {
	ctx.Text(getIsHoliday(carbon.Tomorrow(getLocationStr())))
}

func getIsHoliday(carbon carbon.Carbon) string {
	d := carbon.Format("yy-m-d", getLocationStr())
	info, err := getDateInfo(d)
	if err != nil {
		return err.Error()
	}
	return strconv.FormatBool(info.Holiday)
}

func dataInfo(ctx iris.Context) {
	d := ctx.Params().GetString("date")
	info, err := getDateInfo(d)

	var result model.ResponseInfo
	if err != nil {
		result = model.ResponseInfo{
			Code: -1,
			Msg:  err.Error(),
		}
	} else {
		result = model.ResponseInfo{
			Code: 0,
			Data: info,
		}
	}
	ctx.JSON(&result)
}

func updateYear(ctx iris.Context) {
	year, _ := ctx.Params().GetInt("year")
	initYear(int64(year))
	ctx.JSON(&model.ResponseInfo{
		Code: 0,
		Msg:  "success",
	})
}
