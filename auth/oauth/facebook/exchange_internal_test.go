package facebook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(srv *httptest.Server) *Facebook {
	f := New(Config{ClientID: "cid", ClientSecret: "sec"})
	f.endpoint.TokenURL = srv.URL + "/token"
	f.endpoint.AuthURL = srv.URL + "/auth"
	f.userInfo = srv.URL + "/me?fields=id,name,email"
	f.httpClient = srv.Client()
	return f
}

func TestExchange_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"at","token_type":"bearer"}`))
		case "/me":
			_, _ = w.Write([]byte(`{"id":"fb-1","name":"Bob","email":"bob@x.com","picture":{"data":{"url":"http://p"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	res, err := newTestProvider(srv).Exchange(context.Background(), "code", "https://app/cb")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if res.User.ID != "fb-1" || res.User.Email != "bob@x.com" {
		t.Errorf("user = %+v", res.User)
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
