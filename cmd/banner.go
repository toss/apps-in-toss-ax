package cmd

import "fmt"

var bannerLines = []struct {
	line  string
	color [3]int
}{
	{"   ████       ██      ██", [3]int{102, 178, 255}},
	{"  ██  ██       ██    ██ ", [3]int{77, 159, 255}},
	{" ██    ██       ██  ██  ", [3]int{51, 139, 255}},
	{" ████████        ████   ", [3]int{26, 120, 255}},
	{"██      ██      ██  ██  ", [3]int{0, 100, 255}},
	{"██      ██     ██    ██ ", [3]int{0, 88, 224}},
	{"██      ██    ██      ██", [3]int{0, 77, 192}},
}

func printBanner() {
	fmt.Println()
	for _, l := range bannerLines {
		fmt.Printf("  \033[38;2;%d;%d;%dm%s\033[0m\n", l.color[0], l.color[1], l.color[2], l.line)
	}
	fmt.Println()
	fmt.Printf("  \033[1mAppsInToss eXperience\033[0m\n")
	fmt.Printf("  \033[38;2;136;136;136mDeveloper Tools & MCP Server\033[0m\n")
	v := GetVersion()
	fmt.Printf("  \033[38;2;102;102;102m%s (%s)\033[0m\n\n", v.Version, v.Hash)
}
