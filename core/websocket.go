package core

// WebSocketUpgrader is the interface that adapters implement
// to provide WebSocket upgrade capabilities.
//
// NestGo does not bundle a WebSocket library — it provides the abstraction.
// Use gorilla/websocket with Gin or fiber/websocket with Fiber via the
// Underlying() escape hatch, or implement this interface for adapter-agnostic code.
//
// Example with gorilla/websocket (Gin):
//
//	func (ctrl *ChatController) RegisterRoutes(r core.Router) {
//	    r.GET("/ws", ctrl.HandleWS)
//	}
//
//	func (ctrl *ChatController) HandleWS(c core.Context) error {
//	    if !c.IsWebSocket() {
//	        return core.ErrBadRequest("expected websocket upgrade")
//	    }
//	    // Escape to adapter-specific WebSocket handling
//	    return ctrl.upgradeWebSocket(c)
//	}
//
// For Gin (gorilla/websocket):
//
//	func (ctrl *ChatController) upgradeWebSocket(c core.Context) error {
//	    ginCtx := c.Underlying().(*gin.Context)
//	    conn, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
//	    if err != nil {
//	        return core.ErrInternalServer("websocket upgrade failed")
//	    }
//	    defer conn.Close()
//	    // ... handle messages
//	    return nil
//	}
//
// For Fiber (fiber/websocket):
//
//	import "github.com/gofiber/contrib/websocket"
//
//	// Register using Fiber's native websocket handler via Underlying()
//	fiberApp := server.Underlying().(*fiber.App)
//	fiberApp.Get("/ws", websocket.New(func(c *websocket.Conn) {
//	    // ... handle messages
//	}))

// WebSocketHandler is the handler signature for WebSocket connections.
// Implement this in your controller to handle WebSocket messages.
type WebSocketHandler func(c Context) error

// IsWebSocketRequest checks if the context represents a WebSocket upgrade request.
// This is a convenience function that wraps c.IsWebSocket() with additional checks.
func IsWebSocketRequest(c Context) bool {
	if !c.IsWebSocket() {
		return false
	}
	// Verify required headers for a valid WebSocket upgrade
	connection := c.GetHeader("Connection")
	if connection == "" {
		return false
	}
	return true
}
