package provider

import (
	"fmt"
	"os"
	"path"
	"skybin/core"
	"skybin/util"
	"time"
)

type Config struct {
	ProviderID     string `json:"providerId"`
	ApiAddr        string `json:"apiAddress"`
	MetaAddr       string `json:"metaServerAddress"`
	PrivateKeyFile string `json:"privateKeyFile"`
	PublicKeyFile  string `json:"publicKeyFile"`

	TotalStorage int64 `json:"totalStorage"`

	// Is this provider registered with metaservice?
	IsRegistered bool `json:"isRegistered"`
}

type Provider struct {
	Config    *Config
	Homedir   string
	contracts []*core.Contract
	stats     Stats
	activity  []Activity
	renters   map[string]*RenterInfo
}

// Provider node statistics
type Stats struct {
	StorageReserved int64 `json:"storageReserved"`
	StorageUsed     int64 `json:"storageUsed"`
}

type BlockInfo struct {
	BlockId string `json:"blockId"`
	Size    int64  `json:"blockSize"`
}
type RenterInfo struct {
	StorageReserved int64            `json:"storageReserved"`
	StorageUsed     int64            `json:"storageUsed"`
	Contracts       []*core.Contract `json:"contracts"`
	Blocks          []*BlockInfo     `json:"blocks"`
}

type Activity struct {
	RequestType string         `json:"requestType,omitempty"`
	BlockId     string         `json:"blockId,omitempty"`
	RenterId    string         `json:"renterId,omitempty"`
	TimeStamp   time.Time      `json:"time,omitempty"`
	Contract    *core.Contract `json:"contract,omitempty"`
}

const (
	// Max activity feed size
	maxActivity = 10
)

const (
	// Activity types
	negotiateType   = "NEGOTIATE CONTRACT"
	postBlockType   = "POST BLOCK"
	getBlockType    = "GET BLOCK"
	deleteBlockType = "DELETE BLOCK"
)

type snapshot struct {
	Contracts []*core.Contract       `json:"contracts"`
	Stats     Stats                  `json:"stats"`
	Renters   map[string]*RenterInfo `json:"renters"`
}

func (provider *Provider) saveSnapshot() error {
	s := snapshot{
		Contracts: provider.contracts,
		Stats:     provider.stats,
		Renters:   provider.renters,
	}
	return util.SaveJson(path.Join(provider.Homedir, "snapshot.json"), &s)
}

// Loads configuration and snapshot information
func LoadFromDisk(homedir string) (*Provider, error) {
	provider := &Provider{
		Homedir:   homedir,
		contracts: make([]*core.Contract, 0),
		renters:   make(map[string]*RenterInfo, 0),
	}

	config := &Config{}
	err := util.LoadJson(path.Join(homedir, "config.json"), config)
	if err != nil {
		return nil, err
	}
	provider.Config = config

	snapshotPath := path.Join(homedir, "snapshot.json")
	if _, err := os.Stat(snapshotPath); err == nil {
		var s snapshot
		err := util.LoadJson(snapshotPath, &s)
		if err != nil {
			return nil, fmt.Errorf("Unable to load snapshot. Error: %s", err)
		}

		provider.contracts = s.Contracts
		provider.stats = s.Stats
		provider.renters = s.Renters
	}

	return provider, err
}

func (provider *Provider) addActivity(activity Activity) {
	provider.activity = append(provider.activity, activity)
	if len(provider.activity) > maxActivity {

		// Drop the oldest activity.
		// O(N) but fine for small feed.
		provider.activity = provider.activity[1:]
	}
}

func (provider *Provider) negotiateContract(contract *core.Contract) (*core.Contract, error) {

	// TODO determine if contract is amiable for provider here
	// avail := provider.Config.TotalStorage - provider.stats.StorageReserved
	// if contract.StorageSpace > avail {
	// 	msg := fmt.Sprintf("Contract was not accepted.")
	// 	return nil, fmt.Errorf(msg)
	// }

	// Sign contract
	contract.ProviderSignature = "signature"

	// Add storage space to the renter
	renter, exists := provider.renters[contract.RenterId]
	if !exists {
		renter = &RenterInfo{
			Contracts: []*core.Contract{},
			Blocks:    []*BlockInfo{},
		}
		provider.renters[contract.RenterId] = renter
	}
	renter.StorageReserved += contract.StorageSpace
	renter.Contracts = append(renter.Contracts, contract)

	provider.stats.StorageReserved += contract.StorageSpace
	provider.contracts = append(provider.contracts, contract)

	activity := Activity{
		RequestType: negotiateType,
		Contract:    contract,
		TimeStamp:   time.Now(),
		RenterId:    contract.RenterId,
	}
	provider.addActivity(activity)

	provider.saveSnapshot()
	return contract, nil
}
