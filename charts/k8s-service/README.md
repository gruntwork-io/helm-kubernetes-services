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

If you're using the chart to deploy to [GKE](https://cloud.google.com/kubernetes-engine/), you can also use the chart to deploy a [Google Managed SSL Certificate](https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs) and associate it with the Ingress.


## How to use this chart?

* See the [root README](/README.adoc) for general instructions on using Gruntwork Helm Charts.
* See the [examples](/examples) folder for example usage.
* See the provided [values.yaml](./values.yaml) file for the required and optional configuration values that you can set
  on this chart.

back to [root README](/README.adoc#core-concepts)

## What resources does this Helm Chart deploy?

The following resources will be deployed with this Helm Chart, depending on which configuration values you use:

- `Deployment`: The main `Deployment` controller that will manage the application container image specified in the
                `containerImage` input value.
- Secondary `Deployment` for use as canary: An optional `Deployment` controller that will manage a [canary deployment](https://martinfowler.com/bliki/CanaryRelease.html) of the application container image specified in the `canary.containerImage` input value. This is useful for testing a new application tag, in parallel to your stable tag, prior to rolling the new tag out. Created only if you configure the `canary.containerImage` values (and set `canary.enabled = true`).
- `Service`: The `Service` resource providing a stable endpoint that can be used to address to `Pods` created by the
             `Deployment` controller. Created only if you configure the `service` input (and set
             `service.enabled = true`).
- `ServiceMonitor`: The `ServiceMonitor` describes the set of targets to be monitored by Prometheus. Created only if you configure the service input and set `serviceMonitor.enabled = true`.
- `Ingress`: The `Ingress` resource providing host and path routing rules to the `Service` for the deployed `Ingress`
             controller in the cluster. Created only if you configure the `ingress` input (and set
             `ingress.enabled = true`).
- `Horizontal Pod Autoscaler`: The `Horizontal Pod Autoscaler` automatically scales the number of pods in a replication
                                controller, deployment, replica set or stateful set based on observed CPU or memory utilization.
                                Created only if the user sets `horizontalPodAutoscaler.enabled = true`.
- `PodDisruptionBudget`: The `PodDisruptionBudget` resource that specifies a disruption budget for the `Pods` managed by
                         the `Deployment`. This manages how many pods can be disrupted by a voluntary disruption (e.g
                         node maintenance). Created if you specify a non-zero value for the `minPodsAvailable` input
                         value.
- `ManagedCertificate`: The `ManagedCertificate` is a [GCP](https://cloud.google.com/) -specific resource that creates a Google Managed SSL certificate. Google-managed SSL certificates are provisioned, renewed, and managed for your domain names. Read more about Google-managed SSL certificates [here](https://cloud.google.com/load-balancing/docs/ssl-certificates#managed-certs). Created only if you configure the `google.managedCertificate` input (and set
                         `google.managedCertificate.enabled = true` and `google.managedCertificate.domainName = your.domain.name`).

back to [root README](/README.adoc#core-concepts)

## How do I deploy additional services not managed by the chart?

You can create custom Kubernetes resources, that are not directly managed by the chart, within the `customResources`
key. You provide each resource manifest directly as a value under `customResources.resources` and set
`customResources.enabled` to `true`. For examples of custom resources, take a look at the examples in
[test/fixtures/custom_resources_values.yaml](../../test/fixtures/custom_resources_values.yaml) and
[test/fixtures/multiple_custom_resources_values.yaml](../../test/fixtures/multiple_custom_resources_values.yaml).

back to [root README](/README.adoc#day-to-day-operations)

## How do I expose my application internally to the cluster?

In general, `Pods` are considered ephemeral in Kubernetes. `Pods` can come and go at any point in time, either because
containers fail or the underlying instances crash. In either case, the dynamic nature of `Pods` make it difficult to
consistently access your application if you are individually addressing the `Pods` directly.

Traditionally, this is solved using service discovery, where you have a stateful system that the `Pods` would register
to when they are available. Then, your other applications can query the system to find all the available `Pods` and
access one of the available ones.

Kubernetes provides a built in mechanism for service discovery in the `Service` resource. `Services` are an abstraction
that groups a set of `Pods` behind a consistent, stable endpoint to address them. By creating a `Service` resource, you
can provide a single endpoint to other applications to connect to the `Pods` behind the `Service`, and not worry about
the dynamic nature of the `Pods`.

You can read a more detailed description of `Services` in [the official
documentation](https://kubernetes.io/docs/concepts/services-networking/service/). Here we will cover just enough to
understand how to access your app.

By default, this Helm Chart will deploy your application container in a `Pod` that exposes ports 80. These will
be exposed to the Kubernetes cluster behind the `Service` resource, which exposes port 80. You can modify this behavior
by overriding the `containerPorts` input value and the `service` input value. See the corresponding section in the
`values.yaml` file for more details.

Once the `Service` is created, you can check what endpoint the `Service` provides by querying Kubernetes using
`kubectl`. First, retrieve the `Service` name that is outputted in the install summary when you first install the Helm
Chart. If you forget, you can get the same information at a later point using `helm status`. For example, if you had
previously installed this chart under the name `edge-service`, you can run the following command to see the created
resources:

```bash
$ helm status edge-service
LAST DEPLOYED: Fri Feb  8 16:25:49 2019
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/Service
NAME                AGE
edge-service-nginx  24m

==> v1/Deployment
edge-service-nginx  24m

==> v1/Pod(related)

NAME                                 READY  STATUS   RESTARTS  AGE
edge-service-nginx-844c978df7-f5wc4  1/1    Running  0         24m
edge-service-nginx-844c978df7-mln26  1/1    Running  0         24m
edge-service-nginx-844c978df7-rdsr8  1/1    Running  0         24m
```

This will show you some metadata about the release, the deployed resources, and any notes provided by the Helm Chart. In
this example, the service name is `edge-service-nginx` so we will use that to query the `Service`:

```bash
$ kubectl get service edge-service-nginx
NAME                 TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
edge-service-nginx   ClusterIP   172.20.186.176   <none>        80/TCP    27m
```

Here you can see basic information about the `Service`. The important piece of information is the `CLUSTER-IP` and
`PORT` fields, which tell you the available endpoint for the `Service`, and any exposed ports. Given that, any `Pod` in
your Kubernetes cluster can access the `Pods` of this application by hitting `{CLUSTER-IP}:{PORT}`. So for this example,
that will be `172.20.186.176:80`.

But what if you want to automatically find a `Service` by name? The name of the `Service` created by this Helm Chart is
always `{RELEASE_NAME}-{applicationName}`, where `applicationName` is provided in the input value and `RELEASE_NAME` is
set when you install the Helm Chart. This means that the name is predictable, while the allocated IP address may not be.

To address the `Service` by name, Kubernetes provides two ways:

- environment variables
- DNS

### Addressing Service by Environment Variables

For each active `Service` that a `Pod` has access to, Kubernetes will automatically set a set of environment variables
in the container. These are `{SVCNAME}_SERVICE_HOST` and `{SVCNAME}_SERVICE_PORT` to get the host address (ip address)
and port respectively, where `SVCNAME` is the name of the `Service`. Note that `SVCNAME` will be the all caps version
with underscores of the `Service` name.

Using the previous example where we installed this chart with a release name `edge-service` and `applicationName`
`nginx`, we get the `Service` name `edge-service-nginx`. Kubernetes will expose the following environment variables to
all containers that can access the `Service`:

```
EDGE_SERVICE_NGINX_SERVICE_HOST=172.20.186.176
EDGE_SERVICE_NGINX_SERVICE_PORT=80
```

Note that environment variables are set when the container first boots up. This means that if you already had `Pods`
deployed in your system before the `Service` was created, you will have to cycle the `Pods` in order to get the
environment variables. If you wish to avoid ordering issues, you can use the DNS method to address the `Service`
instead, if that is available.

### Addressing Service by DNS

If your Kubernetes cluster is deployed with the DNS add-on (this is automatically installed for EKS and GKE), then you
can rely on DNS to address your `Service`. Every `Service` in Kubernetes will register the domain
`{SVCNAME}.{NAMESPACE}.svc.cluster.local` to the DNS service of the cluster. This means that all your `Pods` in the
cluster can get the `Service` host by hitting that domain.

The `NAMESPACE` in the domain refers to the `Namespace` where the `Service` was created. By default, all resources are
created in the `default` namespace. This is configurable at install time of the Helm Chart using the `--namespace`
option.

In our example, we deployed the chart to the `default` `Namespace`, and the `Service` name is `edge-service-nginx`. So in
this case, the domain of the `Service` will be `edge-service-nginx.default.svc.cluster.local`. When any `Pod` addresses
that domain, it will get the address `172.20.186.176`.

Note that DNS does not resolve ports, so in this case, you will have to know which port the `Service` uses. So in your
`Pod`, you will have to know that the `Service` exposes port `80` when you address it in your code for the container as
`edge-service-nginx.default.svc.cluster.local:80`. However, like the `Service` name, this should be predictable since it
is specified in the Helm Chart input value.

back to [root README](/README.adoc#day-to-day-operations)

## How do I expose my application externally, outside of the cluster?

Similar to the previous section ([How do I expose my application internally to the
cluster?](#how-do-i-expose-my-application-internally-to-the-cluster), you can use a `Service` resource to expose your
application externally. The primary service type that facilitates external access is the `NodePort` `Service` type.

The `NodePort` `Service` type will expose the `Service` by binding an available port on the network interface of the
physical machines running the `Pod`. This is different from a network interface internal to Kubernetes, which is only
accessible within the cluster. Since the port is on the host machine network interface, you can access the `Service` by
hitting that port on the node.

For example, suppose you had a 2 node Kubernetes cluster deployed on EC2. Suppose further that all your EC2 instances
have public IP addresses that you can access. For the sake of this example, we will assign random IP addresses to the
instances:

- 54.219.117.250
- 38.110.235.198

Now let's assume you deployed this helm chart using the `NodePort` `Service` type. You can do this by setting the
`service.type` input value to `NodePort`:

```yaml
service:
  enabled: true
  type: NodePort
  ports:
    app:
      port: 80
      targetPort: 80
      protocol: TCP
```

When you install this helm chart with this input config, helm will deploy the `Service` as a `NodePort`, binding an
available port on the host machine to access the `Service`. You can confirm this by querying the `Service` using
`kubectl`:

```bash
$ kubectl get service edge-service-nginx
NAME                TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
edge-service-nginx  NodePort   10.99.244.96   <none>        80:31035/TCP   33s
```

In this example, you can see that the `Service` type is `NodePort` as expected. Additionally, you can see that the there
is a port binding between port 80 and 31035. This port binding refers to the binding between the `Service` port (80 in
this case) and the host port (31035 in this case).

One thing to be aware of about `NodePorts` is that the port binding will exist on all nodes in the cluster. This means
that, in our 2 node example, both nodes now have a port binding of 31035 on the host network interface that routes to
the `Service`, regardless of whether or not the node is running the `Pods` backing the `Service` endpoint. This means
that you can reach the `Service` on both of the following endpoints:

- `54.219.117.250:31035`
- `38.110.235.198:31035`

This means that no two `Service` can share the same `NodePort`, as the port binding is shared across the cluster.
Additionally, if you happen to hit a node that is not running a `Pod` backing the `Service`, Kubernetes will
automatically hop to one that is.

You might use the `NodePort` if you do not wish to manage load balancers through Kubernetes, or if you are running
Kubernetes on prem where you do not have native support for managed load balancers.

To summarize:

- `NodePort` is the simplest way to expose your `Service` to externally to the cluster.
- You have a limit on the number of `NodePort` `Services` you can have in your cluster, imposed by the number of open ports
  available on your host machines.
- You have potentially inefficient hopping if you happen to route to a node that is not running the `Pod` backing the
  `Service`.

Additionally, Kubernetes provides two mechanisms to manage an external load balancer that routes to the `NodePort` for
you. The two ways are:

- [Using a `LoadBalancer` `Service` type](#loadbalancer-service-type)
- [Using `Ingress` resources with an `Ingress Controller`](#ingress-and-ingress-controllers)

### LoadBalancer Service Type

The `LoadBalancer` `Service` type will expose the `Service` by allocating a managed load balancer in the cloud that is
hosting the Kubernetes cluster. On AWS, this will be an ELB, while on GCP, this will be a Cloud Load Balancer. When the
`LoadBalancer` `Service` is created, Kubernetes will automatically create the underlying load balancer resource in the
cloud for you, and create all the target groups so that they route to the `Pods` backing the `Service`.

You can deploy this helm chart using the `LoadBalancer` `Service` type by setting the `service.type` input value to
`LoadBalancer`:

```yaml
service:
  enabled: true
  type: LoadBalancer
  ports:
    app:
      port: 80
      targetPort: 80
      protocol: TCP
```

When you install this helm chart with this input config, helm will deploy the `Service` as a `LoadBalancer`, allocating
a managed load balancer in the cloud hosting your Kubernetes cluster. You can get the attached load balancer by querying
the `Service` using `kubectl`. In this example, we will assume we are using EKS:

```
$ kubectl get service edge-service-nginx
NAME                 TYPE           CLUSTER-IP    EXTERNAL-IP        PORT(S)        AGE
edge-service-nginx   LoadBalancer   172.20.7.35   a02fef4d02e41...   80:32127/TCP   1m
```

Now, in this example, we have an entry in the `EXTERNAL-IP` field. This is truncated here, but you can get the actual
output when you describe the service:

```
$ kubectl describe service edge-service-nginx
Name:                     edge-service-nginx
Namespace:                default
Labels:                   app.kubernetes.io/instance=edge-service
                          app.kubernetes.io/managed-by=helm
                          app.kubernetes.io/name=nginx
                          gruntwork.io/app-name=nginx
                          helm.sh/chart=k8s-service-0.1.0
Annotations:              <none>
Selector:                 app.kubernetes.io/instance=edge-service,app.kubernetes.io/name=nginx,gruntwork.io/app-name=nginx
Type:                     LoadBalancer
IP:                       172.20.7.35
LoadBalancer Ingress:     a02fef4d02e4111e9891806271fc7470-173030870.us-west-2.elb.amazonaws.com
Port:                     app  80/TCP
TargetPort:               80/TCP
NodePort:                 app  32127/TCP
Endpoints:                10.0.3.19:80
Session Affinity:         None
External Traffic Policy:  Cluster
Events:
  Type    Reason                Age   From                Message
  ----    ------                ----  ----                -------
  Normal  EnsuringLoadBalancer  2m    service-controller  Ensuring load balancer
  Normal  EnsuredLoadBalancer   2m    service-controller  Ensured load balancer
```

In the describe output, there is a field named `LoadBalancer Ingress`. When you have a `LoadBalancer` `Service` type,
this field contains the public DNS endpoint of the associated load balancer resource in the cloud provider. In this
case, we have an AWS ELB instance, so this endpoint is the public endpoint of the associated ELB resource.

**Note:** Eagle eyed readers might also notice that there is an associated `NodePort` on the resource. This is because under the
hood, `LoadBalancer` `Services` utilize `NodePorts` to handle the connection between the managed load balancer of the
cloud provider and the Kubernetes `Pods`. This is because at this time, there is no portable way to ensure that the
network between the cloud load balancers and Kubernetes can be shared such that the load balancers can route to the
internal network of the Kubernetes cluster. Therefore, Kubernetes resorts to using `NodePort` as an abstraction layer to
connect the `LoadBalancer` to the `Pods` backing the `Service`. This means that `LoadBalancer` `Services` share the same
drawbacks as using a `NodePort` `Service`.

To summarize:

- `LoadBalancer` provides a way to set up a cloud load balancer resource that routes to the provisioned `NodePort` on
  each node in your Kubernetes cluster.
- `LoadBalancer` can be used to provide a persistent endpoint that is robust to the ephemeral nature of nodes in your
  cluster. E.g it is able to route to live nodes in the face of node failures.
- `LoadBalancer` does not support weighted balancing. This means that you cannot balance the traffic so that it prefers
  nodes that have more instances of the `Pod` running.
- Note that under the hood, `LoadBalancer` utilizes a `NodePort` `Service`, and thus shares the same limits as `NodePort`.

### Ingress and Ingress Controllers

`Ingress` is a mechanism in Kubernetes that abstracts externally exposing a `Service` from the `Service` config itself.
`Ingress` resources support:

- assigning an externally accessible URL to a `Service`
- perform hostname and path based routing of `Services`
- load balance traffic using customizable balancing rules
- terminate SSL

You can read more about `Ingress` resources in [the official
documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/). Here, we will cover the basics to
understand how `Ingress` can be used to externally expose the `Service`.

At a high level, the `Ingress` resource is used to specify the configuration for a particular `Service`. In turn, the
`Ingress Controller` is responsible for fulfilling those configurations in the cluster. This means that the first
decision to make in using `Ingress` resources, is selecting an appropriate `Ingress Controller` for your cluster.

#### Choosing an Ingress Controller

Before you can use an `Ingress` resource, you must install an `Ingress Controller` in your Kubernetes cluster. There are
many kinds of `Ingress Controllers` available, each with different properties. You can see [a few examples listed in the
official documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-controllers).

When you use an external cloud `Ingress Controller` such as the [GCE Ingress
Controller](https://github.com/kubernetes/ingress-gce) or [AWS ALB Ingress
Controller](https://github.com/kubernetes-sigs/aws-alb-ingress-controller), Kubernetes will allocate an externally
addressable load balancer (for GCE this will be a Cloud Load Balancer and for AWS this will be an ALB) that fulfills the
`Ingress` rules. This includes routing the domain names and paths to the right `Service` as configured by the `Ingress`
rules. Additionally, Kubernetes will manage the target groups of the load balancer so that they are up to date with
the latest `Ingress` configuration. However, in order for this to work, there needs to be some way for the load balancer
to connect to the `Pods` servicing the `Service`. Since the `Pods` are internal to the Kubernetes network and the load
balancers are external to the network, there must be a `NodePort` that links the two together. As such, like the
`LoadBalancer` `Service` type, these `Ingress Controllers` also require a `NodePort` under the hood.

<!-- TODO: include commentary on how to associate host domains to the Ingress Controller -->

Alternatively, you can use an internal `Ingress Controller` that runs within Kubernetes as `Pods`. For example, the
official `nginx Ingress Controller` will launch `nginx` as `Pods` within your Kubernetes cluster. These `nginx` `Pods`
are then configured using `Ingress` resources, which then allows `nginx` to route to the right `Pods`. Since the `nginx`
`Pods` are internal to the Kubernetes network, there is no need for your `Services` to be `NodePorts` as they are
addressable within the network by the `Pods`. However, this means that you need some other mechanism to expose `nginx`
to the outside world, which will require a `NodePort`. The advantage of this approach, despite still requiring a
`NodePort`, is that you can have a single `NodePort` that routes to multiple services using hostnames or paths as
managed by `nginx`, as opposed to requiring a `NodePort` per `Service` you wish to expose.

Which `Ingress Controller` type you wish to use depends on your infrastructure needs. If you have relatively few
`Services`, and you want the simplicity of a managed cloud load balancer experience, you might opt for the external
`Ingress Controllers` such as GCE and AWS ALB controllers. On the other hand, if you have thousands of micro services
that push you to the limits of the available number of ports on a host machine, you might opt for an internal `Ingress
Controller` approach. Whichever approach you decide, be sure to document your decision where you install the particular
`Ingress Controller` so that others in your team know and understand the tradeoffs you made.

#### Configuring Ingress for your Service

Once you have an `Ingress Controller` installed and configured on your Kuberentes cluster, you can now start creating
`Ingress` resources to add routes to it. This helm chart supports configuring an `Ingress` resource to complement the
`Service` resource that is created in the chart.

To add an `Ingress` resource, first make sure you have a `Service` enabled on the chart. Depending on the chosen
`Ingress Controller`, the `Service` type should be `NodePort` or `ClusterIP`. Here, we will create a `NodePort`
`Service` exposing port 80:

```yaml
service:
  enabled: true
  type: NodePort
  ports:
    app:
      port: 80
      targetPort: 80
      protocol: TCP
```

Then, we will add the configuration for the `Ingress` resource by specifying the `ingress` input value. For this
example, we will assume that we want to route `/app` to our `Service`, with the domain hosted on `app.yourco.com`:

```yaml
ingress:
   enabled: true
   path: /app
   servicePort: 80
   hosts:
     - app.yourco.com
```

This will configure the load balancer backing the `Ingress Controller` that will route any traffic with host and path
prefix `app.yourco.com/app` to the `Service` on port 80. If `app.yourco.com` is configured to point to the `Ingress
Controller` load balancer, then once you deploy the helm chart you should be able to start accessing your app on that
endpoint.

#### Registering additional paths

Sometimes you might want to add additional path rules beyond the main service rule that is injected to the `Ingress`
resource. For example, you might want a path that routes to the sidecar containers, or you might want to reuse a single
`Ingress` for multiple different `Service` endpoints because to share load balancers. For these situations, you can use
the `additionalPaths` and `additionalPathsHigherPriority` input values.

Consider the following `Service`, where we have the `app` served on port 80, and the `sidecarMonitor` served on port
3000:

```yaml
service:
  enabled: true
  type: NodePort
  ports:
    app:
      port: 80
      targetPort: 80
      protocol: TCP
    sidecarMonitor:
      port: 3000
      targetPort: 3000
      protocol: TCP
```

To route `/app` to the `app` service endpoint and `/sidecar` to the `sidecarMonitor` service endpoint, we will configure
the `app` service path rules as the main service route and the `sidecarMonitor` as an additional path rule:

```yaml
ingress:
   enabled: true
   path: /app
   servicePort: 80
   additionalPaths:
     - path: /sidecar
       servicePort: 3000
```

Now suppose you had a sidecar service that will return a fixed response indicating server maintainance and you want to
temporarily route all requests to that endpoint without taking down the pod. You can do this by creating a route that
catches all paths as a higher priority path using the `additionalPathsHigherPriority` input value.

Consider the following `Service`, where we have the `app` served on port 80, and the `sidecarFixedResponse` served on
port 3000:

```yaml
service:
  enabled: true
  type: NodePort
  ports:
    app:
      port: 80
      targetPort: 80
      protocol: TCP
    sidecarFixedResponse:
      port: 3000
      targetPort: 3000
      protocol: TCP
```

To route all traffic to the fixed response port:

```yaml
ingress:
   enabled: true
   path: /app
   servicePort: 80
   additionalPathsHigherPriority:
     - path: /*
       servicePort: 3000
```

The `/*` rule which routes to port 3000 will always be used even when accessing the path `/app` because it will be
evaluated first when routing requests.

back to [root README](/README.adoc#day-to-day-operations)

### How do I expose additional ports?

By default, this Helm Chart will deploy your application container in a Pod that exposes ports 80. Sometimes you might 
want to expose additional ports in your application - for example a separate port for Prometheus metrics. You can expose 
additional ports for your application by overriding `containerPorts` and `service` input values:

```yaml

containerPorts:
  http:
    port: 80
    protocol: TCP
  prometheus:
    port: 2020
    protocol: TCP

service:
  enabled: true
  type: NodePort
  ports:
    app:
      port: 80
      targetPort: 80
      protocol: TCP
    prometheus:
      port: 2020
      targetPort: 2020
      protocol: TCP

```


## How do I deploy a worker service?

Worker services typically do not have a RPC or web server interface to access it. Instead, worker services act on their
own and typically reach out to get the data they need. These services should be deployed without any ports exposed.
However, by default `k8s-service` will deploy an internally exposed service with port 80 open.

To disable the default port, you can use the following `values.yaml` inputs:

```
containerPorts:
  http:
    disabled: true

service:
  enabled: false
```

This will override the default settings such that only the `Deployment` resource is created, with no ports exposed on
the container.

back to [root README](/README.adoc#day-to-day-operations)

## How do I check the status of the rollout?

This Helm Chart packages your application into a `Deployment` controller. The `Deployment` controller will be
responsible with managing the `Pods` of your application, ensuring that the Kubernetes cluster matches the desired state
configured by the chart inputs.

When the Helm Chart installs, `helm` will mark the installation as successful when the resources are created. Under the
hood, the `Deployment` controller will do the work towards ensuring the desired number of `Pods` are up and running.

For example, suppose you set the `replicaCount` variable to 3 when installing this chart. This will configure the
`Deployment` resource to maintain 3 replicas of the `Pod` at any given time, launching new ones if there is a deficit or
removing old ones if there is a surplus.

To see the current status of the `Deployment`, you can query Kubernetes using `kubectl`. The `Deployment` resource of
the chart are labeled with the `applicationName` input value and the release name provided by helm. So for example,
suppose you deployed this chart using the following `values.yaml` file and command:

```yaml
applicationName: nginx
containerImage:
  repository: nginx
  tag: stable
```

```bash
$ helm install -n edge-service gruntwork/k8s-service
```

In this example, the `applicationName` is set to `nginx`, while the release name is set to `edge-service`. This chart
will then install a `Deployment` resource in the default `Namespace` with the following labels that uniquely identifies
it:

```
app.kubernetes.io/name: nginx
app.kubernetes.io/instance: edge-service
```

So now you can query Kubernetes for that `Deployment` resource using these labels to see the state:

```bash
$ kubectl get deployments -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=edge-service"
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
edge-service-nginx   3         3         3            1           24s
```

This includes a few useful information:

- `DESIRED` lists the number of `Pods` that should be running in your cluster.
- `CURRENT` lists how many `Pods` are currently created in the cluster.
- `UP-TO-DATE` lists how many `Pods` are running the desired image.
- `AVAILABLE` lists how many `Pods` are currently ready to serve traffic, as defined by the `readinessProbe`.

When all the numbers are in sync and equal, that means the `Deployment` was rolled out successfully and all the `Pods`
are passing the readiness healthchecks.

In the example output above, note how the `Available` count is `1`, but the others are `3`. This means that all 3 `Pods`
were successfully created with the latest image, but only `1` of them successfully came up. You can dig deeper into the
individual `Pods` to check the status of the unavailable `Pods`. The `Pods` are labeled the same way, so you can pass in
the same label query to get the `Pods` managed by the deployment:

```bash
$ kubectl get pods -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=edge-service"
NAME                                  READY     STATUS    RESTARTS   AGE
edge-service-nginx-844c978df7-f5wc4   1/1       Running   0          52s
edge-service-nginx-844c978df7-mln26   0/1       Pending   0          52s
edge-service-nginx-844c978df7-rdsr8   0/1       Pending   0          52s
```

This will show you the status of each individual `Pod` in your deployment. In this example output, there are 2 `Pods`
that are in the `Pending` status, meaning that they have not been scheduled yet. We can look into why the `Pod` failed
to schedule by getting detailed information about the `Pod` with the `describe` command. Unlike `get pods`, `describe
pod` requires a single `Pod` so we will grab the name of one of the failing `Pods` above and feed it to `describe pod`:

```bash
$ kubectl describe pod edge-service-nginx-844c978df7-mln26
Name:               edge-service-nginx-844c978df7-mln26
Namespace:          default
Priority:           0
PriorityClassName:  <none>
Node:               <none>
Labels:             app.kubernetes.io/instance=edge-service
                    app.kubernetes.io/name=nginx
                    gruntwork.io/app-name=nginx
                    pod-template-hash=4007534893
Annotations:        <none>
Status:             Pending
IP:
Controlled By:      ReplicaSet/edge-service-nginx-844c978df7
Containers:
  nginx:
    Image:        nginx:stable
    Ports:        80/TCP
    Host Ports:   0/TCP
    Environment:  <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-mgkr9 (ro)
Conditions:
  Type           Status
  PodScheduled   False
Volumes:
  default-token-mgkr9:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-mgkr9
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute for 300s
                 node.kubernetes.io/unreachable:NoExecute for 300s
Events:
  Type     Reason            Age               From               Message
  ----     ------            ----              ----               -------
  Warning  FailedScheduling  1m (x25 over 3m)  default-scheduler  0/2 nodes are available: 2 Insufficient pods.
```

This will output detailed information about the `Pod`, including an event log. In this case, the roll out failed because
there is not enough capacity in the cluster to schedule the `Pod`.

back to [root README](/README.adoc#day-to-day-operations)

## How do I set and share configurations with the application?

While you can bake most application configuration values into the application container, you might need to inject
dynamic configuration variables into the container. These are typically values that change depending on the environment,
such as the MySQL database endpoint. Additionally, you might also want a way to securely share secrets with the
container such that they are not hard coded in plain text in the container or in the Helm Chart values yaml file. To
support these use cases, this Helm Chart provides three ways to share configuration values with the application
container:

- [Directly setting environment variables](#directly-setting-environment-variables)
- [Using ConfigMaps](#using-configmaps)
- [Using Secrets](#using-secrets)

### Directly setting environment variables

The simplest way to set a configuration value for the container is to set an environment variable for the container
runtime. These variables are set by Kubernetes before the container application is booted, which can then be looked up
using the standard OS lookup functions for environment variables.

You can use the `envVars` input value to set an environment variable at deploy time. For example, the following entry in
a `values.yaml` file will set the `DB_HOST` environment variable to `mysql.default.svc.cluster.local` and the `DB_PORT`
environment variable to `3306`:

```yaml
envVars:
  DB_HOST: "mysql.default.svc.cluster.local"
  DB_PORT: 3306
```

One thing to be aware of when using environment variables is that they are set at start time of the container. This
means that updating the environment variables require restarting the containers so that they propagate.

### Using ConfigMaps

While environment variables are an easy way to inject configuration values, what if you want to share the configuration
across multiple deployments? If you wish to use the direct environment variables approach, you would have no choice but
to copy paste the values across each deployment. When this value needs to change, you are now faced with going through
each deployment and updating the reference.

For this situation, `ConfigMaps` would be a better option. `ConfigMaps` help decouple configuration values from the
`Deployment` and `Pod` config, allowing you to share the values across the deployments. `ConfigMaps` are dedicated
resources in Kubernetes that store configuration values as key value pairs.

For example, suppose you had a `ConfigMap` to store the database information. You might store the information as two key
value pairs: one for the host (`dbhost`) and one for the port (`dbport`). You can create a `ConfigMap` directly using
`kubectl`, or by using a resource file.

To directly create the `ConfigMap`:

```
kubectl create configmap my-config --from-literal=dbhost=mysql.default.svc.cluster.local --from-literal=dbport=3306
```

Alternatively, you can manage the `ConfigMap` as code using a kubernetes resource config:

```yaml
# my-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  dbhost: mysql.default.svc.cluster.local
  dbport: 3306
```

You can then apply this resource file using `kubectl`:

```
kubectl apply -f my-config.yaml
```

`kubectl` supports multiple ways to seed the `ConfigMap`. You can read all the different ways to create a `ConfigMap` in
[the official
documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#create-a-configmap).

Once the `ConfigMap` is created, you can access the `ConfigMap` within the `Pod` by configuring the access during
deployment. This Helm Chart provides the `configMaps` input value to configure what `ConfigMaps` should be shared with
the application container. There are two ways to inject the `ConfigMap`:

- [Accessing the `ConfigMap` as Environment Variables](#accessing-the-configmap-as-environment-variables)
- [Accessing the `ConfigMap` as Files](#accessing-the-configmap-as-files)

**NOTE**: It is generally not recommended to use `ConfigMaps` to store sensitive data. For those use cases, use
`Secrets` or an external secret store.

##### Accessing the ConfigMap as Environment Variables

You can set the values of the `ConfigMap` as environment variables in the application container. To do so, you set the
`as` attribute of the `configMaps` input value to `environment`. For example, to share the `my-config` `ConfigMap` above
using the same environment variables as the example in [Directly setting environment
variables](#directly-settings-environment-variables), you would set the `configMaps` as follows:

```yaml
configMaps:
  my-config:
    as: environment
    items:
      dbhost:
        envVarName: DB_HOST
      dbport:
        envVarName: DB_PORT
```

In this configuration for the Helm Chart, we specify that we want to share the `my-config` `ConfigMap` as environment
variables with the main application container. Additionally, we want to map the `dbhost` config value to the `DB_HOST`
environment variable, and similarly map the `dbport` config value to the `DB_PORT` environment variable.

Note that like directly setting environment variables, these are set at container start time, and thus the containers
need to be restarted when the `ConfigMap` is updated for the new values to be propagated. You can use files instead if
you wish the `ConfigMap` changes to propagate immediately.

##### Accessing the ConfigMap as Files

You can mount the `ConfigMap` values as files on the container filesystem. To do so, you set the `as` attribute of the
`configMaps` input value to `volume`.

For example, suppose you wanted to share the `my-config` `ConfigMap` above as the files `/etc/db/host` and
`/etc/db/port`. For this case, you would set the `configMaps` input value to:

```yaml
configMaps:
  my-config:
    as: volume
    mountPath: /etc/db
    items:
      dbhost:
        filePath: host
      dbport:
        filePath: port
```

In the container, now the values for `dbhost` is stored as a text file at the path `/etc/db/host` and `dbport` is stored
at the path `/etc/db/port`. You can then read these files in in your application to get the values.

Unlike environment variables, using files has the advantage of immediately reflecting changes to the `ConfigMap`. For
example, when you update `my-config`, the files at `/etc/db` are updated automatically with the new values, without
needing a redeployment to propagate the new values to the container.

### Using Secrets

In general, it is discouraged to store sensitive information such as passwords in `ConfigMaps`. Instead, Kubernetes
provides `Secrets` as an alternative resource to store sensitive data. Similar to `ConfigMaps`, `Secrets` are key value
pairs that store configuration values that can be managed independently of the `Pod` and containers. However, unlike
`ConfigMaps`, `Secrets` have the following properties:

- A secret is only sent to a node if a pod on that node requires it. They are automatically garbage collected when there
  are no more `Pods` referencing it on the node.
- A secret is stored in `tmpfs` on the node, so that it is only available in memory.
- Starting with Kubernetes 1.7, they can be encrypted at rest in `etcd` (note: this feature was in alpha state until
  Kubernetes 1.13).

You can read more about the protections and risks of using `Secrets` in [the official
documentation](https://kubernetes.io/docs/concepts/configuration/secret/#security-properties).

Creating a `Secret` is very similar to creating a `ConfigMap`. For example, suppose you had a `Secret` to store the
database password. Like `ConfigMaps`, you can create a `Secret` directly using `kubectl`:

```
kubectl create secret generic my-secret --from-literal=password=1f2d1e2e67df
```

The `generic` keyword indicates the `Secret` type. Almost all use cases for your application should use this type. Other
types include `docker-registry` for specifying credentials for accessing a private docker registry, and `tls` for
specifying TLS certificates to access the Kubernetes API.

You can also manage the `Secret` as code, although you may want to avoid this for `Secrets` to avoid leaking them in
unexpected locations (e.g source control). Unlike `ConfigMaps`, `Secrets` require values to be stored as base64 encoded
values when using resource files. So the configuration for the above example will be:

```yaml
# my-secret.yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: my-secret
data:
  password: MWYyZDFlMmU2N2Rm
```

Note that `MWYyZDFlMmU2N2Rm` is the base 64 encoded version of `1f2d1e2e67df`. You can then apply this resource config
using `kubectl`:

```
kubectl apply -f my-secret.yaml
```

Similar to `ConfigMaps`, this Helm Chart supports two ways to inject `Secrets` into the application container: as
environment variables, or as files. The syntax to share the values is very similar to the `configMaps` input value, only
you use the `secrets` input value. The properties of each approach is very similar to `ConfigMaps`. Refer to [the
previous section](#using-configmaps) for more details on each approach. Here, we show you examples of the input values
to use for each approach.

**Mounting secrets as environment variables**: In this example, we mount the `my-secret` `Secret` created above as the
environment variable `DB_PASSWORD`.

```yaml
secrets:
  my-secret:
    as: environment
    items:
      password:
        envVarName: DB_PASSWORD
```

**Mounting secrets as files**: In this example, we mount the `my-secret` `Secret` as the file `/etc/db/password`.

```yaml
secrets:
  my-secret:
    as: volume
    mountPath: /etc/db
    items:
      password:
        filePath: password
```

**Mounting secrets with CSI**: In this example, we mount the `my-secret` `Secret` as the file `/etc/db`, and specify that the secret will sync with Secret Manager store (AWS, GCP, Vault) secret named `my-secret`. We also details the csi block were we define the driver and secreteProviderClass.

```yaml
secrets:
  my-secret:
    as: csi
    mountPath: /etc/db
    readOnly: true
    csi:
      driver: secrets-store.csi.k8s.io
      secretProviderClass: secret-provider-class
    items:
      my-secret:
        envVarName: SECRET_VAR
```

**NOTE**: The volumes are different between `secrets` and `configMaps`. This means that if you use the same `mountPath`
for different secrets and config maps, you can end up with only one. It is undefined which `Secret` or `ConfigMap` ends
up getting mounted. To be safe, use a different `mountPath` for each one.

**NOTE**: If you want mount the volumes created with `secrets` or `configMaps` on your init or sidecar containers, you will 
have to append `-volume` to the volume name in . In the example above, the resulting volume will be `my-secret-volume`.

**Note** When installing the CSI driver on your cluster you have an option to activate syncing of secrets 

```yaml
sideCarContainers:
  sidecar:
    image: sidecar/container:latest
    volumeMounts:
    - name: my-secret-volume
      mountPath: /etc/db
```

### Which configuration method should I use?

Which configuration method you should use depends on your needs. Here is a summary of the pro and con of each
approach:

##### Directly setting environment variables

**Pro**:

- Simple setup
- Manage configuration values directly with application deployment config
- Most application languages support looking up environment variables

**Con**:

- Tightly couple configuration settings with application deployment
- Requires redeployment to update values
- Must store in plain text, and easy to leak into VCS

**Best for**:

- Iterating different configuration values during development
- Sotring non-sensitive values that are unique to each environment / deployment

##### Using ConfigMaps

**Pro**:

- Keep config DRY by sharing a common set of configurations
- Independently update config values from the application deployment
- Automatically propagate new values when stored as files

**Con**:

- More overhead to manage the configuration
- Stored in plain text
- Available on all nodes automatically

**Best for**:

- Storing non-sensitive common configuration that are shared across environments
- Storing non-sensitive dynamic configuration values that change frequently

##### Using Secrets

**Pro**:

- All the benefits of using `ConfigMaps`
- Can be encrypted at rest
- Opaque by default when viewing the values (harder to remember base 64 encoded version of "admin")
- Only available to nodes that use it, and only in memory

**Con**:

- All the challenges of using `ConfigMaps`
- Configured in plain text, making it difficult to manage as code securely
- Less safe than using dedicated secrets manager / store like HashiCorp Vault.

**Best for**:

- Storing sensitive configuration values

back to [root README](/README.adoc#day-to-day-operations)

## How do you update the application to a new version?

To update the application to a new version, you can upgrade the Helm Release using updated values. For example, suppose
you deployed `nginx` version 1.15.4 using this Helm Chart with the following values:

```yaml
containerImage:
  repository: nginx
  tag: 1.15.4

applicationName: nginx
```

In this example, we will further assume that you deployed this chart with the above values using the release name
`edge-service`, using a command similar to below:

```bash
$ helm install -f values.yaml --name edge-service gruntwork/k8s-service
```

Now let's try upgrading `nginx` to version 1.15.8. To do so, we will first update our values file:

```yaml
containerImage:
  repository: nginx
  tag: 1.15.8

applicationName: nginx
```

The only difference here is the `tag` of the `containerImage`.

Next, we will upgrade our release using the updated values. To do so, we will use the `helm upgrade` command:

```bash
$ helm upgrade -f values.yaml edge-service gruntwork/k8s-service
```

This will update the created resources with the new values provided by the updated `values.yaml` file. For this example,
the only resource that will be updated is the `Deployment` resource, which will now have a new `Pod` spec that points to
`nginx:1.15.8` as opposed to `nginx:1.15.4`. This automatically triggers a rolling deployment internally to Kubernetes,
which will launch new `Pods` using the latest image, and shut down old `Pods` once those are ready.

You can read more about how changes are rolled out on `Deployment` resources in [the official
documentation](https://kubernetes.io/docs/concepts/workloads/controllers/deployment).

Note that certain changes will lead to a replacement of the `Deployment` resource. For example, updating the
`applicationName` will cause the `Deployment` resource to be deleted, and then created. This can lead to down time
because the resources are replaced in an uncontrolled fashion.

## How do I create a canary deployment?

You may optionally configure a [canary deployment](https://martinfowler.com/bliki/CanaryRelease.html) of an arbitrary tag that will run as an individual deployment behind your configured service. This is useful for ensuring a new application tag runs without issues prior to fully rolling it out.

To configure a canary deployment, set `canary.enabled = true` and define the `containerImage` values. Typically, you will want to specify the tag of your next release candidate:

```yaml
canary:
    enabled: true
    containerImage:
        repository: nginx
        tag: 1.15.9
```
Once deployed, your service will route traffic across both your stable and canary deployments, allowing you to monitor for and catch any issues early.

back to [root README](/README.adoc#major-changes)

## How do I verify my canary deployment?

Canary deployment pods have the same name as your stable deployment pods, with the additional `-canary` appended to the end, like so:

```bash
$ kubectl get pods -l "app.kubernetes.io/name=nginx,app.kubernetes.io/instance=edge-service"
NAME                                          READY     STATUS    RESTARTS   AGE
edge-service-nginx-844c978df7-f5wc4           1/1       Running   0          52s
edge-service-nginx-844c978df7-mln26           0/1       Pending   0          52s
edge-service-nginx-844c978df7-rdsr8           0/1       Pending   0          52s
edge-service-nginx-canary-844c978df7-bsr8     0/1       Pending   0          52s
```

Therefore, in this example, you could monitor your canary by running `kubectl logs -f edge-service-nginx-canary-844c978df7-bsr8`

back to [root README](/README.adoc#day-to-day-operations)

## How do I roll back a canary deployment?

Update your values.yaml file, setting `canary.enabled = false` and then upgrade your helm installation:

```bash
$ helm upgrade -f values.yaml edge-service gruntwork/k8s-service
```
Following this update, Kubernetes will determine that your canary deployment is no longer desired and will delete it.

back to [root README](/README.adoc#day-to-day-operations)

## How do I ensure a minimum number of Pods are available across node maintenance?

Sometimes, you may want to ensure that a specific number of `Pods` are always available during [voluntary
maintenance](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).
This chart exposes an input value `minPodsAvailable` that can be used to specify a minimum number of `Pods` to maintain
during a voluntary maintenance activity. Under the hood, this chart will create a corresponding `PodDisruptionBudget` to
ensure that a certain number of `Pods` are up before attempting to terminate additional ones.

You can read more about `PodDisruptionBudgets` in [our blog post covering the
topic](https://blog.gruntwork.io/avoiding-outages-in-your-kubernetes-cluster-using-poddisruptionbudgets-ef6a4baa5085)
and in [the official
documentation](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#how-disruption-budgets-work).


back to [root README](/README.adoc#major-changes)

## Why does the Pod have a preStop hook with a Shutdown Delay?

When a `Pod` is removed from a Kubernetes cluster, the control plane notifies all nodes to remove the `Pod` from
registered addresses. This includes removing the `Pod` from the list of available `Pods` to service a `Service`
endpoint. However, because Kubernetes is a distributed system, there is a delay between the shutdown sequence and the
`Pod` being removed from available addresses. As a result, the `Pod` could still get traffic despite it having already
been shutdown on the node it was running on.

Since there is no way to guarantee that the deletion has propagated across the cluster, we address this eventual
consistency issue by adding an arbitrary delay between the `Pod` being deleted and the initiation of the `Pod` shutdown
sequence. This is accomplished by adding a `sleep` command in the `preStop` hook.

You can control the length of time to delay with the `shutdownDelay` input value. You can also disable this behavior by
setting the `shutdownDelay` to 0.

You can read more about this topic in [our blog post
"Delaying Shutdown to Wait for Pod Deletion
Propagation"](https://blog.gruntwork.io/delaying-shutdown-to-wait-for-pod-deletion-propagation-445f779a8304).


back to [root README](/README.adoc#day-to-day-operations)

## What is a sidecar container?

In Kubernetes, `Pods` are one or more tightly coupled containers that are deployed together. The containers in the `Pod`
share, amongst other things, the network stack, the IPC namespace, and in some cases the PID namespace. You can read
more about the resources that the containers in a `Pod` share in [the official
documentation](https://kubernetes.io/docs/concepts/workloads/pods/pod/#what-is-a-pod).

Sidecar Containers are additional containers that you wish to deploy in the `Pod` housing your application container.
This helm chart supports deploying these containers by configuring the `sideCarContainers` input value. This input value
is a map between the side car container name and the values of the container spec. The spec is rendered directly into
the `Deployment` resource, with the `name` being set to the key. For example:

```yaml
sideCarContainers:
  datadog:
    image: datadog/agent:latest
    env:
      - name: DD_API_KEY
        value: ASDF-1234
      - name: SD_BACKEND
        value: docker
  nginx:
    image: nginx:1.15.4
```

This input will be rendered in the `Deployment` resource as:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  ... Snipped for brevity ...
spec:
  ... Snipped for brevity ...
  template:
    spec:
      containers:
        ... The first entry relates to the application ...
        - name: datadog
          image: datadog/agent:latest
          env:
            - name: DD_API_KEY
              value: ASDF-1234
            - name: SD_BACKEND
              value: docker
        - name: nginx
          image: nginx:1.15.4
```

In this config, the side car containers are rendered as additional containers to deploy alongside the main application
container configured by the `containerImage`, `ports`, `livenessProbe`, etc input values. Note that the
`sideCarContainers` variable directly renders the spec, meaning that the additional values for the side cars such as
`livenessProbe` should be rendered directly within the `sideCarContainers` input value.

back to [root README](/README.adoc#core-concepts)

## How do I use a private registry?

To pull container images from a private registry, the Kubernetes cluster needs to be able to authenticate to the docker
registry with a registry key. On managed Kubernetes clusters (e.g EKS, GKE, AKS), this is automated through the server
IAM roles that are assigned to the instance VMs. In most cases, if the instance VM IAM role has the permissions to
access the registry, the Kubernetes cluster will automatically be able to pull down images from the respective managed
registry (e.g ECR on EKS or GCR on GKE).

Alternatively, you can specify docker registry keys in the Kubernetes cluster as `Secret` resources. This is helpful in
situations where you do not have the ability to assign registry access IAM roles to the node itself, or if you are
pulling images off of a different registry (e.g accessing GCR from EKS cluster).

You can use `kubectl` to create a `Secret` in Kubernetes that can be used as a docker registry key:

```
kubectl create secret docker-registry NAME \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-password=DOCKER_PASSWORD \
  --docker-email=DOCKER_EMAIL
```

This command will create a `Secret` resource named `NAME` that holds the specified docker registry credentials. You can
then specify the cluster to use this `Secret` when pulling down images for the service `Deployment` in this chart by
using the `imagePullSecrets` input value:

```
imagePullSecrets:
  - NAME
```

You can learn more about using private registries with Kubernetes in [the official
documentation](https://kubernetes.io/docs/concepts/containers/images/#using-a-private-registry).

back to [root README](/README.adoc#day-to-day-operations)
