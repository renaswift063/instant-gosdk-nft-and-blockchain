package conf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONReader(t *testing.T) {

	reader, _ := NewReaderFromJSON(`{
		"chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
		"signature_scheme" : "bls0chain",
		"block_worker" : "https://dev.0chain.net/dns",
		"min_submit" : -20,
		"min_confirmation" : 10,
		"confirmation_chain_length" : 0,
		"preferred_blobbers":["http://one.devnet-0chain.net:31051",
		"http://one.devnet-0chain.net:31052",
		"http://one.devnet-0chain.net:31053"
		]
}`)

	tests := []struct {
		name string
		run  func(*require.Assertions)
	}{
		{
			name: "Test_JSONReader_GetString",
			run: func(r *require.Assertions) {
				r.Equal("0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe", reader.GetString("chain_id"))
				r.Equal("bls0chain", reader.GetString("signature_scheme"))
				r.Equal("https://dev.0chain.net/dns", reader.GetString("block_worker"))
			},
		},
		{
			name: "Test_JSONReader_GetInt",

			run: func(r *require.Assertions) {
				r.Equal(-20, reader.GetInt("min_submit"))
				r.Equal(10, reader.GetInt("min_confirmation"))
				r.Equal(0, reader.GetInt("confirmation_chain_length"))
			},
		},
		{
			name: "Test_JSONReader_GetStringSlice",
			run: func(r *require.Assertions) {

				preferredBlobbers := reader.GetStringSlice("preferred_blobbers")

				r.Equal(3, len(preferredBlobbers))
				r.Equal(preferredBlobbers[0], "http://one.devnet-0chain.net:31051")
				r.Equal(preferredBlobbers[1], "http://one.devnet-0chain.net:31052")
				r.Equal(preferredBlobbers[2], "http://one.devnet-0chain.net:31053")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			require := require.New(t)

			tt.run(require)

		})
	}
}
