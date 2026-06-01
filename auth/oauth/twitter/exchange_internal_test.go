package twitter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(srv *httptest.Server) *Twitter {
	tw := New(Config{ClientID: "cid", ClientSecret: "sec"})
	tw.endpoint.TokenURL = srv.URL + "/token"
	tw.endpoint.AuthURL = srv.URL + "/auth"
	tw.userInfo = srv.URL + "/me"
	tw.httpClient = srv.Client()
	return tw
}

func TestExchange_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"at","token_type":"bearer"}`))
		case "/me":
			_, _ = w.Write([]byte(`{"data":{"id":"t-1","name":"Dave","username":"dave","profile_image_url":"http://img"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	res, err := newTestProvider(srv).Exchange(context.Background(), "code", "https://app/cb")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if res.User.ID != "t-1" || res.User.Name != "Dave" {
		t.Errorf("user = %+v", res.User)
	}
	if res.User.Meta["username"] != "dave" {
		t.Errorf("username meta = %v", res.User.Meta["username"])
	}
}

func TestExchange_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	if _, err := newTestProvider(srv).Exchange(context.Background(), "bad", "https://app/cb"); err == nil {
		t.Error("expected error")
	}
}
