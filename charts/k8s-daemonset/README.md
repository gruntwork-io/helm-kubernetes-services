# Kubernetes DaemonSet Helm Chart

This Helm Chart can be used to deploy your application container under a
[DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) resource onto your Kubernetes
cluster. You can use this Helm Chart to run and deploy a long-running container, such as a web service or backend
microservice. The container will be packaged into
[Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/) that are managed by the `DaemonSet`
controller.

This Helm Chart can be used to deploy the pod under a `DaemonSet` resource onto your Kubernetes cluster.


## How to use this chart?

* See the [root README](/README.adoc) for general instructions on using Gruntwork Helm Charts.
* See the [examples](/examples) folder for example usage.
* See the provided [values.yaml](./values.yaml) file for the required and optional configuration values that you can set
  on this chart.

back to [root README](/README.adoc#core-concepts)

## What resources does this Helm Chart deploy?

The following resources will be deployed with this Helm Chart, depending on which configuration values you use:

- `DaemonSet`: The main `DaemonSet` controller that will manage the application container image specified in the
                `containerImage` input value.


back to [root README](/README.adoc#core-concepts)


## Useful helm commands:
	1. List helm charts 

  ```bash

  helm list

  NAME                    NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                   APP VERSION
  my-daemonset-chart      default         1               2022-04-27 17:52:38.166977136 +0000 UTC deployed        k8s-service-0.1.0       1.16.0
  
  ```

	2. List daemonsets and pods

  ```bash

    # List Daemonsets
    kubectl get daemonset
    NAME      DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
    fluentd   1         1         1       1            1           <none>          2m45s

    # List pods
    kubectl get pods
    NAME            READY   STATUS    RESTARTS   AGE
    fluentd-l9g2s   1/1     Running   0          3m14s
  ```


back to [root README](/README.adoc#day-to-day-operations)
