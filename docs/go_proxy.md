
切换模式
写文章
登录/注册
Golang设置网络代理
喵个咪
喵个咪
10 人赞同了该文章
Golang设置网络代理
打开模块支持
go env -w GO111MODULE=on
取消代理
go env -w GOPROXY=direct
取消校验
go env -w GOSUMDB=off
设置不走 proxy 的私有仓库或组，多个用逗号相隔（可选）
go env -w GOPRIVATE=git.mycompany.com,github.com/my/private
设置代理
国内常用代理列表
提供者	地址
官方全球代理	https://proxy.golang.com.cn
七牛云	https://goproxy.cn
阿里云	https://mirrors.aliyun.com/goproxy/
GoCenter	https://gocenter.io
百度	https://goproxy.bj.bcebos.com/
“direct” 为特殊指示符，用于指示 Go 回源到模块版本的源地址去抓取(比如 GitHub 等)，当值列表中上一个 Go module proxy 返回 404 或 410 错误时，Go 自动尝试列表中的下一个，遇见 “direct” 时回源，遇见 EOF 时终止并抛出类似 “invalid version: unknown revision...” 的错误。

官方全球代理
go env -w GOPROXY=https://proxy.golang.com.cn,direct
go env -w GOPROXY=https://goproxy.io,direct
go env -w GOSUMDB=gosum.io+ce6e7565+AY5qEHUk/qmHc5btzW45JVoENfazw8LielDsaI+lEbq6
go env -w GOSUMDB=sum.golang.google.cn
七牛云
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=goproxy.cn/sumdb/sum.golang.org
阿里云
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
# GOSUMDB 不支持
GoCenter
go env -w GOPROXY=https://gocenter.io,direct
# 不支持 GOSUMDB
百度
go env -w GOPROXY=https://goproxy.bj.bcebos.com/,direct
# 不支持 GOSUMDB
参考资料
goproxy.io文档
goproxy.io文档
七牛云文档
阿里云文档
开发者头条文档
百度文档
Go 国内加速：Go 国内加速镜像
发布于 2022-10-10 14:39
Go 语言
Go 编程学习
网络代理
​赞同 10​
​3 条评论
​分享
​喜欢
​收藏
​申请转载
​
写下你的评论...

3 条评论
默认
最新
TriticumSpp
TriticumSpp
测试，全球代理可能不太好使。这个代码比较好使：go env -w GOPROXY=goproxy.cn,direct

2023-07-17 · 上海
​回复
​3
被窝探险家
被窝探险家

感谢了，这个是真的好使啊[惊喜][大哭]

2023-12-13 · 宁夏
​回复
​1
7262
7262
有用

2024-05-09 · 中国香港
​回复
​喜欢
关于作者
喵个咪
喵个咪
回答
67
文章
94
关注者
149
​关注他
​发私信
推荐阅读
APP服务器，你想知道的都在这里
说到服务器，一般大家都知道有web服务器的存在，也就是网站服务器。因为不管是大家现在使用的智能手机还是电脑上面，都有浏览器的存在，对应就需要使用到web服务器。但是说到APP服务器，相…

鼎峰-凤凰
APP连接服务器，所需要知道的一切事情
APP连接服务器，所需要知道的一切事情
邵励治
使用代理服务器最简单的步骤
互联数据在线代理服务器是网上提供转接功能的服务器，就好比你想看前女友的朋友圈，由于种种原因不方便直接访问，你不想直接关注她，你可以换一个账号。代理服务器也是这样(通过代理服务器…

知乎用户VjeBGT
玩转云服务：手把手带你薅一台腾讯云服务器，公网 IP
玩转云服务：手把手带你薅一台腾讯云服务器，公网 IP
AI码上来
发表于AI工具


选择语言
登录即可查看 超5亿 专业优质内容
超 5 千万创作者的优质提问、专业回答、深度文章和精彩视频尽在知乎。
立即登录/注册