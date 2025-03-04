//go:build small

package formatter

import (
	"strings"
	"testing"
)

func TestDefaultFormatter_Format(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "空のファイルマップ",
			files: map[string]string{},
			want:  "",
		},
		{
			name: "単一ファイル",
			files: map[string]string{
				"spec.md": "# 製品仕様\n- 機能A: Xを行う\n- 機能B: Yを行う",
			},
			want: "[//]: # \"filepath: spec.md\"\n# 製品仕様\n- 機能A: Xを行う\n- 機能B: Yを行う",
		},
		{
			name: "複数ファイル",
			files: map[string]string{
				"spec.md":  "# 製品仕様\n- 機能A: Xを行う\n- 機能B: Yを行う",
				"rules.md": "# コーディング規則\n1. 変数名はcamelCaseを使用\n2. 公開関数にはコメントを追加",
			},
			want: strings.Join([]string{
				"[//]: # \"filepath: rules.md\"",
				"# コーディング規則",
				"1. 変数名はcamelCaseを使用",
				"2. 公開関数にはコメントを追加",
				"",
				"[//]: # \"filepath: spec.md\"",
				"# 製品仕様",
				"- 機能A: Xを行う",
				"- 機能B: Yを行う",
			}, "\n"),
		},
		{
			name: "ファイル名の順序が確定的であること",
			files: map[string]string{
				"c_file.md": "C content",
				"a_file.md": "A content",
				"b_file.md": "B content",
			},
			want: strings.Join([]string{
				"[//]: # \"filepath: a_file.md\"",
				"A content",
				"",
				"[//]: # \"filepath: b_file.md\"",
				"B content",
				"",
				"[//]: # \"filepath: c_file.md\"",
				"C content",
			}, "\n"),
		},
		{
			name: "空のファイルコンテンツを含む",
			files: map[string]string{
				"empty.md": "",
				"test.md":  "テストコンテンツ",
			},
			want: strings.Join([]string{
				"[//]: # \"filepath: empty.md\"",
				"",
				"",
				"[//]: # \"filepath: test.md\"",
				"テストコンテンツ",
			}, "\n"),
		},
		{
			name: "特殊文字を含むファイル内容",
			files: map[string]string{
				"special.md": "# 特殊文字\n* `バッククォート`\n* \"ダブルクォート\"\n* 'シングルクォート'\n* \\バックスラッシュ",
			},
			want: "[//]: # \"filepath: special.md\"\n# 特殊文字\n* `バッククォート`\n* \"ダブルクォート\"\n* 'シングルクォート'\n* \\バックスラッシュ",
		},
	}

	formatter := NewDefaultFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatter.Format(tt.files)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
				// デバッグ用に詳細な差異を表示
				t.Errorf("Got:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}

func TestSortStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "空の配列",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "すでにソート済み",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "逆順",
			input: []string{"c", "b", "a"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "ランダム順",
			input: []string{"b", "a", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "重複を含む",
			input: []string{"b", "a", "b", "c"},
			want:  []string{"a", "b", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 入力値のコピーを作成（sortStringsは破壊的変更を行うため）
			input := make([]string, len(tt.input))
			copy(input, tt.input)

			sortStrings(input)

			// 配列全体を比較
			if len(input) != len(tt.want) {
				t.Errorf("sortStrings() length = %v, want %v", len(input), len(tt.want))
				return
			}

			for i := range input {
				if input[i] != tt.want[i] {
					t.Errorf("sortStrings() at index %d = %v, want %v", i, input[i], tt.want[i])
				}
			}
		})
	}
}
