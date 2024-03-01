package discovery

import "fmt"

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Weight  int    `json:"weight"`
	Version string `json:"version"`
	Ttl     int64  `json:"ttl"`
}

func (s Server) BuildRegisterKey() string {
	// name
	if 0 == len(s.Version) {
		return fmt.Sprint("/%s/%s", s.Name, s.Addr)
	}
	// name/Version
	return fmt.Sprint("/%s/%s/%s", s.Name, s.Version, s.Addr)
}
