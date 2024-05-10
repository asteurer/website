package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadSQLFiles loads SQL queries from files in the specified directory.
func LoadSQLFiles(directory string) error {
	// Open the directory
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	// Process each file
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			err := loadSQLFile(directory + "/" + file.Name())
			if err != nil {
				return fmt.Errorf("error loading %s: %v", file.Name(), err)
			}
		}
	}
	return nil
}

// loadSQLFile reads an SQL file and stores its contents in a map.
func loadSQLFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentQueryName string
	queryMap := make(map[string]string)
	var sb strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--name:") {
			if currentQueryName != "" && sb.Len() > 0 {
				queryMap[currentQueryName] = sb.String()
				sb.Reset()
			}
			currentQueryName = strings.TrimSpace(line[len("--name:"):])
		} else {
			sb.WriteString(line + "\n")
		}
	}

	if currentQueryName != "" && sb.Len() > 0 {
		queryMap[currentQueryName] = sb.String()
	}

	sqlQueries[filePath] = queryMap
	return scanner.Err()
}

func validateToken(tokenString string) (bool, error) {
	SecretKey, exists := os.LookupEnv("TOKEN")
	if !exists {
		return false, fmt.Errorf("environment variable 'TOKEN' does not exist")
	}

	return tokenString == SecretKey, nil
}
