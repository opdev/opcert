package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/opdev/opcert/pkg/opcert"
	"github.com/savaki/jq"

	scapiv1alpha3 "github.com/operator-framework/api/pkg/apis/scorecard/v1alpha3"
)

func main() {
	entrypoint := os.Args[1:]
	if len(entrypoint) == 0 {
		log.Fatal("Test name argument is required")
	}

	// Read the pod's untar'd bundle from a well-known path.
	// cfg, err := apimanifests.GetBundleFromDir(PodBundleRoot)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	img := os.Args[2]
	builder := "docker"

	fmt.Println(img)
	var result scapiv1alpha3.TestStatus

	opcert := opcert.OpCert{}

	opcert.Init(builder, img)

	// Names of the custom tests which would be passed in the
	// `operator-sdk` command.
	switch entrypoint[0] {
	case IsRHELTest:
		result = IsRHEL(&opcert)
	case HasLabelsTest:
		result = HasLabels(img)
	default:
		result = printValidTests()
	}

	// Convert scapiv1alpha3.TestResult to json.
	prettyJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Printf("%s\n", string(prettyJSON))

}

// printValidTests will print out full list of test names to give a hint to the end user on what the valid tests are.
func printValidTests() scapiv1alpha3.TestStatus {
	result := scapiv1alpha3.TestResult{}
	result.State = scapiv1alpha3.FailState
	result.Errors = make([]string, 0)
	result.Suggestions = make([]string, 0)

	str := fmt.Sprintf("Valid tests for this image include: %s %s",
		IsRHELTest,
		HasLabelsTest)
	result.Errors = append(result.Errors, str)
	return scapiv1alpha3.TestStatus{
		Results: []scapiv1alpha3.TestResult{result},
	}
}

const (
	IsRHELTest    = "is_rhel"
	HasLabelsTest = "has_labels"
)

// Mandatory tests with possible fail results:

// 1. Container must be use a base image provided by Red Hat. (verify FROM clause (buildah inspect) against Red Hat catalog list or maybe the ubis at first)
// Test names: is_rhel, has_base_rh_image, repo_list_successful
// Why? So the application's runtime dependencies, such as operating system components and libraries, are fully
// supported.
// How? Go to the Red Hat Container Catalog and select a base image to build upon. Use this image's name in
// the FROM clause in your dockerfile. We recommend using one of the images that are part of the Red Hat
// Universal Base Image (UBI) set, such as ubi7/ubi or ubi7/ubi-minimal.

func IsRHEL(o *opcert.OpCert) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = "IsRHEL"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	//  pull base image from catalog
	err := o.PullImage(o.BaseImage)
	if err != nil {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "couldn't pull image from registry.access.redhat.com")
		r.Suggestions = append(r.Suggestions, "Verify if the base image is provided by Red Hat and available at registry.access.redhat.com")
	}

	// get Red Hat base image layer sha256 digests from manifest
	rhImgLayers, err := o.GetImageLayers(o.BaseImage)
	if err != nil {
		fmt.Println(err)
	}

	// compare partner base image layers with Red Hat's base image layers
	for i, layer := range rhImgLayers {
		if o.LayerDigests[i] != layer {
			r.State = scapiv1alpha3.FailState
			r.Errors = append(r.Errors, "Base image layer "+o.LayerDigests[i]+" doesn't match Red Hat's layer "+layer)
		}
	}
	if r.State == scapiv1alpha3.FailState {
		r.Suggestions = append(r.Suggestions, "Verify that base image is provided by Red Hat.")
		r.Suggestions = append(r.Suggestions, "Make sure base layers weren't changed.")
	}
	return wrapResult(r)
}

// 2. Container image to be distributed through non-Red Hat registries does not include Red Hat Enterprise
// Linux (RHEL) kernel packages. (oscap-podman with xccdf profile)
// Test name: ubi_content_ok
// Why? To ensure, for a UBI type project, RPM packages present in a container image are only from UBI and RHEL
// user space. Red Hat allows redistribution of UBI content as per UBI EULA. Red Hat allows redistribution of
// RHEL user space packages as per Red Hat Container Certification Appendix. Presence of any kernel package will
// cause the test to fail.f
// How? Confirm all Red Hat RPMs included in the container image are from UBI and RHEL user space.

// 3. Container image must include the following metadata: (buildah inspect)

// name: Name of the image
// vendor: Company name
// version: Version of the image
// release: A number used to identify the specific build for this image
// summary: A short overview of the application or component in this image
// description: A long description of the application or component in this image

// Test names: name_label_exists, vendor_label_exists, version_label_exists, release_label_exists,
// summary_label_exists, description_label_exists
// Why? Providing metadata in consistent format helps customers inspect and manage images
// How? Define these as LABELs in your dockerfile.

func HasLabels(img string) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = "Has Labels"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	cmd := exec.Command("podman", "pull", img)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	cmd = exec.Command("podman", "inspect", img)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	labels := []string{"name", "vendor", "version", "release", "summary", "description"}
	for _, label := range labels {
		op, _ := jq.Parse(".[0].Config.Labels." + label)
		value, _ := op.Apply(cmdOutput.Bytes())
		if string(value) == "" {
			r.Errors = append(r.Errors, fmt.Sprintf("Label %s not present.", label))
			r.State = scapiv1alpha3.FailState
			r.Suggestions = append(r.Suggestions, fmt.Sprintf("Please include label %s", label))
		}
		// fmt.Println(string(value))
	}

	return wrapResult(r)
}

// 4. Container image cannot modify content provided by Red Hat packages or layers, except for files that are
// meant to be modified by end users, such as configuration files (oscap-podman xccdf profile)
// Test names: rpm_verify_successful, rpm_list_successful
// Why? Unauthorized changes to Red Hat components would impact or invalidate their support
// How? Don’t modify content in the base image or in Red Hat packages

// 5. Red Hat components in the container image cannot contain any critical or important vulnerabilities, as
// defined at https://access.redhat.com/security/updates/classification (oscap-podman oval eval)
// Test name: free_of_critical_vulnerabilities
// Why? These vulnerabilities introduce risk to your customers
// How? The certification report will indicate if such vulnerabilities are present in your image. It is recommended to
// use the most recent version of a layer or package, and to update your image content using the following
// command:
// yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

// ***** 6. should not modify, replace or combine the Red Hat base layer(s) (buildah inspect ???)
// Test name: good_layer_count
// Why? So the base layer provided by Red Hat can still be identified and inspected.
// How? Typically not an issue. Do not use any tools that attempt to or actually modify, replace, combine (aka
// squash) or otherwise obfuscate layers after the image has been built.

// ****** 7. The uncompressed container image should have less than 40 layers. (podman history -q <imagenameorid> | wc -l)
// Test name: has_under_40_layers
// Why? To ensure that an uncompressed container image has less than 40 layers. Too many layers within a
// container image can degrade container performance. Red Hat atomic errors out trying to mount an image with
// more than 40 layers.
// https://access.redhat.com/security/updates/classification
// How? Confirm that an uncompressed container image has less than 40 layers. You can leverage following
// commands to display layers and their size within a container image:
// podman history <container image name> or docker history <container image name> .

func has_under_40_layers(o *opcert.OpCert) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = "Has Under 40 Layers"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	layers, err := o.GetImageLayers(o.Image)
	if err != nil {
		fmt.Printf("Couldn't get image layers %v", err)
	}
	if len(layers) >= 40 {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "Image has 40 or more layers.")
		r.Suggestions = append(r.Suggestions, "Reduce the number of layers by optimizing the container file.")
	}
	return wrapResult(r)
}

// ****** 8. Image should include a tag, other than latest
// Test name: good_tags
// Why? So the image can be uniquely identified
// How? Use the docker tag command to add a tag. A common tag is the image version. The latest tag will be
// automatically added to the most recent image, so it should not be set explicitly.

// ****** 9. Image must include Partner’s software terms and conditions
// Test name: has_licenses
// Why? So the end user is aware of the terms and conditions applicable to the software. Including opens source
// licensing information, if open source components are included in the image.
// How? Create a directory named /licenses and include all relevant licensing and/or terms and conditions as text
// file(s) in that directory.

// Recommendation tests that may require manual approval:

// Test name: not_running_as_root
// Why? Running a container as root could create a security risk, as any process that breaks out of the container will
// retain the same privileges on the host machine.
// How? Indicate a specific USER in the dockerfile
// A container that does not specify a non-root user will fail the automatic certification, and will be subject to a
// manual review before the container can be approved for publication.

// 11. Do not request host-level privileges
// Test name: not_running_privileged
// Why? A container that requires special host-level privileges to function may not be suitable in environments
// where the application deployer does not have full control over the host infrastructure
// A container that requires special privileges will fail the automatic certification, and will be subject to a manual
// review before the container can be approved for publication.

func wrapResult(r scapiv1alpha3.TestResult) scapiv1alpha3.TestStatus {
	return scapiv1alpha3.TestStatus{
		Results: []scapiv1alpha3.TestResult{r},
	}
}
