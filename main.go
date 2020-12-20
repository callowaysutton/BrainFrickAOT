package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
)

var (
	mutex   sync.Mutex
	current = 0
	result  = make(map[int]string)
	input   int
	err     error
)

func main() {
	if len(os.Args) >= 2 {
		input, err = strconv.Atoi(os.Args[1])
	}
	if err != nil || input == 0 {
		input = 15
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	go number(input, &waitGroup)
	go fizzbuzz(input, &waitGroup)
	go fizz(input, &waitGroup)
	go buzz(input, &waitGroup)
	waitGroup.Wait()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()

		b := bufio.NewWriter(os.Stdout)
		defer b.Flush()

		for k := range sortMap(result) {
			fmt.Fprintf(b, "Number: %d\t\tResult: %s\n", k, result[k])
		}
	}()
	waitGroup.Wait()
}

func sortMap(input map[int]string) []int {
	keys := make([]int, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func fizz(n int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		mutex.Lock()
		if current > n {
			mutex.Unlock()
			return
		}
		if current%3 == 0 && current%5 != 0 {
			result[current] = "Fizz"
			current++
		}
		mutex.Unlock()
	}

}

func buzz(n int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		mutex.Lock()
		if current > n {
			mutex.Unlock()
			return
		}
		if current%3 != 0 && current%5 == 0 {
			result[current] = "Buzz"
			current++
		}
		mutex.Unlock()
	}

}

func fizzbuzz(n int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		mutex.Lock()
		if current > n {
			mutex.Unlock()
			return
		}
		if current%3 == 0 && current%5 == 0 {
			result[current] = "FizzBuzz"
			current++
		}
		mutex.Unlock()
	}
}

func number(n int, waitGroup *sync.WaitGroup) {
	for current <= n {
		if current%3 != 0 && current%5 != 0 {
			mutex.Lock()
			result[current] = strconv.Itoa(current)
			current++
			mutex.Unlock()
		}
	}
	waitGroup.Done()

}
