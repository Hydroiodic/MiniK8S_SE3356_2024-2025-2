#!/bin/bash
#SBATCH --job-name=test-job-namespace_job-example1
#SBATCH --output=job-example1.out
#SBATCH --error=job-example1.err
#SBATCH --partition=dgx2
#SBATCH -n 1
#SBATCH --ntasks-per-node=1
#SBATCH --gres=gpu:1
#SBATCH --cpus-per-task=6
module load gcc cuda
nvcc matrix_add.cu -o matrix_add
nvcc matrix_multiply.cu -o matrix_multiply
./matrix_add
./matrix_multiply
