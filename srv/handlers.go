package srv

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/logproto"
	"go.uber.org/zap"
)

func (s *Service) handleLokiPush(c *gin.Context) {

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {

		s.log.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Decompress Snappy-compressed data
	decompressed, err := snappy.Decode(nil, body)
	if err != nil {

		s.log.Error("Failed to decompress snappy data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Unmarshal protobuf
	var req logproto.PushRequest
	err = req.Unmarshal(decompressed)
	if err != nil {

		s.log.Error("Failed to unmarshal protobuf", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log structured data
	for i, stream := range req.Streams {

		labels := parseLabels(stream.Labels)

		containerName, okContainerName := labels["container_name"]
		serviceName, okServiceName := labels["service_name"]
		if !okContainerName && !okServiceName {

			s.log.Warn("container_name or service_name not found", zap.Int("index", i))
		}

		for _, entry := range stream.Entries {

			if strings.Contains(entry.Line, "error") || strings.Contains(entry.Line, "warning") || strings.Contains(entry.Line, "fatal") {

				go s.handleEntry(containerName, serviceName, stream, entry)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func parseLabels(labelStr string) map[string]string {

	labels := make(map[string]string)

	// Remove the curly braces
	labelStr = strings.Trim(labelStr, "{}")

	// Regex to match label="value" pairs (handles escaped quotes)
	re := regexp.MustCompile(`(\w+)=\\"([^"\\]*(?:\\.[^"\\]*)*)\\"|(\w+)="([^"]*)"`)
	matches := re.FindAllStringSubmatch(labelStr, -1)

	for _, match := range matches {
		if match[1] != "" {
			// Escaped quote format: label=\"value\"
			labels[match[1]] = match[2]
		} else if match[3] != "" {
			// Normal format: label="value"
			labels[match[3]] = match[4]
		}
	}

	return labels
}

func (s *Service) handleEntry(containerName, serviceName string, stream logproto.Stream, entry logproto.Entry) {

	// find telegram token and chat id for the channel
	var telegramToken string
	var telegramChatId int64
	for _, channel := range s.cfg.Channels {

		if strings.Contains(containerName, channel.Needle) || strings.Contains(serviceName, channel.Needle) {

			telegramToken = channel.TelegramToken
			telegramChatId = channel.TelegramChatId
			break
		}
	}

	// set default telegram token and chat id if not found
	if telegramToken == "" || telegramChatId == 0 {

		telegramToken = s.cfg.Telegram.BotToken
		telegramChatId = s.cfg.Telegram.ChatId
	}

	s.sendTelegramMessage(containerName, serviceName, telegramToken, telegramChatId, stream, entry)
}

func (s *Service) sendTelegramMessage(containerName string, serviceName string, telegramToken string, telegramChatId int64, stream logproto.Stream, entry logproto.Entry) {

	labels := parseLabels(stream.Labels)

	var message string

	if level := labels["level"]; level != "" {

		message += fmt.Sprintf("*Level:* `%s`\n", escapeMarkdownV2(level))
	}

	if containerName == "" && serviceName == "" {

		message += fmt.Sprintf("*Labels:* `%s`\n", escapeMarkdownV2(stream.Labels))
	}

	if containerName != "" {

		message += fmt.Sprintf("*Container:* `%s`\n", escapeMarkdownV2(containerName))
	}

	if serviceName != "" {

		message += fmt.Sprintf("*Service:* `%s`\n", escapeMarkdownV2(serviceName))
	}

	message += fmt.Sprintf("```\n%s\n```\n", entry.Line)

	if fileName := labels["filename"]; fileName != "" {

		message += fmt.Sprintf("*File:* `%s`\n", escapeMarkdownV2(fileName))
	}

	if host := labels["host"]; host != "" {

		message += fmt.Sprintf("*Host:* `%s`\n", escapeMarkdownV2(host))
	}

	if ip := labels["ip"]; ip != "" {

		message += fmt.Sprintf("*IpAddress:* `%s`\n", escapeMarkdownV2(ip))
	}

	if time := labels["time"]; time != "" {

		message += fmt.Sprintf("*Time:* `%s`\n", escapeMarkdownV2(time))
	}

	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {

		s.log.Error("Failed to create telegram bot", zap.Error(err))
		return
	}

	msg := tgbotapi.NewMessage(telegramChatId, message)
	msg.ParseMode = "MarkdownV2"

	_, err = bot.Send(msg)
	if err != nil {
		s.log.Error("Failed to send telegram message", zap.Error(err))
	}
}

// escapeMarkdownV2 escapes special characters for Telegram MarkdownV2
func escapeMarkdownV2(text string) string {

	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)

	return replacer.Replace(text)
}
