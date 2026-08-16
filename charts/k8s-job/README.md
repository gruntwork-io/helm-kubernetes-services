# Kubernetes Job Helm Chart

This Helm Chart can be used to deploy your job container under a
[Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) resource onto your Kubernetes
cluster. You can use this Helm Chart to run and deploy a one time job or periodic task such as a security scanner application or data science pipeline job.


## How to use this chart?

* See the [root README](/README.adoc) for general instructions on using Gruntwork Helm Charts.
* See the [examples](/examples) folder for example usage.
* See the provided [values.yaml](./values.yaml) file for the required and optional configuration values that you can set
  on this chart.

back to [root README](/README.adoc#core-concepts)

## What resources does this Helm Chart deploy?

The following resources will be deployed with this Helm Chart, depending on which configuration values you use:

- `Job`: A standalone `Job` running the image specified in the
                `containerImage` input value.

back to [root README](/README.adoc#core-concepts)

## How do I deploy additional resources not managed by the chart?

You can create custom Kubernetes resources, that are not directly managed by the chart, within the `customResources`
key. You provide each resource manifest directly as a value under `customResources.resources` and set
`customResources.enabled` to `true`. For examples of custom resources, take a look at the examples in
[test/fixtures/custom_resources_values.yaml](../../test/fixtures/custom_resources_values.yaml) and
[test/fixtures/multiple_custom_resources_values.yaml](../../test/fixtures/multiple_custom_resources_values.yaml).

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
the application container. With a single-standing Job there is one way to access a `ConfigMap`:

- [Accessing the `ConfigMap` as Environment Variables](#accessing-the-configmap-as-environment-variables)

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

**NOTE**: The volumes are different between `secrets` and `configMaps`. This means that if you use the same `mountPath`
for different secrets and config maps, you can end up with only one. It is undefined which `Secret` or `ConfigMap` ends
up getting mounted. To be safe, use a different `mountPath` for each one.

**NOTE**: If you want mount the volumes created with `secrets` or `configMaps` on your init or sidecar containers, you will 
have to append `-volume` to the volume name in . In the example above, the resulting volume will be `my-secret-volume`.

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
