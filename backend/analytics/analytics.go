package analytics

import (
	"connect-four/game"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

type Service struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	db       interface{} // Can be *sql.DB if needed
}

func NewService() (*Service, error) {
	brokers := getKafkaBrokers()
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		producer.Close()
		return nil, err
	}

	service := &Service{
		producer: producer,
		consumer: consumer,
	}

	// Start consumer in background
	go service.startConsumer()

	return service, nil
}

func (s *Service) startConsumer() {
	partitionConsumer, err := s.consumer.ConsumePartition("game-events", 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Error creating partition consumer: %v", err)
		return
	}
	defer partitionConsumer.Close()

	for message := range partitionConsumer.Messages() {
		var event map[string]interface{}
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}
		s.processEvent(event)
	}
}

func (s *Service) TrackGameStart(game *game.Game) {
	if s == nil || s.producer == nil {
		return
	}
	event := map[string]interface{}{
		"type":         "game_start",
		"gameId":       game.ID,
		"player1":      game.Player1.Username,
		"player2":      game.Player2.Username,
		"player2IsBot": game.Player2.IsBot,
		"timestamp":    game.StartedAt.Format(time.RFC3339),
	}
	s.sendEvent(event)
}

func (s *Service) TrackMove(game *game.Game, column, row int) {
	if s == nil || s.producer == nil {
		return
	}
	player := "bot"
	if game.CurrentPlayer == game.Player1.ID {
		player = game.Player1.Username
	}

	event := map[string]interface{}{
		"type":       "move",
		"gameId":     game.ID,
		"player":     player,
		"column":     column,
		"row":        row,
		"moveNumber": len(game.Moves),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	s.sendEvent(event)
}

func (s *Service) TrackGameEnd(game *game.Game) {
	if s == nil || s.producer == nil {
		return
	}
	var duration *int
	if game.EndedAt != nil {
		d := int(game.EndedAt.Sub(game.StartedAt).Seconds())
		duration = &d
	}

	winner := "draw"
	if game.Winner != "draw" && game.Winner != "" {
		if game.Winner == "bot" {
			winner = "bot"
		} else {
			winner = "player"
		}
	}

	event := map[string]interface{}{
		"type":       "game_end",
		"gameId":     game.ID,
		"winner":     winner,
		"duration":   duration,
		"totalMoves": len(game.Moves),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	if game.EndedAt != nil {
		event["timestamp"] = game.EndedAt.Format(time.RFC3339)
	}
	s.sendEvent(event)
}

func (s *Service) sendEvent(event map[string]interface{}) {
	if s == nil || s.producer == nil {
		return
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	gameID, _ := event["gameId"].(string)
	if gameID == "" {
		gameID = "system"
	}

	msg := &sarama.ProducerMessage{
		Topic: "game-events",
		Key:   sarama.StringEncoder(gameID),
		Value: sarama.ByteEncoder(eventJSON),
	}

	_, _, err = s.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending event to Kafka: %v", err)
	}
}

func (s *Service) processEvent(event map[string]interface{}) {
	// Store event or process analytics
	// This can be extended to store in database
	log.Printf("Processing analytics event: %+v", event)
}

func getKafkaBrokers() []string {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		return []string{}
	}
	// Simple split - in production, handle more complex cases
	return []string{brokersStr}
}
