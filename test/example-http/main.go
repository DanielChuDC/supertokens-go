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

	"github.com/supertokens/supertokens-go/supertokens"
	"github.com/supertokens/supertokens-go/supertokens/core"
)

var noOfTimesGetSessionCalledDuringTest int = 0
var noOfTimesRefreshCalledDuringTest int = 0

func main() {
	supertokens.Config(supertokens.ConfigMap{
		Hosts:          "http://localhost:9000",
		CookieSameSite: "lax",
	})
	http.HandleFunc("/login", login)
	http.HandleFunc("/testUserConfig", testUserConfig)
	http.HandleFunc("/multipleInterceptors", multipleInterceptors)
	http.HandleFunc("/", supertokens.Middleware(defaultHandler))
	http.HandleFunc("/beforeeach", beforeeach)
	http.HandleFunc("/testing", testing)
	http.HandleFunc("/logout", supertokens.Middleware(logout))
	http.HandleFunc("/revokeAll", supertokens.Middleware(revokeAll))
	http.HandleFunc("/refresh", supertokens.Middleware(refresh))
	http.HandleFunc("/refreshCalledTime", refreshCalledTime)
	http.HandleFunc("/getSessionCalledTime", getSessionCalledTime)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/testHeader", testHeader)
	http.HandleFunc("/checkDeviceInfo", checkDeviceInfo)
	http.HandleFunc("/checkAllowCredentials", checkAllowCredentials)
	http.HandleFunc("/testError", testError)
	http.HandleFunc("/index.html", index)
	http.HandleFunc("/fail", fail)
	http.HandleFunc("/update-jwt", supertokens.Middleware(updateJwt))
	supertokens.OnTryRefreshToken(customOnTryRefreshTokenError)
	supertokens.OnUnauthorized(customOnUnauthorizedError)
	supertokens.OnGeneralError(customOnGeneralError)
	port := "8080"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

func fail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte(""))
}

func index(w http.ResponseWriter, r *http.Request) {
	dat, _ := ioutil.ReadFile("./static/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(dat)
}

func options(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Headers", "content-type")
	response.Header().Set("Access-Control-Allow-Methods", "*")
	supertokens.SetRelevantHeadersForOptionsAPI(response)
	response.Write([]byte(""))
}

func login(response http.ResponseWriter, request *http.Request) {

	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
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
	_, err = supertokens.CreateNewSession(response, userID)

	if err != nil {
		supertokens.HandleErrorAndRespond(err, response)
		return
	}
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.Write([]byte(userID))

}

func testUserConfig(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	response.Write([]byte(""))

}
func multipleInterceptors(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
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

func defaultHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	noOfTimesGetSessionCalledDuringTest++
	session := supertokens.GetSessionFromRequest(request)
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.Write([]byte(session.GetUserID()))
}

func updateJwt(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
	} else if request.Method == "GET" {
		response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
		response.Header().Set("Access-Control-Allow-Credentials", "true")
		session := supertokens.GetSessionFromRequest(request)
		json.NewEncoder(response).Encode(session.GetJWTPayload())
	} else if request.Method == "POST" {
		var body map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			response.Write([]byte("error when parsing the body"))
			return
		}
		session := supertokens.GetSessionFromRequest(request)
		session.UpdateJWTPayload(body)
		response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
		response.Header().Set("Access-Control-Allow-Credentials", "true")
		json.NewEncoder(response).Encode(session.GetJWTPayload())
	} else {
		response.Write([]byte("incorrect Method, requires POST or GET"))
	}
}

func beforeeach(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	noOfTimesRefreshCalledDuringTest = 0
	noOfTimesGetSessionCalledDuringTest = 0
	core.ResetHandshakeInfo()
	response.Write([]byte(""))
}

func testing(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	}
	value := request.Header.Get("testing")
	if value != "" {
		response.Header().Set("testing", value)
	}
	response.Write([]byte("success"))
}

func logout(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}

	session := supertokens.GetSessionFromRequest(request)
	err := session.RevokeSession()
	if err != nil {
		supertokens.HandleErrorAndRespond(err, response)
		return
	}
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.Write([]byte("success"))

}

func revokeAll(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	session := supertokens.GetSessionFromRequest(request)
	userID := session.GetUserID()
	supertokens.RevokeAllSessionsForUser(userID)
	response.Write([]byte("success"))
}

func refresh(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	noOfTimesRefreshCalledDuringTest++
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.Write([]byte("refresh success"))
}

func refreshCalledTime(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Write([]byte(strconv.Itoa(noOfTimesRefreshCalledDuringTest)))
}

func getSessionCalledTime(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Write([]byte(strconv.Itoa(noOfTimesGetSessionCalledDuringTest)))
}

func ping(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.Write([]byte(""))
}

func testHeader(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	testheader := request.Header.Get("st-custom-header")
	success := testheader != ""
	json.NewEncoder(response).Encode(map[string]interface{}{
		"success": success,
	})
}

func checkDeviceInfo(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	sdkName := request.Header.Get("supertokens-sdk-name")
	sdkVersion := request.Header.Get("supertokens-sdk-version")
	response.Write([]byte(strconv.FormatBool(sdkName == "website" && sdkVersion != "")))
}

func checkAllowCredentials(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "POST" {
		response.Write([]byte("incorrect Method, requires POST"))
		return
	}
	response.Write([]byte(strconv.FormatBool(request.Header.Get("allow-credentials") != "")))
}

func testError(response http.ResponseWriter, request *http.Request) {
	if request.Method == "OPTIONS" {
		options(response, request)
		return
	} else if request.Method != "GET" {
		response.Write([]byte("incorrect Method, requires GET"))
		return
	}
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte("test error message"))
}

func customOnTryRefreshTokenError(err error, response http.ResponseWriter) {
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.WriteHeader(401)
	response.Write([]byte(""))

}

func customOnUnauthorizedError(err error, response http.ResponseWriter) {
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost.org:8080")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.WriteHeader(401)
	response.Write([]byte(""))
}

func customOnGeneralError(err error, response http.ResponseWriter) {
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte("Something went wrong"))
}
