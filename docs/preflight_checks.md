# Preflight checks

Preflight checks provide pre-installation cluster conformance testing and validation before we install qliksense on the cluster. We gather a suite of conformance tests that can be easily written and run on the target cluster to verify that cluster-specific requirements are met.

We support the following tests at the moment as part of preflight checks, and the range of the suite will be expanded in future.

Run the following command to view help about the commands supported by preflight at any moment:
```shell
$ qliksense preflight
perform preflight checks on the cluster

Usage:
  qliksense preflight [command]

Examples:
qliksense preflight <preflight_check_to_run>

Available Commands:
  all            perform all checks
  authcheck      preflight authcheck
  clean          perform preflight clean
  deployment     perform preflight deployment check
  dns            perform preflight dns check
  kube-version   check kubernetes version
  mongo          preflight mongo OR preflight mongo --url=<url>
  pod            perform preflight pod check
  role           preflight create role check
  rolebinding    preflight create rolebinding check
  service        perform preflight service check
  serviceaccount preflight create ServiceAccount check

Flags:
  -h, --help   help for preflight
  -v, --verbose   verbose mode
```

### DNS check
Run the following command to perform preflight DNS check. We setup a kubernetes deployment and try to reach it as part of establishing DNS connectivity in this check. 
The expected output should be similar to the one shown below.
```shell
$ qliksense preflight dns -v

Preflight DNS check
---------------------
Created deployment "dep-dns-preflight-check"
Created service "svc-dns-pf-check"
Created pod: pf-pod-1
Fetching pod: pf-pod-1
Fetching pod: pf-pod-1
Exec-ing into the container...
Preflight DNS check: PASSED
Completed preflight DNS check
Cleaning up resources...
Deleted pod: pf-pod-1
Deleted service: svc-dns-pf-check
Deleted deployment: dep-dns-preflight-check

```

### Kubernetes version check
We check the version of the target kubernetes cluster and ensure that it falls in the valid range of kubernetes versions that are supported by qliksense. 
The command to run this check and the expected similar output are as shown below:
```shell
$ qliksense preflight k8s-version -v

Preflight kubernetes minimum version check
------------------------------------------
Kubernetes API Server version: v1.15.5
Current K8s Version: 1.15.5
Current 1.15.5 is greater than minimum required version:1.11.0, hence good to go
Preflight minimum kubernetes version check: PASSED
Completed Preflight kubernetes minimum version check

```

### Service check
We use the commmand below to test if we are able to create a service in the cluster.
```shell
$ qliksense preflight service -v

Preflight service check
-----------------------

Preflight service check: 
Created service "svc-pf-check"
Preflight service creation check: PASSED
Cleaning up resources...
Deleted service: svc-pf-check
Completed preflight service check
```

### Deployment check
We use the commmand below to test if we are able to create a deployment in the cluster. After the test exexutes, we wait until the created deployment terminates before we exit the command. 
```shell
$ qliksense preflight deployment -v

Preflight deployment check
-----------------------
Preflight deployment check: 
Created deployment "deployment-preflight-check"
Preflight Deployment check: PASSED
Cleaning up resources...
Deleted deployment: deployment-preflight-check
Completed preflight deployment check
```

### Pod check
We use the commmand below to test if we are able to create a pod in the cluster.
```shell
$ qliksense preflight pod -v

Preflight pod check
--------------------

Preflight pod check: 
Created pod: pod-pf-check
Preflight pod creation check: PASSED
Cleaning up resources...
Deleted pod: pod-pf-check
Completed preflight pod check
```

### Role check
We use the command below to test if we are able to create a role in the cluster
```shell
$ qliksense preflight role -v
Preflight role check
---------------------------
Preflight role check: 
Created role: role-preflight-check
Preflight role check: PASSED
Cleaning up resources...
Deleted role: role-preflight-check

Completed preflight role check
```

### RoleBinding check
We use the command below to test if we are able to create a role binding in the cluster
```shell
$ qliksense preflight rolebinding -v

Preflight rolebinding check
---------------------------
Preflight rolebinding check: 
Created RoleBinding: role-binding-preflight-check
Preflight rolebinding check: PASSED
Cleaning up resources...
Deleting RoleBinding: role-binding-preflight-check
Deleted RoleBinding: role-binding-preflight-check

Completed preflight rolebinding check
```

### Create-ServiceAccount check
We use the command below to test if we are able to create a service account in the cluster
```shell
$ qliksense preflight serviceaccount -v

Preflight ServiceAccount check
-------------------------------------
Preflight serviceaccount check: 
Created Service Account: preflight-check-test-serviceaccount
Preflight serviceaccount check: PASSED
Cleaning up resources...
Deleting ServiceAccount: preflight-check-test-serviceaccount
Deleted ServiceAccount: preflight-check-test-serviceaccount

Completed preflight serviceaccount check
```

### Auth check
We use the command below to combine creation of role, role binding, and service account tests
```shell
$ qliksense preflight authcheck -v

Preflight auth check
-------------------------------------
Preflight create-role check: 
Created role: role-preflight-check
Preflight create-role check: PASSED
Cleaning up resources...
Deleted role: role-preflight-check

Completed preflight create-role check

Preflight create RoleBinding check: 
Created RoleBinding: role-binding-preflight-check
Preflight create RoleBinding check: PASSED
Cleaning up resources...
Deleted RoleBinding: role-binding-preflight-check

Completed preflight create RoleBinding check

Preflight createServiceAccount check: 
Created Service Account: preflight-check-test-serviceaccount
Preflight createServiceAccount check: PASSED
Cleaning up resources...
Deleted ServiceAccount: preflight-check-test-serviceaccount

Completed preflight createServiceAccount check
Completed preflight auth check
```

### Mongodb check
We can check if we are able to connect to an instance of mongodb on the cluster by either supplying the mongodbUri as part of the command or infer it from the current context.

```shell
qliksense preflight mongo --url=<url> -v OR
qliksense preflight mongo -v
qliksense preflight mongo --url=<mongo-server url> --ca-cert=<path to ca-cert file> -v
```
```shell
Preflight mongo check
---------------------
Preflight mongodb check: 
Created pod: pf-mongo-pod
stdout: MongoDB shell version v4.2.5
connecting to: <url>/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("...") }
MongoDB server version: 4.2.5
bye
stderr: 
Preflight mongo check: PASSED
Deleted pod: pf-mongo-pod
Completed preflight mongodb check
```

#### Mongodb check with mutual tls
In order to perform mutual tls with mongo we need to:
- append client certificate to the beginning/end of CA certificate. Make sure to include the beginning and end tags on each certificate. 
The CA certificate file should look like this in the end:
```shell
<existing contents of CA cert>
...
-----BEGIN RSA PRIVATE KEY-----
<private key>
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
<public key>
-----END CERTIFICATE-----
```
- Run the command below to set the ca certificate into the CR
```shell
cat <path_to_ca.crt> | base64 | qliksense config set-secrets qliksense.caCertificates --base64
```

Next, run: 
```shell
qliksense preflight mongo -v
```

### Running all checks
Run the command shown below to execute all preflight checks.
```shell
$ qliksense preflight all --mongodb-url=<url> -v OR
$ qliksense preflight all --mongodb-url=<mongo-server url> --mongodb-ca-cert=<path to ca-cert file> -v

Running all preflight checks

Preflight DNS check
-------------------
Created deployment "dep-dns-preflight-check"
Created service "svc-dns-pf-check"
Created pod: pf-pod-1
Fetching pod: pf-pod-1
Fetching pod: pf-pod-1
Exec-ing into the container...
Preflight DNS check: PASSED
Completed preflight DNS check
Cleaning up resources...
Deleted pod: pf-pod-1
Deleted service: svc-dns-pf-check
Deleted deployment: dep-dns-preflight-check

Preflight kubernetes minimum version check
------------------------------------------
Kubernetes API Server version: v1.15.5
Current K8s Version: 1.15.5
Current 1.15.5 is greater than minimum required version:1.11.0, hence good to go
Preflight minimum kubernetes version check: PASSED
Completed Preflight kubernetes minimum version check
...
...
All preflight checks have PASSED
Completed running all preflight checks

```

### Clean
Run the command below to cleanup entities that were created for the purpose of running preflight checks and left behind in the cluster.
```shell
$ qliksense preflight clean -v

Preflight clean
----------------
Removing deployment...
Removing service...
Removing pod...
Removing role...
Removing rolebinding...
Removing serviceaccount...
Removing DNS check components...
Removing mongo check components...
Preflight cleanup complete

```
