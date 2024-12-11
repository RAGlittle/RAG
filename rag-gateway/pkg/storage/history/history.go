package history

import (
	"context"
	"encoding/json"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type ChatModel struct {
	History []llms.ChatMessageModel `json:"history,omitempty"`
}

type kvChatHistory struct {
	kv     storage.KeyValueStore
	chatId string
}

func NewKVChatHistory(kv storage.KeyValueStore, chatId string) schema.ChatMessageHistory {
	return &kvChatHistory{
		kv:     kv,
		chatId: chatId,
	}
}

func (k *kvChatHistory) createChat(ctx context.Context, message llms.ChatMessage) error {
	data := &ChatModel{
		History: []llms.ChatMessageModel{
			{
				Type: string(message.GetType()),
				Data: llms.ChatMessageModelData{
					Content: message.GetContent(),
					Type:    string(message.GetType()),
				},
			},
		},
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return k.kv.Put(ctx, k.chatId, raw, storage.WithRevision(0))
}

// AddMessage adds a message to the store.
func (k *kvChatHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	// TODO : retry logic for conflicts, in case of concurrent updates

	var revision int64
	res, err := k.kv.Get(ctx, k.chatId, storage.WithRevisionOut(&revision))
	if storage.IsNotFound(err) {
		return k.createChat(ctx, message)
	}

	model := &ChatModel{}
	if err := json.Unmarshal(res, model); err != nil {
		return err
	}

	model.History = append(model.History, llms.ChatMessageModel{
		Type: string(message.GetType()),
		Data: llms.ChatMessageModelData{
			Content: message.GetContent(),
			Type:    string(message.GetType()),
		},
	})

	raw, err := json.Marshal(model)
	if err != nil {
		return err
	}

	return k.kv.Put(ctx, k.chatId, raw, storage.WithRevision(revision))
}

// AddUserMessage is a convenience method for adding a human message string
// to the store.
func (k *kvChatHistory) AddUserMessage(ctx context.Context, message string) error {
	return k.AddMessage(ctx, llms.HumanChatMessage{
		Content: message,
	})
}

// AddAIMessage is a convenience method for adding an AI message string to
// the store.
func (k *kvChatHistory) AddAIMessage(ctx context.Context, message string) error {
	return k.AddMessage(ctx, llms.AIChatMessage{
		Content: message,
	})
}

// Clear removes all messages from the store.
func (k *kvChatHistory) Clear(ctx context.Context) error {
	return k.kv.Delete(ctx, k.chatId)
}

// Messages retrieves all messages from the store
func (k *kvChatHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	res, err := k.kv.Get(ctx, k.chatId)
	if storage.IsNotFound(err) {
		return []llms.ChatMessage{}, nil
	} else if err != nil {
		return nil, err
	}
	model := &ChatModel{}
	if err := json.Unmarshal(res, model); err != nil {
		return nil, err
	}
	messages := []llms.ChatMessage{}
	for _, m := range model.History {
		switch m.Type {
		case string(llms.ChatMessageTypeHuman):
			messages = append(messages, llms.HumanChatMessage{
				Content: m.Data.Content,
			})
		case string(llms.ChatMessageTypeAI):
			messages = append(messages, llms.AIChatMessage{
				Content: m.Data.Content,
			})
		}
	}
	return messages, nil
}

// SetMessages replaces existing messages in the store
func (k *kvChatHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	panic("implement me")
}
