package main

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/ddtmachado/lb-dxp/traefik"
)

var configBackends = make(map[string][]net.IP)

func lookupEnv(env string) string {
	envValue, hasEnv := os.LookupEnv(env)
	if !hasEnv {
		log.Fatalf("%s variable not set", env)
	}
	return envValue
}

func main() {
	serviceName := lookupEnv("SERVICE_NAME")
	backendPort := lookupEnv("BACKEND_PORT")
	traefikAddress := lookupEnv("TRAEFIK_ADDRESS")
	waitInterval, err := strconv.Atoi(lookupEnv("WAIT_INTERVAL"))
	if err != nil {
		log.Fatalf("Wrong value for WAIT_INTERVAL variable %v", waitInterval)
	}

	traefikBinary, lookErr := exec.LookPath("traefik")
	if lookErr != nil {
		log.Fatal("Could not find the traefik binary")
	}

	cmd := exec.Command(traefikBinary,
		"--api",
		"--rest",
		"--logLevel=INFO",
		"--accessLog",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Unable to start traefik process: %v", err)
	}

	for {
		time.Sleep(time.Duration(waitInterval) * time.Second)
		serverIPs, err := net.LookupIP(serviceName)
		if err != nil {
			log.Fatal("Couldn't get service addresses")
		}

		sort.Slice(serverIPs, func(i, j int) bool {
			return bytes.Compare(serverIPs[i], serverIPs[j]) < 0
		})

		currentIPs := configBackends[serviceName]
		if reflect.DeepEqual(serverIPs, currentIPs) {
			continue
		}
		configBackends[serviceName] = serverIPs

		data, err := traefik.NewJsonConfig(serviceName, backendPort, serverIPs)
		log.Printf("Traefik json payload: %s", data)

		payload := bytes.NewReader(data)
		request, err := http.NewRequest(http.MethodPut, traefikAddress, payload)
		client := &http.Client{}
		_, err = client.Do(request)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Traefik config updated")
	}
}
