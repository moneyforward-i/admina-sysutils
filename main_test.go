package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	// 標準出力を元に戻す
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	// 出力を確認
	expected := "Azure for Go Developers\n"
	if string(out) != expected {
		t.Errorf("期待される出力: %q, 実際の出力: %q", expected, string(out))
	}
}

func TestMainWithDifferentName(t *testing.T) {
	// mainパッケージ内の変数を変更するためのモンキーパッチ
	oldName := name
	name = "Cloud Engineers"
	defer func() { name = oldName }()

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	// 標準出力を元に戻す
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	// 出力を確認
	expected := "Azure for Cloud Engineers\n"
	if string(out) != expected {
		t.Errorf("期待される出力: %q, 実際の出力: %q", expected, string(out))
	}
}
