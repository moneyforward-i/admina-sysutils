package mock

import (
	"context"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
)

// MockClient は Client インターフェースを実装するモックです
type Client struct {
	Identities   []admina.Identity
	Cursor       string
	Error        error                  // エラーケースのテスト用
	MergeResults []admina.MergeIdentity // マージ結果を保持するフィールド
}

func (m *Client) GetIdentities(ctx context.Context, cursor string) ([]admina.Identity, string, error) {
	if m.Error != nil {
		return nil, "", m.Error
	}
	return m.Identities, m.Cursor, nil
}

func (c *Client) MergeIdentities(ctx context.Context, fromPeopleID, toPeopleID int) (admina.MergeIdentity, error) {
	if c.Error != nil {
		return admina.MergeIdentity{}, c.Error
	}

	result := admina.MergeIdentity{
		FromPeopleID: fromPeopleID,
		ToPeopleID:   toPeopleID,
	}
	c.MergeResults = append(c.MergeResults, result)
	return result, nil
}
