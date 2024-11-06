package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NodeInfo struct {
	Software struct {
		Name string `json:"name"`
	} `json:"software"`
}

func GetNodeInfo(instance string) (*NodeInfo, error) {
	endpoint := fmt.Sprintf("https://%s/nodeinfo/2.0", instance)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get nodeinfo: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeinfo NodeInfo
	if err := json.Unmarshal(body, &nodeinfo); err != nil {
		return nil, err
	}

	return &nodeinfo, nil
}
