# qliksense about

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
