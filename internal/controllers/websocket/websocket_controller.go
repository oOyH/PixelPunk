package websocket

import (
	"net/http"
	"net/url"
	"pixelpunk/internal/services/auth"
	ws "pixelpunk/internal/websocket"
	"pixelpunk/pkg/common"
	"pixelpunk/pkg/errors"
	"pixelpunk/pkg/logger"
	"pixelpunk/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return isWebSocketOriginAllowed(r)
		},
	}

	globalManager *ws.Manager
)

func isWebSocketOriginAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		// Non-browser clients typically won't set Origin.
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	originHost := strings.ToLower(originURL.Hostname())
	if originHost == "" {
		return false
	}

	requestHost := strings.ToLower(hostnameFromHostPort(r.Host))
	if requestHost != "" && requestHost == originHost {
		return true
	}

	baseURL := strings.TrimSpace(utils.GetBaseUrl())
	if baseURL != "" {
		if baseParsed, err := url.Parse(baseURL); err == nil {
			baseHost := strings.ToLower(baseParsed.Hostname())
			if baseHost != "" && baseHost == originHost {
				return true
			}
		}
	}

	isLocal := func(h string) bool { return h == "localhost" || h == "127.0.0.1" }
	if isLocal(originHost) && isLocal(requestHost) {
		return true
	}

	return false
}

func hostnameFromHostPort(hostport string) string {
	hostport = strings.TrimSpace(hostport)
	if hostport == "" {
		return ""
	}
	// IPv6: [::1]:9520
	if strings.HasPrefix(hostport, "[") {
		if idx := strings.Index(hostport, "]"); idx > 1 {
			return hostport[1:idx]
		}
	}
	if idx := strings.Index(hostport, ":"); idx > 0 {
		return hostport[:idx]
	}
	return hostport
}

func InitWebSocketManager() {
	config := ws.DefaultConfig()
	globalManager = ws.NewManager(config)
	globalManager.Start()
}

func GetWebSocketManager() *ws.Manager {
	return globalManager
}

func HandleWebSocket(c *gin.Context) {
	claims, exists := c.Get("payload")
	if !exists {
		errors.HandleError(c, errors.New(errors.CodeUnauthorized, "User payload not found"))
		return
	}

	jwtClaims, ok := claims.(*auth.JWTClaims)
	if !ok {
		errors.HandleError(c, errors.New(errors.CodeInvalidRequest, "Invalid user payload format"))
		return
	}

	if jwtClaims.Role != common.UserRoleAdmin && jwtClaims.Role != common.UserRoleSuperAdmin {
		errors.HandleError(c, errors.New(errors.CodeForbidden, "Admin permission required"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed: %v", err)
		return
	}

	isAdmin := jwtClaims.Role == common.UserRoleAdmin || jwtClaims.Role == common.UserRoleSuperAdmin
	client := ws.NewClient(conn, jwtClaims.UserID, isAdmin)

	globalManager.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump(globalManager)
}

func BroadcastMessage(msgType ws.MessageType, data interface{}) {
	if globalManager == nil {
		return
	}

	msg := ws.NewMessage(msgType, data)
	globalManager.BroadcastMessage(msg)
}

func BroadcastToAdmins(msgType ws.MessageType, data interface{}) {
	if globalManager == nil {
		return
	}

	msg := ws.NewMessage(msgType, data)
	globalManager.SendToAdmins(msg)

}

func SendToClient(clientID string, msgType ws.MessageType, data interface{}) error {
	if globalManager == nil {
		return errors.New(errors.CodeInternal, "WebSocket manager not initialized")
	}

	msg := ws.NewMessage(msgType, data)
	return globalManager.SendToClient(clientID, msg)
}

func GetStats(c *gin.Context) {
	if globalManager == nil {
		errors.HandleError(c, errors.New(errors.CodeInternal, "WebSocket manager not initialized"))
		return
	}

	stats := globalManager.GetStats()
	errors.ResponseSuccess(c, stats, "Get stats successfully")
}
