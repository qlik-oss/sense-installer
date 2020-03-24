##Preflight checks
Preflight checks provide pre-installation cluster conformance testing and validation before we install qliksense on the cluster. We gather a suite of conformance tests that can be easily written and run on the target cluster to verify that cluster-specific requirements are met.
The suite consists of a set of `collectors` which run the specifications of every test and `analyzers` which analyze the results of every test run by the collector.
We support the following tests at the moment as part of preflight checks, and the range of the suite will be expanded in future.

Run the following command to view help about the commands supported by preflight at any moment:
```console
$ qliksense preflight
perform preflight checks on the cluster

Usage:
  qliksense preflight [command]

Examples:
qliksense preflight <preflight_check_to_run>

Available Commands:
  all         perform all checks
  dns         perform preflight dns check
  k8s-version check k8s version

Flags:
  -h, --help   help for preflight
```

### DNS check
Run the following command to perform preflight DNS check. We setup a kubernetes deployment and try to reach it as part of establishing DNS connectivity in this check. 
The expected output is also shown below.
```console
$ qliksense preflight dns

Creating resources to run preflight checks
deployment.apps/qnginx001 created
service/qnginx001 created
pod/qnginx001-6db5fc95c5-s9sl2 condition met
  Running Preflight checks ⠇
--- PASS DNS check
      --- DNS check passed
--- PASS   cluster-preflight-checks
PASS

DNS check completed, cleaning up resources now
service "qnginx001" deleted
deployment.extensions "qnginx001" deleted

```

### Kubernetes version check
We check the version of the target kubernetes cluster and ensure that it falls in the valid range of kubernetes versions that are supported by qliksense. 
The command to run this check and the expected output are as shown below:
```console
$ qliksense preflight k8s-version

Minimum Kubernetes version supported: 1.11.0
Client Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.3", GitCommit:"06ad960bfd03b39c8310aaf92d1e7c12ce618213", GitTreeState:"clean", BuildDate:"2020-02-13T18:08:14Z", GoVersion:"go1.13.8", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.5", GitCommit:"20c265fef0741dd71a66480e35bd69f18351daea", GitTreeState:"clean", BuildDate:"2019-10-15T19:07:57Z", GoVersion:"go1.12.10", Compiler:"gc", Platform:"linux/amd64"}
  Running Preflight checks ⠇
--- PASS Required Kubernetes Version
      --- Good to go.
--- PASS   cluster-preflight-checks
PASS

Minimum kubernetes version check completed
```

### Running all checks
Run the command below to execute all preflight checks.
```console
$ qliksense preflight all

Running all preflight checks

Running DNS check...
Creating resources to run preflight checks
deployment.apps/qnginx001 created
service/qnginx001 created
pod/qnginx001-6db5fc95c5-grwv2 condition met
  Running Preflight checks ⠇
--- PASS DNS check
      --- DNS check passed
--- PASS   cluster-preflight-checks
PASS

DNS check completed, cleaning up resources now
service "qnginx001" deleted
deployment.extensions "qnginx001" deleted

Running minimum kubernetes version check...
Minimum Kubernetes version supported: 1.11.0
Client Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.3", GitCommit:"06ad960bfd03b39c8310aaf92d1e7c12ce618213", GitTreeState:"clean", BuildDate:"2020-02-13T18:08:14Z", GoVersion:"go1.13.8", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.5", GitCommit:"20c265fef0741dd71a66480e35bd69f18351daea", GitTreeState:"clean", BuildDate:"2019-10-15T19:07:57Z", GoVersion:"go1.12.10", Compiler:"gc", Platform:"linux/amd64"}
  Running Preflight checks ⠧
--- PASS Required Kubernetes Version
      --- Good to go.
--- PASS   cluster-preflight-checks
PASS

Minimum kubernetes version check completed
Completed running all preflight checks
```