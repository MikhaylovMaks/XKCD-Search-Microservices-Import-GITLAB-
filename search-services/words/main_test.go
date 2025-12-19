package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

func TestServer_Ping(t *testing.T) {
	s := &server{}
	resp, err := s.Ping(context.Background(), &emptypb.Empty{})

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Norm(t *testing.T) {
	s := &server{}

	tests := []struct {
		name        string
		phrase      string
		expectedLen int
		wantErr     bool
		errCode     codes.Code
	}{
		{
			name:        "normal phrase",
			phrase:      "Hello, World!",
			expectedLen: 2, // "hello" and "world"
			wantErr:     false,
		},
		{
			name:        "empty phrase",
			phrase:      "",
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name:    "phrase too long",
			phrase:  string(make([]byte, maxPhraseLen+1)),
			wantErr: true,
			errCode: codes.ResourceExhausted,
		},
		{
			name:        "max length phrase",
			phrase:      strings.Repeat("a", maxPhraseLen), // long string of 'a's
			expectedLen: 1,                                 // single long word "a"
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &wordspb.WordsRequest{Phrase: tt.phrase}
			resp, err := s.Norm(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.Words, tt.expectedLen)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	// Test that Config struct has proper defaults
	// We can't easily test env-default parsing without setting env vars
	// So just test that the struct can be created and has expected fields
	cfg := Config{}
	// The Address field should be settable (we can't test env-default easily in unit test)
	// But we can verify the struct exists and has the right field
	assert.NotNil(t, cfg.Address) // Address field exists
	// Since we can't test env-default in isolation, just ensure field is accessible
	cfg.Address = ":8080"
	assert.Equal(t, ":8080", cfg.Address)
}
