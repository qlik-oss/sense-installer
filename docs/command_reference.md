# qliksense command reference

## qliksense apply

`qliksense apply` command takes input a cr file

- `qliksense apply -f cr-file.yaml`

the content of `cr-file.yaml` should be something similar

```yaml
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-test
  labels:
    version: v0.0.2
spec:
  configs:
    qliksense:
    - name: acceptEULA
      value: "yes"
  secrets:
    qliksense:
    - name: mongoDbUri
      value: mongodb://qlik-test-mongodb:27017/qliksense?ssl=false
  profile: docker-desktop
  rotateKeys: "yes"
  ```

After executing the above command, cli will set the current context to the cr name and install the qliksense into the cluster. 
Make sure you dont have a context (cr name = context name) with the same name already. It will though an error if the same context name already exists. 
It will encrypt the secrets from CR while writing CR into the disk.

## qliksense load

`qliksense load` command takes input a cr file.

- `qliksense load -f cr-file.yaml`

Executing the above command will load the cr into `~/.qliksense` folder and create context structure with for the CR and set the current context with the cr name. 
Make sure you dont have a context (cr name = context name) with the same name already. It will though an error if the same context name already exists. 
It will encrypt the secrets from CR while writing the CR into the disk.

## qliksense about

About action will display inside information regarding [qliksense-k8](https://github.com/qlik-oss/qliksense-k8s) release.

it will support following flags

- `qliksense about 1.0.0` display default profile for tag `1.0.0`.
- `qliksense about 1.0.0 --profile=docker-desktop`
- `qliksense about` 
  - assuming current directory has `manifests/docker-desktop`
  - or get version information from pull of `qliksense-k8s` `master`

using other supported commands user might have built the CR into the location `~/.qliksense/myqliksense.yaml`

```yaml
apiVersion: qlik.com/v1
kind: QlikSense
metadata:
  name: myqliksense
spec:
  profile: docker-desktop
  manifestsRoot: /Usr/ddd/my-k8-repo/manifests
  namespace: myqliksense
  storageClassName: efs
  configs:
    qliksense:
    - name: acceptEULA
      value: "yes"
  secrets:
    qliksense:
    - name: mongoDbUri
      value: "mongo://mongo:3307"
    - name: messagingPassword
      valueFromKey: messagingPassword
```

In that case the command would be

- `qliksense about`
   - display from `/Usr/ddd/my-k8-repo/manifests/docker-desktop` location
   - pull from `master` if directory invalid/empty


## qliksense config

Config action will perform operations on configurations and contexts regarding the [qliksense-k8](https://github.com/qlik-oss/qliksense-k8s) release.

it will support following commands:

- `qliksense config apply` - generate the patchs and apply manifests to k8s
- `qliksense config list-contexts` - retrieves the contexts and lists them
- `qliksense config set` - configure a key value pair into the current context
- `qliksense config set-configs` - set configurations into the qliksense context as key-value pairs
- `qliksense config set-context` - sets the context in which the Kubernetes cluster and resources live in
- `qliksense config set-secrets <service_name>.<attribute>="<value>" --secret=false` - set secrets configurations into the qliksense context as key-value pairs and show encrypted value as part of CR
- `qliksense config set-secrets <service_name>.<attribute>="<value>" --secret=true` - set secrets configurations into the qliksense context as key-value pairs and show a key reference to the created Kubernetes secret resource as part of the CR
- `qliksense config view` - view the qliksense operator CR
- `qliksense config delete-context` - deletes a specific context locally (not in-cluster). Deletes context in spec of `config.yaml` and locally deletes entire folder of specified context (does not delete in-cluster secrets)


the global file that abstracts all the contexts is `config.yaml`, located at:  `~/.qliksense/config.yaml`:
```yaml
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: QliksenseConfigMetadata
spec:
  contexts:
  - name: qlik-default
    crFile: /Users/fff/.qliksense/contexts/qlik-default/qlik-default.yaml
  - name: myqliksense
    crFile: /Users/fff/.qliksense/contexts/myqliksense/myqliksense.yaml
  - name: hello
    crFile: /Users/fff/.qliksense/contexts/hello/hello.yaml
  currentContext: hello
```