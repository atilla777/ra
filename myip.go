package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ExteranlIP struct {
	Ip      string `json:"ip"`
	Country string `json:"country"`
	Cc      string `json:"cc"`
}

func myIP() (r ExteranlIP, e error) {
	var result ExteranlIP
	resp, err := http.Get("https://api.myip.com")
	if err != nil {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("External ip get error: %s", err)}
		return result, fmt.Errorf("External ip get error: http response status %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("External ip get error: http response status %s", resp.StatusCode)}
		return result, fmt.Errorf("External ip get error: http response status %s", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("External ip get error: %s", err)}
		return result, fmt.Errorf("External ip get error: %s", err)
	}

	return result, nil
}
