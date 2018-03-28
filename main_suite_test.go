package main_test

import (
	"os/exec"
	"strconv"
	"testing"

	"code.cloudfoundry.org/localip"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var (
	serverBinPath string
	process       ifrit.Process
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

var (
	url string
)

var _ = JustBeforeEach(func() {
	runCmd := exec.Command(serverBinPath)
	port, err := localip.LocalPort()
	Expect(err).NotTo(HaveOccurred())
	runner := ginkgomon.New(ginkgomon.Config{
		Name:       "sample-http-serveer",
		Command:    runCmd,
		StartCheck: "Serving on port",
	})
	portStr := strconv.Itoa(int(port))
	runner.Command.Env = append(runner.Command.Env, "PORT="+portStr)
	process = ginkgomon.Invoke(runner)
	url = "http://localhost:" + portStr
})

var _ = AfterEach(func() {
	ginkgomon.Kill(process)
})
