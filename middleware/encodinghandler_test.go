package middleware

import (
	"io"
	"reflect"
	"testing"
)

func Test_bodyEncorder_BodyAll(t *testing.T) {
	type fields struct {
		bodies     []bodyRW
		acceptEncs []string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]EncordedBody
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := &bodyEncorder{
				bodies:     tt.fields.bodies,
				acceptEncs: tt.fields.acceptEncs,
			}
			if got := be.BodyAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bodyEncorder.BodyAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyEncorder_Encording(t *testing.T) {
	type fields struct {
		bodies     []bodyRW
		acceptEncs []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := &bodyEncorder{
				bodies:     tt.fields.bodies,
				acceptEncs: tt.fields.acceptEncs,
			}
			if got := be.Encording(); got != tt.want {
				t.Errorf("bodyEncorder.Encording() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyEncorder_Length(t *testing.T) {
	type fields struct {
		bodies     []bodyRW
		acceptEncs []string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := &bodyEncorder{
				bodies:     tt.fields.bodies,
				acceptEncs: tt.fields.acceptEncs,
			}
			if got := be.Length(); got != tt.want {
				t.Errorf("bodyEncorder.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyEncorder_Reader(t *testing.T) {
	type fields struct {
		bodies     []bodyRW
		acceptEncs []string
	}
	tests := []struct {
		name   string
		fields fields
		want   io.Reader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := &bodyEncorder{
				bodies:     tt.fields.bodies,
				acceptEncs: tt.fields.acceptEncs,
			}
			if got := be.Reader(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bodyEncorder.Reader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyRW_Length(t *testing.T) {
	type fields struct {
		enc string
		len int
		w   io.Writer
		r   io.Reader
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bodyRW{
				enc: tt.fields.enc,
				len: tt.fields.len,
				w:   tt.fields.w,
				r:   tt.fields.r,
			}
			if got := body.Length(); got != tt.want {
				t.Errorf("bodyRW.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyRW_Encording(t *testing.T) {
	type fields struct {
		enc string
		len int
		w   io.Writer
		r   io.Reader
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bodyRW{
				enc: tt.fields.enc,
				len: tt.fields.len,
				w:   tt.fields.w,
				r:   tt.fields.r,
			}
			if got := body.Encording(); got != tt.want {
				t.Errorf("bodyRW.Encording() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyRW_Reader(t *testing.T) {
	type fields struct {
		enc string
		len int
		w   io.Writer
		r   io.Reader
	}
	tests := []struct {
		name   string
		fields fields
		want   io.Reader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bodyRW{
				enc: tt.fields.enc,
				len: tt.fields.len,
				w:   tt.fields.w,
				r:   tt.fields.r,
			}
			if got := body.Reader(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bodyRW.Reader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBodyEncorder(t *testing.T) {
	tests := []struct {
		name string
		want Middleware
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBodyEncorder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBodyEncorder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetContentEncodingDefault(t *testing.T) {
	type args struct {
		acceptEnc string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no accept-encordings",
			args: args{""},
			want: EncIdentity,
		},
		{
			name: "not support deflate",
			args: args{"deflate, compress"},
			want: EncIdentity,
		},
		{
			name: "not support compress",
			args: args{"compress"},
			want: EncIdentity,
		},
		{
			name: "not support x-gzip",
			args: args{"x-gzip"},
			want: EncIdentity,
		},
		{
			name: "ignore quality",
			args: args{"br;q=0.2, gzip;q=0.8"},
			want: EncBrotli,
		},
		{
			name: "ignore unknown parameter",
			args: args{"br; foo=bar; PI=3.14"},
			want: EncBrotli,
		},
		{
			name: "ignore unknown encording",
			args: args{"unknown"},
			want: EncIdentity,
		},
		{
			name: "ignore unknown encording",
			args: args{"unknown, gzip"},
			want: EncGzip,
		},
		{
			name: "ignore order of accept-encodings",
			args: args{"gzip, br"},
			want: EncBrotli,
		},
		{
			name: "ignore order of accept-encodings 2",
			args: args{"br, gzip"},
			want: EncBrotli,
		},
		{
			name: "br",
			args: args{"br"},
			want: EncBrotli,
		},
		{
			name: "gzip",
			args: args{"gzip"},
			want: EncGzip,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetContentEncodingDefault(tt.args.acceptEnc); got != tt.want {
				t.Errorf("GetContentEncodingDefault(%v) = %v, want %v", tt.args.acceptEnc, got, tt.want)
			}
		})
	}
}

func TestGetContentEncording(t *testing.T) {
	type args struct {
		acceptEnc   string
		acceptOrder []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetContentEncording(tt.args.acceptEnc, tt.args.acceptOrder); got != tt.want {
				t.Errorf("GetContentEncording() = %v, want %v", got, tt.want)
			}
		})
	}
}
