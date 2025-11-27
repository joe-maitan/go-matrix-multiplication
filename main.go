package main

import (
	"os"
	"fmt"
	"time"
	"sync"
	"runtime"
	"strconv"
	"math/rand"
)

type Job struct {
	RowIndex int
	ColIndex int
	Row1 []int  
	Row2 []int
	ProductMatrix *Matrix // pointer to product matrix
} // End Job struct

type Matrix struct {
	Data     [][]int
	Size     int
	Name     string
	JobQueue chan Job
} // End Matrix struct

func (m *Matrix) GenerateMatrix() {
	// fmt.Printf("Generating matrix %v ...\n", m.Name)

	for row := 0; row < m.Size; row++ {
		newRow := make([]int, m.Size)
		for col := 0; col < len(newRow); col++ {
			newRow[col] = rand.Intn(10 - 0 + 1)
		}

		m.Data[row] = newRow
	}
} // End GenerateMatrix() func

func (m *Matrix) Transpose() {
	for row := 0; row < m.Size; row++ {
		for col := row + 1; col < m.Size; col++ {
			temp := m.Data[row][col]
			m.Data[row][col] = m.Data[col][row]
			m.Data[col][row] = temp
		}
	}
} // End Transpose() func

func QueueJobs(m1 *Matrix, m2 *Matrix) Matrix {
	numJobs := m1.Size * m2.Size

	productMatrix := Matrix{
		Data: make([][]int, m1.Size), 
		Size: m1.Size, 
		Name: "x", 
		JobQueue: make(chan Job, numJobs),
	}

	for i := 0; i < m1.Size; i++ {
		productMatrix.Data[i] = make([]int, m2.Size) // initialize rows in the matrix
		
		for j := 0; j < m2.Size; j++ {
			newJob := Job{
				RowIndex: i, 
				ColIndex: j,
				Row1: m1.Data[i], 
				Row2: m2.Data[j], 
				ProductMatrix: &productMatrix,
			}

			productMatrix.JobQueue <- newJob // Send data to a channel
		}
	}

	close(productMatrix.JobQueue) // Signal no more jobs
	return productMatrix
} // End QueueJobs(m1, m2, tileSize) func

func (m *Matrix) Multiply() {
	for job := range m.JobQueue {
		rowIndex := job.RowIndex
		colIndex := job.ColIndex
		row1 := job.Row1
		row2 := job.Row2

		product := 0
		for i := 0; i < m.Size; i++ {
			product += row1[i] * row2[i]
		}

		job.ProductMatrix.Data[rowIndex][colIndex] = product
	} // End for-each job loop
} // End Multiply() func

func (m *Matrix) String() string{
	var toString string
	toString = fmt.Sprintf("Matrix %v:\n", m.Name)

	for i := 0; i < len(m.Data); i++ {
		toString += fmt.Sprintf("%v\n", m.Data[i])
	}

	return toString
}

func main() {
	var grandTotal time.Duration
	grandTotal = 0
	matrixSize, err := strconv.Atoi(os.Args[1])
	numCores := runtime.NumCPU()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Dimensionality of the square matrices is: %v \n", matrixSize)
	fmt.Printf("The thread pool size has been initialized to: %v\n\n", numCores)

	a := Matrix{
		Data: make([][]int, matrixSize), 
		Size: matrixSize, 
		Name: "a",
	}

	a.GenerateMatrix()

	b := Matrix{
		Data: make([][]int, matrixSize), 
		Size: matrixSize, 
		Name: "b",
	}

	b.GenerateMatrix()
	b.Transpose()

	c := Matrix{
		Data: make([][]int, matrixSize), 
		Size: matrixSize, 
		Name: "c",
	}

	c.GenerateMatrix()

	d := Matrix{
		Data: make([][]int, matrixSize), 
		Size: matrixSize, 
		Name: "d",
	}

	d.GenerateMatrix()
	d.Transpose()

	// fmt.Println(a.String())
	// fmt.Println(b.String())
	
	x := QueueJobs(&a, &b)
	y := QueueJobs(&c, &d)
	
	var wg sync.WaitGroup
    start := time.Now()
    for i := 0; i < numCores; i++ {
        wg.Add(1)
        go func(workerID int) {
            x.Multiply()
			defer wg.Done()
        }(i)
    }

	wg.Wait()
	// fmt.Println(x.String())
	fin := time.Now()
	elapsed := fin.Sub(start)
	grandTotal += elapsed
	fmt.Printf("Calculation of X (Product of A and B) complete. Time to compute matrix %v\n", elapsed)

	start = time.Now()
	for i := 0; i < numCores; i++ {
        wg.Add(1)
        go func(workerID int) {
			y.Multiply()
            defer wg.Done()
        }(i)
    }
	
	wg.Wait()
    
	fin = time.Now()
	elapsed = fin.Sub(start)
	grandTotal += elapsed
    fmt.Printf("Calculation of Y (Product of C and D) complete. Time to compute matrix %v\n", elapsed)
   
	y.Transpose()
	
	z := QueueJobs(&x, &y)
	start = time.Now()
	for i := 0; i < numCores; i++ {
        wg.Add(1)
        go func(workerID int) {
            z.Multiply()
			defer wg.Done()
        }(i)
    }
	
	wg.Wait()
	
	fin = time.Now()
	elapsed = fin.Sub(start)
	grandTotal += elapsed
    fmt.Printf("Calculation of Z (Product of X and Y) complete. Time to compute matrix %v\n", elapsed)

	fmt.Printf("Finished! Total time taken = %v\n", grandTotal)
} // End main() func
