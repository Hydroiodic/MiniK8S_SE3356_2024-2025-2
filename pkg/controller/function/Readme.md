1、首先需要启动一个本地 registry：
docker run -d -p 5000:5000 --name registry registry:2

2、使用job-container下的dockerfile构建镜像：
docker build -t baseserver:latest /home/ubuntu/wang/MiniK8S_SE3356_2024-2025-2/assets/func_container

3、将创建好的镜像推到本地registry下
docker tag baseserver:latest localhost:5000/baseserver:latest
docker push localhost:5000/baseserver:latest

4（可选）可以通过 curl http://localhost:5000/v2/_catalog进行检验

5 http trigger test:
func test:
curl -X POST http://localhost:8060/triggerFunction/default/x_y_add  -H "Content-Type: application/json"      -d '{"x": 5, "y": 3}'

curl -X POST http://localhost:8060/triggerFunction/default/fibonaccifunc -H "Content-Type: application/json" -d '{"x": 0, "y": 1, "i": 1}'

workflow test:
curl -X POST localhost:8060/triggerWorkflow/default/FibonacciWorkflow

6 complex test
./kubectl.sh apply -f examples/api/complex-event/load_image_from_url.yaml
./kubectl.sh apply -f examples/api/complex-event/preprocess_image.yaml 
./kubectl.sh apply -f examples/api/complex-event/generate_description.yaml 
./kubectl.sh apply -f examples/api/complex-event/outputresult.yaml 
./kubectl.sh apply -f examples/api/complex-event/complex-workflow.yaml 

curl -X POST http://localhost:8060/triggerFunction/default/load_image_from_url -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerFunction/default/preprocess_image -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerFunction/default/generate_description -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerFunction/default/outputresult -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerWorkflow/default/imageProcessWorkflow -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'