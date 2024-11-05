package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
)

// IdentityCommand handles identity-related operations
type IdentityCommand struct {
	flags        *flag.FlagSet
	outputFormat *string
	parentDomain *string
	childDomains *string
	dryRun       *bool
	autoApprove  *bool
	noMask       *bool
}

// NewIdentityCommand creates a new identity command handler
func NewIdentityCommand() *IdentityCommand {
	cmd := &IdentityCommand{
		flags: flag.NewFlagSet("identity", flag.ExitOnError),
	}

	cmd.outputFormat = cmd.flags.String("output", "json", "出力フォーマット (json, markdown, pretty)")
	cmd.parentDomain = cmd.flags.String("parent-domain", "", "マージ先となる親ドメイン (例: example.com)")
	cmd.childDomains = cmd.flags.String("child-domains", "", "マージ元となる子ドメイン（カンマ区切り）(例: sub1.example.com,sub2.example.com)")
	cmd.dryRun = cmd.flags.Bool("dry-run", false, "マージ操作のシミュレーションを実行")
	cmd.autoApprove = cmd.flags.Bool("y", false, "確認プロンプトをスキップ")
	cmd.noMask = cmd.flags.Bool("nomask", false, "ログとファイル出力でメールアドレスをマスクしない")

	return cmd
}

func (c *IdentityCommand) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("サブコマンドが必要です")
	}

	subCmd := args[0]
	subArgs := args[1:]
	c.flags.SetOutput(c.flags.Output())

	switch subCmd {
	case "matrix":
		if err := c.flags.Parse(subArgs); err != nil {
			return err
		}
		return c.runMatrix()
	case "samemerge":
		if err := c.flags.Parse(subArgs); err != nil {
			return err
		}
		return c.runSameMerge()
	case "help":
		fmt.Println(c.Help())
		return nil
	default:
		return fmt.Errorf("不明なサブコマンド: %s", subCmd)
	}
}

// Help returns detailed usage information
func (c *IdentityCommand) Help() string {
	helpText := `MoneyForward Admina アイデンティティ管理ユーティリティ

使用方法:
  admina-sysutils identity [オプション] <サブコマンド>

サブコマンド:
  matrix      アイデンティティマトリックスを表示します
              組織内の全アイデンティティの関係を可視化します

  samemerge   同じメールローカルパートを持つアイデンティティをマージします
              異なるドメイン間で同一ユーザーのアイデンティティを統合します

  help        このヘルプメッセージを表示します

グローバルオプション:
  --output format   出力フォーマットを指定します (デフォルト: json)
                   指定可能な値: json, markdown, pretty

  --debug          デバッグモードを有効にします
                   詳細なログ出力が表示されます

Samemergeサブコマンドのオプション:
  --parent-domain  マージ先となる親ドメインを指定します
                   例: example.com

  --child-domains  マージ元となる子ドメインをカンマ区切りで指定します
                   例: sub1.example.com,sub2.example.com

  --dry-run       実際のマージを実行せずシミュレーションを行います
                   変更内容の確認に使用します

  -y              マージの確認プロンプトをスキップします
                   自動化スクリプトでの使用に適しています

  --nomask        ログとファイル出力でメールアドレスをマスクしない

使用例:
  # マトリックスの表示
  admina-sysutils identity matrix --output markdown

  # アイデンティティのマージ
  admina-sysutils identity samemerge \
    --parent-domain example.com \
    --child-domains sub1.example.com,sub2.example.com \
    --dry-run

環境変数:
  ADMINA_API_KEY          MoneyForward Admina APIキー
  ADMINA_ORGANIZATION_ID  組織ID
`
	return helpText
}

func (c *IdentityCommand) runMatrix() error {
	client := c.newIdentityClient()
	if client == nil {
		return fmt.Errorf("クライアントの初期化に失敗しました")
	}

	return identity.PrintIdentityMatrix(client, *c.outputFormat)
}

func (c *IdentityCommand) runSameMerge() error {
	if *c.parentDomain == "" {
		return fmt.Errorf("--parent-domain オプションは必須です")
	}
	if *c.childDomains == "" {
		return fmt.Errorf("--child-domains オプションは必須です")
	}

	childDomainList := strings.Split(*c.childDomains, ",")
	for i := range childDomainList {
		childDomainList[i] = strings.TrimSpace(childDomainList[i])
	}

	client := c.newIdentityClient()
	if client == nil {
		return fmt.Errorf("クライアントの初期化に失敗しました")
	}

	mergeConfig := &identity.MergeConfig{
		ParentDomain: *c.parentDomain,
		ChildDomains: childDomainList,
		DryRun:       *c.dryRun,
		AutoApprove:  *c.autoApprove,
		OutputFormat: *c.outputFormat,
		NoMask:       *c.noMask,
	}

	return identity.MergeIdentities(client, mergeConfig)
}

type identityClientAdapter struct {
	client *admina.Client
}

func (a *identityClientAdapter) GetIdentities(ctx context.Context, cursor string) ([]admina.Identity, string, error) {
	identities, nextCursor, err := a.client.GetIdentities(ctx, cursor)
	return identities, nextCursor, err
}

func (a *identityClientAdapter) MergeIdentities(ctx context.Context, fromPeopleID, toPeopleID int) error {
	return a.client.MergeIdentities(ctx, fromPeopleID, toPeopleID)
}

func (c *IdentityCommand) newIdentityClient() identity.Client {
	client := admina.NewClient()
	if client == nil {
		return nil
	}

	if err := client.Validate(); err != nil {
		return nil
	}

	return &identityClientAdapter{client: client}
}
