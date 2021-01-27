package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"os"
	"sync"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize()

	code := m.Run()	
	os.Exit(code)
}

func TestPage(t *testing.T) {
	ip := "1.1.1.1"
	requestCount := 100
	normalPrefix := "counter_normal_1_"
	errorPrefix := "counter_error_1_"

	sendRequests(normalPrefix, errorPrefix, requestCount, ip, t)

	//arrange response values
	iter := a.RedisClient.Scan(0, normalPrefix + "*", 0).Iterator()
	allValues := make(map[string]bool)
	for iter.Next() {
		allValues[iter.Val()] = true
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	errorCounter := 0
	iter = a.RedisClient.Scan(0, errorPrefix + "*", 0).Iterator()
	for iter.Next() {
		errorCounter++
	}

	//verify
	if errorCounter == 40 {
		for i := 1; i <= 60; i++ {
			if allValues[normalPrefix + strconv.Itoa(i)] == false {
				t.Error("fail")		
			}
		}
		t.Log("success")
	} else {
		t.Error("fail")
	}
}

func TestPageResetLimit(t *testing.T) {
	ip := "2.2.2.2"
	requestCount := 60
	normalPrefix := "counter_normal_2_"
	errorPrefix := "counter_error_2_"

	sendRequests(normalPrefix, errorPrefix, requestCount, ip, t)

	time.Sleep(61 * time.Second)

	sendRequests(normalPrefix, errorPrefix, requestCount, ip, t)

	//arrange response values
	iter := a.RedisClient.Scan(0, normalPrefix + "*", 0).Iterator()
	allValues := make(map[string]bool)
	for iter.Next() {
		allValues[iter.Val()] = true
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	errorCounter := 0
	iter = a.RedisClient.Scan(0, errorPrefix + "*", 0).Iterator()
	for iter.Next() {
		errorCounter++
	}

	//verify
	if errorCounter == 0 {
		for i := 1; i <= 60; i++ {
			if allValues[normalPrefix + strconv.Itoa(i)] == false {
				t.Error("fail")		
			}
		}
		t.Log("success")
	} else {
		t.Error("fail")
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func sendRequests(normalPrefix string, errorPrefix string, requestCount int, ip string, t *testing.T) {
	req, _ := http.NewRequest("GET", "/page", nil)
	req.Header.Add("X-Forwarded-For", ip)

	var myWaitGroup sync.WaitGroup
	myWaitGroup.Add(requestCount)

	for i := 0; i < requestCount; i++ {
		go func(w *sync.WaitGroup) {
			body, _ := ioutil.ReadAll(executeRequest(req).Result().Body)
			response := string(body)
			
			if response != "Error" {
				a.RedisClient.Set(normalPrefix + response, response, 0)
			} else {
				a.RedisClient.Set(errorPrefix + uuid.New().String() + "_" + response, response, 0)
			}
			
			w.Done()
		}(&myWaitGroup)
	}

	myWaitGroup.Wait()

	t.Cleanup(func() {
		clearRedisData(normalPrefix, errorPrefix)
	})
}

func clearRedisData(normalPrefix, errorPrefix string) {
	iter := a.RedisClient.Scan(0, normalPrefix + "*", 0).Iterator()
	for iter.Next() {
		a.RedisClient.Del(iter.Val())
	}
	
	iter = a.RedisClient.Scan(0, errorPrefix + "*", 0).Iterator()
	for iter.Next() {
		a.RedisClient.Del(iter.Val())
	}
 }