package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"rag-vector/db"
	"rag-vector/qianwen_api"
	"strconv"

	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	VECTOR_TABLE = "public.vectors"
)

type Services struct {
	PgxDB         *pgx.Conn
	VectorService *qianwen_api.VectorService
	Postgresql    *sql.DB
	MongoDB       *mongo.Client
}

type ServicesConfig struct {
	//pgx
	PostgresqlConfig *db.PostgresqlConfig
	MongoDBConfig    *db.MongoDBConfig
	ApiKey           string
	Dimension        int //vectors table dimension 维度
}

func NewServicesConfig() *ServicesConfig {
	return &ServicesConfig{
		ApiKey: os.Getenv("API_KEY"),
		PostgresqlConfig: &db.PostgresqlConfig{
			Host: os.Getenv("POSTGRESQL_HOST"),
			Port: func() int {
				p, _ := strconv.Atoi(os.Getenv("POSTGRESQL_PORT"))
				return p
			}(),
			User:     os.Getenv("POSTGRESQL_USER"),
			Password: os.Getenv("POSTGRESQL_PASSWORD"),
			Dbname:   os.Getenv("POSTGRESQL_DBNAME"),
		},
		MongoDBConfig: &db.MongoDBConfig{
			Host: os.Getenv("MONGODB_HOST"),
			Port: func() int {
				p, _ := strconv.Atoi(os.Getenv("MONGODB_PORT"))
				return p
			}(),
			User:     os.Getenv("MONGODB_USER"),
			Password: os.Getenv("MONGODB_PASSWORD"),
		},
	}
}

func NewServices(config *ServicesConfig) (services *Services, err error) {
	services = &Services{}

	services.VectorService = qianwen_api.NewVectorService(config.ApiKey)

	if config.PostgresqlConfig.Host == "" {
		config.PostgresqlConfig = &db.PostgresqlConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "example",
			Password: "example",
			Dbname:   "example",
		}
	}
	if config.MongoDBConfig.Host == "" {
		config.MongoDBConfig = &db.MongoDBConfig{
			Host:     "localhost",
			Port:     27017,
			User:     "root",
			Password: "example",
		}
	}
	services.PgxDB, err = db.NewPgxDB(config.PostgresqlConfig)
	if err != nil {
		return nil, err
	}
	services.MongoDB, err = db.NewMongoDB(config.MongoDBConfig)
	if err != nil {
		return nil, err
	}

	err = services.InitServices(config.Dimension)
	if err != nil {
		return nil, err
	}

	return services, nil
}

// 维度
func (services *Services) InitServices(dimension int) error {
	PGDDL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS vectors (id bigserial PRIMARY KEY, embedding vector(%v));`, dimension)
	_, err := services.PgxDB.Exec(context.Background(), PGDDL)
	if err != nil {
		return err
	}
	return nil
}

// GetSimilarVector 接收单个向量，返回相似向量列表
func (s *Services) GetSimilarVectorByDistance(inputVector []float64) ([]VectorResult, error) {
	vectorStr := vectorToVectorString(inputVector)
	query := fmt.Sprintf(`
		SELECT id, embedding,mark,
		embedding <-> %s as distance
		FROM %s
		ORDER BY distance
	`, vectorStr, VECTOR_TABLE)

	rows, err := s.PgxDB.Query(context.Background(), query)
	if err != nil {
		log.Printf("查询相似向量失败: %v", err)
		log.Printf("query: %s", query)
		return nil, err
	}
	defer rows.Close()

	var similarVectors []VectorResult
	for rows.Next() {
		var vr VectorResult
		err := rows.Scan(&vr.ID, &vr.Embedding, &vr.Mark, &vr.Index)
		if err != nil {
			log.Printf("读取行数据失败: %v", err)
			continue
		}
		similarVectors = append(similarVectors, vr)
	}

	return similarVectors, nil
}

// GetSimilarVectors 接收多个向量，返回相似向量列表
func (s *Services) GetSimilarVectorsByDistance(inputVectors [][]float64, topK int) ([][]VectorResult, error) {
	var results [][]VectorResult

	for _, vec := range inputVectors {
		similarVectors, err := s.GetSimilarVectorByDistance(vec)
		if err != nil {
			return nil, err
		}
		results = append(results, similarVectors)
	}

	return results, nil
}

// 余弦相似度
// 1 表示完全相似
// 0 表示完全不相似
// -1 表示方向相反
// 余弦相似度越大，表示越相似
func (s *Services) GetSimilarVectorByCosineSimilarity(inputVector []float64) ([]VectorResult, error) {
	vectorStr := vectorToVectorString(inputVector)
	query := fmt.Sprintf(`
        SELECT id, embedding, mark,
        1 - (embedding <=> %s) as cosine_similarity
        FROM %s
        ORDER BY cosine_similarity DESC
    `, vectorStr, VECTOR_TABLE)

	rows, err := s.PgxDB.Query(context.Background(), query)
	if err != nil {
		log.Printf("查询余弦相似度向量失败: %v", err)
		log.Printf("query: %s", query)
		return nil, err
	}
	defer rows.Close()

	var similarVectors []VectorResult
	for rows.Next() {
		var vr VectorResult
		err := rows.Scan(&vr.ID, &vr.Embedding, &vr.Mark, &vr.Index)
		if err != nil {
			log.Printf("读取行数据失败: %v", err)
			continue
		}
		similarVectors = append(similarVectors, vr)
	}

	return similarVectors, nil
}

// get record by id
func (s *Services) GetRecordById(id int) (*VectorResult, error) {
	vr := new(VectorResult)
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", VECTOR_TABLE, id)
	err := s.PgxDB.QueryRow(context.Background(), query).Scan(&vr.ID, &vr.Embedding, &vr.Mark, &vr.Index)
	if err != nil {
		log.Fatal(err, "query postgresql error", query)
	}
	return vr, nil
}
