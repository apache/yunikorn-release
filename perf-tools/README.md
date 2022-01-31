# Performance tool
It contains following performance scenarios.
1. Throughput
2. Node-fairness

## How to work
1. Run `go mod tidy` to pull and check related package.(one time)
2. In `conf.yaml`,remove `#` before `- yunikorn` in scenarios to run certain scenarios with yunikorn.
3. Use `go run main.go` to start performance tests.
## Experiment setting
### Common
| Key word		| Description				| Default	|
| ---			| ---					| ---		|
| `queue`		| The queue contains applications	| root.default	|
| `outputrootpath`	| Output path				| /tmp		|
### Scenarios
Append scheduler names to `schedulerNames` to add performance test with schedulers in the certain scenario.
A scenario can contain one or more performance cases.
The parameters in a scenario is following.
| Key word              | Description                           |
| ---                   | ---                                   |
| schedulerNames	| Schedulers in the scenario		|
| `numPods`		| Pod number in a deployment		|
| `repeat`		| Deployments number in a case		|
