package api

import (
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"bbs-go/pkg/errs"
	"bbs-go/services"

	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/mlogclub/simple/web"
)

type PresenceController struct {
	Ctx iris.Context
}

func (c *PresenceController) PostTicket() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	if !services.PresenceService.Available() {
		return web.JsonErrorMsg("当前在线功能暂不可用")
	}
	ticket, err := services.PresenceService.IssueTicket(user.Id, clientIP(c.Ctx.Request()))
	if err != nil {
		return web.JsonErrorMsg(err.Error())
	}
	return web.JsonData(map[string]string{"ticket": ticket})
}

func (c *PresenceController) GetOnline() *web.JsonResult {
	if services.UserTokenService.GetCurrent(c.Ctx) == nil {
		return web.JsonError(errs.NotLogin)
	}
	snapshot, err := services.PresenceService.OnlineUsers()
	if err != nil {
		return web.JsonData(services.OnlineSnapshot{Users: []services.OnlineUser{}})
	}
	return web.JsonData(snapshot)
}

func RegisterPresence(app *mvc.Application) {
	app.Handle(new(PresenceController))
}

var presenceUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		for _, allowed := range services.AllowedPresenceOrigins() {
			if strings.EqualFold(strings.TrimRight(origin, "/"), strings.TrimRight(allowed, "/")) {
				return true
			}
		}
		return false
	},
}

func PresenceWebSocket(ctx iris.Context) {
	if !services.PresenceService.Available() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		return
	}
	if !presenceUpgrader.CheckOrigin(ctx.Request()) {
		ctx.StatusCode(http.StatusForbidden)
		return
	}
	ip := clientIP(ctx.Request())
	userId, err := services.PresenceService.ConsumeTicket(ctx.URLParam("ticket"), ip)
	if err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	user := services.UserTokenService.GetCurrent(ctx)
	if user == nil || user.Id != userId || !services.PresenceService.IsActiveUser(userId) {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	connectionId, err := services.PresenceService.AddConnection(userId, ip)
	if err != nil {
		ctx.StatusCode(http.StatusTooManyRequests)
		return
	}
	conn, err := presenceUpgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		services.PresenceService.Remove(userId, ip, connectionId)
		return
	}
	defer conn.Close()
	defer services.PresenceService.Remove(userId, ip, connectionId)
	client := &presenceClient{conn: conn, snapshots: make(chan *services.OnlineSnapshot, 1), done: make(chan struct{})}
	presenceBroadcastHub.add(client)
	defer presenceBroadcastHub.remove(client)
	go client.write(userId, ip, connectionId)
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(45 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(45 * time.Second))
		return services.PresenceService.Refresh(userId, ip, connectionId)
	})
	for {
		if _, _, err = conn.ReadMessage(); err != nil {
			return
		}
		_ = services.PresenceService.Refresh(userId, ip, connectionId)
	}
}

type presenceClient struct {
	conn      *websocket.Conn
	snapshots chan *services.OnlineSnapshot
	done      chan struct{}
}

func (c *presenceClient) write(userId int64, ip, connectionId string) {
	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()
	for {
		select {
		case snapshot := <-c.snapshots:
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			message := map[string]interface{}{"type": "unavailable"}
			if snapshot != nil {
				users := snapshot.Users
				if users == nil {
					users = []services.OnlineUser{}
				}
				message = map[string]interface{}{"type": "snapshot", "users": users, "total": snapshot.Total}
			}
			if err := c.conn.WriteJSON(message); err != nil {
				_ = c.conn.Close()
				return
			}
		case <-pingTicker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				_ = c.conn.Close()
				return
			}
			_ = services.PresenceService.Refresh(userId, ip, connectionId)
		case <-c.done:
			return
		}
	}
}

type presenceHub struct {
	sync.RWMutex
	clients map[*presenceClient]struct{}
	once    sync.Once
	last    services.OnlineSnapshot
}

var presenceBroadcastHub = &presenceHub{clients: make(map[*presenceClient]struct{})}

func (h *presenceHub) add(client *presenceClient) {
	h.Lock()
	h.clients[client] = struct{}{}
	h.Unlock()
	h.once.Do(func() { go h.run() })
	h.broadcast(true)
}

func (h *presenceHub) remove(client *presenceClient) {
	h.Lock()
	if _, exists := h.clients[client]; exists {
		delete(h.clients, client)
		close(client.done)
	}
	h.Unlock()
}

func (h *presenceHub) run() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		h.broadcast(false)
	}
}

func (h *presenceHub) broadcast(force bool) {
	h.RLock()
	if len(h.clients) == 0 {
		h.RUnlock()
		return
	}
	h.RUnlock()
	snapshot, err := services.PresenceService.OnlineUsers()
	if err != nil {
		h.RLock()
		for client := range h.clients {
			select {
			case client.snapshots <- nil:
			default:
			}
		}
		h.RUnlock()
		return
	}
	h.Lock()
	defer h.Unlock()
	if !force && reflect.DeepEqual(h.last, snapshot) {
		return
	}
	h.last = snapshot
	for client := range h.clients {
		current := snapshot
		select {
		case client.snapshots <- &current:
		default:
		}
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if services.IsTrustedProxy(host) {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			candidate := strings.TrimSpace(strings.Split(forwarded, ",")[0])
			if net.ParseIP(candidate) != nil {
				return candidate
			}
		}
	}
	return host
}
