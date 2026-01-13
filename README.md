<h1 align="center">Red Panda 🐼</h1>
<p align="center">
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham-learning/redpanda101/test.yaml?style=for-the-badge&logo=github">
</p>

## Introduction

In certain scenarios, we may seek a lighter alternative to Kafka, and that alternative could be [Red Panda](https://redpanda.com/). In this repository,
I explore using Red Panda as a Kafka replacement with Go. For Kafka integration in Go,
I rely on [franz-go](https://github.com/twmb/franz-go). Additionally, other alternatives include [Confluent Kafka](https://github.com/confluentinc/confluent-kafka-go) and [Sarama](https://github.com/IBM/sarama).

## How to

To evaluate the new event driven architecture, I used [redpanda101](https://github.com/1995parham-learning/redpanda101) to demostrate the architecture and Go improvements. I used Redpanda because it is easier to run using docker and we don't want to compare Redpanda with Kafka in any way.

## How to Run

First, Running the requirements including database, promethues, etc:

```bash
cd deployment
docker compose pull
docker compose up
```

Redpanda requires to create topic manually, so before moving forward, create the `orders` topic through its UI.

```text
http://192.168.73.4:8080

http://127.0.0.1:8080
```

and create `orders` table using built-in migrate command:

```bash
./redpanda101 migrate
```

Then run consumer and producer applications:

```bash
./redpanda101 -c configs/producer.toml produce
./redpanda101 -c configs/consumer.toml consume
```

At the end, you can use k6 to do the load test:

```bash
k6 run api/k6/script.js
```

or use requests in the `demo.http` to try out APIs.

## How to Monitor

Because the sole purpose of the project is evaluation, redpanda101 has lots tools for monitoring and tracing available.

- Prometheus

```text
http://192.168.73.4:9090

http://127.0.0.1:9090
```

- Jeager

```text
http://192.168.73.4:16686

http://127.0.0.1:16686
```

- Grafana

```text
username: parham.alvani@gmail.com
password: P@ssw0rd

http://192.168.73.4:3000

http://127.0.0.1:3000
```

## Parameters

There are lots of moving parts in here, just mention a few:

- k6 script
  - target
  - sleep
- consumer
  - number of workers

The `k6` script has two target first one is the ramp-up target and the second one indicate the steady target.

---

We plan to rewrite the Order Management System (OMS). In this rewrite, we aim to fully utilize memory and no longer rely on a database. Based on the architectures we've reviewed, I think we can start by using Kafka.

Given that Kafka has been rewritten in Redpanda, I will use Redpanda to begin with, as it is better prepared for cloud environments. For example, I will develop a project using the Go language around Redpanda, so we can discuss it as a demo on Sunday. On the other hand, we can also test it under load and report the results.

I used _k6_ for load testing.

~~Regarding in-memory databases, I think Postgres might not be a bad idea, but I need further investigation.~~

In the first step, an implementation was done using the Go language. In this implementation, Redpanda was used, and only messages were exchanged between the consumer and producer. The results were very promising. However, after adding Postgres, the results were not good at all. There are several issues with using Postgres:

- The Postgres engine cannot operate in-memory.
- By default, insert commands in Postgres are very slow.

Another important issue here is, the events that we failed to process properly due to errors. Can we resend these events back to Redpanda?

Since using Postgres did not go well, I decided to review other databases for in-memory usage. Among the available options, MongoDB, CouchDB, and Influx are very suitable, but all require payment for in-memory use.

If we want to use time-series databases, there are not many free options. QuestDB is an open-source option, but unfortunately, its high-availability (HA) features require payment.

## Load Test

- Producer load testing for inserting in Redpanda.
- Producer load testing for inserting in Kafka.
- Consumer load testing for inserting in Postgres:
  - In memory (?)
  - Using Disk
- Consumer load testing for inserting in Redis

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
