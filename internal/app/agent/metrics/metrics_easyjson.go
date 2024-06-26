// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package metrics

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson2220f231DecodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics(in *jlexer.Lexer, out *RequestMetricSlice) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(RequestMetricSlice, 0, 1)
			} else {
				*out = RequestMetricSlice{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 RequestMetric
			(v1).UnmarshalEasyJSON(in)
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson2220f231EncodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics(out *jwriter.Writer, in RequestMetricSlice) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			(v3).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v RequestMetricSlice) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson2220f231EncodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v RequestMetricSlice) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson2220f231EncodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RequestMetricSlice) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson2220f231DecodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *RequestMetricSlice) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson2220f231DecodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics(l, v)
}
func easyjson2220f231DecodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics1(in *jlexer.Lexer, out *RequestMetric) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = string(in.String())
		case "type":
			out.MType = string(in.String())
		case "delta":
			out.Delta = int64(in.Int64())
		case "value":
			out.Value = float64(in.Float64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson2220f231EncodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics1(out *jwriter.Writer, in RequestMetric) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.String(string(in.ID))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.String(string(in.MType))
	}
	{
		const prefix string = ",\"delta\":"
		out.RawString(prefix)
		out.Int64(int64(in.Delta))
	}
	{
		const prefix string = ",\"value\":"
		out.RawString(prefix)
		out.Float64(float64(in.Value))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RequestMetric) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson2220f231EncodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v RequestMetric) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson2220f231EncodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RequestMetric) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson2220f231DecodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *RequestMetric) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson2220f231DecodeGithubComWebkimruGoYandexMetricsInternalAppAgentMetrics1(l, v)
}
