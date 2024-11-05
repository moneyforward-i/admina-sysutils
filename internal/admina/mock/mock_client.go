package mock

import (
	"context"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
)

// MockClient は Client インターフェースを実装するモックです
type Client struct {
	Identities []admina.Identity
	Cursor     string
	Error      error // エラーケースのテスト用
}

func (m *Client) GetIdentities(ctx context.Context, cursor string) ([]admina.Identity, string, error) {
	if m.Error != nil {
		return nil, "", m.Error
	}
	return m.Identities, m.Cursor, nil
}

func (m *Client) MergeIdentities(ctx context.Context, fromPeopleID, toPeopleID int) error {
	if m.Error != nil {
		return m.Error
	}
	return nil
}
