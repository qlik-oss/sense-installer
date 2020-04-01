# CLI reference

### qliksense apply

`qliksense apply` command takes input from a file or from pipe

- `qliksense apply -f cr-file.yaml`
- `cat cr-file.yaml | qliksense apply -f -`

The content of `cr-file.yaml` should be something like the following:

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

`qliksense apply` does everything `qliksense load` does but will install Qlik Sense into the cluster as well

### qliksense load

`qliksense load` command takes input from a file or from pipe

- `qliksense load -f cr-file.yaml`
- `cat cr-file.yaml | qliksense load -f -`

This will load the Custom Resource (CR) into `${QLIKSENSE_HOME}` folder, create context structure and set the current context to that CR.

This will also encrypt the secrets from CR while writing the CR into the disk.

### qliksense about

`qliksense about` command will display information about [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) release.

It supports the following flags:

- `qliksense about 1.0.0` display default profile for tag `1.0.0`.
- `qliksense about 1.0.0 --profile=docker-desktop`
- If `qliksense about` is ran without flags, then it displays
    - Information of the release defined in `manifests/docker-desktop` (if it exist), otherwise
    - Get version information from `master` branch in `qliksense-k8s` repository

Using other supported commands user might have built the CR into the location `~/.qliksense/myqliksense.yaml`

```yaml
apiVersion: qlik.com/v1
kind: QlikSense
metadata:
  name: myqliksense
spec:
  profile: docker-desktop
  manifestsRoot: /Usr/xyz/my-k8-repo/manifests
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

In this case, the result of `qliksense about` command would display information from:

- `/Usr/xyz/my-k8-repo/manifests/docker-desktop` location, or
- Pull and show information from `master` branch if the directory is invalid or empty


### qliksense config

`qliksense config` will perform operations on configurations and contexts regarding the [qliksense-k8](https://github.com/qlik-oss/qliksense-k8s) release.

It supports the following flags:

- `qliksense config apply` - generate the patches and apply manifests to K8s
- `qliksense config list-contexts` - get and list contexts
- `qliksense config set` - configure a key-value pair into the current context
- `qliksense config set-configs` - set configurations into qliksense context as key-value pairs
- `qliksense config set-context` - sets the Kubernetes context where resources are located
- `qliksense config set-secrets <service_name>.<attribute>="<value>" --secret=false` - set secrets configurations into qliksense context as key-value pairs and show encrypted value as part of CR
- `qliksense config set-secrets <service_name>.<attribute>="<value>" --secret=true` - set secrets configurations into qliksense context as key-value pairs and show a key reference to the created Kubernetes secret resource as part of the CR
- `qliksense config view` - view the qliksense operator CR
- `qliksense config delete-context` - deletes a specific context locally (not in-cluster). Deletes context in spec of `config.yaml` and locally deletes entire folder of specified context (does not delete secrets from cluster)


The global file which abstracts all contexts is `~/.qliksense/config.yaml`
```yaml
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: QliksenseConfigMetadata
spec:
  contexts:
  - name: qlik-default
    crFile: /Users/xyz/.qliksense/contexts/qlik-default/qlik-default.yaml
  - name: myqliksense
    crFile: /Users/xyz/.qliksense/contexts/myqliksense/myqliksense.yaml
  - name: hello
    crFile: /Users/xyz/.qliksense/contexts/hello/hello.yaml
  currentContext: hello
```
