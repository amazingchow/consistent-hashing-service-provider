package version

import "fmt"

var (
	Version   = "None"
	GitBranch = "None"
	GitHash   = "None"
	BuildTS   = "None"
)

// PrintFullVersionInfo prints full version info.
func PrintFullVersionInfo() {
	fmt.Println("Version          : ", Version)
	fmt.Println("Git Branch       : ", GitBranch)
	fmt.Println("Git Commit       : ", GitHash)
	fmt.Println("Build Time (UTC) : ", BuildTS)
}
