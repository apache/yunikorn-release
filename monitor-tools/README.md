<!--
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to you under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Tool for Monitoring During Job Deployment

## Description
This command-line tool is crafted for monitoring tasks while jobs are being deployed. It enables users to track the number of pods in a designated namespace, recording either throughput or resource usage metrics. Upon reaching the expected pod count, the tool saves these metrics as a .csv file. Interacting with Prometheus for data collection, it also provides customization options.

We're using these queries to keep an eye on certain things:

Throughput - We monitor how fast pods are getting scheduled using `count(kube_pod_status_scheduled_time)`.

CPU Usage - We track the total CPU usage over time using `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate)`.

Memory Usage - We watch the amount of memory pods are actively using with `sum(node_namespace_pod_container:container_memory_working_set_bytes)`.

## Installation
The tool is a Python script. To use it, follow these steps:
1. Clone the repository or download the tool script.
2. Ensure Python is installed on your system.

## Usage
The tool provides command-line options for customization:
```
$ python3 ./perf-monitor.py -h
usage: perf-monitor.py [-h] [-n] [-o] [-u] [-t] [-m] podCount

Usage: Tools for monitoring during job deployment

positional arguments:
  podCount              Estimated number of pods in the namespace

options:
  -h, --help            show this help message and exit
  -n , --namespace      Monitors the namespace for the pod count (default is 'default')
  -o , --output-path    Specifies the location of the output CSV file (default is ./result.csv)
  -u , --prometheus-url 
                        Sets the URL for Prometheus operation (default is http://localhost:9090)
  -t , --monitor-type   Specifies the type of monitoring indicators (throughput|resource-usage, default is 'throughput')
  -m , --master-node    Defines the master node for monitoring in the case of 'resource-usage' (required when resource-usage is selected).
```

# Example
## Throughput

1. Ensure Prometheus accessibility:
    Before beginning, confirm that Prometheus is accessible. This tool relies on data from Prometheus, so ensure it is reachable. By default, the tool uses the URL http://localhost:9090. If you wish to modify this URL, use the `-u, --prometheus-url` option to set a different one.
    ```
    # add helm repo
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    # create k8s namespace
    kubectl create namespace prometheus
    # deploy chart
    helm install prometheus prometheus-community/kube-prometheus-stack -n prometheus -f /tmp/values.yaml
    kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
    ```

2. Start the monitoring tool:
    Run `python3 ./perf-monitor.py $podCount` to initiate the tool. It observes the performance metrics and records the throughput if the number of pods in the default namespace surpasses the value of $podCount.
    ```
    $ python3 ./perf-monitor.py 5000
    Check prometheus URL...
    Prometheus is accessible
    ============================================================================
    Wait for pods count in "default" is higher then expect count: 5000
    [21:23:19] Number of pods in "default" namespace: 0
    [21:23:22] Number of pods in "default" namespace: 0
    ```

3. Deploy pods and check their count:
    In a separate terminal, add pods to the default namespace. Confirm that the number of pods you add in the YAML file matches or exceeds the number you've set with $podCount in the tool. This ensures the tool monitors performance when the specified number or more pods are active.
    ```
    $ cat ./default-job.yaml
        apiVersion: apps/v1
        kind: Deployment
        ...
        spec:
            replicas: 5000
        ...
    $ kubectl apply -f ./default-job.yaml
    ```

4. Wait for the pod count in the "default" namespace equal to the expected count.
    ```
    $ python3 ./perf-monitor.py 5000
    Check prometheus URL...
    Prometheus is accessible
    ============================================================================
    Wait for pods count in "default" is higher then expect count: 5000
    [21:23:19] Number of pods in "default" namespace: 0
    [21:23:22] Number of pods in "default" namespace: 0
    [21:23:25] Number of pods in "default" namespace: 3570
    [21:23:29] Number of pods in "default" namespace: 5000
    The pod count in the "default" namespace is 5000, which exceeds the expected count.
    ============================================================================
    Collecting "throughput" during scheduling.
    The file is saved in ./result/output-gnrhza.csv
    ```

5. Check output!
    ```$ cat ./result/output-gnrhza.csv
    Timestamp,Value
    1698931401.402,0
    1698931402.402,1020
    1698931403.402,2040
    1698931404.402,3570
    1698931405.402,4590
    1698931406.402,4590
    1698931407.402,5100
    ```
## Resource Usage
The process of monitoring resource usage is the same as monitoring throughput, with the only difference being the requirement to specify the master node using `-m, --master-node`.

```
$ python3 ./perf-monitor.py 5000 -t resource-usage -m ws980t
Check prometheus URL...
Prometheus is accessible
============================================================================
Wait for pods count in "default" is higher then expect count: 5000
[21:35:39] Number of pods in "default" namespace: 0
[21:35:42] Number of pods in "default" namespace: 0
[21:35:46] Number of pods in "default" namespace: 380
[21:35:49] Number of pods in "default" namespace: 4489
[21:35:52] Number of pods in "default" namespace: 5000
The pod count in the "default" namespace is 5000, which exceeds the expected count.
============================================================================
Collecting "resource usage" during scheduling.
The file is saved in ./result/output-fklhtv.csv
```

### Output
```
$ cat ./result/output-fklhtv.csv
Timestamp,Pod count,kube-system-cpu-usage,kube-system-memory-usage,yunikorn-cpu-usage,yunikorn-memory-usage
1698932143.422,0,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932144.422,380,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932145.422,1566,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932146.422,1566,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932147.422,3339,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932148.422,4489,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932149.422,5000,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932150.422,5000,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932151.422,5000,0.36333826465717495,3399712768,0.028113757806607527,613601280
1698932152.422,5000,0.36333826465717495,3399712768,0.028113757806607527,613601280
```