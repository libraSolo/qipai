package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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

func ParseValue(value []byte) (Server, error) {
	var server Server
	if err := json.Unmarshal(value, &server); err != nil {
		return server, err
	}
	return server, nil
}

func ParseKey(key string) (Server, error) {
	// user/v1/127.0.0.1:8080
	// user/127.0.0.1:8080
	split := strings.Split(key, "/")
	if len(split) == 2 {
		return Server{
			Name: split[0],
			Addr: split[1],
		}, nil
	} else if len(split) == 3 {
		return Server{
			Name:    split[0],
			Addr:    split[2],
			Version: split[1],
		}, nil
	}
	return Server{}, errors.New("invalid key")
}
