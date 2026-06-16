# README
1、使用 go + wails  构建的一个boce小工具, apiKey 可以从www.boce.com 获取.目前支持的功能有: 污染检测、qq拦截检测、微信拦截检测、备案查询、备案黑名单查询、背墙检测
2、编译命令: https://wails.io/zh-Hans/docs/gettingstarted/installation
wails build -platform darwin/arm64
wails build -platform darwin/amd64
wails build -platform windows/amd64
wails build -platform linux/amd64

3、已经编译好的在 build/bin
        .exe  windows 系统
        darwin_amd64  mac intel 芯片
        darwin_arm64  mac m 芯片