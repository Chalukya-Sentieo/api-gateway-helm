package main

import (
	"fmt"
	"bytes"
	"strings"
	"net/http"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

type Config struct {
	AuthServerURL string	//Mandatory
	HeaderName string		//Optional	Defaults to `authorization` (bearer token)
	CookieName string		//Optional	Defaults to `x-api-token`
	ParamName  string		//Optional	Defaults to `x-api-token`
	AuthMethod string		//Optional	Defaults to `GET`
	AuthTokenKey string		//Optional	Defaults to `auth_token`
}

func New() interface{} {
	return &Config{}
}

var headersToSend = make(map[string][]string)

func (config Config) Access(kong *pdk.PDK) {
	_ = kong.Response.SetHeader("content-type", "application/json")
	/**
	Check if Header has the authentication token
	*/
	headerName := config.HeaderName
	if headerName == "" {
		headerName = "authorization"
	}
	token, _ := kong.Request.GetHeader(headerName)

	/**
	Check if the query arg has the token in case header doesn't have it
	*/
	if token == "" {
		paramName := config.ParamName
		if paramName == "" {
			paramName = "x-api-token"
		}
		token, _ = kong.Request.GetQueryArg(paramName)
	} else {
		/**
		In case the token header was found
		*/
		valSplit := strings.Split(token, " ")
		token = valSplit[len(valSplit) - 1]	//Get Bearer Token
	}

	/**
	Check if the cookie has the token in case header and query args don't have it
	*/
	if token == "" {
		cookieHeader, _ := kong.Request.GetHeader("cookie")
		
		cookies := strings.Split(cookieHeader, "; ")
		for i := 0; i < len(cookies); i++ {
			cookieName := config.CookieName
			if cookieName == "" {
				cookieName = "x-api-token"
			}
			if strings.HasPrefix(strings.ToLower(cookies[i]), cookieName) {
				token = strings.Split(cookies[i], "=")[1]
			}
		}
	}

	if token == "" {
		kong.Response.Exit(401, "{\"message\": \"Token Not Found !\"}", headersToSend)
	} else {
		/**
		Call Auth Server with the fetched token
		*/
		AuthServiceURL := config.AuthServerURL
		AuthMethod := "GET"

		AuthTokenKey := "auth_token"

		if config.AuthMethod != "" {
			AuthMethod = config.AuthMethod
		}
		if config.AuthTokenKey != "" {
			AuthTokenKey = config.AuthTokenKey
		}

		if AuthMethod == "GET" {
			_, err := http.Get(fmt.Sprintf("%s?%s=%s", AuthServiceURL, AuthTokenKey, token))

			if err != nil {
				kong.Response.Exit(401, "{\"message\": \"Token Could Not Be Authenticated !\"}", headersToSend)
			}
		} else {
			_, err := http.Post(
				AuthServiceURL,
				"application/json",
				bytes.NewBuffer(
					[]byte(fmt.Sprintf(`{"%s":"%s"}`, AuthTokenKey, token)),
				),
			)

			if err != nil {
				kong.Response.Exit(401, "{\"message\": \"Token Could Not Be Authenticated !\"}", headersToSend)
			}
		}
		kong.Log.Notice("Token Authenticated !")
	}
}

func main() {
	Version := "1.5"
	Priority := 1000
	_ = server.StartServer(New, Version, Priority)
}
