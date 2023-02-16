package utils

import (
	"os"
	"strings"

	"os/exec"
	"path/filepath"
	"runtime"
)

// EnvironmentVariable represents the name of an environment variable used to configure Photoview
type EnvironmentVariable string

// General options
const (
	EnvDevelopmentMode           EnvironmentVariable = "PHOTOVIEW_DEVELOPMENT_MODE"
	EnvServeUI                   EnvironmentVariable = "PHOTOVIEW_SERVE_UI"
	EnvUIPath                    EnvironmentVariable = "PHOTOVIEW_UI_PATH"
	EnvMediaCachePath            EnvironmentVariable = "PHOTOVIEW_MEDIA_CACHE"
	EnvFaceRecognitionModelsPath EnvironmentVariable = "PHOTOVIEW_FACE_RECOGNITION_MODELS_PATH"
)

// Network related
const (
	EnvListenIP    EnvironmentVariable = "PHOTOVIEW_LISTEN_IP"
	EnvListenPort  EnvironmentVariable = "PHOTOVIEW_LISTEN_PORT"
	EnvAPIEndpoint EnvironmentVariable = "PHOTOVIEW_API_ENDPOINT"
	EnvUIEndpoint  EnvironmentVariable = "PHOTOVIEW_UI_ENDPOINT"
)

// Database related
const (
	EnvDatabaseDriver EnvironmentVariable = "PHOTOVIEW_DATABASE_DRIVER"
	EnvMysqlURL       EnvironmentVariable = "PHOTOVIEW_MYSQL_URL"
	EnvPostgresURL    EnvironmentVariable = "PHOTOVIEW_POSTGRES_URL"
	EnvSqlitePath     EnvironmentVariable = "PHOTOVIEW_SQLITE_PATH"
)

// Feature related
const (
	EnvDisableFaceRecognition EnvironmentVariable = "PHOTOVIEW_DISABLE_FACE_RECOGNITION"
	EnvDisableVideoEncoding   EnvironmentVariable = "PHOTOVIEW_DISABLE_VIDEO_ENCODING"
	EnvDisableRawProcessing   EnvironmentVariable = "PHOTOVIEW_DISABLE_RAW_PROCESSING"
)

// GetName returns the name of the environment variable itself
func (v EnvironmentVariable) GetName() string {
	return string(v)
}

// GetValue returns the value of the environment
func (v EnvironmentVariable) GetValue() string {
	return os.Getenv(string(v))
}

// GetBool returns the environment variable as a boolean (defaults to false if not defined)
func (v EnvironmentVariable) GetBool() bool {
	value := strings.ToLower(os.Getenv(string(v)))
	trueValues := []string{"1", "true"}

	for _, x := range trueValues {
		if value == x {
			return true
		}
	}

	return false
}

// ShouldServeUI whether or not the "serve ui" option is enabled
func ShouldServeUI() bool {
	return EnvServeUI.GetBool()
}

// DevelopmentMode describes whether or not the server is running in development mode,
// and should thus print debug informations and enable other features related to developing.
func DevelopmentMode() bool {
	return EnvDevelopmentMode.GetBool()
}

// UIPath returns the value from where the static UI files are located if SERVE_UI=1
func UIPath() string {
	if path := EnvUIPath.GetValue(); path != "" {
		return path
	}
	exepath, err := GetCurrentPath()
	if err != nil {

		os.Exit(1)
	} else {
		exepath := exepath + "ui"
		return exepath
	}

	return "./ui"
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	//fmt.Println("path111:", path)
	if runtime.GOOS == "windows" {
		path = strings.Replace(path, "\\", "/", -1)
	}
	//fmt.Println("path222:", path)
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "", err
	}
	//fmt.Println("path333:", path)
	return string(path[0 : i+1]), nil
}
