package utilities

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv(logger *slog.Logger, path string) (map[string]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), 0777)
		if err != nil {
			return nil, err
		}
		envMap := make(map[string]string, 0)
		err := godotenv.Write(envMap, path)
		if err != nil {
			return nil, err
		}
		logger.Info("created .env at", "path", path)
	}

	envmap, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}

	for k, v := range envmap {
		if v == "" {
			logger.Warn("empty value in .env", "value", k)
		}
	}

	return envmap, nil
}
