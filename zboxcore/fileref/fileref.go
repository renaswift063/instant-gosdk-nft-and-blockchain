package fileref

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/encryption"
)

const CHUNK_SIZE = 64 * 1024

const (
	FILE      = "f"
	DIRECTORY = "d"
)

type FileRef struct {
	Ref                 `json:",squash"`
	CustomMeta          string `json:"custom_meta"`
	ContentHash         string `json:"content_hash"`
	Size                int64  `json:"size"`
	MerkleRoot          string `json:"merkle_root"`
	ActualFileSize      int64  `json:"actual_file_size"`
	ActualFileHash      string `json:"actual_file_hash"`
	ActualThubnailSize  int64  `json:"actual_thumnail_size"`
	ActualThumbnailHash string `json:"actual_thumbnail_hash"`
	MimeType            string `json:"mimetype"`
}

type RefEntity interface {
	GetNumBlocks() int64
	GetHash() string
	CalculateHash() string
	GetType() string
	GetPathHash() string
	GetPath() string
	GetName() string
}

type Ref struct {
	Type         string      `json:"type"`
	AllocationID string      `json:"allocation_id"`
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	Hash         string      `json:"hash"`
	NumBlocks    int64       `json:"num_of_blocks"`
	PathHash     string      `json:"path_hash"`
	Children     []RefEntity `json:"-"`
}

func GetReferenceLookup(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

func (r *Ref) CalculateHash() string {
	if len(r.Children) == 0 {
		return r.Hash
	}
	for _, childRef := range r.Children {
		childRef.CalculateHash()
	}
	childHashes := make([]string, len(r.Children))
	childPathHashes := make([]string, len(r.Children))
	var refNumBlocks int64
	for index, childRef := range r.Children {
		childHashes[index] = childRef.GetHash()
		childPathHashes[index] = childRef.GetPathHash()
		refNumBlocks += childRef.GetNumBlocks()
	}
	//fmt.Println("ref name and path, hash :" + r.Name + " " + r.Path + " " + r.Hash)
	//fmt.Println("ref hash data: " + strings.Join(childHashes, ":"))
	r.Hash = encryption.Hash(strings.Join(childHashes, ":"))
	//fmt.Println("ref hash : " + r.Hash)
	r.NumBlocks = refNumBlocks
	//fmt.Println("Ref Path hash: " + strings.Join(childPathHashes, ":"))
	r.PathHash = encryption.Hash(strings.Join(childPathHashes, ":"))
	return r.Hash
}

func (r *Ref) GetHash() string {
	return r.Hash
}

func (r *Ref) GetType() string {
	return r.Type
}

func (r *Ref) GetNumBlocks() int64 {
	return r.NumBlocks
}

func (r *Ref) GetPathHash() string {
	return r.PathHash
}

func (r *Ref) GetPath() string {
	return r.Path
}

func (r *Ref) GetName() string {
	return r.Name
}

func (r *Ref) AddChild(child RefEntity) {
	if r.Children == nil {
		r.Children = make([]RefEntity, 0)
	}
	r.Children = append(r.Children, child)
	sort.SliceStable(r.Children, func(i, j int) bool {
		return strings.Compare(GetReferenceLookup(r.AllocationID, r.Children[i].GetPath()), GetReferenceLookup(r.AllocationID, r.Children[j].GetPath())) == -1
	})

}

func (r *Ref) RemoveChild(idx int) {
	if idx < 0 {
		return
	}
	r.Children = append(r.Children[:idx], r.Children[idx+1:]...)
	sort.SliceStable(r.Children, func(i, j int) bool {
		return strings.Compare(GetReferenceLookup(r.AllocationID, r.Children[i].GetPath()), GetReferenceLookup(r.AllocationID, r.Children[j].GetPath())) == -1
	})
}

func (fr *FileRef) GetHashData() string {
	hashArray := make([]string, 0)
	hashArray = append(hashArray, fr.AllocationID)
	hashArray = append(hashArray, fr.Type)
	hashArray = append(hashArray, fr.Name)
	hashArray = append(hashArray, fr.Path)
	hashArray = append(hashArray, strconv.FormatInt(fr.Size, 10))
	hashArray = append(hashArray, fr.ContentHash)
	hashArray = append(hashArray, fr.MerkleRoot)
	hashArray = append(hashArray, strconv.FormatInt(fr.ActualFileSize, 10))
	hashArray = append(hashArray, fr.ActualFileHash)
	return strings.Join(hashArray, ":")
}

func (fr *FileRef) GetHash() string {
	return fr.Hash
}

func (fr *FileRef) CalculateHash() string {
	//fmt.Println("fileref name , path, hash", fr.Name, fr.Path, fr.Hash)
	//fmt.Println("Fileref hash data: " + fr.GetHashData())
	fr.Hash = encryption.Hash(fr.GetHashData())
	//fmt.Println("Fileref hash : " + fr.Hash)
	fr.NumBlocks = int64(math.Ceil(float64(fr.Size*1.0) / CHUNK_SIZE))
	fr.PathHash = GetReferenceLookup(fr.AllocationID, fr.Path)
	return fr.Hash
}

func (fr *FileRef) GetType() string {
	return fr.Type
}

func (fr *FileRef) GetNumBlocks() int64 {
	return fr.NumBlocks
}

func (fr *FileRef) GetPathHash() string {
	return fr.PathHash
}

func (fr *FileRef) GetPath() string {
	return fr.Path
}
func (fr *FileRef) GetName() string {
	return fr.Name
}
