# opcert - The Operator Enablement Test Suite for Certified Operators

opcert is a tool to run operator image and bundle certification tests. This project has a companion container image that has the opcert binary in order to run tests from the container. That will be used to integrate with operator-sdk's scorecard subcommand in order to provide partners a way to test locally their operators without having to use yet another tool.

This very first alpha 0.0.1 version has only one test that already complies with the operator-sdk scorecard standard. It can be tested as below:

To use it locally in your own laptop or personal computer download it from the releases section on https://github.com/opdev/opcert.

The opcert tool defaults to podman when testing images.

So for Docker users it's necessary to use the builder flag like below:
```
opcert -builder docker -test has_labels -image <YOUR IMAGE:TAG HERE>
```
For a succeed image you should see something like this:

`opcert -test has_labels -image registry.access.redhat.com/ubi8:latest`

```
registry.access.redhat.com/ubi8:latest
{
    "results": [
        {
            "name": "Has Labels",
            "state": "pass"
        }
    ]
}
```

For a failed test you should see something like this:

`opcert -test has_labels -image centos:latest`

```
centos:latest
{
    "results": [
        {
            "name": "Has Labels",
            "state": "fail",
            "errors": [
                "Label name not present.",
                "Label vendor not present.",
                "Label version not present.",
                "Label release not present.",
                "Label summary not present.",
                "Label description not present."
            ],
            "suggestions": [
                "Please include label name",
                "Please include label vendor",
                "Please include label version",
                "Please include label release",
                "Please include label summary",
                "Please include label description"
            ]
        }
    ]
}
```

Finally if you want to run from a container image, you don't need to clone the project. Below is what you may try (keep in mind that it may take a while to pull the image from the container):

`docker run -it --privileged quay.io/opdev/opcert:0.0.1 /scorecard/certified/opcert -test has_labels -image centos:latest`

or

`podman run -it --privileged quay.io/opdev/opcert:0.0.1 /scorecard/certified/opcert -test has_labels -image centos:latest`

```
centos:latest
{
    "results": [
        {
            "name": "Has Labels",
            "state": "fail",
            "errors": [
                "Label name not present.",
                "Label vendor not present.",
                "Label version not present.",
                "Label release not present.",
                "Label summary not present.",
                "Label description not present."
            ],
            "suggestions": [
                "Please include label name",
                "Please include label vendor",
                "Please include label version",
                "Please include label release",
                "Please include label summary",
                "Please include label description"
            ]
        }
    ]
}
```
