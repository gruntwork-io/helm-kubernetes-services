# Tests

<!-- TODO: Replace with helm context -->

This folder contains automated tests for this Module. All of the tests are written in [Go](https://golang.org/).
Most of these are "integration tests" that deploy real infrastructure using Terraform and verify that infrastructure
works as expected using a helper library called [Terratest](https://github.com/gruntwork-io/terratest).



## WARNING WARNING WARNING

**Note #1**: Many of these tests create real resources in an AWS account and then try to clean those resources up at 
the end of a test run. That means these tests may cost you money to run! When adding tests, please be considerate of 
the resources you create and take extra care to clean everything up when you're done!

**Note #2**: Never forcefully shut the tests down (e.g. by hitting `CTRL + C`) or the cleanup tasks won't run!

**Note #3**: We set `-timeout 60m` on all tests not because they necessarily take that long, but because Go has a
default test timeout of 10 minutes, after which it forcefully kills the tests with a `SIGQUIT`, preventing the cleanup
tasks from running. Therefore, we set an overlying long timeout to make sure all tests have enough time to finish and 
clean up.



## Running the tests

### Prerequisites

- Install the latest version of [Go](https://golang.org/).
- Install [dep](https://github.com/golang/dep) for Go dependency management.
- Install [Terraform](https://www.terraform.io/downloads.html).
- Install [pyenv](https://github.com/pyenv/pyenv).
- Install the following versions of python with pyenv:
    - 2.7.12
    - 3.5.2

- Configure your AWS credentials using one of the [options supported by the AWS 
  SDK](http://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html). Usually, the easiest option is to
  set the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.


### One-time setup

Download Go dependencies using dep:

```
cd test
dep ensure
```

Download python dependencies using pip:

```
cd test/script_tests
pip install -r requirements.txt
```


### Run all the tests

#### Terratest

```bash
cd test
go test -v -timeout 60m
```

#### Python scripts

```bash
cd test/script_tests
tox
```


### Run a specific test

#### Terratest

To run a specific test called `TestFoo`:

```bash
cd test
go test -v -timeout 60m -run TestFoo
```

#### Python scripts

TODO


## Known instabilities in test

- `TestEKSCluster` will sometimes fail with:

  ```
  kubernetes_config_map.eks_to_k8s_role_mapping: Post https://5D6DB4CE12C35AD89507342004154717.yl4.us-east-1.eks.amazonaws.com/api/v1/namespaces/kube-system/configmaps: dial tcp 23.23.30.22:443: i/o timeout
  ```

  This is a known issue where the EKS API does not come up immediately and there is a delay between the creation
  completing, and the API actually being available. This is inconsistent, so try again to work around this.
