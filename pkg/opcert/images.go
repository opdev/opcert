package opcert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/savaki/jq"
)

type OpCert struct {
	Version         string
	Builder         string
	Image           string
	LayerCount      int
	LayerDigests    []string
	Tags            []string
	BaseImage       string
	BaseImageLayers []string
	HasLicenses     bool
	Labels          map[string]string
}

func (o *OpCert) Init(builder string, img string) error {
	o.Builder = builder

	if err := o.PullImage(img); err != nil {
		if err != nil {
			fmt.Printf("Error pulling image %s", img)
			fmt.Printf("Error: %v", err)
		}
		return err
	}

	baseImage, err := o.GetBaseImage(img)
	if err != nil {
		err = fmt.Errorf("opcert couldn't read a proper base image tag from %v manifest", img)
		return err
	}
	o.BaseImage = baseImage

	labels, err := o.GetLabels(img)
	if err != nil {
		err = fmt.Errorf("opcert couldn't read labels from %v manifest", img)
		return err
	}
	o.Labels = labels

	o.LayerDigests, err = o.GetImageLayers(img)
	if err != nil {
		err = fmt.Errorf("opcert couldn't read image layers from %v", img)
		return err
	}

	o.Tags, err = o.GetTags(img)
	if err != nil {
		err = fmt.Errorf("opcert couldn't read image tags from %v", img)
		return err
	}

	// o.HasLicenses, err = o.CheckLicenses(img)
	// if err != nil {
	// 	err = fmt.Errorf("opcert couldn't read directory structure from %v", img)
	// 	return err
	// }

	return nil
}

func (o *OpCert) PullImage(img string) error {

	cmd := exec.Command(o.Builder, "pull", img)

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		if strings.Contains(string(stderr.Bytes()), "Repo not found") {
			err = fmt.Errorf("Couldn't find repository for image %v", img)
			return err
		} else {
			err = fmt.Errorf("Unexpected error: %v", err)
			return err
		}
	}
	return nil
}

func (o *OpCert) GetLabels(img string) (map[string]string, error) {

	cmd := exec.Command(o.Builder, "inspect", img)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	op, _ := jq.Parse(".[0].Config.Labels")
	byteLabels, _ := op.Apply(cmdOutput.Bytes())

	labels := map[string]string{}
	err = json.Unmarshal(byteLabels, &labels)
	if err != nil {
		fmt.Println(err)
		return map[string]string{}, err
	}

	return labels, nil
}

func (o *OpCert) GetBaseImage(img string) (string, error) {

	cmd := exec.Command(o.Builder, "inspect", img)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return "", err
	}

	// determining the base image using the Label name in the ContainerConfig struct
	// this is a quick way if the container manifest was not changed
	// partner may put all labels on the Config struct for now
	// TODO: make the checks based only on 256 SHA hashes

	op, _ := jq.Parse(".[0].ContainerConfig.Labels.name")
	name, _ := op.Apply(cmdOutput.Bytes())

	op, _ = jq.Parse(".[0].ContainerConfig.Labels.version")
	version, _ := op.Apply(cmdOutput.Bytes())

	op, _ = jq.Parse(".[0].ContainerConfig.Labels.release")
	release, _ := op.Apply(cmdOutput.Bytes())

	var baseImage string

	// If the label name doesn't exist or is empty let the baseImage field empty
	// That will fail the IsRedHat test

	if string(name) != "" {
		baseImage = "registry.access.redhat.com/" + strings.Trim(string(name), "\"") + ":" + strings.Trim(string(version), "\"") + "-" + strings.Trim(string(release), "\"")
		o.BaseImage = baseImage
	} else {
		baseImage = ""
		o.BaseImage = ""
	}

	return baseImage, nil
}

func (o *OpCert) GetImageLayers(img string) ([]string, error) {

	cmd := exec.Command(o.Builder, "inspect", img)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return []string{}, err
	}

	op, _ := jq.Parse(".[0].RootFS.Layers")
	byteLayers, _ := op.Apply(cmdOutput.Bytes())

	imageLayers := []string{}
	err = json.Unmarshal(byteLayers, &imageLayers)
	if err != nil {
		fmt.Println(err)
	}

	return imageLayers, nil
}

func (o *OpCert) GetTags(img string) ([]string, error) {

	cmd := exec.Command("docker", "inspect", img)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error %v", stderr)
		return []string{}, err
	}

	op, _ := jq.Parse(".[0].RepoTags")
	byteTags, _ := op.Apply(cmdOutput.Bytes())

	tags := []string{}
	err = json.Unmarshal(byteTags, &tags)
	if err != nil {
		fmt.Println(err)
	}

	return tags, nil
}
func (o *OpCert) CheckLicenses(img string) (bool, error) {

	cmd := exec.Command(o.Builder, "create", img)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v", stderr)
		return false, err
	}

	op, _ := jq.Parse(".")
	containerID, _ := op.Apply(cmdOutput.Bytes())

	cmd = exec.Command("mkdir", "-p", "container_fs")

	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		if strings.Contains(err.Error(), "File exists") {

		}
		fmt.Printf("Error: %v", stderr)
		return false, err
	}

	cmd = exec.Command("sudo", o.Builder, "cp", strings.TrimSuffix(string(containerID), "\n")+":/", "./container_fs")
	cmd.Stderr = os.Stderr

	stdin := &bytes.Buffer{}
	cmd.Stdin = stdin

	err = cmd.Run()

	if err != nil {
		fmt.Printf("Error: %v", err)
		return false, err
	}

	files, err := ioutil.ReadDir("./container_fs")
	if err != nil {
		log.Fatal(err)
	}

	hasLicenses := false

	for _, f := range files {
		if f.IsDir() && f.Name() == "licenses" {
			hasLicenses = true
		}
	}

	cmd = exec.Command("rm", "-rf", "container_fs")

	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v", stderr)
		return false, err
	}

	return hasLicenses, nil
}
