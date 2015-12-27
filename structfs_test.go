package structfs

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

type DigitalOceanMeta struct {
	DropletID  int64    `json:"droplet_id"`
	Hostname   string   `json:"hostname"`
	VendorData string   `json:"vendor_data"`
	PublicKeys []string `json:"public_keys"`
	Region     string   `json:"region"`
	Interfaces struct {
		Private []struct {
			IPv4 struct {
				Address string `json:"ip_address"`
				Netmask string `json:"netmask"`
				Gateway string `json:"gateway"`
			}
			Mac  string `json:"mac"`
			Type string `json:"type"`
		} `json:"private"`
		Public []struct {
			IPv4 struct {
				Address string `json:"ip_address"`
				Netmask string `json:"netmask"`
				Gateway string `json:"gateway"`
			} `json:"ipv4"`
			IPv6 struct {
				Address string `json:"ip_address"`
				CIDR    int    `json:"cidr"`
				Gateway string `json:"gateway"`
			} `json:"ipv6"`
			Mac  string `json:"mac"`
			Type string `json:"type"`
		} `json:"public"`
	} `json:"interfaces"`
	FloatingIP struct {
		IPv4 struct {
			Active bool `json:"active"`
		} `json:"ipv4"`
	} `json:"floating_ip"`
	DNS struct {
		Nameservers []string `json:"nameservers"`
	} `json:"dns"`
}

var js = []byte(`{
  "droplet_id":2756294,
  "hostname":"sample-droplet",
  "vendor_data":"#cloud-config\ndisable_root: false\nmanage_etc_hosts: true\n\ncloud_config_modules:\n - ssh\n - set_hostname\n - [ update_etc_hosts, once-per-instance ]\n\ncloud_final_modules:\n - scripts-vendor\n - scripts-per-once\n - scripts-per-boot\n - scripts-per-instance\n - scripts-user\n",
  "public_keys":["ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCcbi6cygCUmuNlB0KqzBpHXf7CFYb3VE4pDOf/RLJ8OFDjOM+fjF83a24QktSVIpQnHYpJJT2pQMBxD+ZmnhTbKv+OjwHSHwAfkBullAojgZKzz+oN35P4Ea4J78AvMrHw0zp5MknS+WKEDCA2c6iDRCq6/hZ13Mn64f6c372JK99X29lj/B4VQpKCQyG8PUSTFkb5DXTETGbzuiVft+vM6SF+0XZH9J6dQ7b4yD3sOder+M0Q7I7CJD4VpdVD/JFa2ycOS4A4dZhjKXzabLQXdkWHvYGgNPGA5lI73TcLUAueUYqdq3RrDRfaQ5Z0PEw0mDllCzhk5dQpkmmqNi0F sammy@digitalocean.com"],
  "region":"nyc3",
  "interfaces":{
    "private":[
      {
        "ipv4":{
          "ip_address":"10.132.255.113",
          "netmask":"255.255.0.0",
          "gateway":"10.132.0.1"
        },
        "mac":"04:01:2a:0f:2a:02",
        "type":"private"
      }
    ],
    "public":[
      {
        "ipv4":{
          "ip_address":"104.131.20.105",
          "netmask":"255.255.192.0",
          "gateway":"104.131.0.1"
        },
        "ipv6":{
          "ip_address":"2604:A880:0800:0010:0000:0000:017D:2001",
          "cidr":64,
          "gateway":"2604:A880:0800:0010:0000:0000:0000:0001"
        },
        "mac":"04:01:2a:0f:2a:01",
        "type":"public"}
    ]
  },
  "floating_ip": {
    "ipv4": {
      "active": false
    }
  },
  "dns":{
    "nameservers":[
      "2001:4860:4860::8844",
      "2001:4860:4860::8888",
      "8.8.8.8"
    ]
  }
}
`)

func server() {
	stfs := DigitalOceanMeta{}
	json.Unmarshal(js, &stfs)
	http.Handle("/metadata/v1/", http.StripPrefix("/metadata/v1/", FileServer(&stfs, "json", time.Now())))
	http.Handle("/", &stfs)
	go func() {
		panic(http.ListenAndServe(":8080", nil))
	}()
	time.Sleep(2 * time.Second)
}

func (stfs *DigitalOceanMeta) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fs := FileServer(stfs, "json", time.Now())
	idx := strings.Index(r.URL.Path[1:], "/")
	r.URL.Path = strings.Replace(r.URL.Path[idx+1:], "/metadata/v1/", "", 1)
	r.RequestURI = r.URL.Path
	fs.ServeHTTP(w, r)
}

func get(path string) ([]byte, error) {
	res, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

var tests = []struct {
	in  string
	out string
}{
	{"http://127.0.0.1:8080/metadata/v1/", "droplet_id\nhostname\nvendor_data\npublic_keys\nregion\ninterfaces\nfloating_ip\ndns"},
	{"http://127.0.0.1:8080/metadata/v1/droplet_id", "2756294"},
	{"http://127.0.0.1:8080/metadata/v1/dns/", "nameservers"},
	{"http://127.0.0.1:8080/metadata/v1/dns/nameservers", "2001:4860:4860::8844\n2001:4860:4860::8888\n8.8.8.8"},
	{"http://127.0.0.1:8080/127.0.0.1/metadata/v1/dns/nameservers", "2001:4860:4860::8844\n2001:4860:4860::8888\n8.8.8.8"},
}

func TestAll(t *testing.T) {
	server()

	for _, tt := range tests {
		buf, _ := get(tt.in)
		if string(buf) != tt.out {
			t.Errorf("%s get %s want %s", tt.in, string(buf), tt.out)
		}
	}

}
