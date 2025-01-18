package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
)

//go:embed public
var embeddedFS embed.FS

var (
	host     string
	user     string
	password string
)

func init() {
	flag.StringVar(&host, "host", "", "host")
	flag.StringVar(&user, "user", "admin", "user")
	flag.StringVar(&password, "password", "", "password")
}

func main() {

	flag.Parse()
	r := NewRouterOSConnection(host, user, password)

	api := NewAPI(r)

	// clients := api.GetWifiClients()

	// for _, client := range clients {
	// 	fmt.Printf("%-20s %-20s %-20s %-20s %-20s %-20s\n", client.ActiveAddress, client.HostName, client.MacAddress, client.RSSI, client.Cap, client.SSID)
	// }

	http.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		graph := api.GetWifiClients()
		m, _ := json.Marshal(graph)

		w.Write(m)
	})

	serverRoot, err := fs.Sub(embeddedFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(serverRoot)))

	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}
