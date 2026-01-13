package handler

import "log"

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleMessage(message []byte) error {
	log.Println(message)
	return nil
}
