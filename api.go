package main

import (
	"strings"
	"sync"
	"time"
)

const (
	RADIO_UPDATE_INTERVAL          = 30 * time.Minute
	LEASE_UPDATE_INTERVAL          = 1 * time.Minute
	WIFI_UPDATE_INTERVAL           = 1 * time.Minute
	WIFI_INTERFACE_UPDATE_INTERVAL = 2 * time.Minute
)

type API struct {
	lock         *sync.RWMutex
	wifiClients  map[string]WifiClient
	ros          *RouterOSConnection
	firstRunLock *sync.RWMutex
	firstRun     bool

	// keyed by MAC
	leasesByMac      map[string]DHCPLeaseEntry
	lastLeasesUpdate time.Time

	// keyed by MAC
	wifiRegistrations map[string]RegistrationTableEntry
	lastWifiUpdate    time.Time

	// keyed by MAC
	wifiRadios      map[string]WifiRadioEntry
	lastRadioUpdate time.Time

	// keyed by interface name
	wifiInterfaces          map[string]WifiInterfaceEntry
	lastWifiInterfaceUpdate time.Time
}

type WifiClient struct {
	ActiveAddress string `json:"activeAddress"`
	HostName      string `json:"hostName"`
	MacAddress    string `json:"macAddress"`
	RSSI          string `json:"rssi"`
	SSID          string `json:"ssid"`
	Cap           string `json:"cap"`
	Band          string `json:"band"`
	Comment       string `json:"comment"`
}

func NewAPI(ros *RouterOSConnection) *API {

	api := &API{
		lock:              &sync.RWMutex{},
		leasesByMac:       map[string]DHCPLeaseEntry{},
		ros:               ros,
		wifiRegistrations: map[string]RegistrationTableEntry{},
		wifiRadios:        map[string]WifiRadioEntry{},
		wifiClients:       map[string]WifiClient{},
		wifiInterfaces:    map[string]WifiInterfaceEntry{},
		firstRunLock:      &sync.RWMutex{},
	}

	api.firstRun = true
	api.firstRunLock.Lock()

	// start the update loop
	go api.updateLoop()

	return api
}

func (a *API) GetWifiClients() map[string]WifiClient {
	a.firstRunLock.RLock()
	defer a.firstRunLock.RUnlock()
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.wifiClients
}

func (a *API) updateLoop() {
	for {
		a.update()
		time.Sleep(1 * time.Second)
	}
}

func (a *API) update() {

	dirty := false

	updated := a.updateLeases()
	dirty = dirty || updated

	updated = a.updateWifiRegistrations()
	dirty = dirty || updated

	updated = a.updateWifiRadios()
	dirty = dirty || updated

	updated = a.updateWifiInterfaces()
	dirty = dirty || updated

	if dirty {
		// recompute the wifi clients table
		newWifiClients := map[string]WifiClient{}

		// for each wifi registration entry retrieve its data from the leases and radio tables
		for _, reg := range a.wifiRegistrations {
			lease, hasLease := a.leasesByMac[reg.MacAddress]

			sanitizedInterface := strings.TrimSuffix(reg.Interface, "-virtual1")
			sanitizedInterface = strings.TrimSuffix(sanitizedInterface, "-virtual")

			wifiInterface, _ := a.wifiInterfaces[sanitizedInterface]

			//lookup the radio based on wifi interface mac address

			radio, hasRadio := a.wifiRadios[wifiInterface.MacAddress]

			client := WifiClient{
				MacAddress: reg.MacAddress,
				RSSI:       reg.Signal,
				SSID:       reg.SSID,
				Band:       reg.Band,
				Comment:    reg.Comment,
			}
			if hasLease {
				client.ActiveAddress = lease.ActiveAddress
				client.HostName = lease.HostName
			}
			if hasRadio {
				client.Cap = radio.GetCapName()
			}

			newWifiClients[client.MacAddress] = client
		}

		a.lock.Lock()
		a.wifiClients = newWifiClients
		a.lock.Unlock()
		if a.firstRun {
			a.firstRun = false
			a.firstRunLock.Unlock()
		}

	}

}

func (a *API) updateLeases() bool {
	dirty := false
	if time.Since(a.lastLeasesUpdate) > LEASE_UPDATE_INTERVAL {
		leases, err := a.ros.GetDHCPLeases()
		if err == nil {

			for _, lease := range leases {
				a.leasesByMac[lease.MacAddress] = lease
			}
			a.lastLeasesUpdate = time.Now()
			dirty = true
		}
	}
	return dirty
}

func (a *API) updateWifiRegistrations() bool {
	dirty := false
	if time.Since(a.lastWifiUpdate) > WIFI_UPDATE_INTERVAL {
		wifi, err := a.ros.GetRegistrationTable()
		if err == nil {
			for _, reg := range wifi {
				a.wifiRegistrations[reg.MacAddress] = reg
			}
			a.lastWifiUpdate = time.Now()
			dirty = true
		}
	}
	return dirty
}

func (a *API) updateWifiRadios() bool {
	dirty := false
	if time.Since(a.lastRadioUpdate) > RADIO_UPDATE_INTERVAL {
		radios, err := a.ros.GetWifiRadios()
		if err == nil {
			for _, radio := range radios {
				a.wifiRadios[radio.RadioMac] = radio
			}
			a.lastRadioUpdate = time.Now()
			dirty = true
		}
	}

	return dirty
}

func (a *API) updateWifiInterfaces() bool {
	dirty := false
	if time.Since(a.lastWifiInterfaceUpdate) > WIFI_INTERFACE_UPDATE_INTERVAL {
		interfaces, err := a.ros.GetWifiInterfaces()
		if err == nil {
			for _, iface := range interfaces {
				a.wifiInterfaces[iface.Name] = iface
			}
			a.lastWifiInterfaceUpdate = time.Now()
			dirty = true
		}
	}
	return dirty
}
