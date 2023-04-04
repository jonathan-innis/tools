package main

import (
	"encoding/base64"
	"fmt"

	"github.com/samber/lo"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	config := controllerruntime.GetConfigOrDie()
	caBundle := lo.Must(GetCABundle(config))
	fmt.Println(caBundle)
}

func GetCABundle(config *rest.Config) (string, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return "", err
	}
	_, err = transport.TLSConfigFor(transportConfig) // fills in CAData!
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(transportConfig.TLS.CAData), nil
}
