package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

func main() {
	log.Infof("Starting script...")

	gradleFile := os.Getenv("gradle_file_path")
	if gradleFile == "" {
		log.Errorf("Gradle file path is not set")
		os.Exit(1)
	}

	versionName, err := extractVersionName(gradleFile)
	if err != nil {
		log.Errorf("Failed to extract versionName: %v", err)
		os.Exit(1)
	}
	log.Infof("Extracted versionName: '%s'", versionName)

	if versionName == "" {
		log.Errorf("Version name is empty")
		os.Exit(1)
	}

	versionNameBase := cleanVersionName(versionName)
	log.Infof("Base versionName: '%s'", versionNameBase)

	tag := fmt.Sprintf("v%s", versionNameBase)
	tagMessage := fmt.Sprintf("Release %s", tag)
	log.Infof("Tag: '%s'", tag)
	log.Infof("Tag message: '%s'", tagMessage)

	if !isValidTagName(tag) {
		log.Errorf("Invalid tag name: '%s' contains spaces or invalid characters", tag)
		os.Exit(1)
	}

	if err := createTag(tag, tagMessage); err != nil {
		log.Errorf("Failed to create tag: %v", err)
		os.Exit(1)
	}

	if err := printLocalTags(); err != nil {
		log.Errorf("Failed to list local tags: %v", err)
		os.Exit(1)
	}

	if err := pushTags(); err != nil {
		log.Errorf("Failed to push tags: %v", err)
		os.Exit(1)
	}

	if err := printLocalTags(); err != nil {
		log.Errorf("Failed to list local tags: %v", err)
		os.Exit(1)
	}

	if err := setEnvVariable("NEW_TAG_NAME", tag); err != nil {
		log.Errorf("Failed to set environment variable: %v", err)
		os.Exit(1)
	}
}

func extractVersionName(gradleFile string) (string, error) {
	file, err := os.Open(gradleFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`versionName\s*=\s*"([^"]+)"`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}
	return "", errors.New("versionName not found")
}

func cleanVersionName(versionName string) string {
	versionName = strings.TrimSpace(versionName)
	versionName = strings.ReplaceAll(versionName, "-DEBUG", "")
	return versionName
}

func isValidTagName(tag string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return re.MatchString(tag)
}

func createTag(tag, tagMessage string) error {
	cmd := exec.Command("git", "tag", "-a", tag, "-m", tagMessage, "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printLocalTags() error {
	cmd := exec.Command("git", "tag")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pushTags() error {
	cmd := exec.Command("git", "push", "--tags", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setEnvVariable(key, value string) error {
	return env.NewCommand("envman", "add", "--key", key, "--value", value).Run()
}
