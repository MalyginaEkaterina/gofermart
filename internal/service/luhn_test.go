package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckNumberByLuhn(t *testing.T) {
	tests := []struct {
		name     string
		inNumber string
		want     bool
	}{
		{name: "positive for even", inNumber: "4561261212345467", want: true},
		{name: "positive for odd", inNumber: "454665453253412", want: true},
		{name: "negative for even", inNumber: "4561261212345464", want: false},
		{name: "negative for odd", inNumber: "454665453253415", want: false},
		{name: "non-integer", inNumber: "454665453253412a", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CheckNumberByLuhn(tt.inNumber))
		})
	}
}
