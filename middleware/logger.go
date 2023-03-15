package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	//取别名
	retalog "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func Logger() gin.HandlerFunc {
	filePath := "log/log"
	//linkName := "latest_log.log"

	//没有的话就新建，有的话追加，这打开的是一个总log
	scr, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("err:", err)
	}
	logger := logrus.New()

	//输出文件，是一个*File
	logger.Out = scr

	logger.SetLevel(logrus.DebugLevel)

	//旋转日志，将里面同一天的取出来当作一个新log文件
	logWriter, _ := retalog.New(
		filePath+"%Y%m%d.log",
		//最大保存时间
		retalog.WithMaxAge(7*24*time.Hour),
		//什么时候分割一次，24小时分割一次
		retalog.WithRotationTime(24*time.Hour),
		//软链接，windows下不好弄
		//retalog.WithLinkName(linkName),
	)

	//钩子
	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}
	Hook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		//时间格式化模板。go的诞生时间
		TimestampFormat: "2006-01-02 15:04:05",
	})
	//添加钩子
	logger.AddHook(Hook)

	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		stopTime := time.Since(startTime).Milliseconds()
		spendTime := fmt.Sprintf("%d ms", stopTime)
		hostName, err := os.Hostname()
		if err != nil {
			hostName = "unknown"
		}
		statusCode := c.Writer.Status()
		//客户端ip
		clientIp := c.ClientIP()
		//客户端的信息
		userAgent := c.Request.UserAgent()
		//请求文件长度
		dataSize := c.Writer.Size()
		if dataSize < 0 {
			dataSize = 0
		}
		//请求方法，method是指GET POST等
		method := c.Request.Method
		//请求路径
		path := c.Request.RequestURI

		entry := logger.WithFields(logrus.Fields{
			"HostName":  hostName,
			"status":    statusCode,
			"SpendTime": spendTime,
			"Ip":        clientIp,
			"Method":    method,
			"Path":      path,
			"DataSize":  dataSize,
			"Agent":     userAgent,
		})
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		}
		if statusCode >= 500 {
			entry.Error()
		} else if statusCode >= 400 {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
