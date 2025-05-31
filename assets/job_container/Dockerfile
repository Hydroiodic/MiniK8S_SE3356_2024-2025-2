FROM golang:1.22.2-alpine

# 将工作目录切换到程序代码所在的目录
WORKDIR /app

# 将当前目录下的所有文件都复制到工作目录下
COPY . .

RUN mkdir -p ./func

# 设置 proxy
RUN go env -w GOPROXY="https://goproxy.io,direct"

# 构建 Go 程序
RUN mkdir -p /root/.ssh 
COPY ./ssh_config/* /root/.ssh/

RUN apk add --no-cache openssh-client
RUN chmod +x /app/start.sh
RUN go build -o /app/job-server /app/main.go

CMD ["/app/start.sh"]

# 构建镜像
# 要构建容器，可以使用以下命令：
# 需要在项目的根路径执行
# docker build -t job-server:latest .
# docker run --entrypoint /bin/job-server musicminion/minik8s-gpu \
#    -jobName job-example1 \
#    -jobNamespace test-job-namespace \
#    -apiServerAddr http://192.168.126.130:8090
