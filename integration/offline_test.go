package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOffline(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when offline", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		it("pack builds and runs the app successfully", func() {
			var logs fmt.Stringer
			var err error
			image, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(offlineMRIBuildpack, offlineBundlerBuildpack, buildPlanBuildpack).
				WithNetwork("none").
				Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred(), logs.String())

			container, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable(), ContainerLogs(container.ID))

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort()))
			Expect(err).NotTo(HaveOccurred())

			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_bundler/bundler/bin/bundler"))
			Expect(string(content)).To(MatchRegexp(`Bundler version 2\.1\.\d+`))

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_mri/mri/bin/ruby"))
			Expect(string(content)).To(MatchRegexp(`ruby 2\.6\.\d+`))
		})
	})
}
