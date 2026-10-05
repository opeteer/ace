package doctor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	provSuccess = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render("✔")
	provInfo    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D7D7")).Render("ℹ")
	provWarn    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5A93C")).Render("!")
	provTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	provDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

// ProvisionRecipe defines how to download and install a specific tool
type ProvisionRecipe struct {
	Name        string
	Binary      string
	Description string
	UserSpace   bool // true if can be installed without sudo/root into ~/.local/bin
	InstallFunc func(localBin string) error
}

// Registry of supported automated tool installers
var ProvisionRecipes = map[string]ProvisionRecipe{
	"composer": {
		Name:        "Composer",
		Binary:      "composer",
		Description: "Dependency Manager for PHP",
		UserSpace:   true,
		InstallFunc: installComposer,
	},
	"bun": {
		Name:        "Bun",
		Binary:      "bun",
		Description: "Fast all-in-one JavaScript/TypeScript runtime",
		UserSpace:   true,
		InstallFunc: installBun,
	},
	"pnpm": {
		Name:        "pnpm",
		Binary:      "pnpm",
		Description: "Fast, disk space efficient package manager",
		UserSpace:   true,
		InstallFunc: installPnpm,
	},
	"rust": {
		Name:        "Rust / Cargo",
		Binary:      "cargo",
		Description: "Rust programming language toolchain",
		UserSpace:   true,
		InstallFunc: installRust,
	},
	"cargo": {
		Name:        "Rust / Cargo",
		Binary:      "cargo",
		Description: "Rust programming language toolchain",
		UserSpace:   true,
		InstallFunc: installRust,
	},
	"dotnet": {
		Name:        ".NET SDK",
		Binary:      "dotnet",
		Description: "Microsoft .NET SDK (v8.0 LTS)",
		UserSpace:   true,
		InstallFunc: installDotnet,
	},
	"node": {
		Name:        "Node.js",
		Binary:      "node",
		Description: "JavaScript runtime environment",
		UserSpace:   false,
		InstallFunc: installNodeSystem,
	},
	"python3": {
		Name:        "Python 3",
		Binary:      "python3",
		Description: "Python programming language & pip",
		UserSpace:   false,
		InstallFunc: installPythonSystem,
	},
	"php": {
		Name:        "PHP",
		Binary:      "php",
		Description: "PHP CLI runtime",
		UserSpace:   false,
		InstallFunc: installPhpSystem,
	},
	"docker": {
		Name:        "Docker Engine",
		Binary:      "docker",
		Description: "Docker container runtime",
		UserSpace:   false,
		InstallFunc: installDockerSystem,
	},
	"git": {
		Name:        "Git",
		Binary:      "git",
		Description: "Distributed version control system",
		UserSpace:   false,
		InstallFunc: installGitSystem,
	},
}

// GetUserLocalBin returns ~/.local/bin, ensuring the directory exists
func GetUserLocalBin() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	localBin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		return "", err
	}
	return localBin, nil
}

// IsToolInstalled checks if a binary is currently available in PATH
func IsToolInstalled(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

// IsProvisionable checks if a tool has an automated provisioner
func IsProvisionable(name string) bool {
	key := strings.ToLower(strings.TrimSpace(name))
	_, exists := ProvisionRecipes[key]
	return exists
}

// InstallSpecificTool downloads and installs a designated tool
func InstallSpecificTool(name string) error {
	key := strings.ToLower(strings.TrimSpace(name))
	recipe, exists := ProvisionRecipes[key]
	if !exists {
		return fmt.Errorf("no automated installer recipe available for '%s'. Available: %s", name, strings.Join(GetAvailableToolNames(), ", "))
	}

	if IsToolInstalled(recipe.Binary) {
		fmt.Printf("  %s %s is already installed on your system.\n", provSuccess, provTitle.Render(recipe.Name))
		return nil
	}

	localBin, err := GetUserLocalBin()
	if err != nil {
		return fmt.Errorf("failed to access ~/.local/bin: %w", err)
	}

	fmt.Printf("  %s %s (%s)...\n", provInfo, provTitle.Render(fmt.Sprintf("Downloading & installing %s", recipe.Name)), provDim.Render(recipe.Description))
	if err := recipe.InstallFunc(localBin); err != nil {
		return fmt.Errorf("failed to install %s: %w", recipe.Name, err)
	}

	fmt.Printf("  %s %s successfully installed and ready to use!\n", provSuccess, provTitle.Render(recipe.Name))
	return nil
}

// InstallMissingTools audits the system and installs all missing tools that have automated recipes
func InstallMissingTools() error {
	diagnostics := DiagnoseSystem()
	var missing []ToolDiagnostic

	for _, d := range diagnostics {
		if d.Status == StatusMissing && IsProvisionable(d.Binary) {
			missing = append(missing, d)
		}
	}

	if len(missing) == 0 {
		fmt.Printf("\n  %s All audited toolchains with automated installers are already installed and healthy!\n", provSuccess)
		return nil
	}

	fmt.Printf("\n%s Found %d missing toolchain(s) to install...\n\n",
		provInfo,
		len(missing),
	)

	localBin, err := GetUserLocalBin()
	if err != nil {
		return fmt.Errorf("failed to access ~/.local/bin: %w", err)
	}

	for _, item := range missing {
		key := strings.ToLower(item.Binary)
		recipe, ok := ProvisionRecipes[key]
		if !ok {
			continue
		}

		fmt.Printf("  %s %s...\n", provInfo, provTitle.Render(fmt.Sprintf("Installing %s", recipe.Name)))
		if err := recipe.InstallFunc(localBin); err != nil {
			fmt.Printf("  %s %s: %v\n", provWarn, provTitle.Render(recipe.Name), err)
		} else {
			fmt.Printf("  %s %s installed successfully.\n", provSuccess, provTitle.Render(recipe.Name))
		}
	}

	fmt.Printf("\n%s Re-run 'ace doctor' to verify all updated toolchain statuses.\n", provInfo)
	return nil
}

// GetAvailableToolNames returns all supported tool names for auto-install
func GetAvailableToolNames() []string {
	keys := make([]string, 0, len(ProvisionRecipes))
	for k := range ProvisionRecipes {
		keys = append(keys, k)
	}
	return keys
}

// ---- Recipe Implementations ----

func installComposer(localBin string) error {
	targetComposer := filepath.Join(localBin, "composer")
	tempComposer := filepath.Join(localBin, "composer.tmp")

	// GitHub releases CDN is significantly faster and more reliable worldwide than getcomposer.org
	urls := []string{
		"https://github.com/composer/composer/releases/latest/download/composer.phar",
		"https://getcomposer.org/download/latest-stable/composer.phar",
		"https://getcomposer.org/composer.phar",
	}

	var downloadErr error
	for _, u := range urls {
		// Prefer curl with IPv4 (-4) and short connect timeout to prevent hanging on broken IPv6 routes
		if curlPath, err := exec.LookPath("curl"); err == nil {
			cmd := exec.Command(curlPath, "-4", "-fsSL", "--connect-timeout", "10", "-o", tempComposer, u)
			if err := cmd.Run(); err == nil {
				// Verify the downloaded file is a valid executable Phar
				if vCmd := exec.Command("php", tempComposer, "--version"); vCmd.Run() == nil {
					_ = os.Rename(tempComposer, targetComposer)
					_ = os.Chmod(targetComposer, 0755)
					return nil
				}
				_ = os.Remove(tempComposer)
			}
		}

		if err := downloadFile(u, tempComposer); err == nil {
			if vCmd := exec.Command("php", tempComposer, "--version"); vCmd.Run() == nil {
				_ = os.Rename(tempComposer, targetComposer)
				_ = os.Chmod(targetComposer, 0755)
				return nil
			}
			_ = os.Remove(tempComposer)
		} else {
			downloadErr = err
		}
	}

	if downloadErr != nil {
		return fmt.Errorf("failed to download composer: %w", downloadErr)
	}

	return fmt.Errorf("downloaded composer was corrupted or failed verification")
}

func installBun(localBin string) error {
	scriptURL := "https://bun.sh/install"
	cmd := exec.Command("bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", scriptURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bun installer failed: %w", err)
	}

	// Symlink from ~/.bun/bin/bun to localBin/bun for instant PATH pickup
	home, _ := os.UserHomeDir()
	bunBin := filepath.Join(home, ".bun", "bin", "bun")
	if _, err := os.Stat(bunBin); err == nil {
		dest := filepath.Join(localBin, "bun")
		_ = os.Remove(dest)
		_ = os.Symlink(bunBin, dest)
	}
	return nil
}

func installPnpm(localBin string) error {
	// Try npm global install with prefix if npm exists
	if npmPath, err := exec.LookPath("npm"); err == nil {
		home, _ := os.UserHomeDir()
		prefixDir := filepath.Join(home, ".local")
		cmd := exec.Command(npmPath, "install", "-g", "--prefix", prefixDir, "pnpm")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to official standalone script
	cmd := exec.Command("bash", "-c", "curl -fsSL https://get.pnpm.io/install.sh | SHELL=$(which bash) bash -")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm installer failed: %w", err)
	}

	home, _ := os.UserHomeDir()
	pnpmBin := filepath.Join(home, ".local", "share", "pnpm", "pnpm")
	if _, err := os.Stat(pnpmBin); err == nil {
		dest := filepath.Join(localBin, "pnpm")
		_ = os.Remove(dest)
		_ = os.Symlink(pnpmBin, dest)
	}
	return nil
}

func installRust(localBin string) error {
	cmd := exec.Command("bash", "-c", "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rustup installer failed: %w", err)
	}

	home, _ := os.UserHomeDir()
	cargoBin := filepath.Join(home, ".cargo", "bin", "cargo")
	rustcBin := filepath.Join(home, ".cargo", "bin", "rustc")
	if _, err := os.Stat(cargoBin); err == nil {
		_ = os.Remove(filepath.Join(localBin, "cargo"))
		_ = os.Symlink(cargoBin, filepath.Join(localBin, "cargo"))
	}
	if _, err := os.Stat(rustcBin); err == nil {
		_ = os.Remove(filepath.Join(localBin, "rustc"))
		_ = os.Symlink(rustcBin, filepath.Join(localBin, "rustc"))
	}
	return nil
}

func installDotnet(localBin string) error {
	cmd := exec.Command("bash", "-c", "curl -sSL https://dot.net/v1/dotnet-install.sh | bash -s -- --channel 8.0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotnet installer failed: %w", err)
	}

	home, _ := os.UserHomeDir()
	dotnetBin := filepath.Join(home, ".dotnet", "dotnet")
	if _, err := os.Stat(dotnetBin); err == nil {
		dest := filepath.Join(localBin, "dotnet")
		_ = os.Remove(dest)
		_ = os.Symlink(dotnetBin, dest)
	}
	return nil
}

func installNodeSystem(localBin string) error {
	pm, args := detectSystemPackageManager("nodejs", "npm")
	if pm == "" {
		return fmt.Errorf("please install Node.js via https://nodejs.org or fnm (curl -fsSL https://fnm.vercel.app/install | bash)")
	}
	return runSudoOrCommand(pm, args)
}

func installPythonSystem(localBin string) error {
	pm, args := detectSystemPackageManager("python3", "python3-pip")
	if pm == "" {
		return fmt.Errorf("please install Python 3 via your operating system package manager")
	}
	return runSudoOrCommand(pm, args)
}

func installPhpSystem(localBin string) error {
	pm, args := detectSystemPackageManager("php", "php-cli", "php-mbstring", "php-xml", "php-curl")
	if pm == "" {
		return fmt.Errorf("please install PHP via your operating system package manager")
	}
	return runSudoOrCommand(pm, args)
}

func installDockerSystem(localBin string) error {
	cmd := exec.Command("bash", "-c", "curl -fsSL https://get.docker.com | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installGitSystem(localBin string) error {
	pm, args := detectSystemPackageManager("git")
	if pm == "" {
		return fmt.Errorf("please install Git via your operating system package manager")
	}
	return runSudoOrCommand(pm, args)
}

func detectSystemPackageManager(packages ...string) (string, []string) {
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew", append([]string{"install"}, packages...)
		}
	}

	// Linux package managers
	managers := []struct {
		bin  string
		args []string
	}{
		{"dnf", append([]string{"install", "-y"}, packages...)},
		{"apt-get", append([]string{"install", "-y"}, packages...)},
		{"pacman", append([]string{"-S", "--noconfirm"}, packages...)},
		{"apk", append([]string{"add", "--no-cache"}, packages...)},
		{"zypper", append([]string{"install", "-y"}, packages...)},
	}

	for _, m := range managers {
		if _, err := exec.LookPath(m.bin); err == nil {
			return m.bin, m.args
		}
	}

	return "", nil
}

func runSudoOrCommand(bin string, args []string) error {
	var cmd *exec.Cmd
	if os.Geteuid() == 0 || bin == "brew" {
		cmd = exec.Command(bin, args...)
	} else {
		cmd = exec.Command("sudo", append([]string{bin}, args...)...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func downloadFile(url, destPath string) error {
	client := &http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AceCLI/0.1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading %s: %s", url, resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
