/*
Chat server is response for storing and retrieving chat histories, on a per-tenant basis
*/
package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/gateway"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/history"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"github.com/gin-gonic/gin"
)

const chatParam = "chatID"

type chatServer struct {
	util.Initializer
	chatKv storage.KeyValueStore
}

func NewChatServer() gateway.HTTPApiExtension {
	return &chatServer{}
}

func (cs *chatServer) apiRoute(route string) string {
	return path.Join("/chat/api/v1alpha1", route)
}

func (cs *chatServer) ConfigureRoutes(router *gin.Engine, cfg gateway.GatewayRuntimeConfig) {
	router.GET(cs.apiRoute("list"), cs.list)
	router.GET(cs.apiRoute(fmt.Sprintf("get/:%s", chatParam)), cs.get)
	router.DELETE(cs.apiRoute(fmt.Sprintf("delete/:%s", chatParam)), cs.delete)
	cs.Initialize(cfg)
}

func (cs *chatServer) Initialize(cfg gateway.GatewayRuntimeConfig) {
	cs.InitOnce(func() {
		cs.chatKv = cfg.ChatKV
	})
}

func (cs *chatServer) list(c *gin.Context) {
	if !cs.Initialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat server not initialized"})
		return
	}
	chats, err := cs.chatKv.ListKeys(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
	})
}

func (cs *chatServer) get(c *gin.Context) {
	if !cs.Initialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat server not initialized"})
		return
	}

	uri := c.Param(chatParam)
	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing chatID"})
		return
	}

	chat, err := cs.chatKv.Get(c.Request.Context(), uri)
	if storage.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	model := &history.ChatModel{}
	if err := json.Unmarshal(chat, model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model)
}

func (cs *chatServer) delete(c *gin.Context) {
	if !cs.Initialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat server not initialized"})
	}

	uri := c.Param(chatParam)
	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing chatID"})
		return
	}
	if err := cs.chatKv.Delete(c.Request.Context(), uri); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "chat deleted"})
}
