# Update Kustomize dependency

To update the kustomize dependency in `go.mod` file, run the following command for the vesion `v0.0.43`

```console
GOPROXY=direct go get github.com/qlik-oss/kustomize/api@qlik/v0.0.43
```

The above command will generate output something like this

```console
go: downloading github.com/qlik-oss/kustomize v3.3.2-0.20200820111149-1a59db58525f+incompatible
go: github.com/qlik-oss/kustomize/api qlik/v0.0.43 => v0.5.2-0.20200820111149-1a59db58525f
go get: github.com/qlik-oss/kustomize/api@v0.5.2-0.20200820111149-1a59db58525f: parsing go.mod:
	module declares its path as: sigs.k8s.io/kustomize/api
	        but was required as: github.com/qlik-oss/kustomize/api
```

So in the `go.mod` file write

```console
sigs.k8s.io/kustomize/api => github.com/qlik-oss/kustomize/api v0.5.2-0.20200820111149-1a59db58525f
```
