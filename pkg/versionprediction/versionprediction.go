package versionprediction

import (
	"strconv"
	"strings"
)

type version struct {
	major int
	minor int
	patch int
}

const (
	dot = "."
)

func GetVersionPredictions(versionString string) []string {
	predictions := []string{}
	versionString = strings.TrimSpace(versionString)
	if versionString == "" {
		return predictions
	}
	numberStrings := strings.Split(versionString, dot)
	if len(numberStrings) > 3 || len(numberStrings) < 1 {
		return predictions
	}
	numbers := []int{}
	for _, numberString := range numberStrings {
		number, err := strconv.Atoi(numberString)
		if err != nil {
			return predictions
		}
		numbers = append(numbers, number)
	}
	currentVersion := &version{
		major: numbers[0],
	}
	if len(numbers) > 1 {
		currentVersion.minor = numbers[1]
		if len(numbers) > 2 {
			currentVersion.patch = numbers[2]
		}
	}
	prediction := currentVersion
	if len(numbers) > 2 {
		prediction.patch += 1
		predictions = append(predictions, prediction.toString())
		prediction = currentVersion
	}
	if len(numbers) > 1 {
		prediction.patch = 0
		prediction.minor += 1
		predictions = append(predictions, prediction.toString())
		prediction = currentVersion
	}
	prediction.patch = 0
	prediction.minor = 0
	prediction.major += 1
	return append(predictions, prediction.toString())
}

func (version version) toString() string {
	versionNumbers := []string{strconv.Itoa(version.major), strconv.Itoa(version.minor), strconv.Itoa(version.patch)}
	return strings.Join(versionNumbers, dot)
}
