# Contribution Guidelines

Contributions to this Package are very welcome! We follow a fairly standard [pull request process](
https://help.github.com/articles/about-pull-requests/) for contributions, subject to the following guidelines:

1. [File a GitHub issue](#file-a-github-issue)
1. [Update the documentation](#update-the-documentation)
1. [Update the tests](#update-the-tests)
1. [Update the code](#update-the-code)
1. [Create a pull request](#create-a-pull-request)
1. [Merge and release](#merge-and-release)

## File a GitHub issue

Before starting any work, we recommend filing a GitHub issue in this repo. This is your chance to ask questions and
get feedback from the maintainers and the community before you sink a lot of time into writing (possibly the wrong)
code. If there is anything you're unsure about, just ask!

## Update the documentation

We recommend updating the documentation *before* updating any code (see [Readme Driven
Development](http://tom.preston-werner.com/2010/08/23/readme-driven-development.html)). This ensures the documentation
stays up to date and allows you to think through the problem at a high level before you get lost in the weeds of
coding.

## Update the tests

We also recommend updating the automated tests *before* updating any code (see [Test Driven
Development](https://en.wikipedia.org/wiki/Test-driven_development)). That means you add or update a test case,
verify that it's failing with a clear error message, and *then* make the code changes to get that test to pass. This
ensures the tests stay up to date and verify all the functionality in this Module, including whatever new
functionality you're adding in your contribution. Check out the
[tests](https://github.com/gruntwork-io/helm-kubernetes-services/tree/master/test) folder for instructions on running
the automated tests.

Note that at Gruntwork, there are three tiers of tests for helm:

- Template tests: These are tests designed to test the logic of the templates. These tests should run `helm template`
  with various input values and parse the yaml to validate any logic embedded in the templates (e.g by reading them in
  using client-go). Since templates are not statically typed, the goal of these tests is to promote fast cycle time
  while catching some of the common bugs from typos or logic errors before getting to the slower integration tests.
- Integration tests: These are tests that are designed to deploy the infrastructure and validate the resource
  configurations. If you consider the templates to be syntactic tests, these are semantic tests that validate the
  behavior of the deployed resources.
- Production tests (helm tests): These are tests that are run with the helm chart after it is deployed to validate the chart
  installed and deployed correctly. These should be smoke tests with minimal validation to ensure that the common
  operator errors are captured as early as possible. Note that because these tests run even on a production system, they
  should be passive and not destructive.

## Update the code

At this point, make your code changes and use your new test case to verify that everything is working. As you work,
keep in mind two things:

1. Backwards compatibility
1. Downtime

### Backwards compatibility

Please make every effort to avoid unnecessary backwards incompatible changes. With Helm charts, this means:

1. Do not delete, rename, or change the type of input variables.
1. If you add an input variable, set a default in `values.yaml`.
1. Do not delete, rename, or change the type of output variables.
1. Do not delete or rename a chart in the `charts` folder.

If a backwards incompatible change cannot be avoided, please make sure to call that out when you submit a pull request,
explaining why the change is absolutely necessary.

### Downtime

Bear in mind that the Helm charts in this Module are used by real companies to run real infrastructure in
production, and certain types of changes could cause downtime. If downtime cannot be avoided, please make sure to call
that out when you submit a pull request.


### Formatting and pre-commit hooks

You must run `helm lint` on the code before committing. You can configure your computer to do this automatically
using pre-commit hooks managed using [pre-commit](http://pre-commit.com/):

1. [Install pre-commit](http://pre-commit.com/#install). E.g.: `brew install pre-commit`.
1. Install the hooks: `pre-commit install`.
1. Make sure you have the helm client installed. See [the official docs](https://docs.helm.sh/using_helm/#install-helm)
   for instructions.

That's it! Now just write your code, and every time you commit, `helm lint` will be run on the charts that you modify.


## Create a pull request

[Create a pull request](https://help.github.com/articles/creating-a-pull-request/) with your changes. Please make sure
to include the following:

1. A description of the change, including a link to your GitHub issue.
1. The output of your automated test run, preferably in a [GitHub Gist](https://gist.github.com/). We cannot run
   automated tests for pull requests automatically due to [security
   concerns](https://circleci.com/docs/fork-pr-builds/#security-implications), so we need you to manually provide this
   test output so we can verify that everything is working.
1. Any notes on backwards incompatibility or downtime.

## Merge and release

The maintainers for this repo will review your code and provide feedback. If everything looks good, they will merge the
code and release a new version, which you'll be able to find in the [releases page](../../releases).
