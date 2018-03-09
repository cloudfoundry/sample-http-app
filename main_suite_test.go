package main_test

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var (
	serverBinPath      string
	process            ifrit.Process
	testWaitTimeString string = "0s"
)

func TestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	serverPath, err := gexec.Build("code.cloudfoundry.org/sample-http-app/", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(serverPath)
}, func(pathsByte []byte) {
	serverBinPath = string(pathsByte)
})

var _ = JustBeforeEach(func() {
	runCmd := exec.Command(serverBinPath)
	runCmd.Env = []string{fmt.Sprintf("WAIT_TIME=%s", testWaitTimeString)}
	runner := ginkgomon.New(ginkgomon.Config{
		Name:       "sample-http-serveer",
		Command:    runCmd,
		StartCheck: "Serving on port",
	})
	process = ginkgomon.Invoke(runner)
})

var _ = AfterEach(func() {
	ginkgomon.Kill(process)
})
