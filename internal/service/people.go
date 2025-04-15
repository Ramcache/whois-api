package service

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
	"whois-api/internal/config"
	"whois-api/internal/models"
)

type PeopleServiceInterface interface {
	EnrichPeople(p models.People) (models.People, error)
}

type PeopleService struct {
	client       *http.Client
	logger       *zap.Logger
	agifyURL     string
	genderizeURL string
	natURL       string
}

func NewPeopleService(logger *zap.Logger, cfg config.Config) *PeopleService {
	timeout, err := strconv.Atoi(cfg.HttpTimeout)
	if err != nil || timeout <= 0 {
		timeout = 5
	}

	return &PeopleService{
		client:       &http.Client{Timeout: time.Duration(timeout) * time.Second},
		logger:       logger,
		agifyURL:     cfg.AgifyURL,
		genderizeURL: cfg.GenderizeURL,
		natURL:       cfg.NationalizeURL,
	}
}

func (s *PeopleService) EnrichPeople(p models.People) (models.People, error) {
	s.logger.Debug("Starting enrichment", zap.String("name", p.Name))

	age, err := s.getAge(p.Name)
	if err != nil {
		s.logger.Error("Failed to get age", zap.Error(err))
		return p, err
	}
	p.Age = age

	gender, err := s.getGender(p.Name)
	if err != nil {
		s.logger.Error("Failed to get gender", zap.Error(err))
		return p, err
	}
	p.Gender = gender

	nat, err := s.getNationality(p.Name)
	if err != nil {
		s.logger.Error("Failed to get nationality", zap.Error(err))
		return p, err
	}
	p.Nationality = nat

	s.logger.Info("Enrichment completed", zap.Any("people", p))
	return p, nil
}

func (s *PeopleService) getAge(name string) (int, error) {
	url := fmt.Sprintf("%s/?name=%s", s.agifyURL, name)
	var resp struct {
		Age int `json:"age"`
	}
	err := s.getJSON(url, &resp)
	if err != nil {
		s.logger.Error("getAge request failed", zap.String("url", url), zap.Error(err))
		return 0, err
	}
	return resp.Age, nil
}

func (s *PeopleService) getGender(name string) (string, error) {
	url := fmt.Sprintf("%s/?name=%s", s.genderizeURL, name)
	var resp struct {
		Gender string `json:"gender"`
	}
	err := s.getJSON(url, &resp)
	if err != nil {
		s.logger.Error("getGender request failed", zap.String("url", url), zap.Error(err))
		return "", err
	}
	return resp.Gender, nil
}

func (s *PeopleService) getNationality(name string) (string, error) {
	url := fmt.Sprintf("%s/?name=%s", s.natURL, name)
	var resp struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}
	err := s.getJSON(url, &resp)
	if err != nil {
		s.logger.Error("getNationality request failed", zap.String("url", url), zap.Error(err))
		return "unknown", err
	}
	if len(resp.Country) == 0 {
		return "unknown", nil
	}
	return resp.Country[0].CountryID, nil
}

func (s *PeopleService) getJSON(url string, target interface{}) error {
	resp, err := s.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}
