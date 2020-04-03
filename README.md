# sanepa

The sane Kubernetes HPA

`sanepa` checks pod resource usage against `deployment.spec.template.spec.containers.resources.limits` values and will scale up and down when a pod is above a set threshold.

This project follows AWS EKS releases, which are well behind Kubernetes releases.

## Running

You can find kubernetes examples in `examples/k8s/`.

Do **not** run multiple SanePA deployments that watch the same deployment.

One SanePA container should watch one deployment. SanePA has a very small footprint, so resource consumption shouldn't be a concern

### Notes

`sanepa` runs its checks every ten seconds. You should take that into consideration when setting `-cooldown` and `-scaledownok` arg values.

## Flags

`-incluster`: Defaults to `true`. Set `-incluster=false` to run outside of cluster. Defaults to `true`.

`-ns`: The namespace to check pods and deployments. Defaults to `""`.

`-dep`: The deployment name to check. Set to `none` to skip checking deployments. Defaults to `""`.

`-cpu`: The percentage of the container CPU limit to scale on. If your container has a limit of `100m` and you set this `-cpu=10`, a scaling event will occur when a container hits 10 mCPU. Defaults to 50.

`-mem`: Same idea as CPU. Defaults to 70.

`-max`: The maximum number of replicas in the deployment. Defaults to 5.

`-min`: The minimum number of replicas in the deployment. Defaults to 1.

`-cooldown`: How much time should pass after a scale up event before checking again. If your application takes 90 seconds to start, you should set this to at least 90.

`-scaledownok`: How many times must all pods be under thresholds before scaling down.

`-scaleupok`: How many times must a pod be above thresholds before scaling up.

## TODO

- Cleanup codebase
- Use a real logger
- Handle all memory and CPU suffixes
  - Details in `kubectl explain PodMetrics.containers.usage`
