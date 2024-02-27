package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	maxWords = 7776
	alphabet = "abcdefghijklmnopqrstuvwxyzæøå"
)

func main() {
	jsonFilePath := "lemma_expanded.json"
	outputFilePath := "output.txt"

	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	elements, err := extractElements(jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting elements: %v\n", err)
		os.Exit(1)
	}

	err = writeElementsToFileWithDicewareNumbering(elements, outputFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Elements extracted and written to '%s'\n", outputFilePath)
}

func extractElements(jsonData []byte) ([]string, error) {
	const minWordLength = 4 // Define a minimum word length for Diceware
	const maxWordLength = 9 // Define a maximum word length for Diceware

	var data [][]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	wordsByLetter := make(map[rune][]string)
	for _, letter := range alphabet {
		wordsByLetter[letter] = make([]string, 0)
	}

	for _, item := range data {
		if len(item) == 0 || len(item) > 2 && (item[2] == "EXPR" || item[2] == "NOUN" || item[2] == "ABBR" || item[2] == "PROPN") {
			continue
		}
		word, ok := item[0].(string)
		if !ok || strings.Contains(word, " ") || strings.Contains(word, "-") || strings.Contains(word, ".") {
			continue
		}

		wordLength := len([]rune(word)) // Get the length of the word in runes, not bytes
		if wordLength < minWordLength || wordLength > maxWordLength {
			continue
		}

		firstLetter := []rune(strings.ToLower(word))[0] // Normalize to lowercase and get the first letter as a rune
		if _, exists := wordsByLetter[firstLetter]; !exists {
			continue // Skip words that do not start with a letter in the Norwegian alphabet
		}

		wordsByLetter[firstLetter] = append(wordsByLetter[firstLetter], word)
	}

	var elements []string
	for len(elements) < maxWords {
		for _, letter := range alphabet {
			if len(elements) >= maxWords {
				break
			}
			if len(wordsByLetter[letter]) > 0 {
				elements = append(elements, wordsByLetter[letter][0])
				wordsByLetter[letter] = wordsByLetter[letter][1:]
			}
		}
	}

	sort.Strings(elements) // Sort the complete list alphabetically

	return elements, nil
}

func writeElementsToFileWithDicewareNumbering(elements []string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	number := [5]int{1, 1, 1, 1, 1}

	for _, element := range elements {
		line := fmt.Sprintf("%d%d%d%d%d\t%s\n", number[0], number[1], number[2], number[3], number[4], element)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("error writing element to file: %w", err)
		}
		incrementNumber(&number)
	}

	return nil
}

func incrementNumber(number *[5]int) {
	for i := 4; i >= 0; i-- {
		if number[i] < 6 {
			number[i]++
			return
		}
		number[i] = 1
	}
}
