docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:management

可能系统版本有一点旧了……
还是用Docker吧

# 下载安装脚本
curl -s https://packagecloud.io/install/repositories/rabbitmq/rabbitmq-server/script.deb.sh | sudo bash
sudo apt install rabbitmq-server

sudo systemctl status rabbitmq-server

见鬼，什么时候已经配好环境了