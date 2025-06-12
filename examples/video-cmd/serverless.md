
2、 演⽰Function的定义和运⾏流程
    2.1、⽤⼾定义的函数内容，通过指令上传给minik8s，上传之后，minik8s能够通过kubect lget之类的指令查看到函数。函数⾄少⽀持Python语⾔
    ./kubectl.sh apply -f examples/api/func.yaml
    ./kubectl.sh get functions
    ./kubectl.sh get  replicasets
    2.2、上传之后，⽤⼾能够通过http请求对函数进⾏调⽤

    2.2.1、演⽰当不存在函数实例时，⾸次调⽤函数会⾃动⽣成新的实例
    ./kubectl.sh get pods 看一下现在没有
    curl -X POST http://localhost:8060/triggerFunction/default/x_y_add  -H "Content-Type: application/json"      -d '{"x": 5, "y": 3}'
    ./kubectl.sh get pods 再看一下有新的pods创建

    2.2.2演示当请求并发数（RPS）增多时，serverless能够扩容⾄多个实例，并且请求可以发送⾄这些实例中的任意⼀个进⾏处理。简要介绍扩容策略。
    多次上述操作

    2.2.3演⽰当⼀段时间内没有新的请求到来时（如30s或1min），函数实例会被清除.

3、演⽰Function的更新和删除
 3.1、对某⼀个函数（以函数名或者id来唯⼀标识），如果有新的函数代码，可以通过update之类的指令对函数进⾏更新。
    ./kubectl.sh update function x_y_add

 3.2、以通过delete之类的指令对函数进⾏删除
    ./kubectl.sh delete function x_y_add

4 演示event
    ./kubectl.sh apply -f examples/api/event.yaml

1、 介绍本⼩组选⽤的复杂应⽤，并基于此应⽤进⾏后续展⽰
    图
5、演⽰Workflow定义和运⾏流程
    5.1、需要展⽰Workflow的定义⽂件（如yaml配置⽂件或其他⽅式），展⽰Workflow的定义和上传过程。展⽰Workflow内的各个函数，建议设置辨识度以区分每个函数在workflow中的位置，同时也要体现出输⼊参数值对返回结果的影响（不可以出现“⽆论传⼊参数是多少，返回结果都是⼀样”的情况）
        ./kubectl.sh apply -f examples/api/complex-event/load_image_from_url.yaml
        ./kubectl.sh apply -f examples/api/complex-event/preprocess_image.yaml 
        ./kubectl.sh apply -f examples/api/complex-event/generate_description.yaml 
        ./kubectl.sh apply -f examples/api/complex-event/outputresult.yaml 
        ./kubectl.sh apply -f examples/api/complex-event/complex-workflow.yaml 
    5.2、介绍Workflow⽀持怎样的分⽀/条件
    运⾏Workflow，演⽰workflow如何运⾏⼀条分⽀上的所有函数的。要求前⼀个函数的输出作为后⼀个函数的输⼊，体现在下游函数的计算过程中
    
curl -X POST http://localhost:8060/triggerFunction/default/load_image_from_url -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'

curl -X POST http://localhost:8060/triggerFunction/default/preprocess_image -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerFunction/default/generate_description -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerFunction/default/outputresult -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'



curl -X POST http://localhost:8060/triggerWorkflow/default/imageProcessWorkflow -H "Content-Type: application/json" -d '{"image_url": "https://bpic.588ku.com/element_origin_min_pic/23/07/11/d32dabe266d10da8b21bd640a2e9b611.jpg!r650"}'

./kubectl.sh get pods
./kubectl.sh get functions
./kubectl.sh get  replicasets


6. 演⽰serverless的并发测试效果
◦
使⽤Jmeter等压⼒测试⼯具进⾏并发量为⼆⼗的并发测试，展⽰并发线程数的设置以及测试的
结果