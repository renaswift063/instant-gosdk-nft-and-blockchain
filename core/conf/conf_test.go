package conf

import (
	"testing"

	"github.com/0chain/gosdk/core/conf/mocks"
	"github.com/stretchr/testify/require"
)

func TestConfigLoad(t *testing.T) {

	tests := []struct {
		name        string
		exceptedErr error

		setup func(*testing.T) ConfigReader
		run   func(*require.Assertions, Config)
	}{
		{
			name:        "Test_Config_Invalid_BlockWorker",
			exceptedErr: ErrInvalidValue,
			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)

				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {

			},
		},
		{
			name: "Test_Config_BlockWorker",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("http://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal("http://127.0.0.1:9091/dns", cfg.BlockWorker)
			},
		},
		{
			name: "Test_Config_Min_Submit_Less_Than_1",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(3, cfg.MinSubmit)
			},
		},
		{
			name: "Test_Config_Min_Confirmation_Less_Than_1",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(50, cfg.MinConfirmation)
			},
		},
		{
			name: "Test_Config_Min_Confirmation_Greater_100",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(101)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(100, cfg.MinConfirmation)
			},
		}, {
			name: "Test_Config_Nax_Txn_Query_Less_Than_1",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(5, cfg.QuerySleepTime)
			},
		}, {
			name: "Test_Config_Max_Txn_Query_Less_Than_1",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(5, cfg.MaxTxnQuery)
			},
		}, {
			name: "Test_Config_Confirmation_Chain_Length_Less_Than_1",

			setup: func(t *testing.T) ConfigReader {

				reader := &mocks.ConfigReader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(3, cfg.ConfirmationChainLength)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			require := require.New(t)

			reader := tt.setup(t)

			cfg, err := Load(reader)

			// test it by predefined error variable instead of error message
			if tt.exceptedErr != nil {
				require.ErrorIs(err, tt.exceptedErr)
			} else {
				require.Equal(nil, err)
			}

			if tt.run != nil {
				tt.run(require, cfg)
			}

		})
	}
}
