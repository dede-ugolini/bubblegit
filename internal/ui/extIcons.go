package ui

import "charm.land/lipgloss/v2"

var extIcons = map[string]fileIcon{
	".c":        {'\ue61e', lipgloss.Color("#4488EE")}, // c        - blue
	".go":       {'\ue627', lipgloss.Color("#00ADD8")}, // go       - blue (Go)
	".js":       {'\ue74e', lipgloss.Color("#F7DF1E")}, // js       - yellow (JS)
	".jsx":      {'\ue74e', lipgloss.Color("#F7DF1E")}, // jsx
	".ts":       {'\ue7ca', lipgloss.Color("#3178C6")}, // ts       - blue (TS)
	".tsx":      {'\ue7ca', lipgloss.Color("#3178C6")}, // tsx
	".py":       {'\ue73c', lipgloss.Color("#3572A5")}, // python   - blue
	".java":     {'\ue738', lipgloss.Color("#B07219")}, // java     - orange
	".rb":       {'\ue739', lipgloss.Color("#701516")}, // ruby     - red
	".rake":     {'\ue739', lipgloss.Color("#701516")}, // rake
	".php":      {'\ue73d', lipgloss.Color("#777BB4")}, // php      - purple
	".swift":    {'\ue755', lipgloss.Color("#F05138")}, // swift    - orange
	".kt":       {'\ue81b', lipgloss.Color("#7F52FF")}, // kotlin   - purple
	".kts":      {'\ue81b', lipgloss.Color("#7F52FF")}, // kts
	".rs":       {'\ue7a8', lipgloss.Color("#DEA584")}, // rust     - tan
	".cpp":      {'\ue7a3', lipgloss.Color("#F34B7D")}, // cpp      - pink
	".cxx":      {'\ue7a3', lipgloss.Color("#F34B7D")}, // cxx
	".cc":       {'\ue7a3', lipgloss.Color("#F34B7D")}, // cc
	".hpp":      {'\ue7a3', lipgloss.Color("#F34B7D")}, // hpp
	".cs":       {'\ue7b2', lipgloss.Color("#68217A")}, // csharp   - purple
	".lua":      {'\ue826', lipgloss.Color("#000080")}, // lua      - navy
	".dart":     {'\ue798', lipgloss.Color("#00B4AB")}, // dart     - teal
	".scala":    {'\ue737', lipgloss.Color("#C22D40")}, // scala    - red
	".r":        {'\ue881', lipgloss.Color("#198CE7")}, // r        - blue
	".hs":       {'\ue777', lipgloss.Color("#5E5086")}, // haskell  - purple
	".lhs":      {'\ue777', lipgloss.Color("#5E5086")}, // lhs
	".clj":      {'\ue768', lipgloss.Color("#DB5855")}, // clojure
	".cljs":     {'\ue768', lipgloss.Color("#DB5855")}, // clojurescript
	".erl":      {'\ue7b1', lipgloss.Color("#A90533")}, // erlang
	".ex":       {'\ue7cd', lipgloss.Color("#6E4A7E")}, // elixir
	".exs":      {'\ue7cd', lipgloss.Color("#6E4A7E")}, // elixir script
	".elm":      {'\ue7ce', lipgloss.Color("#60B5CC")}, // elm
	".ml":       {'\ue84e', lipgloss.Color("#EC6813")}, // ocaml    - orange
	".mli":      {'\ue84e', lipgloss.Color("#EC6813")}, // ocaml interface
	".m":        {'\ue84d', lipgloss.Color("#438EFF")}, // objc     - blue
	".mm":       {'\ue84d', lipgloss.Color("#438EFF")}, // objc++
	".pl":       {'\ue769', lipgloss.Color("#0298C3")}, // perl
	".pro":      {'\ue7a1', lipgloss.Color("#74283C")}, // prolog
	".html":     {'\ue736', lipgloss.Color("#E34F26")}, // html     - orange
	".htm":      {'\ue736', lipgloss.Color("#E34F26")}, // htm
	".css":      {'\ue749', lipgloss.Color("#663399")}, // css      - purple
	".scss":     {'\ue74b', lipgloss.Color("#CD6799")}, // sass     - pink
	".sass":     {'\ue74b', lipgloss.Color("#CD6799")}, // sass
	".vue":      {'\ue7dc', lipgloss.Color("#42B883")}, // vue      - green
	".svelte":   {'\ue8b7', lipgloss.Color("#FF3E00")}, // svelte   - orange red
	".astro":    {'\ue735', lipgloss.Color("#FF5D01")}, // astro    - orange
	".json":     {'\ue80b', lipgloss.Color("#CBCB41")}, // json     - yellow
	".yaml":     {'\ue8eb', lipgloss.Color("#CB171E")}, // yaml     - red
	".yml":      {'\ue8eb', lipgloss.Color("#CB171E")}, // yml
	".xml":      {'\ue8ea', lipgloss.Color("#0060AC")}, // xml      - blue
	".sql":      {'\ue706', lipgloss.Color("#E38C00")}, // database - orange
	".md":       {'\ue73e', lipgloss.Color("#6A737D")}, // markdown - gray
	".markdown": {'\ue73e', lipgloss.Color("#6A737D")}, // markdown
	".sh":       {'\ue760', lipgloss.Color("#4EAA25")}, // shell    - green
	".bash":     {'\ue760', lipgloss.Color("#4EAA25")}, // bash
	".zsh":      {'\ue957', lipgloss.Color("#4EAA25")}, // zsh
	".zig":      {'\ue8ef', lipgloss.Color("#F7A41D")}, // zig      - orange
}
