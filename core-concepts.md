# Background

## What is Kubernetes?

[Kubernetes](https://kubernetes.io) is an open source container management system for deploying, scaling, and managing
containerized applications. Kubernetes is built by Google based on their internal proprietary container management
systems (Borg and Omega). Kubernetes provides a cloud agnostic platform to deploy your containerized applications with
built in support for common operational tasks such as replication, autoscaling, self-healing, and rolling deployments.

You can learn more about Kubernetes from [the official documentation](https://kubernetes.io/docs/tutorials/kubernetes-basics/).


## What is Helm?

[Helm](https://helm.sh/) is a package and module manager for Kubernetes that allows you to define, install, and manage
Kubernetes applications as reusable packages called Charts. Helm provides support for official charts in their
repository that contains various applications such as Jenkins, MySQL, and Consul to name a few. Gruntwork uses Helm
under the hood for the Kubernetes modules in this package.

The Helm client is a command line utility that is the primary interface for installing and managing Charts as releases
in the Helm ecosystem. In addition to providing operational interfaces (e.g install, upgrade, list, etc), the client
also provides utilities to support local development of Charts in the form of a scaffolding command and repository
management (e.g uploading a Chart).


## How do you run applications on Kubernetes?

There are three different ways you can schedule your application on a Kubernetes cluster. In all three, your application
Docker containers are packaged as a [Pod](https://kubernetes.io/docs/concepts/workloads/pods/pod/), which are the
smallest deployable unit in Kubernetes, and represent one or more Docker containers that are tightly coupled. Containers
in a Pod share certain elements of the kernel space that are traditionally isolated between containers, such as the
network space (the containers both share an IP and thus the available ports are shared), IPC namespace, and PIDs in some
cases.

Pods are considered to be relatively ephemeral disposable entities in the Kubernetes ecosystem. This is because Pods are
designed to be mobile across the cluster so that you can design a scalable fault tolerant system. As such, Pods are
generally scheduled with
[Controllers](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/#pods-and-controllers) that manage the
lifecycle of a Pod. Using Controllers, you can schedule your Pods as:

- Jobs, which are Pods with a controller that will guarantee the Pods run to completion. See the [k8s-job
  chart](/charts/k8s-job) for more information.
- Deployments behind a Service, which are Pods with a controller that implement lifecycle rules to provide replication
  and self-healing capabilities. Deployments will automatically reprovision failed Pods, or migrate Pods to healthy
  nodes off of failed nodes. A Service constructs a consistent endpoint that can be used to access the Deployment. See
  the [k8s-service chart](/charts/k8s-service) for more information.
- Daemon Sets, which are Pods that are scheduled on all worker nodes. Daemon Sets schedule exactly one instance of a Pod
  on each node. Like Deployments, Daemon Sets will reprovision failed Pods and schedule new ones automatically on
  new nodes that join the cluster. See the [k8s-daemon-set chart](/charts/k8s-daemon-set) for more information.
