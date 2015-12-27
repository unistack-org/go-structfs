package metadata

type DigitalOceanMetadata struct {
	Metadata struct {
		V1 struct {
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
		} `json:"v1"`
	} `json:"metadata"`
}
