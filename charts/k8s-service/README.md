# Kubernetes Service Helm Chart

This Helm Chart can be used to deploy your application container under a
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) resource onto your Kubernetes
cluster. You can use this Helm Chart to run and deploy a long-running container, such as a web service or backend
microservice. The container will be packaged into
[Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/) that are managed by the `Deployment`
controller.

This Helm Chart can also be used to front the `Pods` of the `Deployment` resource with a
[Service](https://kubernetes.io/docs/concepts/services-networking/service/) to provide a stable endpoint to access the
`Pods`, as well as load balance traffic to them. The Helm Chart can also specify
[Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) rules to further configure complex routing
rules in front of the `Service`.


## How to use this chart?

* See the [root README](/README.md) for general instructions on using Gruntwork Helm Charts.
* See the [examples](/examples) folder for example usage.
* See the provided [values.yaml](./values.yaml) file for the required and optional configuration values that you can set
  on this chart.


## What resources does this Helm Chart deploy?

The following resources will be deployed with this Helm Chart, depending on which configuration values you use:

- `Deployment`: The main `Deployment` controller that will manage the application container image specified in the
                `containerImage` input value.
- `Service`: The `Service` resource providing a stable endpoint that can be used to address to `Pods` created by the
             `Deployment` controller. Created only if you configure the `service` input (and set
             `service.enabled = true`).
- `Ingress`: The `Ingress` resource providing host and path routing rules to the `Service` for the deployed `Ingress`
             controller in the cluster. Created only if you configure the `ingress` input (and set
             `ingress.enabled = true`)
- `PodDisruptionBudget`: The `PodDisruptionBudget` resource that specifies a disruption budget for the `Pods` managed by
                         the `Deployment`. This manages how many pods can be disrupted by a voluntary disruption (e.g
                         node maintenance). Created if you specify a non-zero value for the `minPodsAvailable` input
                         value.


## What are Pods, Deployments, and Services?

TODO


## What is Ingress?

TODO


## What is a PodDisruptionBudget?

TODO


## Why does the Pod have a preStop hook with a Shutdown Delay?

TODO


## What is a sidecar container?

TODO


## How do you update the application to a new version?

TODO
