package types

// ContainerMetrics struct to store container metrics.
type ContainerMetrics struct {
	Active                             bool    `json:"active"`                                 // Status of the API response e.g. "success"
	ContainerID                        string  `json:"container_id"`                           // Container ID e.g. "f3f177b2b3b4"
	ContainerName                      string  `json:"container_name"`                         // Container name e.g. "my-container"
	Timestamp                          string  `json:"timestamp"`                              // Timestamp in RFC3339 format e.g. "2021-09-01T12:34:56Z"
	ContainerCpuUsagePercent           float64 `json:"container_cpu_usage_percent"`            // CPU usage percentage e.g. 0.07
	ContainerMemoryUsageBytes          int64   `json:"container_memory_usage_bytes"`           // Memory usage in bytes e.g. 123456
	ContainerMemoryLimitBytes          int64   `json:"container_memory_limit_bytes"`           // Memory limit in bytes e.g. 123456
	ContainerMemoryUsagePercent        float64 `json:"container_memory_usage_percent"`         // Memory usage percentage e.g. 0.79
	ContainerNetworkReceiveBytesTotal  int64   `json:"container_network_receive_bytes_total"`  // Network receive bytes e.g. 123456
	ContainerNetworkTransmitBytesTotal int64   `json:"container_network_transmit_bytes_total"` // Network transmit bytes e.g. 123456
	ContainerBlockReadBytes            int64   `json:"container_block_read_bytes"`             // Block read bytes e.g. 123456
	ContainerBlockWriteBytes           int64   `json:"container_block_write_bytes"`            // Block write bytes e.g. 123456
	ContainerPIDs                      int     `json:"container_pids"`                         // Number of PIDs e.g. 123
}

// Temporary struct to unmarshal docker stats output.
type DockerStats struct {
	Container string `json:"Container"` // Container ID
	CPUPerc   string `json:"CPUPerc"`   // Format: "0.07%"
	MemUsage  string `json:"MemUsage"`  // Format: "used / total" (e.g. "34.5MiB / 1.945GiB")
	MemPerc   string `json:"MemPerc"`   // Format: "0.79%"
	NetIO     string `json:"NetIO"`     // Format: "rx / tx" (e.g. "1.2MB / 3.4MB")
	BlockIO   string `json:"BlockIO"`   // Format: "read / write" (e.g. "73.7kB / 0B")
	PIDs      string `json:"PIDs"`      // Number of PIDs
}

// APIResponse struct to store API response.
type APIResponse struct {
	Status  string `json:"status"`  // Status of the API response e.g. "success"
	Message string `json:"message"` // Message of the API response e.g. "Container metrics retrieved successfully"
	Data    struct {
		ContainerMetrics []ContainerMetrics `json:"container_metrics"` // List of container metrics
	} `json:"data"` // Data of the API response
}
