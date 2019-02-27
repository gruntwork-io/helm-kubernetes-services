# Quickstart Guide: K8S Service Nginx Example

This quickstart guide uses the `k8s-service` Helm Chart to deploy Nginx with healthchecks defined onto your Kubernetes
cluster. In this guide, we define the input values necessary to set the application container packaged in the
`Deployment` as the `nginx` container.

This guide is meant to demonstrate the defaults set by the Helm Chart to see what you get out of the box.


## Overview

In this guide we will walk through the steps necessary to deploy a vanilla Nginx server using the `k8s-service` Helm
Chart against a Kubernetes cluster. We will use `minikube` for this guide, but the chart is designed to work with many
different Kubernetes clusters (e.g EKS or GKE).

Here are the steps:

1. [Install and setup `minikube`](#setting-up-your-kubernetes-cluster-minikube)
1. [Install and setup `helm`](#setting-up-helm-on-minikube)
1. [Deploy Nginx with `k8s-service`](#deploy-nginx-with-k8s-service)
1. [Check the status of the deployment](#check-the-status-of-the-deployment)
1. [Access Nginx](#accessing-nginx)
1. [Upgrade Nginx to a newer version](#upgrading-nginx-container-to-a-newer-version)

**NOTE:** This guide assumes you are running the steps in this directory. If you are at the root of the repo, be sure to
change directory before starting:

```
cd examples/k8s-service-nginx
```

## Setting up your Kubernetes cluster: Minikube

In this guide, we will use `minikube` as our Kubernetes cluster to deploy Tiller to.
[Minikube](https://kubernetes.io/docs/setup/minikube/) is an official tool maintained by the Kubernetes community to be
able to provision and run Kubernetes locally your machine. By having a local environment you can have fast iteration
cycles while you develop and play with Kubernetes before deploying to production.

To setup `minikube`:

1. [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
1. [Install the minikube utility](https://kubernetes.io/docs/tasks/tools/install-minikube/)
1. Run `minikube start` to provision a new `minikube` instance on your local machine.
1. Verify setup with `kubectl`: `kubectl cluster-info`

## Setting up Helm on Minikube

In order to install Helm Charts, we need to have a working version of Tiller (the Helm server) deployed on our
`minikube` cluster. In this guide, we will use a barebones helm install with the defaults to get up and running quickly.

**WARNING: the barebones Tiller has no security context. Be sure to enable a stronger security context in any production
Kubernetes cluster. Read [our guide on Helm](https://github.com/gruntwork-io/kubergrunt/blob/master/HELM_GUIDE.md) for
more information.**

To setup helm, first install the [`helm` client](https://docs.helm.sh/using_helm/#installing-helm). Make sure the binary
is discoverble in your `PATH` variable. See [this stackoverflow
post](https://stackoverflow.com/questions/14637979/how-to-permanently-set-path-on-linux-unix) for instructions on
setting up your `PATH` on Unix, and [this
post](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows) for instructions on
Windows.

Next, use the `helm` client to setup Tiller. This is done through the `init` command. Run the following command to
deploy Tiller to `minikube`:

```bash
helm init --wait
```

For this guide, we are using the defaults to get up and running quicky on the local environment. In production, you will
want to turn on security features so that you don't expose your system.

The `--wait` option instructs the initializer to wait for Tiller to come up before exiting. When the command finishes
without errors, it means Tiller has been deployed and is available.

When you run this command, you should see output similar to below:

```
Creating /home/ubuntu/.helm
Creating /home/ubuntu/.helm/repository
Creating /home/ubuntu/.helm/repository/cache
Creating /home/ubuntu/.helm/repository/local
Creating /home/ubuntu/.helm/plugins
Creating /home/ubuntu/.helm/starters
Creating /home/ubuntu/.helm/cache/archive
Creating /home/ubuntu/.helm/repository/repositories.yaml
Adding stable repo with URL: https://kubernetes-charts.storage.googleapis.com
Adding local repo with URL: http://127.0.0.1:8879/charts
$HELM_HOME has been configured at /Users/yoriy/.helm.

Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.

Please note: by default, Tiller is deployed with an insecure 'allow unauthenticated users' policy.
To prevent this, run `helm init` with the --tiller-tls-verify flag.
For more information on securing your installation see: https://docs.helm.sh/using_helm/#securing-your-helm-installation
Happy Helming!
```

Verify you can access the server using `helm version`, which will list both the client and server versions.

```bash
$ helm version
Client: &version.Version{SemVer:"v2.11.0", GitCommit:"2e55dbe1fdb5fdb96b75ff144a339489417b146b", GitTreeState:"clean"}
Server: &version.Version{SemVer:"v2.11.0", GitCommit:"2e55dbe1fdb5fdb96b75ff144a339489417b146b", GitTreeState:"clean"}
```

If you have any problems in your setup, the `Server` version will fail to output. Refer to [the official installation
FAQ](https://docs.helm.sh/using_helm/#installation-frequently-asked-questions) for common issues and problems during the
installation step.

## Deploy Nginx with k8s-service

Now that we have a working Kubernetes cluster with Helm installed and ready to go, the next step is to deploy Nginx
using the `k8s-service` chart.

This folder contains predefined input values you can use with the `k8s-service` chart to deploy Nginx. These values
define the container image to use as part of the deployment, and augments the default values of the chart by defining a
`livenessProbe` and `readinessProbe` for the main container (whcih in this case will be `nginx:1.14.2`. Take a look at
the provided [`values.yaml`](./values.yaml) file to see how the values are defined.

We will now instruct helm to install the `k8s-service` chart using these values. To do so, we will use the `helm
install` command:

```
helm install -f values.yaml ../../charts/k8s-service
```

The above command will instruct the `helm` client to install the Helm Chart defined in the relative path
`../../charts/k8s-service`, merging the input values defined in `values.yaml` with the one provided by the chart.

Note that when you install this chart, `helm` will select a random name to use for your release. In Helm, a release
ties together the provided input values with a chart install, tracking the state of the resources that have been
deployed using Helm. The release name is uniquely identifies the release, and can be used to interact with a previous
deployment.

When you run this command, you should see output similar to below:

```
NAME:   queenly-liger
LAST DEPLOYED: Sat Feb 16 09:14:39 2019
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/Service
NAME                 AGE
queenly-liger-nginx  0s

==> v1/Deployment
queenly-liger-nginx  0s

==> v1/Pod(related)

NAME                                READY  STATUS             RESTARTS  AGE
queenly-liger-nginx-7b7bb49d-b8tf8  0/1    ContainerCreating  0         0s
queenly-liger-nginx-7b7bb49d-fgjd4  0/1    ContainerCreating  0         0s
queenly-liger-nginx-7b7bb49d-zxpcm  0/1    ContainerCreating  0         0s


NOTES:
Check the status of your Deployment by running this comamnd:

kubectl get deployments --namespace default -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=queenly-liger"


List the related Pods with the following command:

kubectl get pods --namespace default -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=queenly-liger"


Use the following command to view information about the Service:

kubectl get services --namespace {{ .Release.Namespace }} -l "app.kubernetes.io/name={{ include "k8s-service.name" . }},app.kubernetes.io/instance={{ .Release.Name }}"


Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=queenly-liger" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application container serving port http"
  kubectl port-forward $POD_NAME 8080:80
```

The install command will always output:

- The release name. In this case, the name is `queenly-liger`.
- The namespace where the resources are created. In this case, the namespace is `default`.
- The status of the release. In this case, the release was successfully deployed so the status is `DEPLOYED`.
- A summary of the resources created. Additionally, for certain resources, `helm` will also output the related resource.
  For example, in this case, `helm` outputted all the `Pods` created by the `Deployment` resource.
- Any additional notes provided by the chart maintainer. For `k8s-service`, we output some commands you can use to check on the
  status of the service.

Since we will be referring to this output for the remainder of this guide, it would be a good idea to copy paste the
output somewhere so you can refer to it. If you ever lose the information and want to see the output again, you can use
the `helm status` command to view the output. The `helm status` command takes in the release name, and outputs
information about that release.

Now that the service is installed and deployed, let's verify the deployment!

## Check the Status of the Deployment

In the previous step, we deployed Nginx using the `k8s-service` Helm Chart. Now we want to verify it has deployed
successfully.

Under the hood the Helm Chart creates a `Deployment` resource in Kubernetes. `Deployments` are a controller that can be
used to declaratively manage your application. When you create the `Deployment` resource, it instructs Kubernetes the
desired state of the application deployment (e.g how many `Pods` to use, what container image to use, any volumes to
mount, etc). Kubernetes will then asynchronously create resources to match the desired state. This means that instead of
creating and updating `Pods` on the cluster, you can simply declare that you want 3 Nginx `Pods` deployed and let
Kubernetes handle the details. The downside of this is that the deployment happens asynchronously. In other words, this
means the Helm Chart may install successfully but the deployment could still fail.

So let's look at the status of the deployment to confirm the deployment successfully finished. In the output above, the
`NOTES` section lists out a command that can be used to get information about the `Deployment`. So let's try running
that:

```
$ kubectl get deployments --namespace default -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=queenly-liger"
NAME                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
queenly-liger-nginx   3         3         3            3           5m
```

In the output above, Kubernetes is reporting information about the `Pods` related to the `Deployment`. Each column is a
count of the number of `Pods` that fit that description. In this case, we have the correct number of `Pods` that are up
to date on the latest image (`UP-TO-DATE`) and available to accept traffic (`AVAILABLE`). When those columns diverge
from the `DESIRED` column, then that means either the deployment is still in progress, or something failed in the
process.

You can further dig deeper using `describe`, or querying the different subresources such as the underlying Pods. For
this guide, we are satisfied with the `Deployment` status output above. See [How do I check the status of the
rollout?](/charts/k8s-service/README.md#how-do-i-check-the-status-of-the-rollout) for more details on how to check in on
the detailed status of a rollout, and to help troubleshoot any issues in your environment.

Once you have confirmed the `Deployment` has rolled out successfully, the next step is to verify that Nginx is up and
accessible.

## Accessing Nginx

### Accessing a Pod directly

Let's first try accessing a single Nginx `Pod`. To do so, we will open a tunnel from our local machine that routes
through the Kubernetes control plane to the underlying `Pod` on the worker nodes.

To open the tunnel, we need two pieces of information:

- The name of the `Pod` to open the tunnel to.
- The open ports on the `Pod` and the port we wish to access.

To retrieve the name of a `Pod`, we can inspect the list of `Pods` created by the `Deployment`. As in the previous
section, the `helm install` output notes contains a command we can use to get the list of `Pods` managed by the
`Deployment`, so let's try running that here:

```
$ kubectl get pods --namespace default -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=queenly-liger"
NAME                                 READY     STATUS    RESTARTS   AGE
queenly-liger-nginx-7b7bb49d-b8tf8   1/1       Running   0          13m
queenly-liger-nginx-7b7bb49d-fgjd4   1/1       Running   0          13m
queenly-liger-nginx-7b7bb49d-zxpcm   1/1       Running   0          13m
```

Here you can see that there are 3 `Pods` in the `READY` state that match that criteria. Pick one of them to access from
the list above and record the name.

Next, we need to see what ports are open on the `Pod`. The `k8s-service` Helm Chart will open ports 80 to the
container by default. However, if you do not know which ports are open, you can inspect the `Pod` to a list of the open
ports. To get detailed information about a `Pod`, use `kubectl describe pod NAME`. In our example, we will pull detailed
information about the `Pod` `queenly-liger-nginx-7b7bb49d-b8tf8`:

```
$ kubectl describe pod queenly-liger-nginx-7b7bb49d-b8tf8
Name:               queenly-liger-nginx-7b7bb49d-b8tf8
Namespace:          default
Priority:           0
PriorityClassName:  <none>
Node:               minikube/10.0.2.15
Start Time:         Sat, 16 Feb 2019 09:14:40 -0800
Labels:             app.kubernetes.io/instance=queenly-liger
                    app.kubernetes.io/name=nginx
                    pod-template-hash=7b7bb49d
Annotations:        <none>
Status:             Running
IP:                 172.17.0.6
Controlled By:      ReplicaSet/queenly-liger-nginx-7b7bb49d
Containers:
  nginx:
    Container ID:   docker://ac921c94c8d5f9428815d64bfa541f0481ab37ddaf42a37f2ebec95eb61ef2c0
    Image:          nginx:1.14.2
    Image ID:       docker-pullable://nginx@sha256:d1eed840d5b357b897a872d17cdaa8a4fc8e6eb43faa8ad2febb31ce0c537910
    Ports:          80/TCP
    Host Ports:     0/TCP
    State:          Running
      Started:      Sat, 16 Feb 2019 09:15:09 -0800
    Ready:          True
    Restart Count:  0
    Liveness:       http-get http://:http/ delay=0s timeout=1s period=10s #success=1 #failure=3
    Readiness:      http-get http://:http/ delay=0s timeout=1s period=10s #success=1 #failure=3
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-nskm6 (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Volumes:
  default-token-nskm6:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-nskm6
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute for 300s
                 node.kubernetes.io/unreachable:NoExecute for 300s
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  15m   default-scheduler  Successfully assigned default/queenly-liger-nginx-7b7bb49d-b8tf8 to minikube
  Normal  Pulling    15m   kubelet, minikube  pulling image "nginx:1.14.2"
  Normal  Pulled     14m   kubelet, minikube  Successfully pulled image "nginx:1.14.2"
  Normal  Created    14m   kubelet, minikube  Created container
  Normal  Started    14m   kubelet, minikube  Started container
```

This outputs all the detailed metadata about the running `Pod`, as well as an event log of all the cluster activity
related to the `Pod`. In the output, the `Containers` section shows addtional information about each container deployed in the `Pod`. Since we want to know the open ports for the `nginx` container, we will look at the `Ports` section of the `nginx` container in that output. Here is the specific output we are interested in:

```
Containers:
  nginx:
    Ports:          80/TCP
```

In the output, we confirm that indeed port 80 is open. So let's open a port forward!

In this example, we will open a tunnel from port 8080 on our local machine to port 80 of the `Pod`:

```
$ kubectl port-forward queenly-liger-nginx-7b7bb49d-b8tf8 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```

This command will run in the foreground, and keeps the tunnel open as long as the command is running. You can close the
tunnel at any time by hitting `Ctrl+C`.

Now try accessing `localhost:8080` in the browser. You should get the default nginx welcome page. Assuming you do not
have nginx running locally, this means that you have successfully accessed the `Pod` from your local machine!

### Accessing a Pod through a Service

In the previous step we created a port forward from our local machine to the `Pod` directly. However, normally you would
not want to access your applications this way because `Pods` are ephemeral in Kubernetes. They come and go as nodes
scale up and down. They are also limited to the single resource and thus do not do any form of load balancing. This is
where `Services` come into play.

A `Service` in Kubernetes is used to expose a group of `Pods` that match a given selector under a stable endpoint.
`Service` resources track which `Pods` are live and ready, and only route traffic to those that are in the `READY`
status. The `READY` status is managed using `readinessProbes`: as long as the `Pod` passes the readiness check, the
`Pod` will be marked `READY` and kept in the pool for the `Service`.

There are several different types of `Services`. You can learn more about the different types in the [How do I expose my
application](/charts/k8s-service/README.md#how-do-i-expose-my-application-internally-to-the-cluster) section of the
chart README. For this example, we used the default `Service` resource created by the chart, but overrode the type to be
`NodePort`. A `NodePort` `Service` exposes a port on the Kubernetes worker that routes to the `Service` endpoint. This
endpoint will load balance across the `Pods` that match the selector for the `Service`.

To access a `NodePort` `Service`, we need to first find out what port is exposed. We can do this by querying for the
`Service` using `kubectl`. As before, the `NOTES` output contains a command we can use to find the related `Service`.
However, the `NOTES` output also contains instructions for directly getting the service node port and service node ip.
Here, we will use those commands to extract the endpoint for the `Service`, with one modification. Because we are
running the `Service` on `minikube`, there is one layer of indirection in the `minikube` VM. `minikube` runs in its own
VM on your machine, which means that the ip of the node will be incorrect. So instead of querying for the registered
node IP in Kubernetes, we will instead use `minikube` to get the ip address of the `minikub` VM to use ast the node IP:

```bash
export NODE_PORT=$(kubectl get --namespace default -o jsonpath="{.spec.ports[0].nodePort}" services queenly-liger-nginx)
export NODE_IP=$(minikube ip)
echo http://$NODE_IP:$NODE_PORT
```

The first command query the `Service` resource to find out the node port that was used to expose the service. The second
command queries the ip address of `minikub`. The last command will `echo` out the endpoint where the service is
available. Try hitting that endpoint in your browser and you should see the familiar nginx splash screen.

## Summary

Congratulations! At this point, you have:

- Setup `minikube` to have a local dev environment of Kubernetes.
- Installed and deployed Helm on `minikube`.
- Deployed nginx on to `minikube` using the `k8s-service` Helm Chart.
- Verified the deployment by querying for resources using `kubectl` and opening port forwards to access the endpoints.

To learn more about the `k8s-service` Helm Chart, refer to [the chart documentation](/charts/k8s-service).
