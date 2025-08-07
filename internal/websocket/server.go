package websocket

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/auto-devs/auto-devs/config"
	"github.com/centrifugal/centrifuge"
)

type Server struct {
	node *centrifuge.Node
}

type UserInfo struct {
	UserID string
}

func parseJwtToken(token string) (*UserInfo, error) {
	return nil, errors.New("not implemented")
}

func NewServer(appConfig *config.CentrifugeRedisBrokerConfig) (*Server, error) {
	cfg := centrifuge.Config{
		LogLevel:   centrifuge.LogLevelInfo,
		LogHandler: handleLog,
	}

	node, err := centrifuge.New(cfg)
	if err != nil {
		return nil, err
	}

	setupRedisBroker(node, appConfig)

	node.OnConnecting(func(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		log.Println("on connecting", e.Token)
		claims, err := parseJwtToken(e.Token)
		if err != nil {
			return centrifuge.ConnectReply{
				Credentials: &centrifuge.Credentials{
					UserID: "anonymous",
				},
			}, nil
		}
		log.Println("user_id", claims.UserID)
		return centrifuge.ConnectReply{
			Credentials: &centrifuge.Credentials{
				UserID: claims.UserID,
			},
		}, nil
	})

	node.OnConnect(func(client *centrifuge.Client) {
		log.Printf("------user %s connected", client.UserID())
		transport := client.Transport()
		log.Printf("user %s connected via %s with protocol: %s", client.UserID(), transport.Name(), transport.Protocol())

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			log.Printf("user %s subscribes on %s", client.UserID(), e.Channel)
			// if channel start with $, then it's a private channel,
			// private channel is in format $:user_id:channel_name
			if strings.HasPrefix(e.Channel, "$") {
				// private channel require user_id to be the same as the client.UserID()
				channelParts := strings.Split(e.Channel, ":")
				channelUserId := channelParts[1]
				if client.UserID() != channelUserId {
					cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
					return
				}
				cb(centrifuge.SubscribeReply{}, nil)
			} else {
				cb(centrifuge.SubscribeReply{}, nil)
			}
		})

		client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
			log.Printf("user %s unsubscribed from %s", client.UserID(), e.Channel)
		})

		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			log.Printf("user %s publishes into channel %s: %s", client.UserID(), e.Channel, string(e.Data))
			cb(centrifuge.PublishReply{}, nil)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			log.Printf("user %s disconnected, disconnect: %s", client.UserID(), e.Disconnect)
		})
	})

	return &Server{node: node}, nil
}

func setupRedisBroker(node *centrifuge.Node, appConfig *config.CentrifugeRedisBrokerConfig) {
	redisShardConfigs := []centrifuge.RedisShardConfig{
		{
			Address:  appConfig.Address,
			DB:       appConfig.DB,
			Password: appConfig.Password,
		},
	}
	var redisShards []*centrifuge.RedisShard
	for _, redisConf := range redisShardConfigs {
		log.Printf("Websocket redis broker config: %s/%d\n", redisConf.Address, redisConf.DB)

		redisShard, err := centrifuge.NewRedisShard(node, redisConf)
		if err != nil {
			log.Fatal(err)
		}
		redisShards = append(redisShards, redisShard)
	}

	broker, err := centrifuge.NewRedisBroker(node, centrifuge.RedisBrokerConfig{
		Shards: redisShards,
	})
	if err != nil {
		log.Fatal(err)
	}
	node.SetBroker(broker)
}

func (s *Server) Start() error {
	return s.node.Run()
}

func (s *Server) Shutdown() error {
	return s.node.Shutdown(context.Background())
}

func (s *Server) Publish(channel string, data []byte) error {
	_, err := s.node.Publish(channel, data)
	return err
}

func handleLog(entry centrifuge.LogEntry) {
	log.Printf("[%v] %s", entry.Level, entry.Message)
}
