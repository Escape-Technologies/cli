package escape

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func errorResult(err error) reflect.Value {
	var wrapped error = err
	return reflect.ValueOf(&wrapped).Elem()
}

func TestUnpackExecuteReturnsBody(t *testing.T) {
	t.Parallel()
	want := &v3.UpdateAsset200Response{Id: "asset-1"}
	got, err := unpackExecute([]reflect.Value{reflect.ValueOf(want), errorResult(nil)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestUnpackExecuteSurfacesAPIError(t *testing.T) {
	t.Parallel()
	_, err := unpackExecute([]reflect.Value{
		reflect.ValueOf((*v3.UpdateAsset200Response)(nil)),
		errorResult(errors.New("400 Bad Request")),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400 Bad Request") {
		t.Fatalf("expected api error, got %v", err)
	}
}

func TestUnpackExecuteRejectsTypedNilBody(t *testing.T) {
	t.Parallel()
	_, err := unpackExecute([]reflect.Value{
		reflect.ValueOf((*v3.UpdateAsset200Response)(nil)),
		errorResult(nil),
	})
	if err == nil {
		t.Fatal("expected empty-response error")
	}
	if !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("expected empty response, got %v", err)
	}
}

func TestCreateAssetSchemaRequestRoundTrip(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"asset_type":"SCHEMA","name":"solar-system-openapi","fetch":{"url":"https://tester.tools.escape.tech/solar_system/openapi.json"}}`)
	var payload v3.CreateAssetSchemaRequest
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.CreateSchemaViaFetch == nil {
		t.Fatal("expected fetch variant")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"url":"https://tester.tools.escape.tech/solar_system/openapi.json"`) {
		t.Fatalf("missing fetch url in %s", encoded)
	}
}
