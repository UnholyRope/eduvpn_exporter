package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJson(t *testing.T) {
	output := `
[
  {
      "profile_id": "internet",
      "active_connection_count": 1,
      "max_connection_count": 4093,
      "percentage_in_use": 0,
      "openvpn_max_connection_count": 0,
      "openvpn_active_connection_count": 0,
      "wireguard_max_connection_count": 4093,
      "wireguard_active_connection_count": 1,
      "wireguard_allocated_ip_count": 1,
      "wireguard_free_ip_count": 4092,
      "wireguard_percentage_allocated": 0,
      "connection_list": [
          {
              "user_id": "test",
              "ip_list": [
                  "172.27.0.2",
                  "fd12:3456:789a:1::2"
              ],
              "vpn_proto": "wireguard"
          }
      ]
  },
  {
      "profile_id": "employees",
      "active_connection_count": 0,
      "max_connection_count": 4093,
      "percentage_in_use": 0,
      "openvpn_max_connection_count": 0,
      "openvpn_active_connection_count": 0,
      "wireguard_max_connection_count": 4093,
      "wireguard_active_connection_count": 0,
      "wireguard_allocated_ip_count": 1,
      "wireguard_free_ip_count": 4092,
      "wireguard_percentage_allocated": 0,
      "connection_list": []
  }
]`

	expectedStatus := []EduVPNStatus{
		{
			ProfileID:                      "internet",
			ActiveConnectionCount:          1,
			MaxConnectionCount:             4093,
			OpenVPNMaxConnectionCount:      0,
			OpenVPNActiveConnectionCount:   0,
			WireGuardMaxConnectionCount:    4093,
			WireGuardActiveConnectionCount: 1,
			WireGuardAllocatedIPCount:      1,
			WireGuardFreeIPCount:           4092,
			WireGuardPercentageAllocated:   0,
			ConnectionList: []ConnectionList{
				{
					UserID: "test",
					IPList: []string{
						"172.27.0.2",
						"fd12:3456:789a:1::2",
					},
					VPNProto: "wireguard",
				},
			},
		},
		{
			ProfileID:                      "employees",
			ActiveConnectionCount:          0,
			MaxConnectionCount:             4093,
			OpenVPNMaxConnectionCount:      0,
			OpenVPNActiveConnectionCount:   0,
			WireGuardMaxConnectionCount:    4093,
			WireGuardActiveConnectionCount: 0,
			WireGuardAllocatedIPCount:      1,
			WireGuardFreeIPCount:           4092,
			WireGuardPercentageAllocated:   0,
			ConnectionList:                 []ConnectionList{},
		},
	}

	status, err := parseJson([]byte(output))

	if assert.NoError(t, err) {
		assert.Len(t, status, 2)
		assert.Equal(t, status, expectedStatus)
	}
}
