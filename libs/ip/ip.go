package ip

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var CurrentIp *IP

type IP struct {
	IP     string
	Source SourceInterface
}

type SourceInterface interface {
	ParseResponse([]byte) (string, error)
	GetName() string
	GetURL() string
}

type Source struct {
	Name string
	URL  string
}

func (s *Source) GetName() string {
	return s.Name
}

func (s *Source) GetURL() string {
	return s.URL
}

type ApifyOrg struct {
	Source
}

func (s *ApifyOrg) ParseResponse(body []byte) (string, error) {
	var res struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	return res.IP, nil
}

func NewApifyOrg() *ApifyOrg {
	return &ApifyOrg{
		Source: Source{
			Name: "ApifyOrg",
			URL:  "https://api.ipify.org?format=json",
		},
	}
}

type IpApi struct {
	Source
}

func (s *IpApi) ParseResponse(body []byte) (string, error) {
	var res struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	return res.Query, nil
}

func NewIpApi() *IpApi {
	return &IpApi{
		Source: Source{
			Name: "IpApi",
			URL:  "http://ip-api.com/json/",
		},
	}
}

type IpinfoIo struct {
	Source
}

func (s *IpinfoIo) ParseResponse(body []byte) (string, error) {
	var res struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	return res.IP, nil
}

func NewIpinfoIo() *IpinfoIo {
	return &IpinfoIo{
		Source: Source{
			Name: "IpinfoIo",
			URL:  "https://ipinfo.io/json",
		},
	}
}

type IdentMe struct {
	Source
}

func (s *IdentMe) ParseResponse(body []byte) (string, error) {
	var res struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	return res.Address, nil
}

func NewIdentMe() *IdentMe {
	return &IdentMe{
		Source: Source{
			Name: "IdentMe",
			URL:  "https://ident.me/.json",
		},
	}
}

func GetIP(ctx context.Context) (*IP, error) {
	sources := []SourceInterface{
		NewApifyOrg(),
		NewIpApi(),
		NewIpinfoIo(),
		NewIdentMe(),
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	source := sources[r.Intn(len(sources))]

	req, err := http.NewRequestWithContext(ctx, "GET", source.GetURL(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ipStr, err := source.ParseResponse(body)
	if err != nil {
		return nil, err
	}

	return &IP{
		IP:     ipStr,
		Source: source,
	}, nil
}
