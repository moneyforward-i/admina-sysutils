package cli

import (
	"flag"
	"fmt"
)

func Run(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	helpFlag := flags.Bool("help", false, "ヘルプを表示")

	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("無効なフラグ: %v\nヘルプを表示するには 'admina-sysutils --help' を実行してください", err)
	}

	if *helpFlag {
		printHelp(flags)
		return nil
	}

	if flags.NArg() == 0 {
		return fmt.Errorf("引数が指定されていません\nヘルプを表示するには 'admina-sysutils --help' を実行してください")
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("不明なコマンド: %s\nヘルプを表示するには 'admina-sysutils --help' を実行してください", flags.Arg(0))
	}

	// ここにCLIの主要な機能を実装します

	return nil
}

func printHelp(flags *flag.FlagSet) {
	fmt.Println("使用方法: admina-sysutils [オプション]")
	fmt.Println("オプション:")
	flags.PrintDefaults()
}
