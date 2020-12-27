package checks

import (
	"context"
	"crypto/tls"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/frick/bouncer/pkg/conf"
	"github.com/frick/bouncer/pkg/gpio"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// map tls version ints to their names
var tlsVersions = map[uint16]string{
	0x0301: "TLS 1.0",
	0x0302: "TLS 1.1",
	0x0303: "TLS 1.2",
	0x0304: "TLS 1.3",
}

// CheckState holds the state for our state loop
type CheckState struct {
	NumFailures   int
	TotalFailures int
	LastFailure   time.Time
	NextSite      int // index of next site to use from Config.Sites slice
}

// transport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type transport struct {
	current *http.Request
}

// RoundTrip wraps http.DefaultTransport.RoundTrip to keep track
// of the current request.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.current = req
	return http.DefaultTransport.RoundTrip(req)
}

// handle DNS request completion
func (t *transport) DNSDone(info httptrace.DNSDoneInfo) {
	if info.Err != nil {
		log.WithFields(log.Fields{"addrs": info.Addrs, "err": info.Err}).Error("DNS lookup done, but with error")
	} else {
		log.WithFields(log.Fields{"addrs": info.Addrs}).Debug("DNS lookup complete")
	}
}

// handle connection completion
func (t *transport) ConnectDone(network, addr string, err error) {
	if err != nil {
		log.WithFields(log.Fields{"network": network, "addr": addr, "err": err}).Error("connection done, but with error")
	} else {
		log.WithFields(log.Fields{"network": network, "addr": addr}).Debug("connection complete")
	}
}

// handle TLS handshake completion
func (t *transport) TLSHandshakeDone(connState tls.ConnectionState, err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"version":            tlsVersions[connState.Version],
			"cipherSuite":        tls.CipherSuiteName(connState.CipherSuite),
			"negotiatedProtocol": connState.NegotiatedProtocol,
			"serverName":         connState.ServerName,
			"err":                err,
		}).Error("TLS handshake done, but with error")
	} else {
		log.WithFields(log.Fields{
			"version":            tlsVersions[connState.Version],
			"cipherSuite":        tls.CipherSuiteName(connState.CipherSuite),
			"negotiatedProtocol": connState.NegotiatedProtocol,
			"serverName":         connState.ServerName,
		}).Debug("TLS handshake complete")
	}
}

// jitter returns a random duration within +/- the given duration
func jitter(t time.Duration) time.Duration {
	return time.Duration(t.Seconds()*(rand.Float64()-0.5)*2) * time.Second
}

// performCheck actually sets up and performs a request to the provided site
func performCheck(site string, timeout time.Duration) error {
	log.WithFields(log.Fields{"site": site}).Debug("initiating check of website")

	// create a context with our configured check timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// create our request for the given site
	req, reqErr := http.NewRequestWithContext(ctx, "GET", site, nil)
	if reqErr != nil {
		return reqErr
	}

	// create our transport and httptrace with handlers
	t := &transport{}
	trace := &httptrace.ClientTrace{
		DNSDone:          t.DNSDone,
		ConnectDone:      t.ConnectDone,
		TLSHandshakeDone: t.TLSHandshakeDone,
	}

	// add httptrace to request context
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	req.Close = true

	// create a client with our custom transport and finally make our request
	// we don't care about the response so long as it does not result in error
	client := &http.Client{Transport: t}
	_, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}
	return nil
}

// CheckLoop begins a simple loop to check a site or sites for network connectivity,
// tracking failures, sleeping appropriate amounts of time between checks, etc
func CheckLoop(cfg *conf.Config, relay *gpio.GPIO) error {
	state := &CheckState{
		NumFailures:   0,
		TotalFailures: 0,
		LastFailure:   time.Time{},
		NextSite:      0,
	}

	var timeBeforeNextCheck time.Duration

	for {
		// check the given website
		checkStart := time.Now()
		checkErr := performCheck(cfg.Sites[state.NextSite], cfg.CheckTimeout)

		if checkErr == nil {
			log.WithFields(log.Fields{
				"duration":      time.Since(checkStart),
				"totalFailures": state.TotalFailures,
			}).Debug("website check completed successfully")
			state.NumFailures = 0
			timeBeforeNextCheck = cfg.CheckInterval + jitter(cfg.CheckJitter)
		} else {
			state.LastFailure = time.Now()
			state.NumFailures++
			state.TotalFailures++
			log.WithFields(log.Fields{
				"duration":            time.Since(checkStart),
				"err":                 checkErr,
				"totalFailures":       state.TotalFailures,
				"consecutiveFailures": state.NumFailures,
			}).Error("check failed")
			if state.NumFailures >= cfg.Failures {
				relay.Trigger(cfg.BounceDuration)
				timeBeforeNextCheck = cfg.BounceTimeout
				state.NumFailures = 0
			} else {
				timeBeforeNextCheck = cfg.RetryInterval + jitter(cfg.CheckJitter)
			}
		}

		state.NextSite++
		if state.NextSite >= len(cfg.Sites) {
			state.NextSite = 0
		}
		log.WithFields(log.Fields{"sleep": timeBeforeNextCheck}).Debug("sleeping until next check")
		time.Sleep(timeBeforeNextCheck)
	}
}
