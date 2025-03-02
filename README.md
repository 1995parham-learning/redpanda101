<h1 align="center">Red Panda 🐼</h1>
<p align="center">
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham-learning/redpanda101/test.yaml?style=for-the-badge&logo=github">
</p>

## Introduction

In certain scenarios, we may seek a lighter alternative to Kafka, and that alternative could be [Red Panda](https://redpanda.com/). In this repository,
I explore using Red Panda as a Kafka replacement with Go. For Kafka integration in Go,
I rely on [franz-go](https://github.com/twmb/franz-go). Additionally, other alternatives include [Confluent Kafka](https://github.com/confluentinc/confluent-kafka-go) and [Sarama](https://github.com/IBM/sarama).

## How to

## Load Test

```
         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/

     execution: local
        script: script.js
        output: -

     scenarios: (100.00%) 1 scenario, 35 max VUs, 2m30s max duration (incl. graceful stop):
              * default: Up to 35 looping VUs for 2m0s over 1 stages (gracefulRampDown: 30s, gracefulStop: 30s)


     █ publish

       ✓ success

     checks.........................: 100.00% 606848 out of 606848
     data_received..................: 210 MB  1.8 MB/s
     data_sent......................: 134 MB  1.1 MB/s
     group_duration.................: avg=3.44ms  min=208.79µs med=3.34ms max=87.79ms p(90)=4.92ms p(95)=5.73ms
     http_req_blocked...............: avg=1.25µs  min=0s       med=1µs    max=2.58ms  p(90)=2µs    p(95)=3µs
     http_req_connecting............: avg=8ns     min=0s       med=0s     max=418µs   p(90)=0s     p(95)=0s
   ✓ http_req_duration..............: avg=3.4ms   min=8µs      med=3.29ms max=87.75ms p(90)=4.87ms p(95)=5.68ms
       { expected_response:true }...: avg=3.4ms   min=8µs      med=3.29ms max=87.75ms p(90)=4.87ms p(95)=5.68ms
   ✓ http_req_failed................: 0.00%   0 out of 606848
     http_req_receiving.............: avg=16.88µs min=4µs      med=11µs   max=3.74ms  p(90)=33µs   p(95)=43µs
     http_req_sending...............: avg=4.56µs  min=1µs      med=3µs    max=4.86ms  p(90)=9µs    p(95)=13µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s     max=0s      p(90)=0s     p(95)=0s
     http_req_waiting...............: avg=3.37ms  min=0s       med=3.27ms max=87.74ms p(90)=4.85ms p(95)=5.65ms
     http_reqs......................: 606848  5059.115018/s
     iteration_duration.............: avg=3.45ms  min=212µs    med=3.35ms max=87.79ms p(90)=4.93ms p(95)=5.74ms
     iterations.....................: 606848  5059.115018/s
     vus............................: 34      min=1                max=34
     vus_max........................: 35      min=35               max=35


running (2m00.0s), 00/35 VUs, 606848 complete and 0 interrupted iterations
default ✓ [======================================] 00/35 VUs  2m0s
```
