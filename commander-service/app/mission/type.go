package mission

type Mission struct {
	ID          int64
	Name        string
	Description string
	Status      string
}

type CreateMissionInput struct {
}

type CreateMissionResponse struct {
}

type GetMissionByIDInput struct {
}

type GetMissionByIDResponse struct {
}
