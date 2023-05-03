package session

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/jamesdube/ussd/internal/config"
	"github.com/jamesdube/ussd/internal/utils"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const mapKey = "ussd-sessions"

type HazelcastRepository struct {
	client *hazelcast.Client
}

func NewHazelCast(name string) *HazelcastRepository {

	host := config.Get("HAZELCAST_HOST")
	portS := config.Get("HAZELCAST_PORT")
	port, _ := strconv.Atoi(portS)

	cfg := hazelcast.Config{}
	cc := &cfg.Cluster
	cc.Network.SetAddresses(fmt.Sprintf("%s:%d", host, port))
	cc.Name = name

	client, err := hazelcast.StartNewClientWithConfig(context.TODO(), cfg)
	if err != nil {
		zap.Error(err)
		return nil
	}

	return &HazelcastRepository{
		client: client,
	}
}

func (h *HazelcastRepository) GetSession(id string) (*Session, error) {

	ctx := context.TODO()
	hMap, e := h.client.GetMap(context.TODO(), mapKey)

	if e != nil {
		zap.Error(e)
		return nil, e
	}

	key, err := hMap.ContainsKey(ctx, id)
	if err != nil {
		zap.Error(err)
		return nil, err
	}

	if !key {
		return NewSession(id), nil
	}

	data, err := hMap.Get(ctx, id)
	if err != nil {
		zap.Error(err)
		return nil, err
	}

	var sess Session

	b, err := json.Marshal(data)
	if err != nil {
		utils.Logger.Error(err.Error())
		return nil, err
	}
	s := string(b)

	err = FromJson(s, &sess)
	if err != nil {
		zap.Error(err)
		return nil, err
	}

	return &sess, nil
}

func (h *HazelcastRepository) Save(s *Session) error {

	ctx := context.TODO()
	hMap, e := h.client.GetMap(ctx, mapKey)

	if e != nil {
		utils.Logger.Error(e.Error())
		return e
	}

	err := hMap.SetWithTTL(ctx, s.Id, s, time.Duration(60)*time.Second)
	if err != nil {
		utils.Logger.Error(e.Error())
		return err
	}

	return nil

}

func (h *HazelcastRepository) Delete(id string) {

	ctx := context.TODO()
	hMap, e := h.client.GetMap(ctx, mapKey)

	if e != nil {
		utils.Logger.Error(e.Error())
		return
	}

	err := hMap.Delete(ctx, id)
	if err != nil {
		utils.Logger.Error(err.Error())
		return
	}

}
