import requests
import csv
import time
import datetime
import signal
import os
import random
import string
import argparse

interval = 1 # second
timeout_sec = 60 * 10

class Config:
    pods_num = 0
    namespace = 'default'
    prom_url = 'http://localhost:9090'
    output_path = './result'
    monitor_type = 'throughput'
    master_node = ''

class RecordStrategy:
    def parse(self):
        pass
    def save_result(self, result):
        pass

class RecordThroughputStrategy(RecordStrategy):
    def parse(self, start, end):
        # Construct a query to count the scheduled pods in a specific namespace
        query = 'count(kube_pod_status_scheduled_time{namespace="'+Config.namespace+'"}) by (namespace)'
        params = {
            'query': query,
            'start': start,
            'end': end,
            'step': interval,
        }

        # Make a GET request to Prometheus to get the data
        response = requests.get(f'{Config.prom_url}/api/v1/query_range', params=params)

        if response.status_code == 200:
            data = response.json()
            result = data.get('data', {}).get('result', [])
            if result:
                # Return the values of scheduled pods if available
                return result[0]['values']
            else:
                print('No data returned by the query')
        else:
            print(f'Error: {response.status_code}, {response.text}')  # Display error in case of HTTP status code error
    
    def save_result(self, result):
        # Save the results to a CSV file
        with open(Config.output_path, 'w+', newline='') as csv_file:
            csv_writer = csv.writer(csv_file)
            csv_writer.writerow(['Timestamp', 'Value'])  # Write headers to the CSV file
            firstTimstamp = result[0][0]-interval
            result.insert(0, [firstTimstamp, '0'])
            for item in result:
                _, value = item
                csv_writer.writerow(item)  # Write data to the CSV file
                if int(value)>=Config.pods_num:
                    break  # Break the loop if the pod count exceeds the configured pods_num

            print(f'The file is saved in {Config.output_path}')

class RecordResourceUsageStrategy(RecordStrategy):
    def parse(self, start, end):
        # Define queries for collecting pod count, CPU, and memory usage
        pods_count_query = 'count(kube_pod_status_scheduled_time{namespace="'+Config.namespace+'"}) by (namespace)'
        cpu_query = 'sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{node="'+ Config.master_node +'", namespace!="prometheus"}) by (namespace)'
        memory_query = 'sum(node_namespace_pod_container:container_memory_working_set_bytes{node="'+ Config.master_node +'", namespace!="prometheus"}) by (namespace)'
        queries = [pods_count_query, cpu_query, memory_query]
        all_result = []
        
        for query in queries:
            params = {
                'query': query,
                'start': start,
                'end': end,
                'step': interval,
            }
            # Make GET requests to Prometheus for each query
            response = requests.get(f'{Config.prom_url}/api/v1/query_range', params=params)
            if response.status_code == 200:
                data = response.json()
                result = data.get('data', {}).get('result', [])
                if result:
                    all_result.append(result)
                else:
                    print('No data returned by the query')
            else:
                print(f'Error: {response.status_code}, {response.text}') # Display error in case of HTTP status code error
        
        return all_result

    def save_result(self, all_result):
        # Define monitor namespaces and CSV headers
        monitor_namespace=['kube-system', 'yunikorn']
        csv_headers=['Timestamp', 'Pod count']

        result={
            'count': all_result[0][0]['values'],
            'cpu': {},
            'memory': {},
        }
        for ns in monitor_namespace:
            cpu = self.getResultByNS(all_result[1], ns)
            memory = self.getResultByNS(all_result[2], ns)
            if cpu and memory:
                # Update CSV headers and results for CPU and memory if data is available
                csv_headers.append(ns+'-cpu-usage')
                csv_headers.append(ns+'-memory-usage')
                result['cpu'][ns]=cpu
                result['memory'][ns]=memory
            else:
                monitor_namespace.remove(ns)  # Remove namespace if data is unavailable

        with open(Config.output_path, 'w+', newline='') as csv_file:
            csv_writer = csv.writer(csv_file)
            csv_writer.writerow(csv_headers)   # Write CSV headers
            firstTimstamp = result['count'][0][0]-interval
            result['count'].insert(0, [firstTimstamp, '0'])
            for _, podCount in enumerate(result['count']):
                timestamp, count = podCount
                row = [timestamp, count]
                for ns in monitor_namespace:
                    # Get CPU and memory usage based on the timestamp
                    cpu=self.getUsageByTimestamp(result['cpu'][ns], timestamp)
                    mem=self.getUsageByTimestamp(result['memory'][ns], timestamp)
                    row.append(cpu)
                    row.append(mem)
                csv_writer.writerow(row)
                    
            print(f'The file is saved in {Config.output_path}')
    
    def getResultByNS(self, dictionary, namespace):
        for _, result in enumerate(dictionary):
            if result['metric']['namespace']==namespace:
                return result['values']
    
    def getUsageByTimestamp(self, dictionary, timestamp):
        for _, result in enumerate(dictionary):
            if result[0]==timestamp:
                return result[1]

class Record:
    def __init__(self, RecordStrategy):
        self.RecordStrategy = RecordStrategy

    def execute_record(self, start, end):
        result=self.RecordStrategy.parse(start, end)
        self.RecordStrategy.save_result(result)

def main():
    cli_display()

    start = time.time()
    # Wait for the number of pods in the specified namespace to be greater than or equal to the expected number of pods.
    print("============================================================================")
    print(f'Wait for pods count in "{Config.namespace}" is higher then expect count: {Config.pods_num}')
    while True:
        pods_count = count_pods()
        current=datetime.datetime.fromtimestamp(time.time()).strftime('%H:%M:%S')
        print(f'[{current}] Number of pods in "{Config.namespace}" namespace: {pods_count}')
        if int(pods_count or 0) >= Config.pods_num:
            break
        time.sleep(3)
    print(f'The pod count in the "{Config.namespace}" namespace is {pods_count}, which exceeds the expected count.')
    print("============================================================================")
    end = time.time()
    
    # Select the appropriate monitoring strategy based on the specified monitor type.
    match Config.monitor_type:
        case 'throughput':
            print('Collecting "throughput" during scheduling.')
            throughputStrategy=RecordThroughputStrategy()
            recorder= Record(throughputStrategy)
        case 'resource-usage':
            print('Collecting "resource usage" during scheduling.')
            resourceUsageStrategy=RecordResourceUsageStrategy()
            recorder= Record(resourceUsageStrategy)
        case _:
            print('The monitor-type:', Config.monitor_type, 'is not supported.')

    recorder.execute_record(start, end)

def cli_display():
    """Displays the command-line interface for the tool."""
    
    parser = argparse.ArgumentParser(description="Usage: Tools for monitoring during job deployment")

    # Required argument
    parser.add_argument("podCount", type=int, help="Estimated number of pods in the namespace")
    
    # Optional arguments
    parser.add_argument("-n", "--namespace", metavar="", help="Monitors the namespace for the pod count (default is 'default')")
    parser.add_argument("-o", "--output-path", metavar="", help="Specifies the location of the output CSV file (default is ./result.csv)")
    parser.add_argument("-u", "--prometheus-url", metavar="", help="Sets the URL for Prometheus operation (default is http://localhost:9090)")
    parser.add_argument("-t", "--monitor-type", choices=["throughput", "resource-usage"], metavar="", help="Specifies the type of monitoring indicators (throughput|resource-usage, default is 'throughput')")
    parser.add_argument('-m', "--master-node", metavar="", help="Defines the master node for monitoring in the case of 'resource-usage' (required when resource-usage is selected).")
    args = parser.parse_args()

    # Set the configuration variables based on the command-line arguments
    Config.pods_num=args.podCount
    if args.namespace:
        Config.namespace=args.namespace

    if args.output_path:
        Config.output_path=args.output_path
    else:
        if not os.path.exists(Config.output_path):
            os.makedirs(Config.output_path)
        rand_seq=''.join(random.choice(string.ascii_lowercase) for x in range(6))
        Config.output_path+="/output-"+rand_seq+".csv"

    if args.prometheus_url:
        Config.prom_url=args.prometheus_url
    check_prometheus()

    if args.monitor_type:
        Config.monitor_type=args.monitor_type
        if args.monitor_type=="resource-usage" and args.master_node:
            Config.master_node=args.master_node
        else:
            print("ERROR: When the monitor type is 'resource-usage', the master node is required (use -m, --master).")
            exit(1)

def check_prometheus():
    """Checks if the Prometheus server is accessible."""
    try:
        print("Check prometheus URL...")
        response = requests.get(f'{Config.prom_url}')
        if response.status_code == 200:
            print('Prometheus is accessible')
    except:  # noqa: E722
        print('ERROR: The Prometheus URL is not accessible.')
        exit(1)

def count_pods():
    """Counts the number of pods in the specified namespace."""
    # Define the query parameters
    params = {
        'query': 'count(kube_pod_status_scheduled_time{namespace="'+Config.namespace+'"}) by (namespace)',
    }

    # Make a GET request to Prometheus
    response = requests.get(f'{Config.prom_url}/api/v1/query', params=params)

    if response.status_code == 200:
        data = response.json()
        result = data.get('data', {}).get('result', [])
        if result:
            pods_count = int(result[0]['value'][1])
            return pods_count
        else:
            return 0
    else:
        print(f'Error: {response.status_code}, {response.text}')

def _handle_timeout(signum, frame):
    raise TimeoutError('function timeout')

if __name__=='__main__':
    # handle timeout
    signal.signal(signal.SIGALRM, _handle_timeout)
    signal.alarm(timeout_sec)
    main()