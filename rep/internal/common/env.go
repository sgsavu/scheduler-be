package common

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadEnvVars() {
	file, err := os.Open(".env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		split := strings.Split(line, "=")
		os.Setenv(split[0], split[1])
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}

func getPeriodicPurgeInterval() time.Duration {
	periodicPurge := os.Getenv("PERIODIC_PURGE_INTERVAL")
	Atoi, err := strconv.Atoi(periodicPurge)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return time.Duration(Atoi) * time.Millisecond
}

func getMaxUploadSize() int64 {
	periodicPurge := os.Getenv("MAX_UPLOAD_SIZE")
	Atoi, err := strconv.ParseInt(periodicPurge, 10, 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return Atoi
}
