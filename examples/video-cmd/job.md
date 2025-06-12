 1. 演⽰任务的提交：
准备cuda程序和编译脚本，通过yaml配置⽂件提交给minik8s。
yaml⽂件的格式⽐较灵活，必须字段只有name、kind，任务的配置信息可以⾃⾏设计，包括
slurm脚本要求的⼀些配置信息。
提交之后，能够通过kubectlget之类的命令得到任务的提交情况
CUDA程序需要是矩阵乘法和矩阵加法程序，需要展⽰代码并简单介绍如何利⽤GPU的并发能⼒
./kubectl.sh apply -f examples/api/gpu-job.yaml 
./kubectl.sh get jobs   

2. 演⽰获取GPU任务的返回结果
演⽰minik8s的⽤⼾如何通过kubectl的相关命令来得到返回结果。如果演⽰时任务⼀直处于
pending状态，则推迟到演⽰的最后查看GPU上cuda程序的返回结果。如果仍然pending，那
么就跳过这⼀步