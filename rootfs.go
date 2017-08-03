package container

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/blang/semver"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v1"
)

var (
	latestReleaseURL = "http://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/%s/latest-releases.yaml"
	latestRootfsURL  = "http://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/%s/%s"
	rootfsFlavor     = "alpine-minirootfs"
	rootfsArch       = ""
)

func init() {
	switch runtime.GOARCH {
	case "amd64":
		rootfsArch = "x86_64"
	}

	latestReleaseURL = fmt.Sprintf(latestReleaseURL, rootfsArch)
}

func latestRootfs() (map[interface{}]interface{}, error) {
	res, err := http.Get(latestReleaseURL)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var releases interface{}

	err = yaml.Unmarshal(body, &releases)
	if err != nil {
		return nil, err
	}

	for _, r := range releases.([]interface{}) {
		rootfs := r.(map[interface{}]interface{})

		if rootfs["flavor"] == rootfsFlavor {
			return rootfs, nil
		}
	}

	return nil, fmt.Errorf("%s not found", rootfsFlavor)
}

func currentRootfs() map[interface{}]interface{} {
	data, err := ioutil.ReadFile("rootfs/release.yaml")
	if err != nil {
		logrus.Warn(errors.Wrapf(err, "failed to read release file"))
		return nil
	}

	var rootfs map[interface{}]interface{}

	err = yaml.Unmarshal(data, &rootfs)
	if err != nil {
		logrus.Warn(errors.Wrapf(err, "invalid release file"))
		return nil
	}

	return rootfs
}

func downloadRootfs(rootfs map[interface{}]interface{}) (string, error) {
	res, err := http.Get(fmt.Sprintf(latestRootfsURL, rootfsArch, rootfs["file"].(string)))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("/tmp/%s", rootfs["file"].(string))

	err = ioutil.WriteFile(filename, body, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func unpackRootfs(filename string) error {
	return archiver.TarGz.Open(filename, "rootfs")
}

func pullRootfs() error {
	var err error

	crootfs := currentRootfs()

	if err = os.MkdirAll("rootfs/", 0744); err != nil {
		return err
	}

	lrootfs, err := latestRootfs()
	if err != nil {
		return err
	}

	writeRelease := false

	if crootfs == nil {
		writeRelease = true
	} else {
		v1, _ := semver.Make(crootfs["version"].(string))
		v2, _ := semver.Make(lrootfs["version"].(string))

		if v2.GT(v1) {
			writeRelease = true
		}
	}

	if writeRelease {
		filename, err := downloadRootfs(lrootfs)
		if err != nil {
			return err
		}

		if err = unpackRootfs(filename); err != nil {
			return err
		}

		data, err := yaml.Marshal(lrootfs)
		if err != nil {
			return err
		}

		if err = ioutil.WriteFile("rootfs/release.yaml", data, 0644); err != nil {
			return err
		}
	}

	return nil
}
