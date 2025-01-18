package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RouterOSConnection struct {
	host     string
	user     string
	password string
}

func NewRouterOSConnection(host string, user string, password string) *RouterOSConnection {
	return &RouterOSConnection{
		host:     host,
		user:     user,
		password: password,
	}
}

type RegistrationTableEntry struct {
	ID              string `json:".id"`
	AuthType        string `json:"auth-type"`
	Authorized      string `json:"authorized"`
	Band            string `json:"band"`
	Bytes           string `json:"bytes"`
	Comment         string `json:"comment"`
	Interface       string `json:"interface"`
	MacAddress      string `json:"mac-address"`
	Packets         string `json:"packets"`
	RxBitsPerSecond string `json:"rx-bits-per-second"`
	RxRate          string `json:"rx-rate"`
	Signal          string `json:"signal"`
	SSID            string `json:"ssid"`
	TxBitsPerSecond string `json:"tx-bits-per-second"`
	TxRate          string `json:"tx-rate"`
	Uptime          string `json:"uptime"`
}

func (r *RouterOSConnection) GetRegistrationTable() ([]RegistrationTableEntry, error) {
	body, err := r.apiCall("/interface/wifi/registration-table")
	if err != nil {
		return nil, err
	}

	resp := []RegistrationTableEntry{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type DHCPLeaseEntry struct {
	ID               string `json:".id"`
	ActiveAddress    string `json:"active-address"`
	ActiveMacAddress string `json:"active-mac-address"`
	ActiveServer     string `json:"active-server"`
	Address          string `json:"address"`
	AddressLists     string `json:"address-lists"`
	HostName         string `json:"host-name,ignoreempty"`
	Age              string `json:"age"`
	Blocked          string `json:"blocked"`
	DHCPOption       string `json:"dhcp-option"`
	Disabled         string `json:"disabled"`
	Dynamic          string `json:"dynamic"`
	ExpiresAfter     string `json:"expires-after"`
	LastSeen         string `json:"last-seen"`
	MacAddress       string `json:"mac-address"`
	Radius           string `json:"radius"`
	Server           string `json:"server"`
	Status           string `json:"status"`
}

func (r *RouterOSConnection) GetDHCPLeases() ([]DHCPLeaseEntry, error) {
	body, err := r.apiCall("/ip/dhcp-server/lease?disabled=false&status=bound")
	if err != nil {
		return nil, err
	}

	resp := []DHCPLeaseEntry{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

type WifiRadioEntry struct {
	ID                   string `json:".id"`
	Channels5G           string `json:"5g-channels"`
	Channels2G           string `json:"2g-channels"`
	Bands                string `json:"bands"`
	Cap                  string `json:"cap"`
	Ciphers              string `json:"ciphers"`
	Countries            string `json:"countries"`
	CurrentChannels      string `json:"current-channels"`
	CurrentCountry       string `json:"current-country"`
	CurrentGOPClasses    string `json:"current-gopclasses"`
	CurrentMaxRegPower   string `json:"current-max-reg-power"`
	HWCaps               string `json:"hw-caps"`
	HWType               string `json:"hw-type"`
	Interface            string `json:"interface"`
	MaxInterfaces        string `json:"max-interfaces"`
	MaxPeers             string `json:"max-peers"`
	MaxStationInterfaces string `json:"max-station-interfaces"`
	MaxVlans             string `json:"max-vlans"`
	MinAntennaGain       string `json:"min-antenna-gain"`
	RadioMac             string `json:"radio-mac"`
	RxChains             string `json:"rx-chains"`
	TxChains             string `json:"tx-chains"`
}

func (w WifiRadioEntry) GetCapName() string {
	if strings.Contains(w.Cap, "@") {
		return strings.Split(w.Cap, "@")[0]
	}
	return w.Cap
}

func (r *RouterOSConnection) GetWifiRadios() ([]WifiRadioEntry, error) {
	body, err := r.apiCall("/interface/wifi/radio")
	if err != nil {
		return nil, err
	}

	resp := []WifiRadioEntry{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type WifiInterfaceEntry struct {
	ID                          string `json:".id"`
	ArpTimeout                  string `json:"arp-timeout"`
	Bound                       string `json:"bound"`
	ChannelBand                 string `json:"channel.band"`
	ChannelFrequency            string `json:"channel.frequency"`
	ChannelReselectInterval     string `json:"channel.reselect-interval"`
	ChannelSkipDFSChannels      string `json:"channel.skip-dfs-channels"`
	ChannelWidth                string `json:"channel.width"`
	Configuration               string `json:"configuration"`
	ConfigurationCountry        string `json:"configuration.country"`
	ConfigurationMode           string `json:"configuration.mode"`
	ConfigurationSSID           string `json:"configuration.ssid"`
	DatapathBridge              string `json:"datapath.bridge"`
	Disabled                    string `json:"disabled"`
	Inactive                    string `json:"inactive"`
	MacAddress                  string `json:"mac-address"`
	Master                      string `json:"master"`
	MasterInterface             string `json:"master-interface"`
	Name                        string `json:"name"`
	Running                     string `json:"running"`
	SecurityAuthenticationTypes string `json:"security.authentication-types"`
	SecurityEncryption          string `json:"security.encryption"`
	SecurityFT                  string `json:"security.ft"`
	SecurityFTOverDS            string `json:"security.ft-over-ds"`
	SecurityPassphrase          string `json:"security.passphrase"`
	SteeringRRM                 string `json:"steering.rrm"`
	SteeringWnm                 string `json:"steering.wnm"`
}

func (r *RouterOSConnection) GetWifiInterfaces() ([]WifiInterfaceEntry, error) {
	body, err := r.apiCall("/interface/wifi")
	if err != nil {
		return nil, err
	}

	resp := []WifiInterfaceEntry{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (r *RouterOSConnection) apiCall(path string) ([]byte, error) {
	// we need to perform a HTTP POST to host using user and password
	// to url

	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	url := fmt.Sprintf("http://%s/rest/%s", r.host, path)
	// log.Printf("executing api call %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(r.user, r.password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
