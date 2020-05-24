package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func getConfig() map[string]string {
	var config map[string]string
	configJSON, err := os.Open("config.json")

	if err != nil {
		fmt.Println(err)
	}
	defer configJSON.Close()

	bytesJSON, _ := ioutil.ReadAll(configJSON)

	if err = json.Unmarshal(bytesJSON, &config); err != nil {
		fmt.Println(err)
	}

	return config
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func auth(w http.ResponseWriter, req *http.Request) {
	config := getConfig()
	code := req.URL.Query()["code"]

	// Client provided access code
	if len(code) != 0 {
		// URL to POST to
		url := "https://zoom.us/oauth/token?grant_type=authorization_code&code=" + code[0] + "&redirect_uri=" + config["RedirectURL"]

		// Auth string, in base64 and formatted
		base64auth := "Basic " + b64.StdEncoding.EncodeToString([]byte(config["ClientID"]+":"+config["ClientSecret"]))

		// Create POSt request and set header
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", base64auth)

		// Make request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		// Get response and print
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s\n", body)

		// Return response to client
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		// No access code provided, redirect to auth
		redirectURL := "https://zoom.us/oauth/authorize?response_type=code&client_id=" + config["ClientID"] + "&redirect_uri=" + config["RedirectURL"]
		http.Redirect(w, req, redirectURL, 301)
	}
}

func main() {
	http.HandleFunc("/", auth)
	http.HandleFunc("/headers", headers)

	http.ListenAndServe(":8080", nil)
}
