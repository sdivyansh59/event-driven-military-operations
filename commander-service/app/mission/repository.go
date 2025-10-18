package mission

import (
	"commander-service/app/setup/dbconfig"
	"commander-service/internal-lib/database/crud"
	"commander-service/internal-lib/snowflake"
	"context"
	"time"
)

type IRepository interface {
	CreateMission(ctx context.Context, mission *MissionEntity) error
	GetMissionByID(ctx context.Context, id snowflake.ID) (*MissionEntity, error)
}

type Repository struct {
	handler            *crud.Handler[MissionEntity, snowflake.ID]
	snowflakeGenerator *snowflake.Generator
}

func NewRepository(commanderCampDB *dbconfig.CommandersCampDB, snowflakeGenerator *snowflake.Generator) IRepository {
	return &Repository{
		handler:            crud.NewHandler[MissionEntity, snowflake.ID](commanderCampDB.DB),
		snowflakeGenerator: snowflakeGenerator,
	}
}

func (r *Repository) CreateMission(ctx context.Context, mission *MissionEntity) error {
	mission.ID = r.snowflakeGenerator.Next()
	mission.CreatedAt = time.Now()
	mission.UpdatedAt = time.Now()

	err := r.handler.Create(ctx, mission)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetMissionByID(ctx context.Context, id snowflake.ID) (*MissionEntity, error) {
	entity, err := r.handler.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return entity, nil
}
