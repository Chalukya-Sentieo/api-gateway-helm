package main

import (
	"os"
	"fmt"
	"bytes"
	"strings"
	"net/http"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

type Config struct {
	headerName string
	cookieName string
	paramName  string
	authServerURL string
	authMethod string
	authTokenKey string
}

func New() interface{} {
	return &Config{}
}

var AuthServiceURL = "http://mockbin.org/bin/c4d291ab-eff3-44a9-b901-ef8bf0f4a3d5"
var AuthMethod = "GET"
var AuthURLEnv, authEnvOk = os.LookupEnv("KONG_AUTH_SERVER_URL")
var AuthMethodEnv, AuthMethodOk = os.LookupEnv("KONG_AUTH_METHOD")

var headersToSend = make(map[string][]string)

func (config Config) Access(kong *pdk.PDK) {
	_ = kong.Response.SetHeader("content-type", "application/json")

	/**
	Check if Header has the authentication token
	*/
	headerName := config.headerName
	if headerName == "" {
		headerName = "authorization"
	}
	token, _ := kong.Request.GetHeader(headerName)


	/**
	Check if the query arg has the token in case header doesn't have it
	*/
	if token == "" {
		paramName := config.paramName
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
			cookieName := config.cookieName
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

		getAuthTokenKey := "token"
		postAuthTokenKey := "sid"

		AuthServiceURLUse := AuthServiceURL
		AuthMethodUse := AuthMethod

		if config.authServerURL != "" {
			AuthServiceURLUse = config.authServerURL
		} else if authEnvOk {
			AuthServiceURLUse = AuthURLEnv
		}

		if config.authMethod != "" {
			AuthMethodUse = config.authMethod
		} else if AuthMethodOk {
			AuthMethodUse = AuthMethodEnv
		}

		if config.authTokenKey != "" {
			getAuthTokenKey = config.authTokenKey
			postAuthTokenKey = config.authTokenKey
		}
		
		if AuthMethodUse == "GET" {
			_, err := http.Get(fmt.Sprintf("%s?%s=%s", AuthServiceURLUse, getAuthTokenKey, token))

			if err != nil {
				kong.Response.Exit(401, "{\"message\": \"Token Could Not Be Authenticated !\"}", headersToSend)
			}
		} else {
			_, err := http.Post(
				AuthServiceURLUse,
				"application/json",
				bytes.NewBuffer(
					[]byte(fmt.Sprintf(`{"%s":"%s"}`, postAuthTokenKey, token)),
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
	Version := "1.2"
	Priority := 1000
	_ = server.StartServer(New, Version, Priority)
}
