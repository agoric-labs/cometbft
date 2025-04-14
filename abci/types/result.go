package types

import (
	"bytes"
	"encoding/json"

	"github.com/gogo/protobuf/jsonpb"
)

const (
	CodeTypeOK uint32 = 0
)

// IsOK returns true if Code is OK.
func (r ResponseCheckTx) IsOK() bool {
	return r.Code == CodeTypeOK
}

// IsErr returns true if Code is something other than OK.
func (r ResponseCheckTx) IsErr() bool {
	return r.Code != CodeTypeOK
}

// IsOK returns true if Code is OK.
func (r ResponseDeliverTx) IsOK() bool {
	return r.Code == CodeTypeOK
}

// IsErr returns true if Code is something other than OK.
func (r ResponseDeliverTx) IsErr() bool {
	return r.Code != CodeTypeOK
}

// IsOK returns true if Code is OK.
func (r ResponseQuery) IsOK() bool {
	return r.Code == CodeTypeOK
}

// IsErr returns true if Code is something other than OK.
func (r ResponseQuery) IsErr() bool {
	return r.Code != CodeTypeOK
}

//---------------------------------------------------------------------------
// override JSON marshaling so we emit defaults (ie. disable omitempty)

var (
	jsonpbMarshaller = jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	jsonpbUnmarshaller = jsonpb.Unmarshaler{}
)

func (r *ResponseSetOption) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *ResponseSetOption) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

func (r *ResponseCheckTx) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *ResponseCheckTx) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

func (r *ResponseDeliverTx) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *ResponseDeliverTx) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

func (r *ResponseQuery) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *ResponseQuery) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

func (r *ResponseCommit) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *ResponseCommit) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

func (r EventAttribute) MarshalJSONPB(marshaler *jsonpb.Marshaler) ([]byte, error) {
	return r.MarshalJSON()
}

func (r *EventAttribute) UnmarshalJSONPB(unmarshaler *jsonpb.Unmarshaler, b []byte) error {
	return r.UnmarshalJSON(b)
}

// stringyEventAttribute [AGORIC] avoids base64 encoding of the key and value
// fields in the EventAttribute struct. This is necessary because the
// EventAttribute struct is declared with byte slices, but code in the wild just
// cast to and from strings, without any use of base64.
type stringyEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r EventAttribute) MarshalJSON() ([]byte, error) {
	stringyAttr := stringyEventAttribute{
		Key:   string(r.Key),
		Value: string(r.Value),
	}
	return json.Marshal(stringyAttr)
}

func (r *EventAttribute) UnmarshalJSON(b []byte) error {
	var stringyAttr stringyEventAttribute
	if err := json.Unmarshal(b, &stringyAttr); err != nil {
		return err
	}
	r.Key = []byte(stringyAttr.Key)
	r.Value = []byte(stringyAttr.Value)
	return nil
}

// Some compile time assertions to ensure we don't
// have accidental runtime surprises later on.

// jsonEncodingRoundTripper ensures that asserted
// interfaces implement both MarshalJSON and UnmarshalJSON
type jsonRoundTripper interface {
	json.Marshaler
	json.Unmarshaler
}

type jsonpbRoundTripper interface {
	jsonpb.JSONPBMarshaler
	jsonpb.JSONPBUnmarshaler
}

var _ jsonRoundTripper = (*ResponseCommit)(nil)
var _ jsonRoundTripper = (*ResponseQuery)(nil)
var _ jsonRoundTripper = (*ResponseDeliverTx)(nil)
var _ jsonRoundTripper = (*ResponseCheckTx)(nil)
var _ jsonRoundTripper = (*ResponseSetOption)(nil)

var _ jsonRoundTripper = (*EventAttribute)(nil)
var _ jsonpbRoundTripper = (*EventAttribute)(nil)
var _ jsonpb.JSONPBMarshaler = EventAttribute{}
