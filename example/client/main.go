package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/internal/testdata"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
)

var wait sync.WaitGroup

type Param struct {
	Verbose                         bool
	Quiet                           bool
	Insecure                        bool
	EnableQlog                      bool
	ConnFinishThenSendInitial       bool
	OnlySendInitial                 bool
	ConnFinishThenSendInitialPktNum int
	MaxIdleTimeout                  time.Duration
	ForMaxIdleTimeoutTest           bool
}

func oneTest(pool *x509.CertPool, param Param, keyLogFile *string, urls []string) {
	defer wait.Done()
	logger := utils.DefaultLogger

	if param.Verbose {
		logger.SetLogLevel(utils.LogLevelDebug)
	} else {
		logger.SetLogLevel(utils.LogLevelError)
	}
	logger.SetLogTimeFormat(time.Stamp)

	var keyLog io.Writer
	if len(*keyLogFile) > 0 {
		f, err := os.Create(*keyLogFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		keyLog = f
	}

	var qconf quic.Config
	qconf.OnlySendInitial = param.OnlySendInitial
	qconf.ConnFinishThenSendInitial = param.ConnFinishThenSendInitial
	qconf.ConnFinishThenSendInitialPktNum = param.ConnFinishThenSendInitialPktNum
	qconf.MaxIdleTimeout = param.MaxIdleTimeout
	qconf.ForMaxIdleTimeoutTest = param.ForMaxIdleTimeoutTest
	// qconf.HandshakeIdleTimeout = 2 * time.Second
	if param.EnableQlog {
		qconf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			filename := fmt.Sprintf("client_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
		})
	}
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: param.Insecure,
			KeyLogWriter:       keyLog,
		},
		QuicConfig: &qconf,
	}
	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
		Timeout:   1 * time.Second,
	}
	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, addr := range urls {
		logger.Infof("GET %s", addr)
		go func(addr string) {
			rsp, err := hclient.Get(addr)
			if err != nil {
				wait.Done()
				return
				// log.Fatal(err)
			}
			logger.Infof("Got response for %s: %#v", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				log.Fatal(err)
			}
			if param.Quiet {
				fmt.Printf("Response Body: %d bytes", body.Len())
			} else {
				fmt.Printf("Response Body:")
				fmt.Printf("%s\n", body.Bytes())
			}
			wg.Done()
		}(addr)
	}
	wg.Wait()

	// defer wait.Done()
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	quiet := flag.Bool("q", false, "don't print the data")
	keyLogFile := flag.String("keylog", "", "key log file")
	insecure := flag.Bool("insecure", false, "skip certificate verification")
	enableQlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
	onlySendInitial := flag.Bool("onlySendInitial", false, "only send init packet for test")
	connFinishThenSendInitial := flag.Bool("connFinishThenSendInitial", false, "when connection build finished then send init packet for test")
	connFinishThenSendInitialPktNum := flag.Int("connFinishThenSendInitialPktNum", 1, "send init packet num")
	repeatCnt := flag.Int("repeatCnt", 1, "repeat test count")
	maxIdleTimeout := flag.Duration("maxIdleTimeout", 0, "max_idle_timeout")
	flag.Parse()
	urls := flag.Args()

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}
	testdata.AddRootCA(pool)

	// fmt.Println("send initial packet begin")
	forMaxIdleTimeout := false
	if *maxIdleTimeout != time.Duration(0) {
		forMaxIdleTimeout = true
	}

	param := Param{
		Verbose:                         *verbose,
		Quiet:                           *quiet,
		Insecure:                        *insecure,
		EnableQlog:                      *enableQlog,
		OnlySendInitial:                 *onlySendInitial,
		ConnFinishThenSendInitial:       *connFinishThenSendInitial,
		ConnFinishThenSendInitialPktNum: *connFinishThenSendInitialPktNum,
		MaxIdleTimeout:                  *maxIdleTimeout,
		ForMaxIdleTimeoutTest:           forMaxIdleTimeout,
	}
	wait.Add(*repeatCnt)
	for i := 0; i < *repeatCnt; i++ {
		go oneTest(pool, param, keyLogFile, urls)
	}
	wait.Wait()
	// fmt.Printf("send %d initial packet finished!!!, Please CTRL+C to exit\n", *repeatCnt)
	// var tmp chan int
	// <-tmp
}
