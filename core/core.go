package core

import (
	"time"
)

const (
	DefaultMetaAddr     = "127.0.0.1:8001"
	DefaultRenterAddr   = "127.0.0.1:8002"
	DefaultPublicProviderAddr = ":8003"
	DefaultLocalProviderAddr = "127.0.0.1:8004"
)

type ProviderInfo struct {
	ID          string `json:"id,omitempty"`
	PublicKey   string `json:"publicKey"`
	Addr        string `json:"address"`
	SpaceAvail  int64  `json:"spaceAvail,omitempty"`
	StorageRate int64  `json:"storageRate"`
}

type RenterInfo struct {
	ID        string   `json:"id"`
	Alias     string   `json:"alias"`
	PublicKey string   `json:"publicKey"`
	Files     []string `json:"files"`
	Shared    []string `json:"shared"`
}

type Contract struct {
	ID                string `json:"contractId"`
	RenterId          string `json:"renterId"`
	ProviderId        string `json:"providerId"`
	StorageSpace      int64  `json:"storageSpace"`
	RenterSignature   string `json:"renterSignature"`
	ProviderSignature string `json:"providerSignature"`
}

type BlockLocation struct {
	ProviderId string `json:"providerId"`
	Addr       string `json:"address"`
	ContractId string `json:"contractId"`
}

type Block struct {
	ID string `json:"id"`

	// Offset of the block in the file, relative to the file's other blocks.
	// For the first block, this is zero.
	Num int `json:"num"`

	// Size of the block in bytes
	Size int64 `json:"size"`

	// sha256 hash of the block
	Sha256Hash string `json:"sha256hash"`

	// Location of the provider where the block is stored
	Location BlockLocation `json:"location"`
}

// Permission provides access to a file to a non-owning user
type Permission struct {

	// The renter who this permission grants access to
	RenterId string `json:"renterId"`

	// The file's encryption information encrypted with the user's public key
	AesKey string `json:"aesKey"`
	AesIV  string `json:"aesIV"`
}

type File struct {
	ID         string       `json:"id"`
	OwnerID    string       `json:"ownerId"`
	Name       string       `json:"name"`
	IsDir      bool         `json:"isDir"`
	AccessList []Permission `json:"accessList"`
	AesKey     string       `json:"aesKey"`
	AesIV      string       `json:"aesIV"`
	Versions   []Version    `json:"versions"`
}

type Version struct {
	Num             int       `json:"num"`
	Size            int64     `json:"size"`
	ModTime         time.Time `json:"modTime"`
	UploadSize      int64     `json:"uploadSize"`
	PaddingBytes    int64     `json:"paddingBytes"`
	NumDataBlocks   int       `json:"numDataBlocks"`
	NumParityBlocks int       `json:"numParityBlocks"`
	Blocks          []Block   `json:"blocks"`
}
