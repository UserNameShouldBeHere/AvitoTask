package domain

import (
	"errors"
	"testing"

	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

func TestUserCredsValidation(t *testing.T) {
	testData := []struct {
		TestName string
		UserName string
		Password string
		IsValid  bool
	}{
		{
			"incorrect user name",
			"t",
			"test_password",
			false,
		},
		{
			"incorrect password",
			"test_user",
			"t",
			false,
		},
		{
			"incorrect user name and password",
			"t",
			"t",
			false,
		},
		{
			"empty fields",
			"",
			"",
			false,
		},
		{
			"correct data",
			"test_user",
			"test_password",
			true,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.TestName, func(t *testing.T) {
			transaction := UserCredantials{
				UserName: testCase.UserName,
				Password: testCase.Password,
			}
			err := transaction.Validate()
			if !testCase.IsValid && !errors.Is(err, customErrors.ErrDataNotValid) {
				t.Errorf("unexpected error on case %v", transaction)
			} else if testCase.IsValid && err != nil {
				t.Errorf("missed an error on case %v", transaction)
			}
		})
	}
}

func TestTransactionValidation(t *testing.T) {
	testData := []struct {
		TestName string
		From     string
		To       string
		Amount   int
		IsValid  bool
	}{
		{
			"incorrect from user name",
			"t",
			"test_user_2",
			100,
			false,
		},
		{
			"incorrect to user name",
			"test_user_1",
			"t",
			100,
			false,
		},
		{
			"incorrect both users names",
			"t",
			"t",
			100,
			false,
		},
		{
			"incorrect amount",
			"test_user_1",
			"test_user_2",
			-10,
			false,
		},
		{
			"empty fields",
			"",
			"",
			0,
			false,
		},
		{
			"correct data",
			"test_user_1",
			"test_user_2",
			100,
			true,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.TestName, func(t *testing.T) {
			transaction := Transaction{
				From:   testCase.From,
				To:     testCase.To,
				Amount: testCase.Amount,
			}
			err := transaction.Validate()
			if !testCase.IsValid && !errors.Is(err, customErrors.ErrDataNotValid) {
				t.Errorf("unexpected error on case %v", transaction)
			} else if testCase.IsValid && err != nil {
				t.Errorf("missed an error on case %v", transaction)
			}
		})
	}
}
