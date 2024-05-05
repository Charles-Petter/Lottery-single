package service

import (
	"fmt"
	"lottery_single/internal/model"
	"lottery_single/internal/repo"
)

type BlackIpService struct {
	BlackIpRepo *repo.BlackIpRepo
}

func blackIp(blackIpRepo *repo.BlackIpRepo) *BlackIpService {
	return &BlackIpService{BlackIpRepo: blackIpRepo}
}

func (s *BlackIpService) CreateBlackIp(ip string) error {
	blackIp := &model.BlackIp{Ip: ip}
	err := s.BlackIpRepo.Create(blackIp)
	if err != nil {
		return fmt.Errorf("failed to create black IP: %v", err)
	}
	return nil
}

func (s *BlackIpService) GetBlackIpByID(id uint) (*model.BlackIp, error) {
	return s.BlackIpRepo.GetByID(id)
}

func (s *BlackIpService) GetBlackIpByIP(ip string) (*model.BlackIp, error) {
	return s.BlackIpRepo.GetByIP(ip)
}

func (s *BlackIpService) UpdateBlackIp(blackIp *model.BlackIp) error {
	return s.BlackIpRepo.Update(blackIp)
}

func (s *BlackIpService) DeleteBlackIpByID(id uint) error {
	return s.BlackIpRepo.DeleteByID(id)
}
