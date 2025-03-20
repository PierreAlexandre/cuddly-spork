## Background

This is based on a real-world problem we encountered early on in EdgeDB Cloud
(but to be clear, we have a working solution for this already in place, we are
not asking you to do any work on our actual cloud as part of this interview
process)

## Problem overview

We run Consul in our internal cloud infrastructure.

Consul limits the number of open HTTP connections from clients,
[defaulting to 200](https://developer.hashicorp.com/consul/docs/agent/config/config-files#http_max_conns_per_client).

Consul also [exposes telemetry](https://developer.hashicorp.com/consul/docs/agent/monitor/telemetry)
in standard Prometheus format.

However, "current count of open HTTP connections" is not part of Consul's
telemetry, so we have no way of knowing how close we are to this 200-connection
limit.

The goal here is to gather this count of open connections ourselves and send it
to our Prometheus metrics server, where we can graph it or alert on it alongside
the other metrics that Consul exposes natively.

## Technical details

- All of the connections that we care about are on TCP port 8500 (Consul's
  primary service port).

- All of the connections we currently have are using IPv4, but we try to leave
  ourselves open to IPv6 compatibility. It's up to you whether you want to
  support IPv6 or leave that as a future TODO.

- We run the Prometheus node_exporter on all hosts that run Consul, and have
  its [textfile collector](https://github.com/prometheus/node_exporter?tab=readme-ov-file#textfile-collector)
  enabled with `--collector.textfile.directory=/tmp/node-exporter`. It's up to
  you whether you want to write a full Prometheus metrics collector
  implementation, or write to the node_exporter's textfile directory.

- It's up to you how to collect the metric value itself - calling `netstat` and
  looking at its output, as in the example below; implementing it yourself by
  looking at files under /proc or /sys; using a 3rd-party library that exposes
  the value, etc.

- We've provided a test script in Python, but you **must** write your solution in Go.

## Working example

Setting up a full Consul installation to repro this problem would be non-trivial
and outside the scope of this interview question, so we have provided a simple
Python script that simulates the behavior, by opening a server on a specified
port, then opening a specified number of client connections to the server, and
holding them open until the script is killed.

The script should run on any Python higher than 3.7 and uses the stdlib only
(does not require a virtualenv or `pip install` or anything similar).

```
$ ./port-opener.py -h
usage: port-opener.py [-h] [--port PORT] [--num-connections NUM_CONNECTIONS] [--ipv {4,6}] [--verbose]

options:
  -h, --help          show this help message and exit
  --port PORT, -p PORT  port to listen and make connections on
  --num-connections NUM_CONNECTIONS, -n NUM_CONNECTIONS
                      number of connections to open
  --ipv {4,6}         IP version (4 or 6) to use
  --verbose, -v       enable verbose logging
```

To see it in action:

```
$ ./port-opener.py
started server on 127.0.0.1:8500
opened 200 client connections

```

Then, in a separate terminal window:

```
$ netstat -tn
Active Internet connections (w/o servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
...
tcp        0      0 127.0.0.1:8500          127.0.0.1:34980         ESTABLISHED
tcp        0      0 127.0.0.1:8500          127.0.0.1:35006         ESTABLISHED
...
```

With those 200 connections open, your solution to this problem should emit a
Prometheus metric that looks something like `consul_open_http_connections 200`
or `open_tcp_conns{port=8500} 200` or something similar.

Note that if you run the port-opener script and then kill it, this will close
the connections, but for a few minutes afterwards netstat will still list them
in `TIME_WAIT` state:

```
> netstat -tn
Active Internet connections (w/o servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
...
tcp        0      0 127.0.0.1:55036         127.0.0.1:8500          TIME_WAIT
tcp        0      0 127.0.0.1:54748         127.0.0.1:8500          TIME_WAIT
tcp        0      0 127.0.0.1:54690         127.0.0.1:8500          TIME_WAIT
...
```

We don't care about any `TIME_WAIT` connections, because from Consul's point of
view they don't count towards the 200-connection limit.

## Your solution

Send us:

- Your code implementing the Prometheus metric collection, in **Go**

- A readme with:

  - How to run your code, if there's any non-obvious steps

  - Any design decisions or tradeoffs you made

  - Any test files or scripts you wrote, or modifications to our port-opener
    script

- This take-home question is also new-ish for us, so we would also appreciate
  any feedback you have about:

  - Any challenges you had understanding our description of the problem, or
    getting our test script to run, etc

  - Any feedback you have on this question that we can use to improve it for
    other candidates
