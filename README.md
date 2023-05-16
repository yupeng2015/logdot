# logdot

## 介绍
精简的go日志库
## 使用
```go
logger := logdot.Create(logdot.Option{
    Dir:    "runtime/",  //日志目录位置
    File:   "log/2006-01-02/server/2006-01-02 15.log", //日志文件
    Stdout: true,  //是打印到控制台
    SpecificFile: map[string]logdot.Option{  //独立日志，使用map[string]logdot.Option配置，例如单独保存sql，单独打印一些异常
        "program": {
            Dir:    ProgramDir + "runtime/",
            File:   "program.info",
            Stdout: false,
        },
        "sql": {
            Dir:    ProgramDir + "runtime/",
            File:   "sql/2006-01-02 15.log",
            Stdout: true,
        },
    },
})
logger.Info("日志内容")  //默认使用日志打印 
loggerSql := logger.SpecialLogger("sql")  //使用独立日志
loggerSql.Warn("异常日志 select xxxx from xxx异常")
```