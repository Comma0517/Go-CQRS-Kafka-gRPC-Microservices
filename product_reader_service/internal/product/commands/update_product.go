package commands

import (
	"context"
	"github.com/AleksK1NG/cqrs-microservices/pkg/logger"
	"github.com/AleksK1NG/cqrs-microservices/product_reader_service/config"
	"github.com/AleksK1NG/cqrs-microservices/product_reader_service/internal/models"
	"github.com/AleksK1NG/cqrs-microservices/product_reader_service/internal/product/repository"
)

type UpdateProductCmdHandler interface {
	Handle(ctx context.Context, command *UpdateProductCommand) error
}

type updateProductCmdHandler struct {
	log       logger.Logger
	cfg       *config.Config
	mongoRepo repository.Repository
	redisRepo repository.CacheRepository
}

func NewUpdateProductCmdHandler(log logger.Logger, cfg *config.Config, mongoRepo repository.Repository, redisRepo repository.CacheRepository) *updateProductCmdHandler {
	return &updateProductCmdHandler{log: log, cfg: cfg, mongoRepo: mongoRepo, redisRepo: redisRepo}
}

func (c *updateProductCmdHandler) Handle(ctx context.Context, command *UpdateProductCommand) error {
	product := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
		UpdatedAt:   command.UpdatedAt,
	}

	updated, err := c.mongoRepo.UpdateProduct(ctx, product)
	if err != nil {
		return err
	}

	c.log.Debugf("updated product: %+v", updated)
	c.redisRepo.PutProduct(ctx, updated.ProductID, updated)
	return nil
}