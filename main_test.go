package main_test

import (
	"io/ioutil"
	"net/http"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sample HTTP Server", func() {
	var (
		err  error
		resp *http.Response
	)

	Context("when the endpoint is called", func() {
		It("receives an HTTP OK response", func() {
			resp, err = http.Get("http://0.0.0.0:8080")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("when WAIT_TIME is set", func() {
		var (
			testWaitTime time.Duration
		)

		BeforeEach(func() {
			testWaitTimeString = "2s"
			testWaitTime, err = time.ParseDuration(testWaitTimeString)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should receive an HTTP OK after the WAIT_TIME", func() {
			startTime := time.Now()
			resp, err = http.Get("http://0.0.0.0:8080")
			endTime := time.Now()
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(endTime.Sub(startTime)).To(BeNumerically(">=", testWaitTime))
		})
	})

	Context("with an inflight request", func() {
		var (
			doneChan chan string
		)

		JustBeforeEach(func() {
			doneChan = make(chan string)
			go func() {
				resp, err = http.Get("http://0.0.0.0:8080")
				close(doneChan)
			}()
		})

		Context("when SIGTERM is sent to the server", func() {
			JustBeforeEach(func() {
				time.Sleep(100 * time.Millisecond)
				process.Signal(syscall.SIGTERM)
			})

			It("it finishes the inflight request", func() {
				Eventually(doneChan, 2).Should(BeClosed())
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(readResponseBody(resp)).To(Equal("hello world!\n"))
			})

			It("does not accept new requests", func() {
				time.Sleep(200 * time.Millisecond)
				_, err := http.Get("http://0.0.0.0:8080")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func readResponseBody(resp *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	return string(bodyBytes)
}
