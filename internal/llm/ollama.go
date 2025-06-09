package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Клиент для запросов к Ollama
type Client interface {
	// Возвращает true, если сентимент положительный
	EvaluateSentiment(ctx context.Context, text string) (bool, error)
}

type client struct {
	baseURL string
	model   string
	h       *http.Client
}

// Создаёт клиента. baseURL вроде "http://localhost:11434"
// Если httpClient не передан — используется стандартный с таймаутом 10s
func New(baseURL, model string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &client{baseURL: strings.TrimRight(baseURL, "/"), model: model, h: httpClient}
}

// ----- структуры запроса/ответа ------------------------------------------------

type chatReq struct {
	Model    string    `json:"model"`
	Messages []chatMsg `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
}

// --------------------------------------------------------------------------------

// Отправляет текст модели и анализирует ответ.
// Ожидается одно слово: "positive"/"negative" на любом языке.
func (c *client) EvaluateSentiment(ctx context.Context, text string) (bool, error) {
	prompt := "Определи сентимент отзыва как positive или negative. Ответь одним словом. Отзыв: " + text
	reqBody, _ := json.Marshal(chatReq{
		Model:  c.model,
		Stream: false,
		Messages: []chatMsg{{
			Role:    "user",
			Content: prompt,
		}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.h.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return false, errors.New("ollama: ошибка статуса: " + resp.Status + " " + string(data))
	}

	var out chatResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	if len(out.Choices) == 0 {
		return false, errors.New("ollama: пустой ответ")
	}
	answer := strings.ToLower(out.Choices[0].Message.Content)
	return strings.Contains(answer, "positive") || strings.Contains(answer, "полож"), nil
}
