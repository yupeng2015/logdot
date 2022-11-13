package logdot

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const PathSep = string(os.PathSeparator) // PathSeparator 目录分隔符
const LineSep = "\r\n"                   // LineSeparator 行分隔符

type Logdot struct {
	*Logger
	MultipleLogger map[string]*Logger
	currentLogger  *Logger
}

func (this *Logdot) SpecialLogger(spec string) *Logger {
	return this.MultipleLogger[spec]
}

type Logger struct {
	Log *log.Logger
}

// 原始数据打印
func (this *Logger) Print(s string) {
	this.Log.Output(2, fmt.Sprintln(s))
}

func (this *Logger) Info(s ...any) {
	content := this.SetContent("INFO", s)
	this.Log.Output(2, fmt.Sprintln(content...))
}

func (this *Logger) Infof(s string, v ...any) {
	this.Log.Output(2, fmt.Sprintf(s, v...))
}

func (this *Logger) Warn(s ...any) {
	content := this.SetContent("WARN", s)
	this.Log.Output(2, fmt.Sprintln(content...))
}

func (this *Logger) Error(s ...any) {
	this.Log.Output(2, fmt.Sprintln(this.SetContent("ERROR", s)...))
}

func (this *Logger) Errorf(s string, v ...any) {
	this.Log.Output(2, fmt.Sprintf(s, v...))
}

func (this *Logger) SetContent(level string, s []any) []any {
	arr := make([]any, len(s)+1)
	arr[0] = "[" + level + "]"
	copy(arr[1:], s)
	return arr
}

/*
协程安全的Writer
*/
type SyncWriter struct {
	Chan chan []byte
}

func (w *SyncWriter) Write(p []byte) (n int, err error) {
	p2 := make([]byte, len(p))
	copy(p2, p)
	w.Chan <- p2
	return len(p), nil
}

/*
写入的同时打印到控制台
*/
type PrintWriter struct {
	SyncWriter
}

func NewPrintWriter(c chan []byte) *PrintWriter {
	w := new(PrintWriter)
	w.Chan = c
	return w
}

func (w *PrintWriter) Write(p []byte) (n int, err error) {
	_, _ = os.Stdout.Write(p)
	return w.SyncWriter.Write(p)
}

/*
创建并返回一个写入文件的通道
内部有一个协程负责将通道中的数据写入到文件中
dir是文件目录
format是文件名格式，可以根据写入时间生成文件名
*/
func FileChan(dir string, format string) chan []byte {
	if err := os.MkdirAll(dir, 0777); err != nil {
		panic("创建目录:" + dir + "失败, err:" + err.Error())
	}
	ch := make(chan []byte, 1000)
	go func() {
		name := ""
		var file *os.File
		for msg := range ch {
			if name2 := time.Now().Format(format); name2 != name {
				if file != nil {
					_ = file.Close()
					file = nil
				}
				name = name2
			}
			if file == nil {
				var err error
				file, err = os.OpenFile(dir+name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("打开文件失败", err)
					return
				}
			}
			if _, err := file.Write(msg); err != nil {
				fmt.Println("写入文件失败", file, err)
			}
		}
	}()
	return ch
}

/*
指定写入的文件的路径格式(fixed+variable)，可以根据写入时间控制写入的文件
*/
type DailyWriter struct {
	fixed        string            // 路径中固定的部分
	variable     string            // 路径中可变的部分，遵循时间格式
	SpecificFile map[string]string //特指文件路径
	path         string            // 当前路径
	file         *os.File          // 当前文件
}

func NewDailyWriter(fixed string, variable string) *DailyWriter {
	w := &DailyWriter{
		fixed:    fixed,
		variable: variable,
	}
	w.SwitchPath(w.GenPath())
	return w
}

func (w *DailyWriter) Write(p []byte) (n int, err error) {
	if path := w.GenPath(); path != w.path {
		w.SwitchPath(path)
	}
	return w.file.Write(p)
}

func (w *DailyWriter) GenPath() string {
	return w.fixed + time.Now().Format(w.variable)
}

func (w *DailyWriter) SwitchPath(path string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		panic("创建目录:" + dir + "失败, err:" + err.Error())
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic("打开文件:" + path + "失败, err:" + err.Error())
	}
	w.path = path
	w.file = file
}

func (w *DailyWriter) SetSpecificFile(sf map[string]string) {
	w.SpecificFile = sf
}

/*
写入控制台的同时也可以写入其他地方
*/
type ConsoleWriter []io.Writer

func NewConsoleWriter(ws ...io.Writer) *ConsoleWriter {
	ws = append(ws, os.Stdout)
	return (*ConsoleWriter)(&ws)
}

func (g *ConsoleWriter) Write(p []byte) (n int, err error) {
	for _, w := range *g {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return
}

type Option struct {
	Dir          string
	File         string
	SpecificFile map[string]string
	Stdout       bool
}

func Create(opt Option) (dot *Logdot) {
	dot = &Logdot{}
	dot.MultipleLogger = make(map[string]*Logger)
	for k, v := range opt.SpecificFile {
		MultipleWriter := loadOpt(opt.Dir, v)
		ml := &Logger{
			Log: log.New((*ConsoleWriter)(&MultipleWriter), "", log.LstdFlags|log.Lshortfile),
		}
		dot.MultipleLogger[k] = ml
	}
	//设置默认得日志
	wList := loadOpt(opt.Dir, opt.File)
	if opt.Stdout {
		wList = append(wList, os.Stdout)
	}
	l := &Logger{
		Log: log.New((*ConsoleWriter)(&wList), "", log.LstdFlags|log.Lshortfile),
	}
	dot.Logger = l
	return
}

func loadOpt(dir, file string) []io.Writer {
	daily := NewDailyWriter(dir, file)
	return []io.Writer{daily}
}
