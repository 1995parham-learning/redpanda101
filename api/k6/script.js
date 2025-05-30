import http from "k6/http";
import exec from "k6/execution";
import { check, group, sleep } from "k6";

const baseURL = "http://127.0.0.1:1378";

export let options = {
  stages: [
    {
      target: 10,
      duration: "2m",
    },
    {
      target: 10,
      duration: "5m",
    },
  ],
  thresholds: {
    http_req_duration: ["avg<10000", "p(100)<30000"],
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  group("publish", () => {
    let payload = JSON.stringify({
      channel: 0,
      description: `hello world ${exec.vu.iterationInInstance}`,
      dst_currency: 0,
      src_currency: 0,
    });

    let res = http.post(`${baseURL}/orders/`, payload, {
      headers: {
        "Content-Type": "application/json",
      },
    });

    check(res, {
      success: (res) => res.status == 200,
    });

    sleep(0.001);
  });
}
