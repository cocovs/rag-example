package main

import (
	"context"
	"fmt"
	"log"
	"testing"
)

var MARK_SOFT_WARE_ISSUE = "软件故障维护服务"
var MARK_HOME_DEVICE_ISSUE = "家用设备故障维护服务"
var MARK_HARDWARE_ISSUE = "硬件设备故障维护服务"

var issues = []VectorResult{
	{
		Text: "水管坏了,这件事情非常紧急，请帮我解决",
		Mark: MARK_HOME_DEVICE_ISSUE,
	},
	{
		Text: "水管漏水了,情况非常严重，请尽快处理,需要的更多的信息请联系我",
		Mark: MARK_HOME_DEVICE_ISSUE,
	},
	{
		Text: "我家电梯坏了,急需维修，请帮忙安排",
		Mark: MARK_HARDWARE_ISSUE,
	},
	{
		Text: "厨房水龙头坏了,水流不止，请尽快修理",
		Mark: MARK_HOME_DEVICE_ISSUE,
	},
	{
		Text: "卫生间堵塞了,影响使用，请立即解决",
		Mark: MARK_HOME_DEVICE_ISSUE,
	},
	{
		Text: "冰箱不制冷了,食物快坏了，请尽快解决",
		Mark: MARK_HOME_DEVICE_ISSUE,
	},
	{
		Text: "问题反馈：服务单查询存在问题 。客户:xxx",
		Mark: MARK_SOFT_WARE_ISSUE,
	},
	{
		Text: "选项存在问题，乱码",
		Mark: MARK_SOFT_WARE_ISSUE,
	},
	{
		Text: "用户主页，角色切换，切换不生效",
		Mark: MARK_SOFT_WARE_ISSUE,
	},
	{
		Text: "用户id为空",
		Mark: MARK_SOFT_WARE_ISSUE,
	},
	{
		Text: "用户服务出现重复选项",
		Mark: MARK_SOFT_WARE_ISSUE,
	},
}

func TestEmbeddingInster(t *testing.T) {
	config := NewServicesConfig()
	log.Default().Println(config)
	// 初始化服务
	services, err := NewServices(config)
	if err != nil {
		log.Fatal(err, "newservices error")
	}
	//获取向量
	issuesTexts := make([]string, 0)
	for _, v := range issues {
		issuesTexts = append(issuesTexts, v.Text)
	}
	res, err := services.VectorService.GetEmbeddings(issuesTexts)
	if err != nil {
		log.Fatal(err, "getembeddings error")
	}
	embeddings := res.Output.Embeddings
	//inster 到 postgresql
	for i, v := range embeddings {
		vectorStr := vectorToVectorString(v.Embedding)
		sql := fmt.Sprintf("INSERT INTO  %s (embedding,text,mark) VALUES (%s,'%s','%s')", VECTOR_TABLE, vectorStr, issues[i].Text, issues[i].Mark)
		//log.Default().Println(sql)
		_, err := services.PgxDB.Exec(context.Background(), sql)
		if err != nil {
			log.Fatal(err, "  insert postgresql error")
		}
	}
}

func TestGetSimilarVectors(t *testing.T) {
	config := NewServicesConfig()
	log.Default().Println(config)
	services, err := NewServices(config)
	if err != nil {
		log.Fatal(err, config, "newservices error")
	}
	defer services.PgxDB.Close(context.Background())
	defer services.MongoDB.Disconnect(context.Background())

	embeddings, err := services.VectorService.GetEmbeddingsOnly([]string{"水管坏了,这件事情非常紧急，请帮我解决"})
	if err != nil {
		log.Fatal(err, "getembeddingsonly error")
	}

	results, err := services.GetSimilarVectorByDistance(embeddings)
	if err != nil {
		log.Fatal(err, "getsimilarvectors error")
	}
	for _, vec := range results {
		fmt.Println(vec.ID)
	}
}

func TestGetSimilarVectorByDistance(t *testing.T) {
	config := NewServicesConfig()
	log.Default().Println(config)
	services, err := NewServices(config)
	if err != nil {
		log.Fatal(err, "newservices error")
	}
	//pq get id =1
	vr, err := services.GetRecordById(21)
	if err != nil {
		log.Fatal(err, "getrecordbyid error")
	}
	results, err := services.GetSimilarVectorByDistance(vectorStringToVector(vr.Embedding))
	if err != nil {
		log.Fatal(err, "getsimilarvectors error")
	}
	for _, v := range results {
		fmt.Println(v.ID, v.Index, v.Mark)
	}
}

func TestGetSimilarVectorByCosineSimilarity(t *testing.T) {
	config := NewServicesConfig()
	log.Default().Println(config)
	services, err := NewServices(config)
	if err != nil {
		log.Fatal(err, "newservices error")
	}
	vr, err := services.GetRecordById(21)
	if err != nil {
		log.Fatal(err, "getrecordbyid error")
	}
	results, err := services.GetSimilarVectorByCosineSimilarity(vectorStringToVector(vr.Embedding))
	if err != nil {
		log.Fatal(err, "getcosinesimilarvector error")
	}
	for _, v := range results {
		fmt.Println(v.ID, v.Index, v.Mark)
	}
}
