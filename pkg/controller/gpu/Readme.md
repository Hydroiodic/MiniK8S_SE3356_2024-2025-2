1、首先需要启动一个本地 registry：
docker run -d -p 5000:5000 --name registry registry:2

2、使用job-container下的dockerfile构建镜像：
docker build -t job-server:latest /home/ubuntu/wang/MiniK8S_SE3356_2024-2025-2/assets/job_container

3、将创建好的镜像推到本地registry下
docker tag job-server:latest localhost:5000/job-server:latest
docker push localhost:5000/job-server:latest

4（可选）可以通过 curl http://localhost:5000/v2/_catalog进行检验

