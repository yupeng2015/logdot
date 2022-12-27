package logdot

import (
	"os"
	"path/filepath"
	"time"
)

const PathSep = string(os.PathSeparator) // PathSeparator 目录分隔符
const LineSep = "\r\n"                   // LineSeparator 行分隔符

var ProgramDir = func() string {
	path, _ := filepath.Abs(os.Args[0])
	return filepath.Dir(path) + PathSep
}()                                                                   // 程序所在目录，默认要从这个目录下读取配置文件，写日志
var RuntimeDir = ProgramDir + "runtime" + PathSep                     // 运行时生成的文件保存路径
var DailyDir = RuntimeDir + time.Now().Format("2006-01-02") + PathSep // 按天划分的日志，重启时生成
