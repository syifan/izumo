package main

/*
#cgo LDFLAGS: -L/usr/local/cuda/lib64 -lcudart
#cgo CFLAGS: -I/usr/local/cuda/include/
#include <cuda.h>
#include <cuda_runtime.h>
#include <driver_types.h>
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
	// "github.com/pkg/profile"

	"github.com/sunyifan112358/izumo"
)

var length = 512000
var threadPerBlock = 512

func main() {
	var v1, v2, gpuSum, cpuSum []float32
	v1 = make([]float32, length)
	v2 = make([]float32, length)
	gpuSum = make([]float32, length)
	cpuSum = make([]float32, length)

	initializeVector(v1, v2)
	gpuVectorAdd(v1, v2, gpuSum)
	cpuVectorAdd(v1, v2, cpuSum)
	verify(gpuSum, cpuSum)

}

func initializeVector(v1, v2 []float32) {
	for i := 0; i < length; i++ {
		v1[i] = float32(1)
		v2[i] = float32(i)
	}
}

func gpuVectorAdd(v1, v2, sum []float32) {
	gV1, _ := izumo.NewGpuMem(length * 4)
	gV2, _ := izumo.NewGpuMem(length * 4)
	gSum, _ := izumo.NewGpuMem(length * 4)

	gV1.CopyHostToDevice(unsafe.Pointer(&v1[0]))
	gV2.CopyHostToDevice(unsafe.Pointer(&v2[0]))

	module, err := izumo.LoadModuleFromFile("add_kernel.ptx")
	if err != nil {
		log.Fatal(err)
	}

	function, err := module.GetFunction("add")
	if err != nil {
		log.Fatal(err)
	}

	gridDim := izumo.Dim3{length / threadPerBlock, 1, 1}
	blockDim := izumo.Dim3{threadPerBlock, 1, 1}
	err = function.LaunchKernel(gridDim, blockDim, 0, *gV1, *gV2, *gSum)
	if err != nil {
		log.Fatal(err)
	}

	gSum.CopyDeviceToHost(unsafe.Pointer(&sum[0]))
}

func cpuVectorAdd(v1, v2, sum []float32) {
	for i := 0; i < length; i++ {
		sum[i] = v1[i] + v2[i]
	}
}

func verify(gpuSum, cpuSum []float32) {
	for i := 0; i < length; i++ {
		if gpuSum[i] != cpuSum[i] {
			fmt.Printf("Mismatch at %d, expected %f, but get %f\n", i,
				cpuSum[i], gpuSum[i])
			break
		}
	}
	fmt.Println("Succeed!")
}
