package redis

import (
	"context"
	"encoding/json"
	"github.com/XiovV/selly-server/models"
	"github.com/go-redis/redis/v9"
	"os"
)

type Redis struct {
	ctx context.Context
	db  *redis.Client
}

func New() *Redis {
	r := Redis{}

	r.ctx = context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})

	r.db = rdb
	return &r
}

func (r *Redis) SetOnline(sellyId string) {
	r.db.Set(r.ctx, sellyId, true, 0)
}

func (r *Redis) IsUserOnline(sellyId string) bool {
	online, _ := r.db.Get(r.ctx, sellyId).Bool()
	return online
}

func (r *Redis) DelUser(sellyId string) {
	r.db.Del(r.ctx, sellyId)
}

func (r *Redis) PushMessage(message models.Message) {
	m, _ := json.Marshal(message)

	r.db.LPush(r.ctx, message.Receiver, m)
}

func (r *Redis) GetMessages(sellyId string) []models.Message {
	res, _ := r.db.LRange(r.ctx, sellyId, 0, -1).Result()

	messages := []models.Message{}
	var message models.Message

	for _, m := range res {
		json.Unmarshal([]byte(m), &message)

		messages = append(messages, message)
	}

	return messages
}

//msg1 := models.Message{
//	Sender:   "user1",
//	Receiver: "user2",
//	Message:  "message1",
//}
//
//msg2 := models.Message{
//	Sender:   "user1",
//	Receiver: "user2",
//	Message:  "message2",
//}
//
//msg3 := models.Message{
//	Sender:   "user1",
//	Receiver: "user2",
//	Message:  "message3",
//}
//
//m1, _ := json.Marshal(msg1)
//m2, _ := json.Marshal(msg2)
//m3, _ := json.Marshal(msg3)
//
//db.LPush(ctx, "user1", m1)
//db.LPush(ctx, "user1", m2)
//db.LPush(ctx, "user1", m3)
//
//res := db.LRange(ctx, "user1", 0, -1)
//str, _ := res.Result()
//
//messages := []models.Message{}
//for i, message := range str {
//	json.Unmarshal([]byte(message), &messages[i])
//}
//
//fmt.Println(messages[0].Message)
