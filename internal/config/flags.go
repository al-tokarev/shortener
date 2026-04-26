package config

import "flag"

var Options struct {
	AddrServe string
	AddrResp  string
}

func RunFlags() {
	flag.StringVar(&Options.AddrServe, "a", "localhost:8080", "Address server")
	flag.StringVar(&Options.AddrResp, "b", "localhost:8080", "Address response")

	flag.Parse()
}
