//go:build small

package formatter

import "testing"

func TestDefaultFormatter_Format(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		contents map[string]string
		want     string
	}{
		{
			name:     "空のファイルリスト",
			files:    []string{},
			contents: map[string]string{},
			want:     "",
		},
		{
			name:  "単一ファイル",
			files: []string{"spec.md"},
			contents: map[string]string{
				"spec.md": "# 製品仕様\n- 機能A: Xを行う\n- 機能B: Yを行う",
			},
			want: `[//]: # "filepath: spec.md"
# 製品仕様
- 機能A: Xを行う
- 機能B: Yを行う`,
		},
		{
			name:  "複数ファイル（順序維持）",
			files: []string{"spec.md", "rules.md"},
			contents: map[string]string{
				"spec.md":  "# 製品仕様\n- 機能A: Xを行う\n- 機能B: Yを行う",
				"rules.md": "# コーディング規則\n1. 変数名はcamelCaseを使用\n2. 公開関数にはコメントを追加",
			},
			want: `[//]: # "filepath: spec.md"
# 製品仕様
- 機能A: Xを行う
- 機能B: Yを行う

[//]: # "filepath: rules.md"
# コーディング規則
1. 変数名はcamelCaseを使用
2. 公開関数にはコメントを追加`,
		},
		{
			name:  "順序が指定された場合は指定順を維持",
			files: []string{"c_file.md", "a_file.md", "b_file.md"},
			contents: map[string]string{
				"c_file.md": "C content",
				"a_file.md": "A content",
				"b_file.md": "B content",
			},
			want: `[//]: # "filepath: c_file.md"
C content

[//]: # "filepath: a_file.md"
A content

[//]: # "filepath: b_file.md"
B content`,
		},
		{
			name:  "空のファイルコンテンツを含む",
			files: []string{"empty.md", "test.md"},
			contents: map[string]string{
				"empty.md": "",
				"test.md":  "テストコンテンツ",
			},
			want: `[//]: # "filepath: empty.md"


[//]: # "filepath: test.md"
テストコンテンツ`,
		},
		{
			name:  "特殊文字を含むファイル内容",
			files: []string{"special.md"},
			contents: map[string]string{
				"special.md": "# 特殊文字\n* `バッククォート`\n* \"ダブルクォート\"\n* 'シングルクォート'\n* \\バックスラッシュ",
			},
			want: `[//]: # "filepath: special.md"
# 特殊文字
* ` + "`バッククォート`" + `
* "ダブルクォート"
* 'シングルクォート'
* \バックスラッシュ`,
		},
	}

	formatter := NewDefaultFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatter.Format(tt.files, tt.contents)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Format() got and want differ\nGot:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}
