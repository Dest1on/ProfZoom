package app

import (
	"context"
	"time"

	"profzom/internal/common"
	"profzom/internal/domain/analytics"
	"profzom/internal/domain/application"
	"profzom/internal/domain/message"
	"profzom/internal/domain/vacancy"
)

type MessageService struct {
	messages     message.Repository
	applications application.Repository
	vacancies    vacancy.Repository
	analytics    analytics.Repository
}

const (
	messageMaxLength   = 2000
	messageMinInterval = 2 * time.Second
)

func NewMessageService(messages message.Repository, applications application.Repository, vacancies vacancy.Repository, analytics analytics.Repository) *MessageService {
	return &MessageService{messages: messages, applications: applications, vacancies: vacancies, analytics: analytics}
}

func (s *MessageService) Send(ctx context.Context, applicationID, senderID common.UUID, body string) (*message.Message, error) {
	if body == "" {
		return nil, common.NewError(common.CodeValidation, "message body is required", nil)
	}
	if len(body) > messageMaxLength {
		return nil, common.NewError(common.CodeValidation, "message is too long", nil)
	}
	app, err := s.applications.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	vac, err := s.vacancies.GetByID(ctx, app.VacancyID)
	if err != nil {
		return nil, err
	}
	if senderID != app.StudentID && senderID != vac.CompanyID {
		return nil, common.NewError(common.CodeForbidden, "user is not allowed to send messages", nil)
	}
	latest, err := s.messages.LatestByApplication(ctx, applicationID)
	if err == nil {
		if time.Since(latest.CreatedAt) < messageMinInterval {
			return nil, common.NewError(common.CodeValidation, "messages are sent too frequently", nil)
		}
	} else if !common.Is(err, common.CodeNotFound) {
		return nil, err
	}
	created, err := s.messages.Create(ctx, message.Message{ApplicationID: applicationID, SenderID: senderID, Body: body})
	if err != nil {
		return nil, err
	}
	_ = s.analytics.Create(ctx, analytics.Event{Name: "message.sent", UserID: &senderID, Payload: analyticsPayload(ctx, map[string]string{"application_id": applicationID.String(), "message_id": created.ID.String()})})
	return created, nil
}

func (s *MessageService) List(ctx context.Context, applicationID, userID common.UUID, limit, offset int) ([]message.Message, error) {
	app, err := s.applications.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	vac, err := s.vacancies.GetByID(ctx, app.VacancyID)
	if err != nil {
		return nil, err
	}
	if userID != app.StudentID && userID != vac.CompanyID {
		return nil, common.NewError(common.CodeForbidden, "user is not allowed to view messages", nil)
	}
	return s.messages.ListByApplication(ctx, applicationID, limit, offset)
}
