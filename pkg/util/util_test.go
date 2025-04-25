package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_firstGoProxy(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{
				s: "https://proxy.golang.org",
			},
			want: "https://proxy.golang.org",
		},
		{
			name: "2",
			args: args{
				s: "https://proxy.golang.org,direct",
			},
			want: "https://proxy.golang.org",
		},
		{
			name: "3",
			args: args{
				s: "https://proxy.golang.org|direct",
			},
			want: "https://proxy.golang.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, firstGoProxy(tt.args.s))
		})
	}
}
