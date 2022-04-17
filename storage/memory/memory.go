package memory

import (
	"fmt"

	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/storage"
)

type Memory struct{}

func New(_ config.Config) (Memory, error) {
	return Memory{}, nil
}

func (m Memory) SaveData(tmpTopDataArray []storage.TopData) error {
	for _, td := range tmpTopDataArray {
		fmt.Println(td)
	}
	return nil
}

func (m Memory) Populate(lastID int) ([]storage.Site, error) {
	// returns data only for the fist call
	if lastID == 0 {
		return []storage.Site{
			{Hosts: 0, Hits: 0, ID: 1, CounterID: 1, Digits: true},
			{Hosts: 0, Hits: 0, ID: 2, CounterID: 2, Digits: true},
		}, nil

	}
	return []storage.Site{}, nil
}

func (m Memory) UpdateSites(_ []storage.Site) error {
	return nil
}

func (m *Memory) Close() error {
	return nil
}
