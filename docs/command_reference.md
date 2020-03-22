# qliksense command reference

## qliksense apply

`qliksense apply` command takes input a cr file or input from pipe

- `qliksense apply -f cr-file.yaml`
- `cat cr-file.yaml | qliksense apply -f -`

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

after doing one of the above commands, cli will set the current context to the cr name and install the qliksense into the cluster. so make sure you dont have a context (cr name = context name) with the same name already. It will though error if it same context name already exist.
