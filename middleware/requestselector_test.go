package middleware

import (
	"reflect"
	"testing"
)

func TestNewPattern(t *testing.T) {
	type args struct {
		p string
		v []string
	}
	tests := []struct {
		name     string
		args     args
		want     bool
		wantType reflect.Type
		wantErr  bool
	}{
		{
			name: "full match",
			args: args{
				"/",
				[]string{"/"},
			},
			want:     true,
			wantType: reflect.TypeOf(eqPattern{}),
			wantErr:  false,
		},
		{
			name: "any match",
			args: args{
				"*",
				[]string{"", "/", "hoge", "***", "💩"},
			},
			want:     true,
			wantType: reflect.TypeOf(anyPattern),
			wantErr:  false,
		},
		{
			name: "prefix match",
			args: args{
				"/foo*",
				[]string{"/foo", "/foo/dsa", "/foooooo", "/foo*"},
			},
			want:     true,
			wantType: reflect.TypeOf(prefixPattern{}),
			wantErr:  false,
		},
		{
			name: "suffix match",
			args: args{
				"*.jpg",
				[]string{"/foo.jpg", "a.jpg", "b.jpg.jpg", ".jpg"},
			},
			want:     true,
			wantType: reflect.TypeOf(suffixPattern{}),
			wantErr:  false,
		},
		{
			name: "prefix/suffix match",
			args: args{
				"/images/*.jpg",
				[]string{"/images/foo.jpg", "/images/thumb/a.jpg", "/images/dsa.jpg"},
			},
			want:     true,
			wantType: reflect.TypeOf(presufPattern{}),
			wantErr:  false,
		},
		{
			name: "not support multiple wildcard",
			args: args{
				"*/files/*.jpg",
				[]string{"/user1/files/foo.jpg"},
			},
			want:     false,
			wantType: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := NewPattern(tt.args.p)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewPattern(%v) error = %v, wantErr %v", tt.args.p, err, tt.wantErr)
				}
				return
			}
			gotType := reflect.TypeOf(pattern)
			if gotType != tt.wantType {
				t.Errorf("NewPattern(%v) type = %v, wantType %v", tt.args.p, gotType, tt.wantType)
				return
			}
			for _, v := range tt.args.v {
				got := pattern.Match(v)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewPattern(%v).Match(%v) = %v, want %v, pattern:%#v", tt.args.p, v, got, tt.want, pattern)
					return
				}
			}
		})
	}
}
