package mission

import (
	"commander-service/app/shared"
	"commander-service/internal-lib/messaging"
	"commander-service/internal-lib/snowflake"
	"commander-service/internal-lib/utils"
	"context"
)

type Controller struct {
	*utils.WithLogger
	Converter          *Converter
	repository         IRepository
	snowflakeGenerator *snowflake.Generator
	messageProducer    *messaging.MessageProducer
}

func NewController(logger *utils.WithLogger, convertor *Converter, repository IRepository,
	snowflakeGenerator *snowflake.Generator, messageProducer *messaging.MessageProducer) *Controller {
	return &Controller{
		WithLogger:         logger,
		Converter:          convertor,
		repository:         repository,
		snowflakeGenerator: snowflakeGenerator,
		messageProducer:    messageProducer,
	}
}

func (c *Controller) CreateMission(ctx context.Context, input *CreateMissionInput) (*CreateMissionResponse, error) {
	// authenticate user and validate input here (omitted for brevity)

	entity := &MissionEntity{
		ID:          c.snowflakeGenerator.Next(),
		Name:        input.Name,
		Description: input.Description,
		Status:      shared.MissionStatusCreated,
		CreatedBy:   input.CreatedBy,
	}

	if err := c.repository.CreateMission(ctx, entity); err != nil {
		c.Logger.Error().Err(err).Msg("failed to create mission")
		return nil, err
	}

	entity, err := c.repository.GetMissionByID(ctx, entity.ID)
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to fetch created mission")
		return nil, err
	}

	// Send mission created event to RabbitMQ order_queue
	missionCreatedMsg := &messaging.MissionCreatedMessage{
		MissionID: snowflake.ConvertFromSnowflake(entity.ID),
		Name:      entity.Name,
		Status:    string(entity.Status),
		CreatedAt: entity.CreatedAt,
		CreatedBy: entity.CreatedBy,
	}

	if err := c.messageProducer.PublishMissionCreated(ctx, missionCreatedMsg); err != nil {
		// Log error but don't fail the request - messaging is not critical for mission creation
		c.Logger.Error().Err(err).
			Str("mission_id", missionCreatedMsg.MissionID).
			Msg("failed to publish mission created event to queue")
	}

	return &CreateMissionResponse{
		Body: c.Converter.ToDTO(entity),
	}, nil
}

func (c *Controller) GetMissionByID(ctx context.Context, input *GetMissionByIDInput) (*GetMissionByIDResponse, error) {
	// Convert string ID to snowflake ID
	snowflakeID, err := snowflake.ConvertToSnowflake(input.ID)
	if err != nil {
		c.Logger.Error().Err(err).Str("id", input.ID).Msg("invalid mission ID format")
		return nil, err
	}

	entity, err := c.repository.GetMissionByID(ctx, snowflakeID)
	if err != nil {
		c.Logger.Error().Err(err).Str("id", input.ID).Msg("failed to get mission")
		return nil, err
	}

	return &GetMissionByIDResponse{
		Body: c.Converter.ToDTO(entity),
	}, nil
}
