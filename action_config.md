# qliksense config

Config action will perform operations on configurations and contexts regarding the [qliksense-k8](https://github.com/qlik-oss/qliksense-k8s) release.

it will support following commands:

- `qliksense config apply` - generate the patchs and apply manifests to k8s
- `qliksense config list-contexts` - retrieves the contexts and lists them
- `qliksense config set` - configure a key value pair into the current context
- `qliksense config set-configs` - set configurations into the qliksense context
- `qliksense config set-context` - sets the context in which the Kubernetes cluster and resources live in
- `qliksense config set-secrets` - set secrets configurations into the qliksense context
- `qliksense config set-secrets` - view the qliksense operator CR
- `qliksense config delete-context` - deletes a specific context locally (not in-cluster)
                                    - deletes context in spec of `config.yaml` and locally deletes entire folder of specified context (does not delete in-cluster secrets)

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

