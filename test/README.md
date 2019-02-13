# Tests

This folder contains automated tests for this Module. All of the tests are written in [Go](https://golang.org/).

There are three tiers of tests for helm:

- Template tests: These are tests designed to test the logic of the templates. These tests should run `helm template`
  with various input values and parse the yaml to validate any logic embedded in the templates (e.g by reading them in
  using client-go). Since templates are not statically typed, the goal of these tests is to promote fast cycle time
  while catching some of the common bugs from typos or logic errors before getting to the slower integration tests.
- Integration tests: These are tests that are designed to deploy the infrastructure and validate that it actually
  works as expected. If you consider the template tests to be syntactic tests, these are semantic tests that validate
  the behavior of the deployed resources.
- Production tests (helm tests): These are tests that are run with the helm chart after it is deployed to validate the
  chart installed and deployed correctly. These should be smoke tests with minimal validation to ensure that the common
  operator errors during deployment are captured as early as possible. Note that because these tests run even on a
  production system, they should be passive and not destructive.

This folder contains the "template tests" and "integration tests". Both types of tests use a helper library called
[Terratest](https://github.com/gruntwork-io/terratest). While "template tests" do not need any infrastructure, the
"integration tests" deploy the charts to a Kubernetes cluster.



## WARNING WARNING WARNING

**Note #1**: Many of these tests create real resources in a Kubernetes cluster and then try to clean those resources up at
the end of a test run. That means these tests may potentially pollute your Kubernetes cluster with unnecessary
resources! When adding tests, please be considerate of the resources you create and take extra care to clean everything
up when you're done!

**Note #2**: Never forcefully shut the tests down (e.g. by hitting `CTRL + C`) or the cleanup tasks won't run!

**Note #3**: We set `-timeout 60m` on all tests not because they necessarily take that long, but because Go has a
default test timeout of 10 minutes, after which it forcefully kills the tests with a `SIGQUIT`, preventing the cleanup
tasks from running. Therefore, we set an overlying long timeout to make sure all tests have enough time to finish and
clean up.



## Running the tests

### Prerequisites

- Install the latest version of [Go](https://golang.org/).
- Install [dep](https://github.com/golang/dep) for Go dependency management.
- Setup a Kubernetes cluster. We recommend using a local version for fast iteration:
    - Linux: [minikube](https://github.com/kubernetes/minikube)
    - Mac OSX: [Kubernetes on Docker For Mac](https://docs.docker.com/docker-for-mac/kubernetes/)
    - Windows: [Kubernetes on Docker For Windows](https://docs.docker.com/docker-for-windows/kubernetes/)
- Install and setup [helm](https://docs.helm.sh/using_helm/#installing-helm)

### One-time setup

Download Go dependencies using dep:

```
cd test
dep ensure
```

### Run all the tests

We use build tags to categorize the tests. The tags are:

- `all`: Run all the tests
- `tpl`: Run the template tests
- `integration`: Run the integration tests

You can run all the tests by passing the `all` build tag:

```bash
cd test
go test -v -tags all -timeout 60m
```

### Run a specific test

To run a specific test called `TestFoo`:

```bash
cd test
go test -v -timeout 60m -tags all -run TestFoo
```

### Run just the template tests

Since the integration tests require infrastructure, they can be considerably slower than the unit tests. As such, to
promote fast test cycles, you may want to test just the template tests. To do so, you can pass the `tpl` build tag:

```bash
cd test
go test -v -tags tpl
```
