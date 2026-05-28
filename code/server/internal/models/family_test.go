package models

import "testing"

func TestCreateFamilyRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateFamilyRequest
		wantErr error
	}{
		{
			name:    "valid request with parent name",
			request: CreateFamilyRequest{ParentName: "John Doe"},
			wantErr: nil,
		},
		{
			name:    "valid request without parent name",
			request: CreateFamilyRequest{ParentName: ""},
			wantErr: nil,
		},
		{
			name:    "parent name too long",
			request: CreateFamilyRequest{ParentName: string(make([]byte, 101))},
			wantErr: ErrParentNameTooLong,
		},
		{
			name:    "parent name exactly 100 characters",
			request: CreateFamilyRequest{ParentName: string(make([]byte, 100))},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPairChildRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request PairChildRequest
		wantErr error
	}{
		{
			name:    "valid request",
			request: PairChildRequest{PairingCode: "ABCD1234", ChildName: "Little John"},
			wantErr: nil,
		},
		{
			name:    "empty pairing code",
			request: PairChildRequest{PairingCode: "", ChildName: "Little John"},
			wantErr: ErrPairingCodeRequired,
		},
		{
			name:    "pairing code too short",
			request: PairChildRequest{PairingCode: "ABC123", ChildName: "Little John"},
			wantErr: ErrInvalidPairingCode,
		},
		{
			name:    "pairing code too long",
			request: PairChildRequest{PairingCode: "ABCD12345", ChildName: "Little John"},
			wantErr: ErrInvalidPairingCode,
		},
		{
			name:    "empty child name",
			request: PairChildRequest{PairingCode: "ABCD1234", ChildName: ""},
			wantErr: ErrChildNameRequired,
		},
		{
			name:    "child name too long",
			request: PairChildRequest{PairingCode: "ABCD1234", ChildName: string(make([]byte, 101))},
			wantErr: ErrChildNameTooLong,
		},
		{
			name:    "child name exactly 100 characters",
			request: PairChildRequest{PairingCode: "ABCD1234", ChildName: string(make([]byte, 100))},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
