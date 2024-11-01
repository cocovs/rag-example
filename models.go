package main

// VectorResult 存储向量查询结果
type VectorResult struct {
	ID        int     `json:"id"`
	Embedding string  `json:"embedding"`
	Mark      string  `json:"mark"`
	Index     float64 `json:"index"` //指标 如余弦相似度 欧几里得距离
	Text      string  `json:"text"`
}
