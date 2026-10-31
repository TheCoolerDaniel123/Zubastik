package sms

import "log"

type MockSMSSender struct{}

func NewMockSMSSender() *MockSMSSender {
	return &MockSMSSender{}
}

func (s *MockSMSSender) SendSMS(phone string, code string) error {
	log.Printf("[SMS MOCK] Sending SMS to %s: %s", phone, code)
	return nil
}
