package qianwen_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// VectorService 向量服务结构体
type VectorService struct {
	apiKey  string
	baseURL string
}

// NewVectorService 创建新的向量服务实例
func NewVectorService(apiKey string) *VectorService {
	return &VectorService{
		apiKey:  apiKey,
		baseURL: "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding",
	}
}

// EmbeddingRequest 请求体结构
type EmbeddingRequest struct {
	Model      string `json:"model"`
	Input      Input  `json:"input"`
	Parameters Params `json:"parameters"`
}

type Input struct {
	Texts []string `json:"texts"`
}

type Params struct {
	TextType string `json:"text_type"`
}

// ErrorResponse 错误响应结构体
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// EmbeddingResponse 响应体结构
type EmbeddingResponse struct {
	Output struct {
		Embeddings []EmbeddingResult `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

// EmbeddingResult 向量结果结构
type EmbeddingResult struct {
	TextIndex int       `json:"text_index"`
	Embedding []float64 `json:"embedding"`
}

// GetEmbeddings 获取文本向量
func (s *VectorService) GetEmbeddings(texts []string) (*EmbeddingResponse, error) {

	// 参数验证
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts 不能为空")
	}

	reqBody := EmbeddingRequest{
		Model: "text-embedding-v1",
		Input: Input{
			Texts: texts,
		},
		Parameters: Params{
			TextType: "query",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API错误: %s - %s (RequestID: %s)",
			errResp.Code, errResp.Message, errResp.RequestID)
	}

	// 解析响应数据
	var response EmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &response, nil
}

// only return embedding
func (s *VectorService) GetEmbeddingsOnly(texts []string) ([]float64, error) {
	response, err := s.GetEmbeddings(texts)
	if err != nil {
		return nil, err
	}
	return response.Output.Embeddings[0].Embedding, nil
}
