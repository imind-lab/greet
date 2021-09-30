package main

//import "fmt"
//
//func main() {
//	fmt.Println(RecvDir, SendDir, BothDir)
//}
//
//type ChanDir int
//
//const (
//	RecvDir ChanDir             = 1 << iota // <-chan
//	SendDir                                 // chan<-
//	BothDir = RecvDir | SendDir             // chan
//)
//
//
//func binarySearch(list []int, item int) int {
//	low := 0
//	high := len(list) - 1
//	for {
//		mid := (low + high) / 2
//		guess := list[mid]
//		if guess == item {
//			return mid
//		} else if guess > item {
//			high = mid - 1
//		} else {
//			low = mid + 1
//		}
//		if low > high {
//			break
//		}
//	}
//	return -1
//}
//
//func quickSort(list []int) []int {
//	if len(list) < 2 {
//		return list
//	}
//	pivot := list[0]
//	var less []int
//	var greater []int
//	for _, item := range list[1:] {
//		if item <= pivot {
//			less = append(less, item)
//		} else {
//			greater = append(greater, item)
//		}
//	}
//	return append(append(quickSort(less), pivot), quickSort(greater)...)
//}
