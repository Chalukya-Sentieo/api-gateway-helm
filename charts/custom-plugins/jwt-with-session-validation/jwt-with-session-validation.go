package main

import (
	"fmt"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"strings"
	"net/http"
)

type Config struct {
	WaitTime int
}

func New() interface{} {
	return &Config{}
}

var AuthServiceURL = "http://mockbin.org/bin/c4d291ab-eff3-44a9-b901-ef8bf0f4a3d5"

var headersToSend = make(map[string][]string)

func (config Config) Access(kong *pdk.PDK) {
	_ = kong.Response.SetHeader("content-type", "application/json")

	kong.Log.Notice("Running JWT Session Validation")

	jwtToken, _ := kong.Request.GetHeader("authorization")

	if jwtToken == "" {
		jwtToken, _ = kong.Request.GetQueryArg("x-api-token")
	} else {
		jwtToken = strings.Split(jwtToken, " ")[1]	//Get Bearer Token
	}

	if jwtToken == "" {
		cookieHeader, _ := kong.Request.GetHeader("cookie")
		
		cookies := strings.Split(cookieHeader, "; ")
		for i := 0; i < len(cookies); i++ {
			if strings.HasPrefix(strings.ToLower(cookies[i]), "x-api-token") {
				jwtToken = strings.Split(cookies[i], "=")[1]
			}
		}
	}

	if jwtToken == "" {
		kong.Response.Exit(401, "{\"message\": \"JWT Token Not Found !\"}", headersToSend)
	} else {
		_, err := http.Get(fmt.Sprintf("%s?token=%s", AuthServiceURL, jwtToken))
		if err != nil {
			kong.Response.Exit(401, "{\"message\": \"JWT Token Could Not Be Authenticated !\"}", headersToSend)
		}

		kong.Log.Notice("JWT Authenticated !")
	}
}

func main() {
	Version := "1.1"
	Priority := 1000
	_ = server.StartServer(New, Version, Priority)
}