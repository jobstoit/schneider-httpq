package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

const (
	TestURLKey = "dfjklaflkdidfsofwem"
)

func TestPublish(t *testing.T) {
	wg := sync.WaitGroup{}

	toPublish := []string{"test 1", "test 2"}

	httpq := NewHTTPQ()
	svr := httptest.NewServer(httpq)

	for _, p := range toPublish {
		wg.Add(1)
		go func() {
			defer wg.Done()
			expectErr(t, makePubRequest(svr.URL, p, 100))
		}()
	}

	wg.Wait()
	time.Sleep(time.Second)

	eq(t, 0, httpq.RxBytes)
	eq(t, 2, httpq.PubFails)
}

func TestConsume(t *testing.T) {
	wg := sync.WaitGroup{}
	amount := 2

	httpq := NewHTTPQ()
	svr := httptest.NewServer(httpq)

	for i := 0; i < amount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := makeSubRequest(svr.URL, 100)
			expectErr(t, err)
		}()
	}

	wg.Wait()
	time.Sleep(time.Second)

	eq(t, amount, httpq.SubFails)
}

func TestPublishAndConsume(t *testing.T) {
	wg := sync.WaitGroup{}

	toPublish := []string{"test 1", "test 2", "test 3"}
	totalBytes := 0
	httpq := NewHTTPQ()

	svr := httptest.NewServer(httpq)

	for _, p := range toPublish {
		wg.Add(2)
		go func() {
			defer wg.Done()
			unexpectErr(t, makePubRequest(svr.URL, p, 2000))
		}()

		go func() {
			defer wg.Done()
			res, err := makeSubRequest(svr.URL, 2000)
			unexpectErr(t, err)
			eq(t, p, res)
		}()

		totalBytes += len(p)
	}

	wg.Wait()

	eq(t, totalBytes, httpq.RxBytes)
	eq(t, totalBytes, httpq.TxBytes)
	eq(t, 0, httpq.PubFails)
	eq(t, 0, httpq.SubFails)
}

func TestStats(t *testing.T) {
	httpq := NewHTTPQ()
	httpq.TxBytes = 18
	httpq.RxBytes = 18
	httpq.PubFails = 3
	httpq.SubFails = 5

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	wr := httptest.NewRecorder()
	httpq.Stats(wr, req)

	res := wr.Result()
	eq(t, http.StatusOK, res.StatusCode)

	body := StatsResponse{}
	unexpectErr(t, json.NewDecoder(res.Body).Decode(&body))

	eq(t, httpq.TxBytes, body.PublishedBytes)
	eq(t, httpq.RxBytes, body.ConsumedBytes)
	eq(t, httpq.PubFails, body.PublishedFails)
	eq(t, httpq.SubFails, body.ConsumedFails)
}

func makePubRequest(serverURL, body string, timeoutMiliseconds int) error {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", serverURL, TestURLKey),
		bytes.NewBufferString(body),
	)
	if err != nil {
		return err
	}

	_, err = requestWithTimeout(req, timeoutMiliseconds)
	return err
}

func makeSubRequest(serverURL string, timeoutMiliseconds int) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/%s", serverURL, TestURLKey),
		nil,
	)
	if err != nil {
		return "", err
	}

	res, err := requestWithTimeout(req, timeoutMiliseconds)
	if err != nil {
		return "", err
	}

	buff := &bytes.Buffer{}
	_, err = buff.ReadFrom(res.Body) // nolint: errcheck

	return buff.String(), err
}

func requestWithTimeout(req *http.Request, timeoutMiliseconds int) (*http.Response, error) {
	dur, err := time.ParseDuration(fmt.Sprintf("%dms", timeoutMiliseconds))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func expectErr(t *testing.T, e error) {
	if e == nil {
		t.Error("missing expected error")
	}
}

func unexpectErr(t *testing.T, e error) {
	if e != nil {
		t.Errorf("unexpected error: %v", e)
	}
}

func eq[V string | int](t *testing.T, expected, actual V) {
	if expected != actual {
		t.Errorf("expected: %v, but got: %v", expected, actual)
	}
}
