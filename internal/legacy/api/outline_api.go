package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateOutlineAPI is a default OutlineAPI factory
func CreateOutlineAPI(l core.Logger, url string) *OutlineAPI {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return &OutlineAPI{l, url, client}
}

type OutlineAPI struct {
	l      core.Logger
	url    string
	client *http.Client
}

type OutlineAPIKeys struct {
	AccessKeys []*VpnKey
}

type VpnKey struct {
	ID        string
	Name      string
	Password  string
	Port      int
	Method    string
	AccessURL string
}

func (api *OutlineAPI) GetKeys(chatID int64) ([]*VpnKey, error) {
	res, err := api.client.Get(api.url + "/access-keys")
	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	var keys OutlineAPIKeys
	body, _ := io.ReadAll(res.Body)

	err = json.Unmarshal(body, &keys)
	if err != nil {
		return nil, err
	}

	return keys.AccessKeys, nil
}

func (api *OutlineAPI) CreateKey(chatID int64, name string) (*VpnKey, error) {
	res, err := api.client.Post(api.url+"/access-keys", "application/json", bytes.NewBuffer([]byte{}))

	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	var key VpnKey
	body, _ := io.ReadAll(res.Body)

	err = json.Unmarshal(body, &key)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	values := map[string]string{"name": name}
	data, err := json.Marshal(values)

	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, api.url+"/access-keys/"+key.ID+"/name", bytes.NewBuffer(data))
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res, err = api.client.Do(req)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	if res.StatusCode != 204 {
		api.l.Warn("unexpected response", "response", res)
		return nil, fmt.Errorf("can't rename created key")
	}

	return &key, nil
}

func (api *OutlineAPI) DeleteKey(key *legacy.VpnKey) error {
	req, err := http.NewRequest(http.MethodDelete, api.url+"/access-keys/"+key.ID, bytes.NewBuffer([]byte{}))
	if err != nil {
		api.l.Error(err)
		return err
	}

	res, err := api.client.Do(req)
	if err != nil {
		api.l.Error(err)
		return err
	}

	if res.StatusCode != 204 {
		api.l.Warn("unexpected response", "response", res)
		return fmt.Errorf("can't remove key")
	}

	return nil
}
