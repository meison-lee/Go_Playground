package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

const (
	width  = 20000
	height = 10000
)

type Grid [][]bool

// Clear terminal (Unix)
func clearScreen() {
	cmd := exec.Command("clear") // use "cls" for Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func newGrid() Grid {
	grid := make(Grid, height)
	for i := range grid {
		grid[i] = make([]bool, width)
	}
	return grid
}

func randomize(grid Grid) {
	for y := range grid {
		for x := range grid[y] {
			grid[y][x] = rand.Intn(20) == 0 // 25% alive
		}
	}
}

func printGrid(grid Grid) {
	for _, row := range grid {
		for _, alive := range row {
			if alive {
				fmt.Print("O")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

// Conway's rules
func stepConcurrent(src Grid) Grid {
	dst := newGrid()

	var wg sync.WaitGroup
	for y := range src {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()

			for x := range src[y] {
				live := 0
				// Count neighbors
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dy == 0 && dx == 0 {
							continue
						}
						ny, nx := y+dy, x+dx
						if ny >= 0 && ny < height && nx >= 0 && nx < width && src[ny][nx] {
							live++
						}
					}
				}
				dst[y][x] = (src[y][x] && (live == 2 || live == 3)) || (!src[y][x] && live == 3)
			}
		}(y)
	}
	wg.Wait()
	return dst
}

func stepWorkerPool(src Grid) Grid {
	dst := newGrid()
	var wg sync.WaitGroup
	jobs := make(chan int, height)

	numWorkers := runtime.NumCPU()
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range jobs {
				for x := 0; x < width; x++ {
					live := 0
					for dy := -1; dy <= 1; dy++ {
						for dx := -1; dx <= 1; dx++ {
							if dy == 0 && dx == 0 {
								continue
							}
							ny, nx := y+dy, x+dx
							if ny >= 0 && nx >= 0 && ny < height && nx < width && src[ny][nx] {
								live++
							}
						}
					}
					dst[y][x] = (src[y][x] && (live == 2 || live == 3)) || (!src[y][x] && live == 3)
				}
			}
		}()
	}
	for y := 0; y < height; y++ {
		jobs <- y
	}
	close(jobs)
	wg.Wait()
	return dst
}

func step(src Grid) Grid {
	dst := newGrid()
	for y := range src {

		for x := range src[y] {
			live := 0
			// Count neighbors
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dy == 0 && dx == 0 {
						continue
					}
					ny, nx := y+dy, x+dx
					if ny >= 0 && ny < height && nx >= 0 && nx < width && src[ny][nx] {
						live++
					}
				}
			}
			dst[y][x] = (src[y][x] && (live == 2 || live == 3)) || (!src[y][x] && live == 3)
		}
	}

	return dst
}

func main() {
	grid := newGrid()
	randomize(grid)

	for {
		// clearScreen()
		// printGrid(grid)
		calculate_time := time.Now()
		grid = step(grid)
		fmt.Printf("Time elapsed: %v\n", time.Since(calculate_time))
		time.Sleep(100 * time.Millisecond)
	}
}
