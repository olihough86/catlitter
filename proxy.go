package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	proxyUrlTemplate = "http://%s:%s@%s:%s"
)

type ProxyConfig struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	GeonodeDNS   string `json:"geonode_dns"`
	GeonodePort  string `json:"geonode_port"`
}

func loadProxyConfig(filename string) (*ProxyConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var config ProxyConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	
	return &config, nil
}

func createGeoNodeClient(config *ProxyConfig) (*http.Client, error) {
	proxy := fmt.Sprintf(proxyUrlTemplate, config.Username, config.Password, config.GeonodeDNS, config.GeonodePort)
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to form proxy URL: %v", err)
	}
	
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		MaxIdleConns:        48,
		MaxIdleConnsPerHost: 48,
		IdleConnTimeout:     30 * time.Second,
	}
	
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	
	return client, nil
}
