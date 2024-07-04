package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"
)

func countWords(lines <-chan string, counts chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	wordCounts := make(map[string]int)
	for line := range lines {
		words := strings.FieldsFunc(line, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		for _, word := range words {
			word = strings.ToLower(word)
			wordCounts[word]++
		}
	}
	counts <- wordCounts
}

func mergeCounts(counts <-chan map[string]int, finalCounts map[string]int, done chan<- struct{}) {
	for wordCount := range counts {
		for word, count := range wordCount {
			finalCounts[word] += count
		}
	}
	done <- struct{}{}
}

func main() {

	res, err := readRandomWords("random_words.txt")

	if err != nil {
		fmt.Println(err)
	}

	resC, errC := readRandomWords("common_words.txt")

	if errC != nil {
		fmt.Println(errC)
	} else {
		for key, _ := range resC {
			fmt.Printf("Count for %s is %d\n", key, res[key])
		}
	}

}

func readRandomWords(fileName string) (map[string]int, error) {

	pwd, _ := os.Getwd()

	// Open the file
	file, err := os.Open(pwd + "/cmd/assesment/files/" + fileName)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return make(map[string]int), nil
	}
	defer file.Close()

	lines := make(chan string, 100)
	counts := make(chan map[string]int, 10)
	var wg sync.WaitGroup

	// Start goroutines for counting words
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go countWords(lines, counts, &wg)
	}

	// Start a goroutine to merge the counts
	finalCounts := make(map[string]int)
	done := make(chan struct{})
	go mergeCounts(counts, finalCounts, done)

	// Read the file and send lines to the workers
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
	close(lines)

	// Wait for all workers to finish
	wg.Wait()
	close(counts)

	// Wait for the merge goroutine to finish
	<-done

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	return finalCounts, nil
}
