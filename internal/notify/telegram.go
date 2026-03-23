package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"barber-backend/internal/model"
)

type Notifier struct {
	token  string
	chatID string
}

func NewNotifier() *Notifier {
	return &Notifier{
		token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		chatID: os.Getenv("TELEGRAM_MASTER_CHAT_ID"),
	}
}

func (n *Notifier) Enabled() bool {
	return n.token != "" && n.chatID != ""
}

func (n *Notifier) send(text string) {
	if !n.Enabled() {
		return
	}
	payload := map[string]interface{}{
		"chat_id":                  n.chatID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("[notify] telegram error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// ─── helpers ───────────────────────────────────────────────────────────────

var ruMonths = []string{"янв", "фев", "мар", "апр", "май", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"}
var ruDays = []string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"}

func fmtDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return fmt.Sprintf("%d %s (%s)", t.Day(), ruMonths[t.Month()-1], ruDays[t.Weekday()])
}

func fmtContacts(apt *model.Appointment) string {
	var parts []string
	if apt.Phone != "" {
		parts = append(parts, fmt.Sprintf("📱 %s", apt.Phone))
	}
	if apt.Telegram != "" {
		tg := apt.Telegram
		if !strings.HasPrefix(tg, "@") {
			tg = "@" + tg
		}
		handle := strings.TrimPrefix(tg, "@")
		parts = append(parts, fmt.Sprintf(`✈️ <a href="https://t.me/%s">%s</a>`, handle, tg))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ") + "\n"
}

func fmtComment(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("📝 <b>Комментарий:</b> %s\n", comment)
}

func fmtDuration(apt *model.Appointment) string {
	if apt.DurationMin > 0 && apt.DurationMin != 60 {
		return fmt.Sprintf("⏱ <b>Длительность:</b> %d мин\n", apt.DurationMin)
	}
	return ""
}

func fmtPrice(apt *model.Appointment) string {
	if apt.Price <= 0 {
		return ""
	}
	return fmt.Sprintf("💰 <b>Цена:</b> %s ₽\n", fmtNum(apt.Price))
}

func fmtNum(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var result []byte
	for i, c := range []byte(s) {
		pos := len(s) - i
		if i > 0 && pos%3 == 0 {
			result = append(result, ' ')
		}
		result = append(result, c)
	}
	return string(result)
}

const sep = "──────────────────\n"

// ─── notifications ─────────────────────────────────────────────────────────

// NotifyNewBooking — новая запись от клиента (isNew = первый раз у мастера).
func (n *Notifier) NotifyNewBooking(apt *model.Appointment, isNew bool, clientComment string) {
	header := "📋 <b>НОВАЯ ЗАПИСЬ</b>"
	if isNew {
		header = "🆕 <b>НОВЫЙ КЛИЕНТ</b>"
	}

	text := fmt.Sprintf(
		"%s\n"+sep+
			"👤 <b>Имя:</b> %s\n"+
			"%s"+
			sep+
			"✂️ <b>Услуга:</b> %s\n"+
			"📅 <b>Дата:</b> %s\n"+
			"🕐 <b>Время:</b> %s\n"+
			"%s"+
			"%s"+
			"%s",
		header,
		apt.ClientName,
		fmtContacts(apt),
		apt.Service,
		fmtDate(apt.Date),
		apt.Time,
		fmtDuration(apt),
		fmtPrice(apt),
		fmtComment(clientComment),
	)
	n.send(text)
}

// NotifyRescheduled — запись перенесена (мастер или клиент изменили дату/время).
func (n *Notifier) NotifyRescheduled(apt *model.Appointment, oldDate, oldTime string, clientComment string) {
	text := fmt.Sprintf(
		"🔄 <b>ПЕРЕНОС ЗАПИСИ</b>\n"+sep+
			"👤 <b>Клиент:</b> %s\n"+
			"%s"+
			"✂️ <b>Услуга:</b> %s\n"+
			sep+
			"📅 <b>Было:</b>  %s, %s\n"+
			"📅 <b>Стало:</b> %s, %s\n"+
			"%s",
		apt.ClientName,
		fmtContacts(apt),
		apt.Service,
		fmtDate(oldDate), oldTime,
		fmtDate(apt.Date), apt.Time,
		fmtComment(clientComment),
	)
	n.send(text)
}

// NotifyCancelled — запись отменена.
func (n *Notifier) NotifyCancelled(apt *model.Appointment, clientComment string) {
	text := fmt.Sprintf(
		"❌ <b>ОТМЕНА ЗАПИСИ</b>\n"+sep+
			"👤 <b>Клиент:</b> %s\n"+
			"%s"+
			"✂️ <b>Услуга:</b> %s\n"+
			"📅 %s · %s\n"+
			"%s"+
			"<i>Запись на %s отменена</i>",
		apt.ClientName,
		fmtContacts(apt),
		apt.Service,
		fmtDate(apt.Date), apt.Time,
		fmtComment(clientComment),
		apt.Time,
	)
	n.send(text)
}

// NotifyLate — клиент опаздывает.
func (n *Notifier) NotifyLate(apt *model.Appointment, lateMin int) {
	text := fmt.Sprintf(
		"⏰ <b>КЛИЕНТ ОПАЗДЫВАЕТ</b>\n"+sep+
			"👤 <b>Клиент:</b> %s\n"+
			"%s"+
			"✂️ <b>Услуга:</b> %s\n"+
			"📅 %s · %s\n"+
			sep+
			"⏱ Опаздывает на <b>%d мин</b>",
		apt.ClientName,
		fmtContacts(apt),
		apt.Service,
		fmtDate(apt.Date), apt.Time,
		lateMin,
	)
	n.send(text)
}

// NotifyCompleted — запись завершена.
func (n *Notifier) NotifyCompleted(apt *model.Appointment) {
	var fin string
	if apt.Price > 0 {
		fin = sep + fmt.Sprintf("💰 <b>Оплата:</b> %s ₽\n", fmtNum(apt.Price))
		if apt.Tips > 0 {
			fin += fmt.Sprintf("🎁 <b>Чаевые:</b> %s ₽\n", fmtNum(apt.Tips))
		}
	}
	text := fmt.Sprintf(
		"✅ <b>ЗАПИСЬ ЗАВЕРШЕНА</b>\n"+sep+
			"👤 <b>Клиент:</b> %s\n"+
			"✂️ <b>Услуга:</b> %s\n"+
			"📅 %s · %s\n"+
			"%s",
		apt.ClientName,
		apt.Service,
		fmtDate(apt.Date), apt.Time,
		fin,
	)
	n.send(text)
}
