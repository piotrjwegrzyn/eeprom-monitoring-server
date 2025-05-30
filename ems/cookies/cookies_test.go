package cookies

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStore_Create(t *testing.T) {
	store := NewStore(time.Hour)
	testWriter := httptest.NewRecorder()

	store.Create(context.Background(), testWriter, "test")

	if len(store.cookies) != 1 {
		t.Errorf("expected 1 cookie, got %d", len(store.cookies))
	}

	for _, cookie := range testWriter.Result().Cookies() {
		if cookie.Name != "session_token" {
			t.Errorf("expected cookie name to be 'session_token', got %s", cookie.Name)
		}
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewStore(time.Hour)
	testWriter := httptest.NewRecorder()

	store.Create(context.Background(), testWriter, "test")

	cookie := testWriter.Result().Cookies()[0]

	store.Delete(context.Background(), testWriter, &cookie.Value)

	if len(store.cookies) != 0 {
		t.Errorf("expected 0 cookies, got %d", len(store.cookies))
	}
}

func TestStore_IsSignedIn(t *testing.T) {
	t.Run("is signed in", func(t *testing.T) {
		store := NewStore(time.Hour)
		testWriter := httptest.NewRecorder()

		store.Create(context.Background(), testWriter, "test")

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(testWriter.Result().Cookies()[0])

		got := store.IsSignedIn(request)

		if !got {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("no cookie set", func(t *testing.T) {
		store := NewStore(time.Hour)

		request := httptest.NewRequest(http.MethodGet, "/", nil)

		got := store.IsSignedIn(request)

		if got {
			t.Errorf("expected false, got true")
		}
	})
}
