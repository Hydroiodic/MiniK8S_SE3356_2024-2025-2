#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include "cuda_runtime.h"
#include "device_launch_parameters.h"
using namespace std;
const int M = 8;
const int N = 8;
// 矩阵加法核函数
// 每个线程负责计算C[i][j] = A[i][j] + B[i][j]
// 核启动函数
__global__ void matrix_add(int **A, int **B, int **C) {
    int i = (blockIdx.x * blockDim.x + threadIdx.x);
    int j = (blockIdx.y * blockDim.y + threadIdx.y);
    C[i][j] = A[i][j] + B[i][j];
}

int main() {
    // 初始化主机二维数组
    int nbytes=M*N*sizeof(int);
    int **host_A = (int **) malloc(M * sizeof(int *));
    int **host_B = (int **) malloc(M * sizeof(int *));
    int **host_C = (int **) malloc(M * sizeof(int *));
    int *data_A = (int *) malloc(nbytes);
    int *data_B = (int *) malloc(nbytes);
    int *data_C = (int *) malloc(nbytes);
    for (int i = 0; i < M; i++) {
        host_A[i] = &data_A[i * N];
        host_B[i] = &data_B[i * N];
        host_C[i] = &data_C[i * N];
        for (int j = 0; j < N ; j++) {
            data_A[i*N+j] = i*N+j;
            data_B[i*N+j] = i*N+j;
            data_C[i*N+j] = 0;
        }
    }

    // 声名设备上的指针，它们将保存GPU上矩阵的指针的设备内存地址。
    int **dev_A, **dev_B, **dev_C;。
    int *dev_A1, *dev_B1, *dev_C1;
    cudaMalloc((void **)&dev_A1, nbytes);
    cudaMalloc((void **)&dev_B1, nbytes);
    cudaMalloc((void **)&dev_C1, nbytes);
    // 将已初始化的矩阵数据从主机的连续内存 (data_A, data_B) 复制到设备的连续内存 (dev_A1, dev_B1)
    cudaMemcpy((void *)dev_A1, (void *)data_A, nbytes, cudaMemcpyHostToDevice);
    cudaMemcpy((void *)dev_B1, (void *)data_B, nbytes, cudaMemcpyHostToDevice);
    // 将分配给 dev_C1 的设备内存中的所有字节设置为零。这有效地将设备上的结果矩阵初始化为全零。
    cudaMemset((void *)dev_C1, 0, nbytes);
    // host_A[i] 将存储 &dev_A1[i*N] 的设备地址（设备内存中每行的地址）
    for (int i = 0; i < M; i++) {
        host_A[i] = dev_A1 + i * N;
        host_B[i] = dev_B1 + i * N;
        host_C[i] = dev_C1 + i * N;
    }
    //对行指针进行同样的操作
    cudaMalloc((void **)&dev_A, sizeof(int *) * M);
    cudaMalloc((void **)&dev_B, sizeof(int *) * M);
    cudaMalloc((void **)&dev_C, sizeof(int *) * M);

    cudaMemcpy((void *)dev_A, (void *)host_A, sizeof(int *) * M, cudaMemcpyHostToDevice);
    cudaMemcpy((void *)dev_B, (void *)host_B, sizeof(int *) * M, cudaMemcpyHostToDevice);
    cudaMemcpy((void *)dev_C, (void *)host_C, sizeof(int *) * M, cudaMemcpyHostToDevice);

    // 定义网格维度。这将创建一个二维线程块网格。每个维度是矩阵维度的一半，这意味着每个块将处理 2x2 个元素。
    dim3 grid(M / 2, N / 2);
    // 定义块维度。这将创建一个 2×2=4 个线程的二维块
    dim3 block(2, 2);
    // matrix_add 核函数以指定的 grid 和 block 维度启动
    matrix_add<<<grid, block>>>(dev_A, dev_B, dev_C);
    // 核函数执行完毕后，这行代码将计算得到的结果矩阵从设备的连续内存 (dev_C1) 复制回主机的连续内存 (data_C)
    cudaMemcpy((void *) data_C,(void *) dev_C1, nbytes, cudaMemcpyDeviceToHost);

    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            cout<<data_C[i*N+j]<<" ";
        }
        cout<<endl;
    }
    free(data_A);
    free(data_B);
    free(data_C);
    free(host_A);
    free(host_B);
    free(host_C);
    cudaFree(dev_A);
    cudaFree(dev_B);
    cudaFree(dev_C);
    cudaFree(dev_A1);
    cudaFree(dev_B1);
    cudaFree(dev_C1);

    return 0;
}