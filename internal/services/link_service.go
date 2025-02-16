package services

import (
	"context"
	"fmt"
	"time"

	"github.com/devingoodsell/go-links-free/internal/models"
)

type LinkService struct {
	linkRepo *models.LinkRepository
}

func NewLinkService(linkRepo *models.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

func (s *LinkService) Create(ctx context.Context, userID int64, alias, destinationURL string, expiresAt *time.Time) (*models.Link, error) {
	link := &models.Link{
		Alias:          alias,
		DestinationURL: destinationURL,
		CreatedBy:      userID,
		ExpiresAt:      expiresAt,
	}

	if err := s.linkRepo.Create(ctx, link); err != nil {
		if err == models.ErrDuplicate {
			return nil, err
		}
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	return link, nil
}

func (s *LinkService) GetByAlias(ctx context.Context, alias string) (*models.Link, error) {
	return s.linkRepo.GetByAlias(ctx, alias)
}

func (s *LinkService) List(ctx context.Context, userID int64, opts models.ListOptions) ([]*models.Link, error) {
	return s.linkRepo.ListByUserWithFilters(ctx, userID, opts)
}

func (s *LinkService) Update(ctx context.Context, userID int64, alias string, destinationURL string, expiresAt *time.Time) (*models.Link, error) {
	link, err := s.linkRepo.GetByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}

	if link.CreatedBy != userID {
		return nil, models.ErrUnauthorized
	}

	link.DestinationURL = destinationURL
	link.ExpiresAt = expiresAt

	// TODO: Implement Update in LinkRepository
	return link, nil
}

func (s *LinkService) Delete(ctx context.Context, userID int64, alias string) error {
	link, err := s.GetByAlias(ctx, alias)
	if err != nil {
		return err
	}

	if link.CreatedBy != userID {
		return models.ErrUnauthorized
	}

	return s.linkRepo.Delete(ctx, link.ID, userID)
}

func (s *LinkService) IncrementStats(ctx context.Context, linkID int64) error {
	return s.linkRepo.IncrementStats(ctx, linkID)
}

func (s *LinkService) BulkDelete(ctx context.Context, userID int64, ids []int64) error {
	// Verify all links exist and belong to the user
	for _, id := range ids {
		link, err := s.linkRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link.CreatedBy != userID {
			return models.ErrUnauthorized
		}
	}

	return s.linkRepo.BulkDelete(ctx, userID, ids)
}

func (s *LinkService) BulkUpdateStatus(ctx context.Context, userID int64, ids []int64, isActive bool) error {
	// Verify all links exist and belong to the user
	for _, id := range ids {
		link, err := s.linkRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link.CreatedBy != userID {
			return models.ErrUnauthorized
		}
	}

	return s.linkRepo.BulkUpdateStatus(ctx, userID, ids, isActive)
}
