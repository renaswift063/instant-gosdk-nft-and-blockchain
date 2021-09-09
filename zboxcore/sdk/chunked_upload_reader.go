package sdk

import (
	"io"
	"math"
	"strconv"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
)

type ChunkReader interface {
	// GetChunkDataSize() int64
	Next() (*ChunkData, error)
}

// chunkReader read chunk bytes from io.Reader. see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-fixedmerkletree
type chunkReader struct {
	fileReader io.Reader

	// chunkSize chunk size with encryption header
	chunkSize int64

	// chunkDataSize data size without encryption header in a chunk. It is same as ChunkSize if EncryptOnUpload is false
	chunkDataSize int64

	// chunkDataSizePerRead total size should be read from original io.Reader. It is DataSize * DataShards.
	chunkDataSizePerRead int64

	// nextChunkIndex next index for reading
	nextChunkIndex int64

	dataShards int

	// encryptOnUpload enccrypt data on upload
	encryptOnUpload bool

	uploadMask zboxutil.Uint128
	// erasureEncoder erasuer encoder
	erasureEncoder reedsolomon.Encoder
	// encscheme encryption scheme
	encscheme encryption.EncryptionScheme
	// hash for actual file hash, content hash and challenge hash
	hasher Hasher
}

// createChunkReader create ChunkReader instance
func createChunkReader(fileReader io.Reader, chunkSize int64, dataShards int, encryptOnUpload bool, uploadMask zboxutil.Uint128, erasureEncoder reedsolomon.Encoder, encscheme encryption.EncryptionScheme, hasher Hasher) (ChunkReader, error) {

	if chunkSize <= 0 {
		return nil, errors.Throw(constants.ErrInvalidParameter, "chunkSize: "+strconv.FormatInt(chunkSize, 10))
	}

	if dataShards <= 0 {
		return nil, errors.Throw(constants.ErrInvalidParameter, "dataShards: "+strconv.Itoa(dataShards))
	}

	if erasureEncoder == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "erasureEncoder")
	}

	if hasher == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "hasher")
	}

	r := &chunkReader{
		fileReader:      fileReader,
		chunkSize:       chunkSize,
		nextChunkIndex:  0,
		dataShards:      dataShards,
		encryptOnUpload: encryptOnUpload,
		uploadMask:      uploadMask,
		erasureEncoder:  erasureEncoder,
		encscheme:       encscheme,
		hasher:          hasher,
	}

	if r.encryptOnUpload {
		r.chunkDataSize = chunkSize - 16 - 2*1024
	} else {
		r.chunkDataSize = chunkSize
	}

	r.chunkDataSizePerRead = r.chunkDataSize * int64(dataShards)

	return r, nil
}

// ChunkData data of a chunk
type ChunkData struct {
	// Index current index of chunks
	Index int64
	// IsFinal last chunk or not
	IsFinal bool

	// Size total size read from original reader (un-encoded, un-encrypted)
	Size int64
	// FragmentSize fragment size for a blobber (un-encrypted)
	FragmentSize int64
	// Fragments data shared for bloobers
	Fragments [][]byte
}

// func (r *chunkReader) GetChunkDataSize() int64 {
// 	if r == nil {
// 		return 0
// 	}
// 	return r.chunkDataSize
// }

// Next read next chunks for blobbers
func (r *chunkReader) Next() (*ChunkData, error) {

	if r == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "r")
	}

	chunk := &ChunkData{
		Index:   r.nextChunkIndex,
		IsFinal: false,

		Size:         0,
		FragmentSize: 0,
	}

	chunkBytes := make([]byte, r.chunkDataSizePerRead)
	readLen, err := r.fileReader.Read(chunkBytes)

	if err != nil {

		if !errors.Is(err, io.EOF) {
			return nil, err
		}

		//all bytes are read
		chunk.IsFinal = true
	}

	if readLen == 0 {
		chunk.IsFinal = true
		return chunk, nil
	}

	if readLen < int(r.chunkDataSizePerRead) {
		chunk.FragmentSize = int64(math.Ceil(float64(readLen) / float64(r.dataShards)))
		chunkBytes = chunkBytes[:readLen]
		chunk.IsFinal = true
	}

	chunk.Size = int64(readLen)

	err = r.hasher.WriteToFile(chunkBytes, chunk.Index)
	if err != nil {
		return chunk, err
	}

	fragments, err := r.erasureEncoder.Split(chunkBytes)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, err
	}

	err = r.erasureEncoder.Encode(fragments)
	if err != nil {
		return nil, err
	}

	var pos uint64
	if r.encryptOnUpload {
		for i := r.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := r.encscheme.Encrypt(fragments[pos])
			if err != nil {
				logger.Logger.Error("[upload] Encryption on thumbnail failed:", err.Error())
				return nil, err
			}
			header := make([]byte, 2*1024)
			copy(header[:], encMsg.MessageChecksum+","+encMsg.OverallChecksum)
			fragments[pos] = append(header, encMsg.EncryptedData...)
		}
	}

	chunk.Fragments = fragments
	r.nextChunkIndex++
	return chunk, nil
}
