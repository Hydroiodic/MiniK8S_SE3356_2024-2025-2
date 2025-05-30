FROM localhost:5000/job-server:latest
ENV OUTPUT_FILE job-example1.out
ENV ERROR_FILE job-example1.err
ENV API_SERVER_IP&PORT http://192.168.1.14:8080
ENV JOB_NAME job-example1
ENV JOB_NAMESPACE test-job-namespace
COPY ./job-example1 /app/func
