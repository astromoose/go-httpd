package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Test_NewServer performs basic testing of the HTTP service.
func Test_NewServer(t *testing.T) {
	store := newTestStore()
	s := &testServer{New(":0", store)}
	if s == nil {
		t.Fatal("failed to create HTTP service")
	}

	if err := s.Start(); err != nil {
		t.Fatalf("failed to start HTTP service: %s", err)
	}

	b := doGet(t, s.URL(), "k1")
	if string(b) != `{"k1":""}` {
		t.Fatalf("wrong value received for key k1: %s", string(b))
	}

	doPost(t, s.URL(), "k1", "v1")

	b = doGet(t, s.URL(), "k1")
	if string(b) != `{"k1":"v1"}` {
		t.Fatalf("wrong value received for key k1: %s", string(b))
	}

	store.m["k2"] = "v2"
	b = doGet(t, s.URL(), "k2")
	if string(b) != `{"k2":"v2"}` {
		t.Fatalf("wrong value received for key k2: %s", string(b))
	}

	doDelete(t, s.URL(), "k2")
	b = doGet(t, s.URL(), "k2")
	if string(b) != `{"k2":""}` {
		t.Fatalf("wrong value received for key k2: %s", string(b))
	}

}

// testServer represents a service under test.
type testServer struct {
	*Service
}

// URL returns the URL of the service.
func (t *testServer) URL() string {
	port := strings.TrimLeft(t.Addr().String(), "[::]:")
	return fmt.Sprintf("http://127.0.0.1:%s", port)
}

// testStore represents a mock store, demonstrating the use of interfaces
// to mock out a real store.
type testStore struct {
	m map[string]string
}

// newTestStore returns an initialized mock store.
func newTestStore() *testStore {
	return &testStore{
		m: make(map[string]string),
	}
}

// Get gets the requested key.
func (t *testStore) Get(key string) (string, error) {
	return t.m[key], nil
}

// Set sets the given key to given value.
func (t *testStore) Set(key, value string) error {
	t.m[key] = value
	return nil
}

// Delete delets the given key.
func (t *testStore) Delete(key string) error {
	delete(t.m, key)
	return nil
}

func doGet(t *testing.T, url, key string) string {
	resp, err := http.Get(fmt.Sprintf("%s/key/%s", url, key))
	if err != nil {
		t.Fatalf("failed to GET key: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %s", err)
	}
	return string(body)
}

func doPost(t *testing.T, url, key, value string) {
	b, err := json.Marshal(map[string]string{key: value})
	if err != nil {
		t.Fatalf("failed to encode key and value for POST: %s", err)
	}
	resp, err := http.Post(fmt.Sprintf("%s/key", url), "application-type/json", bytes.NewReader(b))
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("POST request failed: %s", err)
	}
}

func doDelete(t *testing.T, u, key string) {
	ru, err := url.Parse(fmt.Sprintf("%s/key/%s", u, key))
	if err != nil {
		t.Fatalf("failed to parse URL for delete: %s", err)
	}
	req := &http.Request{
		Method: "DELETE",
		URL:    ru,
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to GET key: %s", err)
	}
	defer resp.Body.Close()
}
