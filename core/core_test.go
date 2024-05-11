package core

import (
	"reflect"
	"testing"
)

func TestBubbleSort(t *testing.T) {
	testCases := []struct {
		input    []int
		expected []int
	}{
		{
			input:    []int{5, 3, 8, 6, 1},
			expected: []int{1, 3, 5, 6, 8},
		},
		{
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			input:    []int{},
			expected: []int{},
		},
		{
			input:    []int{-3, 5, 0, -8, 4},
			expected: []int{-8, -3, 0, 4, 5},
		},
	}

	for _, testCase := range testCases {
		bubbleSort(testCase.input)
		if !reflect.DeepEqual(testCase.input, testCase.expected) {
			t.Errorf("bubbleSort(%v) = %v; expected %v", testCase.input, testCase.input, testCase.expected)
		}
	}
}

func TestQuickSort(t *testing.T) {
	testCases := []struct {
		input    []int
		expected []int
	}{
		{[]int{3, 6, 8, 10, 1, 2, 1}, []int{1, 1, 2, 3, 6, 8, 10}},
		{[]int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{[]int{5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5}},
		{[]int{}, []int{}},
		{[]int{1}, []int{1}},
	}

	for _, testCase := range testCases {
		result := quickSort(testCase.input)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("quickSort(%v) = %v; expected %v", testCase.input, result, testCase.expected)
		}
	}
}

func TestInsertionSort(t *testing.T) {
	testCases := []struct {
		input    []int
		expected []int
	}{
		{
			input:    []int{3, 4, 6, 8, 9, 1, 2},
			expected: []int{1, 2, 3, 4, 6, 8, 9},
		},
		{
			input:    []int{1, 1, 1, 1, 1},
			expected: []int{1, 1, 1, 1, 1},
		},
		{
			input:    []int{},
			expected: []int{},
		},
		{
			input:    []int{-5, 3, -8, 6, 2},
			expected: []int{-8, -5, 2, 3, 6},
		},
	}

	for _, testCase := range testCases {
		insertionSort(testCase.input)
		if !reflect.DeepEqual(testCase.input, testCase.expected) {
			t.Errorf("Expected %v, but got %v", testCase.expected, testCase.input)
		}
	}
}

func TestMergeSort(t *testing.T) {
	testCases := []struct {
		input    []int
		expected []int
	}{
		{[]int{5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5}},
		{[]int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{[]int{}, []int{}},
		{[]int{1}, []int{1}},
		{[]int{3, 2, 1}, []int{1, 2, 3}},
	}

	for _, testCase := range testCases {
		result := mergeSort(testCase.input)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("Error: expected %v, got %v", testCase.expected, result)
		}
	}
}

func TestSelectionSort(t *testing.T) {
	testCases := []struct {
		input    []int
		expected []int
	}{
		{
			input:    []int{5, 3, 8, 6, 2},
			expected: []int{2, 3, 5, 6, 8},
		},
		{
			input:    []int{-2, 3, 8, -6, 2},
			expected: []int{-6, -2, 2, 3, 8},
		},
		{
			input:    []int{1, 1, 1, 1, 1},
			expected: []int{1, 1, 1, 1, 1},
		},
		{
			input:    []int{},
			expected: []int{},
		},
	}

	for _, testCase := range testCases {
		selectionSort(testCase.input)
		if !reflect.DeepEqual(testCase.input, testCase.expected) {
			t.Errorf("Error: expected %v, got %v", testCase.expected, testCase.input)
		}
	}
}
