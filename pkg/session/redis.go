package session

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jamesdube/ussd/internal/config"
	"log"
	"strconv"
	"time"
)

type Redis struct {
	client *redis.Client
	ttl    int
}

func NewRedis() *Redis {

	redisHost := config.Get("REDIS_HOST")
	redisPort := config.Get("REDIS_PORT")
	redisDB := config.Get("REDIS_DB")
	ttl := config.Get("SESSION_TTL")

	db, _ := strconv.Atoi(redisDB)

	c := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "", // no password set
		DB:       db, // use default DB
	})

	sTtl, _ := strconv.Atoi(ttl)

	return &Redis{c, sTtl}
}

func (r *Redis) GetSession(id string) (*Session, error) {

	s, err := r.client.Get(generateKey(id)).Result()
	if err != nil && err != redis.Nil {
		fmt.Println(err)
		return nil, err
	}

	if s == "{}" || err == redis.Nil {
		return NewSession(id), nil
	}

	var sess Session
	err2 := FromJson(s, &sess)

	return &sess, err2

}

func (r *Redis) Save(s *Session) error {

	sJson, err := ToJson(s)
	if err != nil {
		log.Println("error converting session to json", err)
	}

	err = r.client.Set(generateKey(s.GetID()), sJson, time.Second*time.Duration(r.ttl)).Err()
	return err
}

func (r *Redis) Delete(id string) {
	r.client.Del(generateKey(id))
}

func ToJson(sess *Session) (string, error) {
	b, e := json.Marshal(sess)
	s := string(b)
	return s, e
}

func FromJson(j string, sess *Session) error {
	err := json.Unmarshal([]byte(j), sess)
	return err
}

func generateKey(id string) string {
	return fmt.Sprintf("sessions::%s", id)
}

func FromJsonArray(j string, sess *[]Session) error {
	err := json.Unmarshal([]byte(j), &sess)
	return err
}
