package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bbs-go/cache"
	"bbs-go/model/constants"
	"bbs-go/pkg/config"

	"github.com/google/uuid"
	"github.com/mediocregopher/radix/v3"
	"github.com/sirupsen/logrus"
)

const (
	presenceTTL    = 75 * time.Second
	ticketTTL      = 30 * time.Second
	presencePrefix = "bbs:presence:"
	// maxOnlineUsersReturned 限制单次快照中携带的用户数量，避免在线人数很多时
	// 每个 WebSocket 客户端都收到过大的消息；真实在线人数通过 Total 单独返回。
	maxOnlineUsersReturned = 200
)

var PresenceService = &presenceService{}

type OnlineUser struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Greeting string `json:"greeting"`
}

// OnlineSnapshot 是在线名单的推送内容。Users 最多包含 maxOnlineUsersReturned
// 条记录，Total 是 Redis 中真实存在的在线用户总数，用于前端显示"还有 N 人在线"。
type OnlineSnapshot struct {
	Users []OnlineUser `json:"users"`
	Total int          `json:"total"`
}

type presenceService struct {
	mu              sync.RWMutex
	initMu          sync.Mutex
	client          radix.Client
	lastInitAttempt time.Time
}

func (s *presenceService) IsActiveUser(userId int64) bool {
	user := cache.UserCache.Get(userId)
	return user != nil && user.Status == constants.StatusOk
}

func (s *presenceService) Init() {
	s.initMu.Lock()
	defer s.initMu.Unlock()
	if s.redis() != nil || time.Since(s.lastInitAttempt) < 10*time.Second {
		return
	}
	s.lastInitAttempt = time.Now()
	if config.Instance.Redis.Url == "" {
		logrus.Warn("Redis is not configured; online presence is disabled")
		return
	}

	dialOpts := make([]radix.DialOpt, 0, 2)
	if config.Instance.Redis.Password != "" {
		dialOpts = append(dialOpts, radix.DialAuthPass(config.Instance.Redis.Password))
	}
	if config.Instance.Redis.DB != 0 {
		dialOpts = append(dialOpts, radix.DialSelectDB(config.Instance.Redis.DB))
	}
	pool, err := radix.NewPool("tcp", config.Instance.Redis.Url, 12, radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr, dialOpts...)
	}))
	if err != nil {
		logrus.WithError(err).Warn("Redis unavailable; online presence is disabled")
		return
	}
	if err = pool.Do(radix.Cmd(nil, "PING")); err != nil {
		_ = pool.Close()
		logrus.WithError(err).Warn("Redis unavailable; online presence is disabled")
		return
	}
	s.mu.Lock()
	s.client = pool
	s.mu.Unlock()
}

func (s *presenceService) redis() radix.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client
}

func (s *presenceService) Available() bool {
	if s.redis() == nil {
		s.Init()
	}
	return s.redis() != nil
}

func (s *presenceService) IssueTicket(userId int64, ip string) (string, error) {
	client := s.redis()
	if client == nil {
		return "", errors.New("当前在线功能暂不可用")
	}
	ticket := uuid.NewString()
	value := strconv.FormatInt(userId, 10) + "|" + ipHash(ip)
	var result string
	if err := client.Do(radix.Cmd(&result, "SET", presencePrefix+"ticket:"+ticket, value, "NX", "EX", strconv.Itoa(int(ticketTTL.Seconds())))); err != nil || result != "OK" {
		return "", errors.New("当前在线功能暂不可用")
	}
	return ticket, nil
}

func (s *presenceService) ConsumeTicket(ticket, ip string) (int64, error) {
	client := s.redis()
	if client == nil || ticket == "" {
		return 0, errors.New("无效的连接凭证")
	}
	const consumeScript = `local v=redis.call('GET',KEYS[1]); if v then redis.call('DEL',KEYS[1]) end; return v`
	var value string
	if err := client.Do(radix.Cmd(&value, "EVAL", consumeScript, "1", presencePrefix+"ticket:"+ticket)); err != nil || value == "" {
		return 0, errors.New("连接凭证已失效")
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 || parts[1] != ipHash(ip) {
		return 0, errors.New("连接凭证与客户端不匹配")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func (s *presenceService) AddConnection(userId int64, ip string) (string, error) {
	client := s.redis()
	if client == nil {
		return "", errors.New("当前在线功能暂不可用")
	}
	connectionId := uuid.NewString()
	now := time.Now().Unix()
	expires := now + int64(presenceTTL.Seconds())
	userKey := presencePrefix + "user:" + strconv.FormatInt(userId, 10)
	ipKey := presencePrefix + "ip:" + ipHash(ip)
	limits := config.Instance.Presence
	const addScript = `
redis.call('ZREMRANGEBYSCORE',KEYS[1],'-inf',ARGV[1])
redis.call('ZREMRANGEBYSCORE',KEYS[2],'-inf',ARGV[1])
redis.call('ZREMRANGEBYSCORE',KEYS[3],'-inf',ARGV[1])
if redis.call('ZCARD',KEYS[1]) >= tonumber(ARGV[3]) then return -1 end
if redis.call('ZCARD',KEYS[2]) >= tonumber(ARGV[4]) then return -2 end
if redis.call('ZCARD',KEYS[3]) >= tonumber(ARGV[5]) then return -3 end
redis.call('ZADD',KEYS[1],ARGV[2],ARGV[6])
redis.call('ZADD',KEYS[2],ARGV[2],ARGV[6])
redis.call('ZADD',KEYS[3],ARGV[2],ARGV[6])
redis.call('ZADD',KEYS[4],ARGV[2],ARGV[7])
redis.call('EXPIRE',KEYS[2],ARGV[8])
redis.call('EXPIRE',KEYS[3],ARGV[8])
redis.call('EXPIRE',KEYS[4],ARGV[8])
return 1`
	var result int
	err := client.Do(radix.Cmd(&result, "EVAL", addScript, "4",
		presencePrefix+"connections", userKey, ipKey, presencePrefix+"users",
		strconv.FormatInt(now, 10), strconv.FormatInt(expires, 10), strconv.Itoa(limits.MaxConnections),
		strconv.Itoa(limits.MaxConnectionsPerUser), strconv.Itoa(limits.MaxConnectionsPerIP), connectionId,
		strconv.FormatInt(userId, 10), strconv.Itoa(int((presenceTTL * 2).Seconds()))))
	if err != nil {
		return "", errors.New("当前在线功能暂不可用")
	}
	switch result {
	case -1:
		return "", errors.New("在线连接数已达上限")
	case -2:
		return "", errors.New("该账号的在线连接数已达上限")
	case -3:
		return "", errors.New("该网络地址的在线连接数已达上限")
	}
	return connectionId, nil
}

func (s *presenceService) Refresh(userId int64, ip, connectionId string) error {
	client := s.redis()
	if client == nil {
		return errors.New("Redis unavailable")
	}
	expires := time.Now().Add(presenceTTL).Unix()
	userIdString := strconv.FormatInt(userId, 10)
	commands := radix.Pipeline(
		radix.Cmd(nil, "ZADD", presencePrefix+"connections", strconv.FormatInt(expires, 10), connectionId),
		radix.Cmd(nil, "ZADD", presencePrefix+"user:"+userIdString, strconv.FormatInt(expires, 10), connectionId),
		radix.Cmd(nil, "ZADD", presencePrefix+"ip:"+ipHash(ip), strconv.FormatInt(expires, 10), connectionId),
		radix.Cmd(nil, "ZADD", presencePrefix+"users", strconv.FormatInt(expires, 10), userIdString),
		radix.Cmd(nil, "EXPIRE", presencePrefix+"user:"+userIdString, strconv.Itoa(int((presenceTTL*2).Seconds()))),
		radix.Cmd(nil, "EXPIRE", presencePrefix+"ip:"+ipHash(ip), strconv.Itoa(int((presenceTTL*2).Seconds()))),
		radix.Cmd(nil, "EXPIRE", presencePrefix+"users", strconv.Itoa(int((presenceTTL*2).Seconds()))),
		radix.Cmd(nil, "EXPIRE", presencePrefix+"connections", strconv.Itoa(int((presenceTTL*2).Seconds()))),
	)
	return client.Do(commands)
}

func (s *presenceService) Remove(userId int64, ip, connectionId string) {
	client := s.redis()
	if client == nil {
		return
	}
	userIdString := strconv.FormatInt(userId, 10)
	userKey := presencePrefix + "user:" + userIdString
	const removeScript = `redis.call('ZREM',KEYS[1],ARGV[1]); redis.call('ZREM',KEYS[2],ARGV[1]); redis.call('ZREM',KEYS[3],ARGV[1]); redis.call('ZREMRANGEBYSCORE',KEYS[2],'-inf',ARGV[2]); if redis.call('ZCARD',KEYS[2]) == 0 then redis.call('ZREM',KEYS[4],ARGV[3]) end; return 1`
	_ = client.Do(radix.Cmd(nil, "EVAL", removeScript, "4", presencePrefix+"connections", userKey,
		presencePrefix+"ip:"+ipHash(ip), presencePrefix+"users", connectionId, strconv.FormatInt(time.Now().Unix(), 10), userIdString))
}

func (s *presenceService) OnlineUsers() (OnlineSnapshot, error) {
	client := s.redis()
	if client == nil {
		return OnlineSnapshot{}, errors.New("Redis unavailable")
	}
	now := strconv.FormatInt(time.Now().Unix(), 10)
	if err := client.Do(radix.Cmd(nil, "ZREMRANGEBYSCORE", presencePrefix+"users", "-inf", now)); err != nil {
		return OnlineSnapshot{}, err
	}
	var total int
	if err := client.Do(radix.Cmd(&total, "ZCARD", presencePrefix+"users")); err != nil {
		return OnlineSnapshot{}, err
	}
	var ids []string
	if err := client.Do(radix.FlatCmd(&ids, "ZRANGEBYSCORE", presencePrefix+"users", now, "+inf", "LIMIT", "0", strconv.Itoa(maxOnlineUsersReturned))); err != nil {
		return OnlineSnapshot{}, err
	}
	users := make([]OnlineUser, 0, len(ids))
	for _, idString := range ids {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			continue
		}
		user := cache.UserCache.Get(id)
		if user != nil && user.Status == constants.StatusOk {
			users = append(users, OnlineUser{Id: user.Id, Nickname: user.Nickname, Avatar: user.Avatar, Greeting: user.Greeting})
		}
	}
	sort.Slice(users, func(i, j int) bool { return users[i].Id < users[j].Id })
	return OnlineSnapshot{Users: users, Total: total}, nil
}

func ipHash(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:12])
}
