# sanepa

The sane Kubernetes HPA

## Dependencies

Go Client: `go get k8s.io/client-go@v0.15.7`

K8s API Machinery: `go get k8s.io/apimachinery@v0.15.7`

## Running

## Flags

`-incluster`: Defaults to `true`. Set `-incluster=false` to run outside of cluster.

`-ns`: The namespace to check pods and deployments

`-dep`: The deployment name to check. Set to `none` to skip checking deployments
