package reporter

import "testing"

// The console title is a jump hint, so the only thing that matters is that a
// binary's path never reaches the resolver as a title: an untitled console
// reports the shell's own executable path, and no window title contains that.
func TestSanitizeConsoleTitleRejectsBinaryPaths(t *testing.T) {
	cases := []struct{ in, want string }{
		{`C:\windows\system32\WindowsPowerShell\v1.0\powershell.EXE`, ""},
		{`/bin/bash`, ""},
		{`C:\Program Files\Git\bin\bash.exe`, ""},
		{"   ", ""},
		{"", ""},
		{"命令提示符 - agy", "命令提示符 - agy"},
		{"Windows PowerShell", "Windows PowerShell"},
		{"OC | 项目交接与下一步计划", "OC | 项目交接与下一步计划"},
		{"  OC | 项目交接  ", "OC | 项目交接"},
	}
	for _, test := range cases {
		if got := sanitizeConsoleTitle(test.in); got != test.want {
			t.Errorf("sanitizeConsoleTitle(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}
