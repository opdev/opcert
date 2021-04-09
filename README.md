# opcert - The Operator Enablement Test Suite for Certified Operators

opcert is a tool to run image and bundle certification tests. This project has a companion container image that has opcert as an entry point that will be used to integrate with operator-sdk's scorecard subcommand in order to provide partners a way to test locally their operators.

This very first alpha 0.0.1 version has only one test that already complies with the operator-sdk scorecard standard. It can be tested as below:

```
git clone https://github.com/opdev/opcert.git
cd opcert/build
```

To use it locally in your own laptop follow the instructions below:

For Linux users:
```
build/linux/opcert has_labels <YOUR IMAGE:TAG HERE>
```
For Mac users:
```
build/macos/opcert has_labels <YOUR IMAGE:TAG HERE>
```
For a succeed image you should see something like this:

`macos/opcert has_labels registry.access.redhat.com/ubi8:latest`

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

`macos/opcert has_labels centos:latest`

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

Finally if you want to run it in a container image here is what you may try something like the command below (keep in mind that it may take a while to pull the image from the container):

`docker run -it --privileged quay.io/opdev/opcert:0.0.1 /scorecard/certified/opcert has_labels centos:latest`

or

`podman run -it --privileged quay.io/opdev/opcert:0.0.1 /scorecard/certified/opcert has_labels centos:latest`

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
