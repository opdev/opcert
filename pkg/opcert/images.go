package opcert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/savaki/jq"
)

type OpCert struct {
	Version      string
	Builder      string
	Image        string
	LayerCount   int
	LayerDigests []string
	BaseImage    string
}

func (o *OpCert) Init(builder string, img string) error {
	o.Builder = builder

	if err := o.PullImage(img); err != nil {
		err := fmt.Errorf("opcert wasn't able to pull image from %v", img)
		return err
	}

	baseImage, err := o.GetBaseImage(img)
	if err != nil {
		err = fmt.Errorf("opcert couldn't read a proper base image tag from %v manifest", img)
		return err
	}
	o.BaseImage = baseImage

	o.LayerDigests, err = o.GetImageLayers(img)
	if err != nil {
		err = fmt.Errorf("opcert couldn't read image layers from %v", img)
		return err
	}

	return nil
}

func (o *OpCert) PullImage(img string) error {

	cmd := exec.Command("docker", "pull", img)

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Couldn't pull image %s %s\n", img, err)
		return err
	}
	return nil
}

func (o *OpCert) GetBaseImage(img string) (string, error) {

	cmd := exec.Command("docker", "inspect", img)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return "", err
	}

	op, _ := jq.Parse(".[0].ContainerConfig.Labels.name")
	name, _ := op.Apply(cmdOutput.Bytes())

	op, _ = jq.Parse(".[0].ContainerConfig.Labels.version")
	version, _ := op.Apply(cmdOutput.Bytes())

	op, _ = jq.Parse(".[0].ContainerConfig.Labels.release")
	release, _ := op.Apply(cmdOutput.Bytes())

	baseImage := "registry.access.redhat.com/" + strings.Trim(string(name), "\"") + ":" + strings.Trim(string(version), "\"") + "-" + strings.Trim(string(release), "\"")

	return baseImage, nil
}

func (o *OpCert) GetImageLayers(img string) ([]string, error) {

	cmd := exec.Command("docker", "inspect", img)

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
