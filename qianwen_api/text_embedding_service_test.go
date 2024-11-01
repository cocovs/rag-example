package qianwen_api

import (
	"fmt"
	"os"
	"testing"
)

func TestVectorService(t *testing.T) {
	service := NewVectorService(os.Getenv("API_KEY"))

	texts := []string{
		"水管坏了,这件事情非常紧急，请帮我解决",
	}

	response, err := service.GetEmbeddings(texts)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("请求ID: %s\n", response.RequestID)
	fmt.Printf("总Token数: %d\n", response.Usage.TotalTokens)
	for _, embedding := range response.Output.Embeddings {
		fmt.Printf("文本索引 %d 的向量长度: %d\n",
			embedding.TextIndex, len(embedding.Embedding))
	}
}
