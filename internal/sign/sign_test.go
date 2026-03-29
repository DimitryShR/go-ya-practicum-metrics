package sign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_signer_Sign(t *testing.T) {
	tests := []struct {
		name string
		key  string
		src  []byte
		want string
	}{
		{
			name: "Success sign data",
			key:  "testkey",
			src:  []byte("testdata"),
			want: "220afe7c01cca398fff2fc2c3687be94ded74f1b853db65707bf8440055217b0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSigner(tt.key)
			got := s.Sign(tt.src)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_signer_Verify(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		sign    string
		src     []byte
		want    bool
		wantErr bool
	}{
		{
			name: "Success verify",
			key:  "testkey",
			sign: "220afe7c01cca398fff2fc2c3687be94ded74f1b853db65707bf8440055217b0",
			src:  []byte("testdata"),
			want: true,
		},
		{
			name:    "Failed verify",
			key:     "testkey",
			sign:    "220afe7c01cca398fff2fc2c3687be94ded74f1b853db65707bf8440055217b1",
			src:     []byte("testdata"),
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid hex sign",
			key:     "testkey",
			sign:    "not-hex",
			src:     []byte("testdata"),
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSigner(tt.key)
			got, gotErr := s.Verify(tt.sign, tt.src)

			if assert.Equal(t, tt.wantErr, gotErr != nil) && gotErr != nil {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_signer_VerifySignedPayload(t *testing.T) {
	s := NewSigner("testkey")
	src := []byte("testdata")
	sign := s.Sign(src)

	got, err := s.Verify(sign, src)

	assert.NoError(t, err)
	assert.True(t, got)
}
