package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
		split := strings.Split(line, "=")
		os.Setenv(split[0], split[1])
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
