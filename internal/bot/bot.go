package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"barber-backend/internal/repository"
)

type tgUpdate struct {
	UpdateID int        `json:"update_id"`
	Message  *tgMessage `json:"message"`
}

type tgMessage struct {
	From *tgUser `json:"from"`
	Chat tgChat  `json:"chat"`
	Text string  `json:"text"`
}

type tgUser struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type Bot struct {
	token    string
	userRepo *repository.TelegramUserRepository
}

func New(token string, userRepo *repository.TelegramUserRepository) *Bot {
	return &Bot{token: token, userRepo: userRepo}
}

// Start запускает long-polling в фоновой горутине
func (b *Bot) Start(ctx context.Context) {
	go func() {
		offset := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			updates, err := b.getUpdates(offset)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			for _, u := range updates {
				b.handleUpdate(u)
				offset = u.UpdateID + 1
			}
			if len(updates) == 0 {
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

func (b *Bot) getUpdates(offset int) ([]tgUpdate, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=20", b.token, offset)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (b *Bot) handleUpdate(u tgUpdate) {
	if u.Message == nil {
		return
	}
	msg := u.Message
	if !strings.HasPrefix(strings.TrimSpace(msg.Text), "/start") {
		return
	}

	chatID := msg.Chat.ID

	if msg.From == nil || msg.From.Username == "" {
		b.send(chatID, "Чтобы получать напоминания, нужен username в Telegram.\nУстановите его в настройках и напишите /start снова.")
		return
	}

	username := strings.ToLower(msg.From.Username)
	if err := b.userRepo.Upsert(chatID, username); err != nil {
		fmt.Printf("[bot] upsert error: %v\n", err)
		b.send(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	b.send(chatID, fmt.Sprintf(
		"✅ Готово, %s!\n\nТеперь за 24 часа до каждого визита вы получите напоминание с датой, временем и адресом.\n\n📍 Большой Головин переулок, 3к2, 4 этаж, кабинет 13",
		msg.From.FirstName,
	))
}

func (b *Bot) send(chatID int64, text string) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("[bot] send error: %v\n", err)
		return
	}
	resp.Body.Close()
}
