/*
 * Copyright (c) 2020, VRAI Labs and/or its affiliates. All rights reserved.
 *
 * This software is licensed under the Apache License, Version 2.0 (the
 * "License") as published by the Apache Software Foundation.
 *
 * You may not use this file except in compliance with the License. You may
 * obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-go/gin/supertokens"
	"github.com/supertokens/supertokens-go/supertokens/core"
)

var noOfTimesGetSessionCalledDuringTest int = 0
var noOfTimesRefreshCalledDuringTest int = 0

func main() {
	supertokens.Config(supertokens.ConfigMap{
		Hosts:          "http://localhost:9000",
		CookieSameSite: "lax",
	})
	r := gin.Default()

	// it's important to set CORS before any route. Otherwise it will not work
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost.org:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "HEAD", "OPTIONS"},
		AllowHeaders:     append([]string{"Content-Type"}, supertokens.GetCORSAllowedHeaders()...),
		AllowCredentials: true,
	}))
	r.Any("/login", login)
	r.Any("/testUserConfig", testUserConfig)
	r.Any("/multipleInterceptors", multipleInterceptors)
	r.Any("/", supertokens.Middleware(), defaultHandler)
	r.Any("/beforeeach", beforeeach)
	r.Any("/testing", testing)
	r.Any("/logout", supertokens.Middleware(), logout)
	r.Any("/revokeAll", supertokens.Middleware(), revokeAll)
	r.Any("/refresh", supertokens.Middleware(), refresh)
	r.Any("/refreshCalledTime", refreshCalledTime)
	r.Any("/getSessionCalledTime", getSessionCalledTime)
	r.Any("/ping", ping)
	r.Any("/testHeader", testHeader)
	r.Any("/checkDeviceInfo", checkDeviceInfo)
	r.Any("/checkAllowCredentials", checkAllowCredentials)
	r.Any("/testError", testError)
	r.Any("/index.html", index)
	r.Any("/fail", fail)
	r.Any("/update-jwt", supertokens.Middleware(), updateJwt)
	supertokens.OnTryRefreshToken(customOnTryRefreshTokenError)
	supertokens.OnUnauthorized(customOnUnauthorizedError)
	supertokens.OnGeneralError(customOnGeneralError)
	port := "8080"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	r.Run("0.0.0.0:" + port)
}

func fail(c *gin.Context) {
	w := c.Writer
	w.WriteHeader(404)
	w.Write([]byte(""))
}

func index(c *gin.Context) {
	w := c.Writer
	dat, _ := ioutil.ReadFile("./static/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(dat)
}

func login(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}

	var body map[string]interface{}
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		response.Write([]byte("error when parsing body"))
		return
	}
	userID := body["userId"].(string)
	_, err = supertokens.CreateNewSession(c, userID)

	if err != nil {
		supertokens.HandleErrorAndRespond(err, c)
		return
	}
	response.Write([]byte(userID))

}

func testUserConfig(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	response.Write([]byte(""))

}
func multipleInterceptors(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	interceptorheader2 := request.Header.Get("interceptorheader2")
	interceptorheader1 := request.Header.Get("interceptorheader1")

	var resp string
	if interceptorheader2 != "" && interceptorheader1 != "" {
		resp = "success"
	} else {
		resp = "failure"
	}
	response.Write([]byte(resp))
}

func defaultHandler(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	noOfTimesGetSessionCalledDuringTest++
	var session *supertokens.Session = supertokens.GetSessionFromRequest(c)
	response.Write([]byte(session.GetUserID()))
}

func updateJwt(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method == "GET" {
		session := supertokens.GetSessionFromRequest(c)
		json.NewEncoder(response).Encode(session.GetJWTPayload())
	} else if request.Method == "POST" {
		var body map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			response.Write([]byte("error when parsing the body"))
			return
		}
		session := supertokens.GetSessionFromRequest(c)
		session.UpdateJWTPayload(body)
		json.NewEncoder(response).Encode(session.GetJWTPayload())
	} else {
		response.Write([]byte("incorrect Method, requires POST or GET"))
	}
}

func beforeeach(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	noOfTimesRefreshCalledDuringTest = 0
	noOfTimesGetSessionCalledDuringTest = 0
	core.ResetHandshakeInfo()
	response.Write([]byte(""))
}

func testing(c *gin.Context) {
	response := c.Writer
	request := c.Request
	value := request.Header.Get("testing")
	if value != "" {
		response.Header().Set("testing", value)
	}
	response.Write([]byte("success"))
}

func logout(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}

	session := supertokens.GetSessionFromRequest(c)
	err := session.RevokeSession()
	if err != nil {
		supertokens.HandleErrorAndRespond(err, c)
		return
	}
	response.Write([]byte("success"))

}

func revokeAll(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	session := supertokens.GetSessionFromRequest(c)
	userID := session.GetUserID()
	supertokens.RevokeAllSessionsForUser(userID)
	response.Write([]byte("success"))
}

func refresh(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	noOfTimesRefreshCalledDuringTest++
	response.Write([]byte("refresh success"))
}

func refreshCalledTime(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.Write([]byte(strconv.Itoa(noOfTimesRefreshCalledDuringTest)))
}

func getSessionCalledTime(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.Write([]byte(strconv.Itoa(noOfTimesGetSessionCalledDuringTest)))
}

func ping(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.Write([]byte(""))
}

func testHeader(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	testheader := request.Header.Get("st-custom-header")
	success := testheader != ""
	json.NewEncoder(response).Encode(map[string]interface{}{
		"success": success,
	})
}

func checkDeviceInfo(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	sdkName := request.Header.Get("supertokens-sdk-name")
	sdkVersion := request.Header.Get("supertokens-sdk-version")
	response.Write([]byte(strconv.FormatBool(sdkName == "website" && sdkVersion != "")))
}

func checkAllowCredentials(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	response.Write([]byte(strconv.FormatBool(request.Header.Get("allow-credentials") != "")))
}

func testError(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte("test error message"))
}

func customOnTryRefreshTokenError(err error, response http.ResponseWriter) {
	response.WriteHeader(401)
	response.Write([]byte(""))

}

func customOnUnauthorizedError(err error, response http.ResponseWriter) {
	response.WriteHeader(401)
	response.Write([]byte(""))
}

func customOnGeneralError(err error, response http.ResponseWriter) {
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte("Something went wrong"))
}
