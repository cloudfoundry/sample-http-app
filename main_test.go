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
			resp, err = http.Get(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("when wait query param is set", func() {
		It("should receive an HTTP OK after the WAIT_TIME", func() {
			startTime := time.Now()
			resp, err = http.Get(url + "?wait=2s")
			endTime := time.Now()
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(endTime.Sub(startTime)).To(BeNumerically(">=", 2*time.Second))
		})
	})

	Context("with an inflight request", func() {
		var (
			doneChan chan string
		)

		JustBeforeEach(func() {
			doneChan = make(chan string)
			go func() {
				resp, err = http.Get(url + "?wait=2s")
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
				_, err := http.Get(url)
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
