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
	token          string
	chatID         string
	masterUsername string
	siteURL        string
}

func NewNotifier() *Notifier {
	return &Notifier{
		token:          os.Getenv("TELEGRAM_BOT_TOKEN"),
		chatID:         os.Getenv("TELEGRAM_MASTER_CHAT_ID"),
		masterUsername: os.Getenv("TELEGRAM_MASTER_USERNAME"),
		siteURL:        os.Getenv("SITE_URL"),
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

func fmtBookingRequest(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("💬 <b>Запрос клиента:</b> %s\n", comment)
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
		fmtBookingRequest(apt.Comment),
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

// NotifyIndividualRequest — клиент хочет записаться на индивидуальное время, указал предпочтительные дни.
func (n *Notifier) NotifyIndividualRequest(apt *model.Appointment, isNew bool, clientComment string) {
	header := "📋 <b>ЗАПРОС НА ЗАПИСЬ</b>"
	if isNew {
		header = "🆕 <b>НОВЫЙ КЛИЕНТ · ЗАПРОС НА ЗАПИСЬ</b>"
	}

	// Format comma-separated dates
	dates := apt.Date
	rawDates := strings.Split(apt.Date, ",")
	if len(rawDates) > 0 {
		var formatted []string
		for _, d := range rawDates {
			d = strings.TrimSpace(d)
			formatted = append(formatted, fmtDate(d))
		}
		dates = strings.Join(formatted, ", ")
	}

	text := fmt.Sprintf(
		"%s\n"+sep+
			"👤 <b>Имя:</b> %s\n"+
			"%s"+
			sep+
			"✂️ <b>Услуга:</b> %s\n"+
			"📅 <b>Желаемые дни:</b> %s\n"+
			"%s"+
			"%s"+
			"⚡ <i>Нужно согласовать время как можно скорее</i>",
		header,
		apt.ClientName,
		fmtContacts(apt),
		apt.Service,
		dates,
		fmtBookingRequest(apt.Comment),
		fmtComment(clientComment),
	)
	n.send(text)
}

// SendToUser отправляет сообщение конкретному клиенту по chat_id
func (n *Notifier) SendToUser(chatID int64, text string) {
	if n.token == "" {
		return
	}
	payload := map[string]interface{}{
		"chat_id":                  chatID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("[notify] send to user error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// NotifyClientReminder — напоминание клиенту за 24ч до записи
func (n *Notifier) NotifyClientReminder(chatID int64, apt *model.Appointment) {
	contactLine := ""
	if n.masterUsername != "" {
		handle := strings.TrimPrefix(n.masterUsername, "@")
		contactLine = fmt.Sprintf("\n📩 <a href=\"https://t.me/%s\">Написать мастеру в Telegram</a>", handle)
	}
	siteLink := ""
	if n.siteURL != "" {
		siteLink = fmt.Sprintf("\n🌐 <a href=\"%s\">Отменить или перенести запись на сайте</a>", n.siteURL)
	}
	text := fmt.Sprintf(
		"⏰ <b>Напоминание о визите завтра</b>\n"+sep+
			"✂️ <b>Услуга:</b> %s\n"+
			"📅 <b>Дата:</b> %s\n"+
			"🕐 <b>Время:</b> %s\n"+
			sep+
			"📍 Большой Головин переулок, 3к2\n"+
			"4 этаж, кабинет 13\n"+
			sep+
			"<i>Ждём вас! Если планы изменились — свяжитесь с мастером заранее.</i>"+
			"%s%s",
		apt.Service,
		fmtDate(apt.Date),
		apt.Time,
		contactLine,
		siteLink,
	)
	n.SendToUser(chatID, text)
}

// ─── daily summary ──────────────────────────────────────────────────────────

// DailySummaryItem — данные по одной записи для вечерней сводки мастеру.
type DailySummaryItem struct {
	Apt          model.Appointment
	ClientCard   *model.Client   // карточка клиента (может быть nil)
	PastComments []string        // master_comment из прошлых завершённых визитов (последние 3)
	LowSupplies  []model.Supply  // расходники под услугу с низким остатком
}

// NotifyDailySummary отправляет сводку по записям на следующий день.
func (n *Notifier) NotifyDailySummary(date string, items []DailySummaryItem, globalLowStock []model.Supply) {
	if !n.Enabled() {
		return
	}
	if len(items) == 0 && len(globalLowStock) == 0 {
		return
	}

	header := fmt.Sprintf("📅 <b>СВОДКА НА %s</b>\n", fmtDate(date))
	if len(items) == 0 {
		n.send(header + sep + "Записей нет.")
		return
	}

	// Счётчик слово «запись»
	wordZapis := "записей"
	switch len(items) {
	case 1:
		wordZapis = "запись"
	case 2, 3, 4:
		wordZapis = "записи"
	}
	msg := fmt.Sprintf("%s%s<b>%d %s</b>\n\n", header, sep, len(items), wordZapis)

	for i, item := range items {
		apt := item.Apt
		// Заголовок записи
		msg += fmt.Sprintf("<b>%d. %s — %s</b>\n", i+1, apt.Time, apt.Service)
		if apt.DurationMin > 0 && apt.DurationMin != 60 {
			msg += fmt.Sprintf("⏱ %d мин\n", apt.DurationMin)
		}
		if apt.Price > 0 {
			msg += fmt.Sprintf("💰 %s ₽\n", fmtNum(apt.Price))
		}

		// Клиент
		msg += fmt.Sprintf("\n👤 <b>%s</b>\n", apt.ClientName)
		msg += fmtContacts(&apt)

		// Комментарий к записи
		if apt.Comment != "" {
			msg += fmt.Sprintf("💬 %s\n", apt.Comment)
		}

		// Карточка клиента
		if item.ClientCard != nil {
			c := item.ClientCard
			hasCard := c.HairType != "" || c.Allergies != "" || c.ColorFormula != "" || c.Comment != ""
			if hasCard {
				msg += "\n🗂 <b>Карточка клиента:</b>\n"
				if c.HairType != "" {
					msg += fmt.Sprintf("• Тип волос: %s\n", c.HairType)
				}
				if c.Allergies != "" {
					msg += fmt.Sprintf("• Аллергии: %s\n", c.Allergies)
				}
				if c.ColorFormula != "" {
					msg += fmt.Sprintf("• Формула: %s\n", c.ColorFormula)
				}
				if c.Comment != "" {
					msg += fmt.Sprintf("• Заметка: %s\n", c.Comment)
				}
			}
		}

		// Прошлые визиты (комментарии мастера)
		if len(item.PastComments) > 0 {
			msg += "\n📖 <b>Прошлые визиты:</b>\n"
			for _, pc := range item.PastComments {
				msg += fmt.Sprintf("• %s\n", pc)
			}
		}

		// Расходники с низким остатком под эту услугу
		if len(item.LowSupplies) > 0 {
			msg += "\n⚠️ <b>Не хватает расходников:</b>\n"
			for _, s := range item.LowSupplies {
				unit := "г"
				if s.Unit == "piece" {
					unit = "шт"
				}
				msg += fmt.Sprintf("• %s %s — %.0f%s (мин %.0f%s)\n",
					s.Brand, s.Name, s.Quantity, unit, s.MinQuantity, unit)
			}
		}

		msg += sep
	}

	// Глобальные расходники с низким остатком (не связанные с конкретной услугой)
	if len(globalLowStock) > 0 {
		msg += "\n🔴 <b>Общий склад — заканчивается:</b>\n"
		for _, s := range globalLowStock {
			unit := "г"
			if s.Unit == "piece" {
				unit = "шт"
			}
			msg += fmt.Sprintf("• %s %s — %.0f%s\n", s.Brand, s.Name, s.Quantity, unit)
		}
	}

	n.send(strings.TrimRight(msg, "\n"))
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
