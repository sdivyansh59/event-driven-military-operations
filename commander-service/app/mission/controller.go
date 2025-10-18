package mission

import (
	"commander-service/app/shared"
	"commander-service/internal-lib/snowflake"
	"commander-service/internal-lib/utils"
	"context"
)

type Controller struct {
	*utils.WithLogger
	Converter          *Converter
	repository         IRepository
	snowflakeGenerator *snowflake.Generator
}

func NewController(logger *utils.WithLogger, convertor *Converter, repository IRepository,
	snowflakeGenerator *snowflake.Generator) *Controller {
	return &Controller{
		WithLogger:         logger,
		Converter:          convertor,
		repository:         repository,
		snowflakeGenerator: snowflakeGenerator,
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

	return &CreateMissionResponse{
		Body: c.Converter.ToDTO(entity),
	}, nil
}

func (c *Controller) GetMissionByID(ctx context.Context, input *GetMissionByIDInput) (*GetMissionByIDResponse, error) {
	// authenticate user here (omitted for brevity)

	// validate mission's id
	missionID, err := snowflake.ConvertToSnowflake(input.ID)
	if err != nil {
		c.Logger.Error().Err(err).Msg("invalid mission ID")
		return nil, err
	}

	// Fetch mission entity from repository
	entity, err := c.repository.GetMissionByID(ctx, missionID)
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to retrieve mission")
		return nil, err
	}

	return &GetMissionByIDResponse{
		Body: c.Converter.ToDTO(entity),
	}, nil
}
