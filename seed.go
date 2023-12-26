package seed

import "net/http"

// mseed is an app driven by Router
type mseed struct {
	Router
	Ser *http.Server
}

// New return *mseed
func New() *mseed {
	return &mseed{
		Router: NewRouter(),
		Ser:    &http.Server{},
	}
}

// Run start http server
func (c *mseed) Run(addr string) error {
	c.Ser.Addr = addr
	c.Ser.Handler = c
	return c.Ser.ListenAndServe()
}

// RunTLS  start https server
func (c *mseed) RunTLS(addr string, certFile, keyFile string) error {
	c.Ser.Addr = addr
	c.Ser.Handler = c
	return c.Ser.ListenAndServeTLS(certFile, keyFile)
}
