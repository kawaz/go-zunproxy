package middleware

import (
	"reflect"
	"testing"
)

func TestNewWildCard(t *testing.T) {
	type args struct {
		p  string
		ss []string
	}
	tests := []struct {
		name     string
		args     args
		want     bool
		wantType reflect.Type
	}{
		{
			name:     "eq match",
			args:     args{"/foo", []string{"/foo"}},
			want:     true,
			wantType: reflect.TypeOf(eqPattern{}),
		},
		{
			name:     "eq match fail",
			args:     args{"/foo", []string{"/", "/foo/", "/fooo", "/foo*", ""}},
			want:     false,
			wantType: reflect.TypeOf(eqPattern{}),
		},
		{
			name:     "any match",
			args:     args{"*", []string{"", "/", "hoge", "***", "💩"}},
			want:     true,
			wantType: reflect.TypeOf(anyPattern),
		},
		{
			name:     "prefix match",
			args:     args{"/foo*", []string{"/foo", "/foo/", "/foo/bar", "/foooooo"}},
			want:     true,
			wantType: reflect.TypeOf(prefixPattern{}),
		},
		{
			name:     "prefix match fail",
			args:     args{"/foo*", []string{"/", "/fo", "/aaa/foo", "/aaa/foo/", "/aaa/foo/bar", ""}},
			want:     false,
			wantType: reflect.TypeOf(prefixPattern{}),
		},
		{
			name:     "suffix match",
			args:     args{"*.go", []string{".go", "foo.go", "/foo.go", "/foo/bar.go"}},
			want:     true,
			wantType: reflect.TypeOf(suffixPattern{}),
		},
		{
			name:     "suffix match fail",
			args:     args{"*.go", []string{".go.jpg", "Xgo", ""}},
			want:     false,
			wantType: reflect.TypeOf(suffixPattern{}),
		},
		{
			name:     "prefix and suffix match",
			args:     args{"/images/*.jpg", []string{"/images/.jpg", "/images/foo.jpg", "/images/thumb/foo.jpg"}},
			want:     true,
			wantType: reflect.TypeOf(presufPattern{}),
		},
		{
			name:     "prefix and suffix match fail",
			args:     args{"/images/*.jpg", []string{"/images/foo.png", "/user/images/foo.jpg", ""}},
			want:     false,
			wantType: reflect.TypeOf(presufPattern{}),
		},
		{
			name:     "multiple wildcard match",
			args:     args{"*/img/*.jpg", []string{"/img/foo.jpg", "/img/x/foo.jpg", "/user/img/1.jpg", "/user/img/1/2.jpg"}},
			want:     true,
			wantType: reflect.TypeOf(regexPattern{}),
		},
		{
			name:     "multiple wildcard match fail",
			args:     args{"*/img/*.jpg", []string{"/img/foo.png", "/images/foo.jpg", "/user/img.jpg"}},
			want:     false,
			wantType: reflect.TypeOf(regexPattern{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewWildCard(tt.args.p)
			gotType := reflect.TypeOf(pattern)
			if gotType != tt.wantType {
				t.Errorf("NewWildCard(%v) type = %v, wantType %v", tt.args.p, gotType, tt.wantType)
				return
			}
			for _, s := range tt.args.ss {
				got := pattern.Match(s)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewWildCard(%v).Match(%v) = %v, want %v, pattern = %#v", tt.args.p, s, got, tt.want, pattern)
					return
				}
			}
		})
	}
}
