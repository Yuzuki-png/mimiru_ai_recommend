package services

import (
	"context"
	"fmt"
	"mimiru-ai/infrastructure/database"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseEvent struct {
	TableName string                 `json:"table_name"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type DatabaseMonitorService struct {
	db            *database.Client
	eventHandlers map[string][]func(DatabaseEvent)
}

func NewDatabaseMonitorService(db *database.Client) *DatabaseMonitorService {
	return &DatabaseMonitorService{
		db:            db,
		eventHandlers: make(map[string][]func(DatabaseEvent)),
	}
}

func (s *DatabaseMonitorService) RegisterEventHandler(tableName string, handler func(DatabaseEvent)) {
	s.eventHandlers[tableName] = append(s.eventHandlers[tableName], handler)
}

func (s *DatabaseMonitorService) StartMonitoring(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, s.db.Pool.Config().ConnString())
	if err != nil {
		return fmt.Errorf("監視接続の作成に失敗しました: %w", err)
	}
	defer conn.Close(ctx)

	monitoredTables := []string{
		"playback_sessions_change",
		"user_ratings_change",
	}

	for _, channel := range monitoredTables {
		_, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel))
		if err != nil {
			return fmt.Errorf("%sのリッスンに失敗しました: %w", channel, err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			notification, err := conn.WaitForNotification(ctx)
			if err != nil {
				continue
			}

			s.handleNotification(notification)
		}
	}
}

func (s *DatabaseMonitorService) handleNotification(notification *pgconn.Notification) {

	var tableName string
	switch notification.Channel {
	case "playback_sessions_change":
		tableName = "playback_sessions"
	case "user_ratings_change":
		tableName = "user_ratings"
	default:
		return
	}

	event := DatabaseEvent{
		TableName: tableName,
		EventType: "CHANGE",
		Data:      map[string]interface{}{"payload": notification.Payload},
		Timestamp: time.Now(),
	}

	if handlers, exists := s.eventHandlers[tableName]; exists {
		for _, handler := range handlers {
			go func(h func(DatabaseEvent)) {
				defer func() {
					if r := recover(); r != nil {
					}
				}()
				h(event)
			}(handler)
		}
	}
}

func (s *DatabaseMonitorService) PollingMonitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	lastCheck := time.Now().Add(-interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkRecentChanges(ctx, lastCheck)
			lastCheck = time.Now()
		}
	}
}

func (s *DatabaseMonitorService) checkRecentChanges(ctx context.Context, since time.Time) {
	query := `
		SELECT id, user_id, audio_content_id, created_at
		FROM playback_sessions
		WHERE created_at > $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := s.db.Pool.Query(ctx, query, since)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, audioContentID int
		var createdAt time.Time
		
		if err := rows.Scan(&id, &userID, &audioContentID, &createdAt); err != nil {
			continue
		}

			event := DatabaseEvent{
			TableName: "playback_sessions",
			EventType: "INSERT",
			Data: map[string]interface{}{
				"id":               id,
				"user_id":          userID,
				"audio_content_id": audioContentID,
				"created_at":       createdAt,
			},
			Timestamp: createdAt,
		}

		if handlers, exists := s.eventHandlers["playback_sessions"]; exists {
			for _, handler := range handlers {
				go handler(event)
			}
		}
	}
}