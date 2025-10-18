package mission

type Converter struct {
}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ToDTO(mission *Mission) *MissionDTO {
	return &MissionDTO{
		ID:          mission.ID.String(),
		Name:        mission.Name,
		Description: mission.Description,
		Status:      mission.Status,
		CreatedBy:   mission.CreatedBy,
		CreatedAt:   mission.CreatedAt,
		UpdatedAt:   mission.UpdatedAt,
	}
}
