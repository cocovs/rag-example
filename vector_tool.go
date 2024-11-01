package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 格式化切片为字符串
// eg [1,2,3] -> "1,2,3"
func formatSlice(vector []float64) string {
	result := ""
	for i, v := range vector {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%.6f", v)
	}
	return result
}

// vectorToVectorString 将向量转换为 pgvector 格式的字符串
// eg [1,2,3] -> '[1,2,3]'
func vectorToVectorString(vector []float64) string {
	return fmt.Sprintf("'[%s]'", formatSlice(vector))
}

// vectorStringToVector 将 pgvector 格式的字符串转换为向量
// eg '[1,2,3]' -> [1,2,3]
func vectorStringToVector(vectorStr string) []float64 {
	vectorStr = strings.Trim(vectorStr, "[]")
	vectorStrs := strings.Split(vectorStr, ",")
	vector := make([]float64, len(vectorStrs))
	for i, v := range vectorStrs {
		vector[i], _ = strconv.ParseFloat(v, 64)
	}
	return vector
}
