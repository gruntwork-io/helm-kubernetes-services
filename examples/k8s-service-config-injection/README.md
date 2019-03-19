# Quickstart Guide: K8S Service Config Injection Example

This quickstart guide uses the `k8s-service` Helm Chart to deploy a sample web app that is configured using environment
variables. In this guide, we will walk through the different ways to set environment variables on the application
container deployed using the `k8s-service` Helm Chart.

This guide is meant to demonstrate how you might pass in external values such as dependent resource URLs and various
secrets that your application needs.


## Prerequisites

This guide assumes that you are familiar with the defaults provided in the `k8s-service` Helm Chart. Please refer to the
[k8s-service-nginx](../k8s-service-nginx) example for an introduction to the core features of the Helm Chart.


## Overview

In this guide, we will walk through the steps to:

- Deploy a dockerized sample app on a Kubernetes cluster. We will use `minikube` for this guide.
- Use the `envVars` input value to set the port that the container listens on.
- Create a [`ConfigMap`](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) that
  provides some server text for the application to use.
- Use the `configMaps` input value to set the server text returned by the application from the `ConfigMap`.
- Create a [`Secret`](https://kubernetes.io/docs/concepts/configuration/secret/) that
  provides some server text for the application to use.
- Use the `secrets` input value to set the server text returned by the application from the `Secret`.

At the end of this guide, you should be familiar with the three ways provided by `k8s-service` to configure your
application.


## Outline

1. [Install and setup `minikube`](#setting-up-your-kubernetes-cluster-minikube)
1. [Install and setup `helm`](#setting-up-helm-on-minikube)
1. [Package the sample app docker container for `minikube`](#package-the-sample-app-docker-container-for-minikube)
1. [Deploy the sample app docker container with `k8s-service`](#deploy-the-sample-app-docker-container-with-k8s-service)
1. [Setting the server text using a ConfigMap](#setting-the-server-text-using-a-configmap)
1. [Setting the server text using a Secret](#setting-the-server-text-using-a-secret)

**NOTE:** This guide assumes you are running the steps in this directory. If you are at the root of the repo, be sure to
change directory before starting:

```
cd examples/k8s-service-config-injection
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


## Package the sample app docker container for Minikube

For this guide, we will need a docker container that provides a web service and is configurable using environment
variables.

We provide a sample app built using [Sinatra](http://sinatrarb.com/) on Ruby that returns some server text set using the
environment variable `SERVER_TEXT`. You can see the full code for the server in [docker/app.rb](./docker/app.rb).

In order to be able to deploy this on Kubernetes, we will need to package the app into a Docker container. To do so, we
need to first authenticate the docker client to be able to access the Docker Daemon running on `minikube`:

```bash
eval $(minikube docker-env)
```

The above step extracts the host information of the Docker Daemon running on your `minikube` virtual machine, and
configures the `docker` client using environment variables.

You can verify that you can reach the `minikube` Docker Daemon by running `docker ps`. You should see output similar to
below, listing a bunch of docker containers related to Kubernetes:

```
CONTAINER ID        IMAGE                                     COMMAND                  CREATED             STATUS              PORTS               NAMES
6ec6576eeb7e        gcr.io/kubernetes-helm/tiller             "/tiller"                29 minutes ago      Up 29 minutes                           k8s_tiller_tiller-deploy-6f4dbc6d67-dthjc_kube-system_a55f2873-41c7-11e9-af90-0800274e6ff3_0
1147669c0de6        k8s.gcr.io/pause:3.1                      "/pause"                 29 minutes ago      Up 29 minutes                           k8s_POD_tiller-deploy-6f4dbc6d67-dthjc_kube-system_a55f2873-41c7-11e9-af90-0800274e6ff3_0
5f6131b6b3ca        gcr.io/k8s-minikube/storage-provisioner   "/storage-provisioner"   About an hour ago   Up About an hour                        k8s_storage-provisioner_storage-provisioner_kube-system_2c2465d6-41c3-11e9-af90-0800274e6ff3_0
481b954a22b6        k8s.gcr.io/pause:3.1                      "/pause"                 About an hour ago   Up About an hour                        k8s_POD_storage-provisioner_kube-system_2c2465d6-41c3-11e9-af90-0800274e6ff3_0
8ec108f9948f        f59dcacceff4                              "/coredns -conf /etc…"   About an hour ago   Up About an hour                        k8s_coredns_coredns-86c58d9df4-rr262_kube-system_2a971ff2-41c3-11e9-af90-0800274e6ff3_0
84fd48fa4fa5        f59dcacceff4                              "/coredns -conf /etc…"   About an hour ago   Up About an hour                        k8s_coredns_coredns-86c58d9df4-kqn7c_kube-system_2a806542-41c3-11e9-af90-0800274e6ff3_0
b1a8616de7f6        98db19758ad4                              "/usr/local/bin/kube…"   About an hour ago   Up About an hour                        k8s_kube-pro
..... SNIPPED FOR BREVITY ...
```

Once your `docker` client is able to talk to the `minikube` Docker Daemon, we can now build our sample app container so
that it is available to `minikube` to use:

```bash
docker build -t gruntwork-io/sample-sinatra-app ./docker
```

This will build a container that has the runtime environment for running sinatra in the `minikube` virtual machine.
Once the container is created, we tag it as `gruntwork-io/sample-sinatra-app` so that it is easy to reference later.
Note that because this is built in the `minikube` virtual machine directly, the image will be cached within the VM. This
is why `minikube` is able to use the built container when you reference it in `k8s-service`.


## Deloy the sample app Docker container with k8s-service

Now that we have a working Kubernetes cluster with Helm installed and a sample Docker container to deploy, we are ready
to deploy our application using the `k8s-service` chart.

This folder contains predefined input values you can use with the `k8s-service` chart to deploy the sample app
container. Like the [k8s-service-nginx](../k8s-service-nginx) example, these values define the container image to use as
part of the deployment, and augments the default values of the chart by defining a `livenessProbe` and `readinessProbe`
for the main container (which in this case will be `gruntwork-io/sample-sinatra-app:latest`, the one we built in the
previous step). Take a look at the provided [`values.yaml`](./values.yaml) file to see how the values are defined.

However, the values in this example also sets an environment variable to configure the application. By default the
application listens for web requests on port 8080. However, most of the default values for the `k8s-service` helm chart
assumes the container listens for requests on port 80. While we can update the port that the chart uses, here we opt to
update the application container instead to provide an example of how you can hard code environment variables to pass
into the container in the `values.yaml` file. We use the `envVars` input map to set the `SERVER_PORT` to `80` in the
container:

```yaml
envVars:
  SERVER_PORT: 80
```

Each key in the `envVars` input map represents an environment variable, with the keys and values directly mapping to the
environment.

We will now instruct helm to install the `k8s-service` chart using these values. To do so, we will use the `helm
install` command:

```
helm install -f values.yaml ../../charts/k8s-service --wait
```

The above command will instruct the `helm` client to install the Helm Chart defined in the relative path
`../../charts/k8s-service`, merging the input values defined in `values.yaml` with the one provided by the chart.
Additionally, we provide the `--wait` keyword to ensure the command doesn't exit until the `Deployment` resource
completes the rollout of the containers.

At the end of this command, you should be able to access the sample web app via the `Service`. To hit the `Service`, get
the selected node port and hit the `minikube` ip on that port:

```bash
# NOTE: you must set RELEASE_NAME to be the chosen name of the release
export NODE_PORT=$(kubectl get --namespace default -o jsonpath="{.spec.ports[0].nodePort}" services "$RELEASE_NAME-sample-sinatra-app")
export NODE_IP=$(minikube ip)
curl "http://$NODE_IP:$NODE_PORT"
```

The above `curl` call should return the default server text set on the application container in JSON format:

```json
{"text":"Hello from backend"}
```


## Setting the server text using a ConfigMap

The previous step showed you how you can hard code environment variable settings into the `values.yaml` file. The
disadvantage of hardcoding the environment values is that you will need separate `vaules.yaml` file for each deployment
environment (e.g dev vs production), and manage them independently. This can be cumbersome if you have a lot of common
settings you want to share between the two environments.

Kubernetes provides [`ConfigMaps`](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) to
help decouple application configuration from the deployment settings. `ConfigMaps` are objects that hold key value pairs
that can then be injected into the application container at deploy time as environment variables, or as files on the
file system.

The `k8s-service` Helm Chart supports both modes of operation. In this guide, we will show you how to set an environment
variable from a `ConfigMap` key. To do so, we first need to create the `ConfigMap`.

### Creating the ConfigMap

For this example we will update the server text of our application using a `ConfigMap`. We will use `kubectl` to create
our `ConfigMap` on our Kubernetes cluster.

Take a look at [the provided resource file in the `kubernetes` folder](./kubernetes/config-map.yaml) that defines a
`ConfigMap`. This resource file will create a `ConfigMap` resource named `sample-sinatra-app-server-text` containing a
single key `server_text` holding the value `Hello! I was configured using a ConfigMap!`. We can create this `ConfigMap`
by using `kubectl apply` to apply the resource file:

```
kubectl apply -f ./kubernetes/config-map.yaml
```

To verify the `ConfigMap` was created, you can use the `kubectl get` command to get a list of available `ConfigMaps` on
your cluster:

```
$ kubectl get configmap
NAME                             DATA      AGE
sample-sinatra-app-server-text   1         57s
```

### Injecting the ConfigMap in to the application

Now that we have created a `ConfigMap` containing the server text config, let's augment our Helm Chart input value to
set the `SERVER_TEXT` environment variable from the `ConfigMap`. Take a look at the
[extensions/config_map_values.yaml](./extensions/config_map_values.yaml) file. This values file defines an entry for the
`configMaps` input map:

```yaml
configMaps:
  sample-sinatra-app-server-text:
    as: environment
    items:
      server_text:
        envVarName: SERVER_TEXT
```

Each key at the root of the `configMaps` map value specifies a `ConfigMap` by name. Then, the value is another map that
specifies how that `ConfigMap` should be included in the application container. You can either include it as a file
(`as: volume`) or environment variable (`as: environmet`). Here we include it as an environment variable, setting the
variable `SERVER_TEXT` to the value of the `server_text` key of the `ConfigMap`. You can refer to the documentation in
the chart's [`values.yaml`](/charts/k8s-service/values.yaml) for details on how to set the input map.

To deploy this, we will pass it in in addition to the root `values.yaml` file to merge the two inputs together. We will
use `helm upgrade` here instead of `helm install` so that we can update our previous deployment:

```
helm upgrade "$RELEASE_NAME" ../../charts/k8s-service -f values.yaml -f ./extensions/config_map_values.yaml --wait
```

When you pass in multiple `-f` options, `helm` will combine all the yaml files into one, preferring the right value over
the left (e.g if there was overlap, then `helm` will choose the value defined in `./extensions/config_map_values.yaml`
over the one defined in `values.yaml`).

When this deployment completes and you hit the server again, you should get the server text defined in the `ConfigMap`:

```
$ curl "http://$NODE_IP:$NODE_PORT"
{"text":"Hello! I was configured using a ConfigMap!"}
```


## Setting the server text using a Secret

`ConfigMaps` and hard coded environment variables are great for application configuration values, but are not very
secure. Hard coding environment variables leak into your code, and thus risk being checked in to source control while
`ConfigMaps` are not stored encrypted on the Kubernetes server and reports the value in plain text in the shell.

Kubernetes provides a built in secrets manager in the form of the
[`Secret`](https://kubernetes.io/docs/concepts/configuration/secret/) resource. Unlike `ConfigMaps`, `Secrets`:

- Can be stored in encrypted form in `etcd`.
- Is only sent to a node if a pod on that node requires it, and is only available in memory (using `tmpfs`).
- Obfuscates the text using `base64` to avoid "shoulder surfing" leakage.

**NOTE: Be aware of [the risks with using `Secrets` as your secrets
manager](https://kubernetes.io/docs/concepts/configuration/secret/#risks).**

Like `ConfigMaps`, `Secrets` can be injected into the application container as environment variables or as files, and
the `k8s-service` Helm Chart supports both modes of operation.

In this guide, we will use a `Secret` to replace the server text config set using a `ConfigMap` in the previous step.

### Creating the Secret

Since `Secrets` contain sensitive information, it is typically recommended to create `Secrets` manually using the command line.

To create a `Secret`, we can use `kubectl create secret`. Here, we will create a new secret containing the key
`server_text` set to `Hello! I was configured using a Secret!`:

```
kubectl create secret generic sample-sinatra-app-server-text --from-literal server_text='Hello! I was configured using a Secret!'
```

To verify the `Secret` was created, you can use the `kubectl get` command to get a list of available `Secrets`:

```
$ kubectl get secrets
NAME                             TYPE                                  DATA      AGE
default-token-wmb57              kubernetes.io/service-account-token   3         27m
sample-sinatra-app-server-text   Opaque                                1         1m
```

### Injecting the Secret in to the application

Now that we have created a `Secret` containing the server text config, let's try to inject it into the application
container. The settings to inject `Secrets` is formulated in a very similar manner to `ConfigMaps`. Take a look at the
[extensions/secret_values.yaml](./extensions/secret_values.yaml) file. This file defines a single input value `secrets`,
which sets the `SERVER_TEXT` environment variable to the `server_text` key on the `sample-sinatra-app-server-text`
`Secret`:

```yaml
secrets:
  sample-sinatra-app-server-text:
    as: environment
    items:
      server_text:
        envVarName: SERVER_TEXT
```

Compare this configuration with [extensions/config_map_values.yaml](./extensions/config_map_values.yaml). Note how the
only thing that differs is the input key is `secrets` as opposed to `configMaps`. This is because both `ConfigMaps` and
`Secrets` behave in very similar manners in Kubernetes, and so the `k8s-service` Helm Chart intentionally exposes a
similar interface to configure the two.

Deploying this config is very similar to how we deployed the `config_map_values.yaml` extension. We need to combine this
with the root `values.yaml` file to get a complete input and update our existing release:

```
helm upgrade "$RELEASE_NAME" ../../charts/k8s-service -f values.yaml -f ./extensions/secret_values.yaml --wait
```

When this deployment completes and you hit the server again, you should get the server text defined in the config map:

```
$ curl "http://$NODE_IP:$NODE_PORT"
{"text":"Hello! I was configured using a Secret!"}
```


## Summary

Congratulations! At this point, you have:

- Setup `minikube` to have a local dev environment of Kubernetes.
- Installed and deployed Helm on `minikube`.
- Packaged a sample application using Docker.
- Deployed the dockerized application on to `minikube` using the `k8s-service` Helm Chart.
- Configured the application using hard coded environment variables in the input values.
- Configured the application using `ConfigMaps`.
- Configured the application using `Secrets`.

To learn more about the `k8s-service` Helm Chart, refer to [the chart documentation](/charts/k8s-service).
