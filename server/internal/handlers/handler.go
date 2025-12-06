package handlers

type Handler struct {
	gameService   GameServiceInterface
	replayService ReplayServiceInterface
}

func NewHandler(gameService GameServiceInterface, replayService ReplayServiceInterface) *Handler {
	return &Handler{
		gameService:   gameService,
		replayService: replayService,
	}
}
