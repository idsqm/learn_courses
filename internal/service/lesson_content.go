package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type LessonContentService interface {
	GetLessonDetail(ctx context.Context, authorID, courseID, lessonID int) (*domain.LessonDetail, error)
	SaveContent(ctx context.Context, authorID, courseID, lessonID int, req domain.SaveLessonContentRequest) error
	CreateQuestion(ctx context.Context, authorID, courseID, lessonID int, req domain.SaveQuizQuestionRequest) (int, error)
	UpdateQuestion(ctx context.Context, authorID, courseID, questionID int, req domain.SaveQuizQuestionRequest) error
	DeleteQuestion(ctx context.Context, authorID, courseID, questionID int) error
}

type lessonContentService struct {
	repo repository.LessonContentRepository
}

func NewLessonContentService(repo repository.LessonContentRepository) LessonContentService {
	return &lessonContentService{repo: repo}
}

func (s *lessonContentService) GetLessonDetail(ctx context.Context, authorID, courseID, lessonID int) (*domain.LessonDetail, error) {
	if err := s.repo.VerifyCourseOwner(ctx, authorID, courseID); err != nil {
		return nil, err
	}
	return s.repo.GetLessonDetail(ctx, lessonID)
}

func (s *lessonContentService) SaveContent(ctx context.Context, authorID, courseID, lessonID int, req domain.SaveLessonContentRequest) error {
	if err := s.repo.VerifyCourseOwner(ctx, authorID, courseID); err != nil {
		return err
	}
	return s.repo.SaveContent(ctx, lessonID, req)
}

func (s *lessonContentService) CreateQuestion(ctx context.Context, authorID, courseID, lessonID int, req domain.SaveQuizQuestionRequest) (int, error) {
	if err := s.repo.VerifyCourseOwner(ctx, authorID, courseID); err != nil {
		return 0, err
	}
	return s.repo.CreateQuestion(ctx, lessonID, req)
}

func (s *lessonContentService) UpdateQuestion(ctx context.Context, authorID, courseID, questionID int, req domain.SaveQuizQuestionRequest) error {
	if err := s.repo.VerifyQuestionOwner(ctx, authorID, questionID); err != nil {
		return err
	}
	return s.repo.UpdateQuestion(ctx, questionID, req)
}

func (s *lessonContentService) DeleteQuestion(ctx context.Context, authorID, courseID, questionID int) error {
	if err := s.repo.VerifyQuestionOwner(ctx, authorID, questionID); err != nil {
		return err
	}
	return s.repo.DeleteQuestion(ctx, questionID)
}
