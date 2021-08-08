package commands

import (
	"context"
	kafkaClient "github.com/AleksK1NG/cqrs-microservices/pkg/kafka"
	"github.com/AleksK1NG/cqrs-microservices/pkg/logger"
	"github.com/AleksK1NG/cqrs-microservices/product_writer_service/config"
	"github.com/AleksK1NG/cqrs-microservices/product_writer_service/internal/models"
	"github.com/AleksK1NG/cqrs-microservices/product_writer_service/internal/product/repository"
	"github.com/AleksK1NG/cqrs-microservices/product_writer_service/mappers"
	kafkaMessages "github.com/AleksK1NG/cqrs-microservices/proto/kafka"
	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	"time"
)

type UpdateProductCmdHandler interface {
	Handle(ctx context.Context, command *UpdateProductCommand) (*models.Product, error)
}

type updateProductHandler struct {
	log           logger.Logger
	cfg           *config.Config
	pgRepo        repository.Repository
	kafkaProducer kafkaClient.Producer
}

func NewUpdateProductHandler(log logger.Logger, cfg *config.Config, pgRepo repository.Repository, kafkaProducer kafkaClient.Producer) *updateProductHandler {
	return &updateProductHandler{log: log, cfg: cfg, pgRepo: pgRepo, kafkaProducer: kafkaProducer}
}

func (c *updateProductHandler) Handle(ctx context.Context, command *UpdateProductCommand) (*models.Product, error) {
	productDto := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
	}

	updatedProduct, err := c.pgRepo.UpdateProduct(ctx, productDto)
	if err != nil {
		return nil, err
	}

	msg := &kafkaMessages.ProductUpdated{Product: mappers.ProductToGrpcMessage(updatedProduct)}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	message := kafka.Message{
		Topic: c.cfg.KafkaTopics.ProductUpdated.TopicName,
		Value: msgBytes,
		Time:  time.Now().UTC(),
	}

	c.log.Infof("updated product: %+v", updatedProduct)
	if err := c.kafkaProducer.PublishMessage(ctx, message); err != nil {
		return nil, err
	}

	return updatedProduct, nil
}