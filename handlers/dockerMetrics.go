package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"vchan.in/doctor-metrics/types"
)

func getMetrics(containerID string) (types.ContainerMetrics, error) {
	/*
		Get container metrics.

		Function input is a container ID like "f3f177b2b3b4".
		Function returns a ContainerMetrics struct with container metrics.
	*/
	var metrics types.ContainerMetrics
	metrics.Timestamp = time.Now().UTC().Format(time.RFC3339)
	metrics.ContainerID = containerID

	// Get container name using docker inspect (remove leading "/" if present)
	containerName, err := exec.Command("docker", "inspect", "--format={{.Name}}", containerID).Output()
	if err == nil {
		name := strings.TrimSpace(string(containerName))
		name = strings.TrimPrefix(name, "/")
		metrics.ContainerName = name
	} else {
		metrics.ContainerName = "N/A"
	}

	// Use docker stats to get container metrics in JSON format.
	statsOutput, err := exec.Command("docker", "stats", containerID, "--no-stream", "--format", "{{json .}}").Output()
	if err != nil {
		return metrics, err
	}

	var ds types.DockerStats
	if err := json.Unmarshal(statsOutput, &ds); err != nil {
		return metrics, err
	}

	// Parse CPU percentage (e.g. "0.07%")
	cpuStr := strings.TrimSuffix(ds.CPUPerc, "%")
	cpuUsage, _ := strconv.ParseFloat(strings.TrimSpace(cpuStr), 64)
	metrics.ContainerCpuUsagePercent = cpuUsage

	// Parse memory usage and limit.
	memParts := strings.Split(ds.MemUsage, "/")
	if len(memParts) == 2 {
		usedStr := strings.TrimSpace(memParts[0])
		limitStr := strings.TrimSpace(memParts[1])
		usedBytes, _ := convertToBytes(usedStr)
		limitBytes, _ := convertToBytes(limitStr)
		metrics.ContainerMemoryUsageBytes = usedBytes
		metrics.ContainerMemoryLimitBytes = limitBytes
	}

	// Parse memory percentage (e.g. "0.79%")
	memPercStr := strings.TrimSuffix(ds.MemPerc, "%")
	memPerc, _ := strconv.ParseFloat(strings.TrimSpace(memPercStr), 64)
	metrics.ContainerMemoryUsagePercent = memPerc

	// Parse network I/O.
	netParts := strings.Split(ds.NetIO, "/")
	if len(netParts) == 2 {
		rxBytes, _ := convertToBytes(strings.TrimSpace(netParts[0]))
		txBytes, _ := convertToBytes(strings.TrimSpace(netParts[1]))
		metrics.ContainerNetworkReceiveBytesTotal = rxBytes
		metrics.ContainerNetworkTransmitBytesTotal = txBytes
	}

	// Parse block I/O.
	blockParts := strings.Split(ds.BlockIO, "/")
	if len(blockParts) == 2 {
		readBytes, _ := convertToBytes(strings.TrimSpace(blockParts[0]))
		writeBytes, _ := convertToBytes(strings.TrimSpace(blockParts[1]))
		metrics.ContainerBlockReadBytes = readBytes
		metrics.ContainerBlockWriteBytes = writeBytes
	}

	// Parse PIDs.
	pids, _ := strconv.Atoi(ds.PIDs)
	metrics.ContainerPIDs = pids

	return metrics, nil
}

func convertToBytes(s string) (int64, error) {
	// Convert a value like "1.2MB" or "3.4KB" to bytes.
	s = strings.ToUpper(strings.TrimSpace(s))
	if strings.HasSuffix(s, "KB") {
		valueStr := strings.TrimSuffix(s, "KB")
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
		return int64(val * 1024), err
	} else if strings.HasSuffix(s, "MB") {
		valueStr := strings.TrimSuffix(s, "MB")
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
		return int64(val * 1024 * 1024), err
	} else if strings.HasSuffix(s, "GB") {
		valueStr := strings.TrimSuffix(s, "GB")
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
		return int64(val * 1024 * 1024 * 1024), err
	} else if strings.HasSuffix(s, "MIB") {
		valueStr := strings.TrimSuffix(s, "MIB")
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
		return int64(val * 1024 * 1024), err
	} else if strings.HasSuffix(s, "GIB") {
		valueStr := strings.TrimSuffix(s, "GIB")
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
		return int64(val * 1024 * 1024 * 1024), err
	}
	return 0, fmt.Errorf("unknown byte unit in %s", s)
}

func GetDockerMetrics(c echo.Context) error {
	/*
		Get metrics for all containers.

		[
			{
			"Container": "f3f177b2b3b4",
			"Name": "my-container",
			"CPUPerc": "0.07%",
			"MemUsage": "34.5MiB / 1.945GiB",
			"MemPerc": "0.79%",
			"NetIO": "1.2MB / 3.4MB",
			"BlockIO": "73.7kB / 0B",
			"PIDs": 123
			},
			...
		]

		Function returns a JSON response with the container metrics.
	*/
	listMetrics := []types.ContainerMetrics{}

	// List container IDs using docker ps
	containerIDsBytes, err := exec.Command("docker", "ps", "-q").Output()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve container list")
	}
	containerIDs := strings.Fields(string(containerIDsBytes))

	var wg sync.WaitGroup
	metricsChan := make(chan types.ContainerMetrics, len(containerIDs))
	errorChan := make(chan error, len(containerIDs))

	// Limit the number of concurrent goroutines
	concurrencyLimit := 10
	sem := make(chan struct{}, concurrencyLimit)

	for _, containerID := range containerIDs {
		wg.Add(1)
		go func(containerID string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			metrics, err := getMetrics(containerID)
			if err != nil {
				errorChan <- err
				return
			}
			metricsChan <- metrics
		}(containerID)
	}

	wg.Wait()
	close(metricsChan)
	close(errorChan)

	if len(errorChan) > 0 {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve container metrics")
	}

	for metrics := range metricsChan {
		listMetrics = append(listMetrics, metrics)
	}
	response := types.APIResponse{
		Status:  "success",
		Message: "Container metrics retrieved successfully",
		Data: struct {
			ContainerMetrics []types.ContainerMetrics `json:"container_metrics"`
		}{
			ContainerMetrics: listMetrics,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func GetMetricsContainerByName(c echo.Context) error {
	/*
		Get metrics for a specific container.

		{
		  "Container": "f3f177b2b3b4",
		  "Name": "my-container",
		  "CPUPerc": "0.07%",
		  "MemUsage": "34.5MiB / 1.945GiB",
		  "MemPerc": "0.79%",
		  "NetIO": "1.2MB / 3.4MB",
		  "BlockIO": "73.7kB / 0B",
		  "PIDs": 123
		}

		Function takes a container name as input.
		Function returns a JSON response with the container metrics.
	*/

	containerName := c.Param("containerName")

	// Get container ID using docker ps
	containerIDBytes, err := exec.Command("docker", "ps", "--filter", "name="+containerName, "--format", "{{.ID}}").Output()
	if err != nil {
		return err
	}
	containerID := strings.TrimSpace(string(containerIDBytes))

	metrics, err := getMetrics(containerID)
	if err != nil {
		return err
	}

	response := types.APIResponse{
		Status:  "success",
		Message: "Container metrics retrieved successfully",
		Data: struct {
			ContainerMetrics []types.ContainerMetrics `json:"container_metrics"`
		}{
			ContainerMetrics: []types.ContainerMetrics{metrics},
		},
	}

	return c.JSON(http.StatusOK, response)
}

func GetMetricsContainerByID(c echo.Context) error {
	/*
		Get metrics for a specific container.

		{
		  "Container": "f3f177b2b3b4",
		  "Name": "my-container",
		  "CPUPerc": "0.07%",
		  "MemUsage": "34.5MiB / 1.945GiB",
		  "MemPerc": "0.79%",
		  "NetIO": "1.2MB / 3.4MB",
		  "BlockIO": "73.7kB / 0B",
		  "PIDs": 123
		}

		Function takes a container ID as input.
		Function returns a JSON response with the container metrics.
	*/

	containerID := c.Param("containerID")

	metrics, err := getMetrics(containerID)
	if err != nil {
		return err
	}

	response := types.APIResponse{
		Status:  "success",
		Message: "Container metrics retrieved successfully",
		Data: struct {
			ContainerMetrics []types.ContainerMetrics `json:"container_metrics"`
		}{
			ContainerMetrics: []types.ContainerMetrics{metrics},
		},
	}

	return c.JSON(http.StatusOK, response)
}
