package mocks

import (
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/stretchr/testify/mock"
)

type MockPantryClient struct {
	mock.Mock
}

func (m *MockPantryClient) AddItem(item model.PantryItem) (int, error) {
	args := m.Called(item)
	return args.Int(0), args.Error(1)
}

func (m *MockPantryClient) UpdateItem(item model.PantryItem) {
	m.Called(item)
}

func (m *MockPantryClient) RemoveItem(id int) {
	m.Called(id)
}

func (m *MockPantryClient) GetItems() []model.PantryItem {
	args := m.Called()
	return args.Get(0).([]model.PantryItem)
}
