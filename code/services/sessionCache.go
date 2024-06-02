package services

import (
	"strings"
	"time"

	myopenai "github.com/blacklee123/feishu-openai/services/openai"

	"github.com/pandodao/tokenizer-go"
	openai "github.com/sashabaranov/go-openai"

	"github.com/patrickmn/go-cache"
)

type SessionMode string
type VisionDetail string
type SessionService struct {
	cache *cache.Cache
}
type PicSetting struct {
	resolution Resolution
	style      PicStyle
}
type Resolution string
type PicStyle string

type SessionMeta struct {
	Mode         SessionMode                    `json:"mode"`
	Msg          []openai.ChatCompletionMessage `json:"msg,omitempty"`
	PicSetting   PicSetting                     `json:"pic_setting,omitempty"`
	AIMode       myopenai.AIMode                `json:"ai_mode,omitempty"`
	VisionDetail VisionDetail                   `json:"vision_detail,omitempty"`
}

type SessionServiceCacheInterface interface {
	GetMsg(sessionId string) []openai.ChatCompletionMessage
	SetMsg(sessionId string, msg []openai.ChatCompletionMessage)
	Clear(sessionId string)
}

var sessionServices *SessionService

func (s *SessionService) GetMsg(sessionId string) (msg []openai.ChatCompletionMessage) {
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return nil
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return sessionMeta.Msg
}

func (s *SessionService) SetMsg(sessionId string, msg []openai.ChatCompletionMessage) {
	maxLength := 4096
	maxCacheTime := time.Hour * 12

	//限制对话上下文长度
	for getStrPoolTotalLength(msg) > maxLength {
		msg = append(msg[:1], msg[2:]...)
	}

	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		sessionMeta := &SessionMeta{Msg: msg}
		s.cache.Set(sessionId, sessionMeta, maxCacheTime)
		return
	}
	sessionMeta := sessionContext.(*SessionMeta)
	sessionMeta.Msg = msg
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

func (s *SessionService) Clear(sessionId string) {
	// Delete the session context from the cache.
	s.cache.Delete(sessionId)
}

func GetSessionCache() SessionServiceCacheInterface {
	if sessionServices == nil {
		sessionServices = &SessionService{cache: cache.New(time.Hour*12, time.Hour*1)}
	}
	return sessionServices
}

func getStrPoolTotalLength(strPool []openai.ChatCompletionMessage) int {
	var total int
	for _, v := range strPool {
		total += CalculateTokenLength(v)
	}
	return total
}

func CalculateTokenLength(msg openai.ChatCompletionMessage) int {
	text := strings.TrimSpace(msg.Content)
	return tokenizer.MustCalToken(text)
}
