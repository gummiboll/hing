package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"regexp"
	"time"
)

// RequestTarget represents a target
type RequestTarget struct {
	URL       string
	Method    string
	Timeout   int
	SleepReq  int
	Stop      bool
	UserAgent string
	Headers   []RequestHeader
}

// RequestHeader contains headers for the target
type RequestHeader struct {
	Name  string
	Value string
}

// RequestResult represents a result from the test
type RequestResult struct {
	Seq                    int
	Time                   float64
	StatusCode             int
	Error                  string
	StartTime              time.Time
	dnsStart               time.Time
	connStart              time.Time
	reqStart               time.Time
	writeDone              time.Time
	resStart               time.Time
	DNSLookupTime          float64
	ConnectionTime         float64
	WriteRequestTime       float64
	WrittenToFirstByteTime float64
	ReadResponseTime       float64
}

// NewRequestTarget returns a RequestTarget with some defaults
func NewRequestTarget() *RequestTarget {
	return &RequestTarget{Method: "GET", Timeout: 2, SleepReq: 5, UserAgent: "hing (https://github.com/gummiboll/hing)"}
}

// Client returns a client with timeout
func (rt RequestTarget) Client() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}
	return &http.Client{Transport: tr, Timeout: time.Duration(time.Duration(rt.Timeout) * time.Second)}
}

// Req returns a request with headers
func (rt RequestTarget) Req() (*http.Request, error) {
	req, err := http.NewRequest(rt.Method, rt.URL, nil)
	if err != nil {
		return &http.Request{}, err
	}

	req.Header.Set("User-Agent", rt.UserAgent)

	// Set headers
	for _, header := range rt.Headers {
		req.Header.Set(header.Name, header.Value)
	}

	return req, nil
}

// Finalize finalizes a RequestResult
func (rr *RequestResult) Finalize(sCode int, err error) {
	rr.Time = time.Now().Sub(rr.StartTime).Seconds()
	// Better handling for timeouts?
	if sCode != 0 {
		rr.ReadResponseTime = time.Now().Sub(rr.resStart).Seconds()
	} else {
		rr.ReadResponseTime = 0
	}

	if err != nil {
		r := regexp.MustCompile(`\((.+)?\)`)
		res := r.FindStringSubmatch(err.Error())
		rr.Error = res[1]
	}

	rr.StatusCode = sCode
}

// DoRequest sends the actual request
func DoRequest(resChan chan RequestResult, rt RequestTarget, seq int) {
	rr := RequestResult{Seq: seq, StartTime: time.Now()}

	client := rt.Client()
	req, err := rt.Req()
	if err != nil {
		panic(err)
	}

	// Setup trace
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			rr.dnsStart = time.Now()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			rr.DNSLookupTime = time.Now().Sub(rr.dnsStart).Seconds()
		},
		GetConn: func(h string) {
			rr.connStart = time.Now()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			rr.ConnectionTime = time.Now().Sub(rr.connStart).Seconds()
			rr.reqStart = time.Now()
		},
		WroteRequest: func(w httptrace.WroteRequestInfo) {
			rr.WriteRequestTime = time.Now().Sub(rr.reqStart).Seconds()
			rr.writeDone = time.Now()
		},
		GotFirstResponseByte: func() {
			rr.WrittenToFirstByteTime = time.Now().Sub(rr.writeDone).Seconds()
			rr.resStart = time.Now()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := client.Do(req)
	if err != nil {
		rr.Finalize(0, err)
		resChan <- rr
		return
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		rr.Finalize(resp.StatusCode, err)
		resChan <- rr
		return
	}

	rr.Finalize(resp.StatusCode, nil)
	resChan <- rr
}
