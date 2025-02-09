package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"vchan.in/doctor-metrics/types"
)

var testContainerID string

func TestMain(m *testing.M) {
	// Pull the alpine image
	cmd := exec.Command("docker", "pull", "alpine")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error pulling alpine image: %v\n", err)
		os.Exit(1)
	}

	// Remove the container if it exists
	cmd = exec.Command("docker", "rm", "-f", "test-alpine-container")
	_ = cmd.Run()
	// Start a container from the alpine image
	cmd = exec.Command("docker", "run", "--name", "test-alpine-container", "-d", "alpine", "sleep", "3600")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error starting alpine container: %v\n", err)
		os.Exit(1)
	}
	testContainerID = strings.TrimSpace(string(output))

	// Run tests
	code := m.Run()

	// Stop and remove the container
	cmd = exec.Command("docker", "rm", "-f", testContainerID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error removing alpine container: %v\n", err)
	}

	os.Exit(code)
}

func TestGetRoot(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	version := "1.0.0"
	if assert.NoError(t, GetRoot(c, version)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response types.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Doctor Metrics API v"+version, response.Message)
	}
}

func TestHandleAuthMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HandleAuthMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	os.Setenv("DM_USERNAME", "user")
	os.Setenv("DM_PASSWORD", "password@123")
	defer os.Unsetenv("DM_USERNAME")
	defer os.Unsetenv("DM_PASSWORD")

	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:password@123"))
	req.Header.Set("Authorization", auth)

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func TestHandleAuthMiddlewareInvalidCredentials(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HandleAuthMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	os.Setenv("DM_USERNAME", "user")
	os.Setenv("DM_PASSWORD", "password@123")
	defer os.Unsetenv("DM_USERNAME")
	defer os.Unsetenv("DM_PASSWORD")

	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:wrongpassword"))
	req.Header.Set("Authorization", auth)

	err := handler(c)
	if assert.Error(t, err) {
		httpError, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusUnauthorized, httpError.Code)
		}
	}
}

func TestGetDockerMetrics(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock the exec.Command function to return a predefined output
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	if assert.NoError(t, GetDockerMetrics(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response types.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Container metrics retrieved successfully", response.Message)
	}
}

func TestGetDockerMetricsError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock exec.Command to simulate an error
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false") // Simulate a command that fails
	}
	defer func() { execCommand = exec.Command }() // Restore original exec.Command

	if assert.Error(t, GetDockerMetrics(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to retrieve container list")
	}
}

func TestGetMetricsContainerByName(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/containerName", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("containerName")
	c.SetParamValues("test-alpine-container")

	// Mock the exec.Command function to return a predefined output
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	if assert.NoError(t, GetMetricsContainerByName(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response types.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Container metrics retrieved successfully", response.Message)
	}
}

func TestGetMetricsContainerByNameError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/containerName", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("containerName")
	c.SetParamValues("nonexistent-container")

	// Mock the exec.Command function to simulate an error
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = exec.Command }()

	err := GetMetricsContainerByName(c)
	if assert.Error(t, err) {
		httpError, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		}
	}
}

func TestGetMetricsContainerByID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/containerID", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("containerID")
	c.SetParamValues(testContainerID)

	// Mock the exec.Command function to return a predefined output
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	if assert.NoError(t, GetMetricsContainerByID(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response types.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Container metrics retrieved successfully", response.Message)
	}
}

func TestGetMetricsContainerByIDError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/containerID", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("containerID")
	c.SetParamValues("nonexistent-container-id")

	// Mock the exec.Command function to simulate an error
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = exec.Command }()

	err := GetMetricsContainerByID(c)
	if assert.Error(t, err) {
		httpError, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		}
	}
}

// Mock exec.Command function
var execCommand = exec.Command

func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	if len(args) > 3 && args[3] == "ps" {
		fmt.Fprintf(os.Stdout, testContainerID)
	} else if len(args) > 3 && args[3] == "inspect" {
		fmt.Fprintf(os.Stdout, `[
			{
				"Id": "%s",
				"Name": "alpine",
				"State": {
					"Status": "running"
				},
				"Config": {
					"Image": "alpine"
				}
			}
		]`, testContainerID)
	} else {
		fmt.Fprintf(os.Stdout, `{"Container":"%s","Name":"alpine","CPUPerc":"0.07%%","MemUsage":"34.5MiB / 1.945GiB","MemPerc":"0.79%%","NetIO":"1.2MB / 3.4MB","BlockIO":"73.7kB / 0B","PIDs":123}`, testContainerID)
	}
	os.Exit(0)
}
