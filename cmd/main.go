package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opdev/opcert/pkg/opcert"

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

	err := opcert.Init(builder, img)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Names of the custom tests which would be passed in the
	// `operator-sdk` command.
	switch entrypoint[0] {

	case IsImageRedHatProvidedTest:
		result = IsImageRedHat(&opcert)

	case HasLabelsTest:
		result = HasLabels(&opcert)

	case HasUnder40LayersTest:
		result = HasUnder40Layers(&opcert)

	case HasGoodTagsTest:
		result = HasGoodTags(&opcert)

	case HasLicensesTest:
		result = HasLicenses(&opcert)

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

const (
	IsImageRedHatProvidedTest = "is_red_hat"
	HasLabelsTest             = "has_labels"
	HasUnder40LayersTest      = "has_under_40_Layers"
	HasGoodTagsTest           = "has_good_tags"
	HasLicensesTest           = "has_licenses"
)

// printValidTests will print out full list of test names to give a hint to the end user on what the valid tests are.
func printValidTests() scapiv1alpha3.TestStatus {
	result := scapiv1alpha3.TestResult{}
	result.State = scapiv1alpha3.FailState
	result.Errors = make([]string, 0)
	result.Suggestions = make([]string, 0)

	str := fmt.Sprintf("Valid tests for this image include: %s %s",
		IsImageRedHatProvidedTest,
		HasLabelsTest)
	result.Errors = append(result.Errors, str)
	return scapiv1alpha3.TestStatus{
		Results: []scapiv1alpha3.TestResult{result},
	}
}

// Mandatory tests with possible fail results:

// ******************************* Image Metadata Tests ***************************************

// IsImageRedHatProvided:

// a. Container must be use a base image provided by Red Hat.
// Why? So the application's runtime dependencies, such as operating system components and libraries, are fully
// supported.
// How? Go to the Red Hat Container Catalog and select a base image to build upon. Use this image's name in
// the FROM clause in your dockerfile. We recommend using one of the images that are part of the Red Hat
// Universal Base Image (UBI) set, such as ubi7/ubi or ubi7/ubi-minimal.

// b. should not modify, replace or combine the Red Hat base layer(s)
// Why? So the base layer provided by Red Hat can still be identified and inspected.
// How? Typically not an issue. Do not use any tools that attempt to or actually modify, replace, combine (aka
// squash) or otherwise obfuscate layers after the image has been built.

func IsImageRedHat(o *opcert.OpCert) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = "Is Base Image Red Hat"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	// Check if base image is Red Hat
	if o.BaseImage == "" {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "Base image not present in Red Hat's catalog.")
		r.Suggestions = append(r.Suggestions, "Verify that base image comes from Red Hat's catalog.")
		r.Suggestions = append(r.Suggestions, "If it's an actual Red Hat based image, verify if the ContainerConfig was changed.")
		r.Suggestions = append(r.Suggestions, "The original Red Hat labels must be preserved in order to verify its origin.")
		r.Suggestions = append(r.Suggestions, "Use the Config field to insert new labels instead.")
		r.Suggestions = append(r.Suggestions, "In case you need any assistance please contact sd-ecosystem@redhat.com")
		return wrapResult(r)
	}
	//  pull base image from catalog
	err := o.PullImage(o.BaseImage)
	if err != nil {
		fmt.Printf("couldn't get image %v from registry.access.redhat.com. %v", o.BaseImage, err)
	}

	// get Red Hat base image layer sha256 digests from manifest
	rhImgLayers, err := o.GetImageLayers(o.BaseImage)
	if err != nil {
		fmt.Printf("couldn't get image layers %v from registry.access.redhat.com. %v", o.BaseImage, err)
	}

	o.BaseImageLayers = rhImgLayers

	// compare partner base image layers with Red Hat's base image layers
	for i, baseImageLayer := range o.BaseImageLayers {
		if o.LayerDigests[i] != baseImageLayer {
			r.State = scapiv1alpha3.FailState
			r.Errors = append(r.Errors, "Base image layer "+o.LayerDigests[i]+" doesn't match Red Hat's layer "+baseImageLayer)
		}
	}

	if r.State == scapiv1alpha3.FailState {
		r.Suggestions = append(r.Suggestions, "Make sure base layers weren't changed.")
	}
	return wrapResult(r)
}

// HasLabels:

// Container image must include the following metadata:
// name: Name of the image
// vendor: Company name
// version: Version of the image
// release: A number used to identify the specific build for this image
// summary: A short overview of the application or component in this image
// description: A long description of the application or component in this image
// Why? Providing metadata in consistent format helps customers inspect and manage images
// How? Define these as LABELs in your dockerfile.

func HasLabels(o *opcert.OpCert) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = "Has Labels"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	requiredLabels := []string{"name", "vendor", "version", "release", "summary", "description"}

	var isTestLabelPresent bool

	// For each of the required label check the label map
	// and report the missing labels
	// NOTE: the labels come from the Config field on the manifest
	// TODO: future work can change this for any label present on the manifest

	for _, requiredLabel := range requiredLabels {
		isTestLabelPresent = false
		for imgLabel := range o.Labels {
			if requiredLabel == imgLabel && requiredLabel != "" {
				isTestLabelPresent = true
			}
		}
		if isTestLabelPresent == false {
			r.Errors = append(r.Errors, fmt.Sprintf("Label %s not present.", requiredLabel))
			r.State = scapiv1alpha3.FailState
			r.Suggestions = append(r.Suggestions, fmt.Sprintf("Please include label %s in the container manifest's Config section.", requiredLabel))
		}
	}
	if r.State == scapiv1alpha3.FailState {
		r.Suggestions = append(r.Suggestions, "In case you need any assistance please contact sd-ecosystem@redhat.com")
	}
	return wrapResult(r)
}

// HasUnder40Layers:

// The uncompressed container image should have less than 40 layers.
// Why? To ensure that an uncompressed container image has less than 40 layers. Too many layers within a
// container image can degrade container performance. Red Hat atomic errors out trying to mount an image with
// more than 40 layers.
// How? Confirm that an uncompressed container image has less than 40 layers. You can leverage following
// commands to display layers and their size within a container image:
// podman history <container image name> or docker history <container image name> .

func HasUnder40Layers(o *opcert.OpCert) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = "Has Under 40 Layers"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	qtLayer := len(o.LayerDigests)
	fmt.Sprintln(qtLayer)
	if len(o.LayerDigests) >= 40 {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "Image has 40 or more layers.")
		r.Suggestions = append(r.Suggestions, "Reduce the number of layers by optimizing the container file.")
		r.Suggestions = append(r.Suggestions, "In case you need any assistance please contact sd-ecosystem@redhat.com")
	}
	return wrapResult(r)
}

// HasGoodTags:

// Image should include a tag, other than latest
// Why? So the image can be uniquely identified
// How? Use the docker tag command to add a tag. A common tag is the image version. The latest tag will be
// automatically added to the most recent image, so it should not be set explicitly.

func HasGoodTags(o *opcert.OpCert) scapiv1alpha3.TestStatus {

	r := scapiv1alpha3.TestResult{}
	r.Name = "Has Good Tags"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	goodTag := false

	for _, tag := range o.Tags {
		fmt.Printf("TAG: %v", tag[strings.Index(tag, ":")+1:])
		if tag[strings.Index(tag, ":")+1:] != "latest" {
			goodTag = true
			break
		}
	}
	if goodTag == false {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, fmt.Sprintf("There are no tags with the SemVer format for the image %v", o.Image))
		r.Suggestions = append(r.Suggestions, fmt.Sprintf("Please add new tags to the image %v in the SemvVer format.", o.Image))
	}
	return wrapResult(r)
}

// ****** 9. Image must include Partner’s software terms and conditions
// Test name: has_licenses
// Why? So the end user is aware of the terms and conditions applicable to the software. Including open source
// licensing information, if open source components are included in the image.
// How? Create a directory named /licenses and include all relevant licensing and/or terms and conditions as text
// file(s) in that directory.

func HasLicenses(o *opcert.OpCert) scapiv1alpha3.TestStatus {

	r := scapiv1alpha3.TestResult{}
	r.Name = "Has Licenses"
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)

	if o.HasLicenses == false {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "There is no /licenses folder and license information in the container image  %v", o.Image)
		r.Suggestions = append(r.Suggestions, "Please add the licenses folder and license files to the image %v.", o.Image)
	}
	return wrapResult(r)
}

// *************************************  Security Tests ************************************************

// 2. Container image to be distributed through non-Red Hat registries does not include Red Hat Enterprise
// Linux (RHEL) kernel packages. (oscap-podman with xccdf profile)
// Test name: ubi_content_ok
// Why? To ensure, for a UBI type project, RPM packages present in a container image are only from UBI and RHEL
// user space. Red Hat allows redistribution of UBI content as per UBI EULA. Red Hat allows redistribution of
// RHEL user space packages as per Red Hat Container Certification Appendix. Presence of any kernel package will
// cause the test to fail.f
// How? Confirm all Red Hat RPMs included in the container image are from UBI and RHEL user space.

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
